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
	"log"
	"sync"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/transform"
	controllerTime "go.opentelemetry.io/otel/sdk/metric/controller/time"
	"go.opentelemetry.io/otel/sdk/resource"
)

const DEFAULT_PROTOCOL_VERSION int32 = 1

type Watcher interface {
	OnInitialConfig(config *Config)
	OnUpdatedConfig(config *Config)
}

// A Notifier monitors a config service for a config changing
// It then lets all it's subscribers know if the config has changed
type Notifier struct {
	// Used to shut down the config checking routine when we stop Notifier
	ch chan struct{}

	// How often we check to see if the config service has changed
	checkFrequency time.Duration

	// Added for testing time-related functionality
	clock controllerTime.Clock

	// Current config
	config *Config

	// Optional if config is non-dynamic. Address of config service host
	configHost string

	lock sync.Mutex

	// Label to associate configs to individual instances
	// Optional if config is non-dynamic. Must be set to read from config service
	resource *resource.Resource

	// Protocol version to check config service with
	protoVersion int32

	// Contains all the notifier's subscribers
	subscribed map[Watcher]bool

	// Controls when we check the config service for a new config
	ticker controllerTime.Ticker

	// Used to wait for the config checking routine to return when we stop the notifier
	wg sync.WaitGroup
}

// Constructor for a Notifier
func NewNotifier(checkFrequency time.Duration, defaultConfig *Config, opts ...Option) *Notifier {
	notifier := &Notifier{
		ch:             make(chan struct{}),
		checkFrequency: checkFrequency,
		clock:          controllerTime.RealClock{},
		config:         defaultConfig,
		subscribed:     make(map[Watcher]bool),
	}

	for _, opt := range opts {
		opt.Apply(notifier)
	}

	return notifier
}

// SetClock supports setting a mock clock for testing.  This must be
// called before Start().
func (n *Notifier) SetClock(clock controllerTime.Clock) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.clock = clock
}

func (n *Notifier) Start() {
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

func (n *Notifier) Stop() {
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

func (n *Notifier) Register(watcher Watcher) {
	n.lock.Lock()
	n.subscribed[watcher] = true

	// Avoid lock starvation with OnInitialConfig by making a copy of the config
	config_copy := *n.config
	n.lock.Unlock()

	watcher.OnInitialConfig(&config_copy)
}

func (n *Notifier) Unregister(watcher Watcher) {
	n.lock.Lock()
	defer n.lock.Unlock()

	delete(n.subscribed, watcher)
}

func (n *Notifier) checkChanges(ch chan struct{}) {
	if n.resource == nil {
		log.Fatalf("Missing Resource: required for reading from config service")
	}

	protoVersion := DEFAULT_PROTOCOL_VERSION
	if n.protoVersion != 0 {
		protoVersion = n.protoVersion
	}

	serviceReader := newServiceReader(n.configHost, protoVersion, transform.Resource(n.resource))

	for {
		select {
		case <-ch:
			n.wg.Done()
			return
		case <-n.ticker.C():
			newConfig, err := serviceReader.readConfig()
			if err != nil {
				log.Printf("Failed to read from config service: %v\n", err)
				break
			}

			if !n.config.Equals(newConfig) {
				n.lock.Lock()
				n.config = newConfig

				// To prevent lock starvation, make a list of the subscribers, then update
				subscribed_copy := make([]Watcher, len(n.subscribed))
				i := 0
				for watcher := range n.subscribed {
					subscribed_copy[i] = watcher
					i++
				}
				n.lock.Unlock()

				for _, watcher := range subscribed_copy {
					watcher.OnUpdatedConfig(newConfig)
				}
			}
		}
	}
}
