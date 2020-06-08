// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dynamicconfig

import (
	"fmt"
	"sync"
	"time"

	controllerTime "go.opentelemetry.io/otel/sdk/metric/controller/time"
)

type MetricConfig struct {
	Period time.Duration
}

// TODO: Replace with something better when we have a solid config spec.
func (config *MetricConfig) equals(otherConfig *MetricConfig) bool {
	if config == nil && otherConfig == nil {
		return true
	}

	if config == nil || otherConfig == nil {
		return false
	}

	configString := fmt.Sprintf("%+v", *config)
	otherConfigString := fmt.Sprintf("%+v", *otherConfig)

	return configString == otherConfigString
}

// TODO: Read the actual config from host.
// We don't do anything with the configHost string right now. TIMES_READ_TEST
// is for testing purposes, so the second time we call readConfig, the
// sampling period changes. This is to mimic a dynamic config service.
var TIMES_CONFIG_READ_TEST = 0

func readConfig(configHost string) *MetricConfig {
	period := 10 * time.Second

	if TIMES_CONFIG_READ_TEST > 0 {
		period = 2 * time.Second
	}

	TIMES_CONFIG_READ_TEST++

	return &MetricConfig{
		Period: period,
	}
}

type ConfigWatcher interface {
	// NOTE: A lock will be held during the execution of both these functions.
	// Please ensure their implementation is not too slow so as to avoid lock-
	// starvation.
	OnInitialConfig(config *MetricConfig)
	OnUpdatedConfig(config *MetricConfig)
}

// A ConfigNotifier monitors a config service for a config changing, then letting
// all its subscribers know if the config has changed.
type ConfigNotifier struct {
	// Used to shut down the config checking routine when we stop ConfigNotifier.
	ch chan struct{}

	// How often we check to see if the config service has changed.
	checkFrequency time.Duration

	// Added for testing time-related functionality.
	clock controllerTime.Clock

	// The current config we used as a default if we cannot read from the remote
	// configuration service.
	config *MetricConfig

	// Optional field for the address of the config service host if the config is
	// non-dynamic.
	configHost string

	lock sync.Mutex

	// Set of all the notifier's subscribers.
	subscribed map[ConfigWatcher]bool

	// Controls when we check the config service for a potential new config. It is
	// set to nil if configHost is empty.
	ticker controllerTime.Ticker

	// This is used to wait for the config checking routine to return when we stop
	// the notifier.
	wg sync.WaitGroup
}

// Constructor for a ConfigNotifier
func New(checkFrequency time.Duration, defaultConfig *MetricConfig, opts ...Option) *ConfigNotifier {
	notifier := &ConfigNotifier{
		ch:             make(chan struct{}),
		checkFrequency: checkFrequency,
		clock:          controllerTime.RealClock{},
		config:         defaultConfig,
		subscribed:     make(map[ConfigWatcher]bool),
	}

	for _, opt := range opts {
		opt.Apply(notifier)
	}

	return notifier
}

// SetClock supports setting a mock clock for testing.  This must be
// called before Start().
func (n *ConfigNotifier) SetClock(clock controllerTime.Clock) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.clock = clock
}

func (n *ConfigNotifier) Start() {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.configHost == "" {
		return
	}

	if n.ticker != nil {
		return
	}

	n.ticker = n.clock.Ticker(n.checkFrequency)
	n.wg.Add(1)
	go n.checkChanges(n.ch)
}

func (n *ConfigNotifier) Stop() {
	n.lock.Lock()

	if n.configHost == "" {
		return
	}

	if n.ch == nil {
		return
	}
	close(n.ch)
	n.ch = nil

	n.lock.Unlock()

	n.wg.Wait()
	n.ticker.Stop()
}

func (n *ConfigNotifier) Register(watcher ConfigWatcher) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.subscribed[watcher] = true
	watcher.OnInitialConfig(n.config)
}

func (n *ConfigNotifier) Unregister(watcher ConfigWatcher) {
	n.lock.Lock()
	defer n.lock.Unlock()

	delete(n.subscribed, watcher)
}

func (n *ConfigNotifier) checkChanges(ch chan struct{}) {
	for {
		select {
		case <-ch:
			n.wg.Done()
			return
		case <-n.ticker.C():
			newConfig := readConfig(n.configHost)

			n.lock.Lock()
			if !n.config.equals(newConfig) {
				n.config = newConfig

				for watcher := range n.subscribed {
					watcher.OnUpdatedConfig(newConfig)
				}
			}
			n.lock.Unlock()
		}
	}
}
