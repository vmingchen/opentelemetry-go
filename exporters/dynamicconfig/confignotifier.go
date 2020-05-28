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
)

type MetricConfig struct {
	Period time.Duration
}

// TODO: Replace with something better when we have a solid config spec
func (config *MetricConfig) equals(otherConfig *MetricConfig) bool {
	if config == nil && otherConfig == nil {
		return true
	}

	if config == nil || otherConfig == nil {
		return false
	}

	configString := fmt.Sprintf("%+v", *config)
	otherConfigString := fmt.Sprintf("%+v", *otherConfig)

	if configString == otherConfigString {
		return true
	}
	return false
}

// TODO: Read actual config from host
// We don't do anything with the configHost string right now
// TIMES_READ_TEST is for testing purposes, to mimic a dynamic service
// The second time we call readConfig, the sampling period changes
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
	OnInitialConfig(config *MetricConfig)
	OnUpdatedConfig(config *MetricConfig)
}

type ConfigNotifier struct {
	checkFrequency time.Duration
	config         *MetricConfig
	configLock     sync.Mutex
	configHost     string
	subscribed     map[ConfigWatcher]bool
	subscribedLock sync.Mutex
}

// Constructor for a ConfigNotifier
// Set configHost to "" if there is no remote config service and the config is not dynamic
func New(checkFrequency time.Duration, defaultConfig *MetricConfig, configHost string) *ConfigNotifier {
	config := defaultConfig
	if configHost != "" {
		config = readConfig(configHost)
	}

	configNotifier := &ConfigNotifier{
		checkFrequency: checkFrequency,
		config:         config,
		configHost:     configHost,
		subscribed:     make(map[ConfigWatcher]bool),
	}

	return configNotifier
}

func (notifier *ConfigNotifier) Start() {
	go notifier.checkChanges()
}

func (notifier *ConfigNotifier) Register(watcher ConfigWatcher) {
	notifier.subscribedLock.Lock()
	notifier.subscribed[watcher] = true
	notifier.subscribedLock.Unlock()

	notifier.configLock.Lock()
	watcher.OnInitialConfig(notifier.config)
	notifier.configLock.Unlock()
}

func (notifier *ConfigNotifier) Unregister(watcher ConfigWatcher) {
	notifier.subscribedLock.Lock()
	defer notifier.subscribedLock.Unlock()

	delete(notifier.subscribed, watcher)
}

func (notifier *ConfigNotifier) checkChanges() {
	if notifier.configHost == "" {
		return
	}

	for {
		time.Sleep(notifier.checkFrequency)

		newConfig := readConfig(notifier.configHost)

		notifier.configLock.Lock()
		if !notifier.config.equals(newConfig) {
			notifier.config = newConfig

			for watcher := range notifier.subscribed {
				watcher.OnUpdatedConfig(newConfig)
			}
		}
		notifier.configLock.Unlock()
	}
}
