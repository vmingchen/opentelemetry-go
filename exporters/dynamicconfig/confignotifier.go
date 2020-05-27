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
	"time"
)

type MetricConfig struct {
	Period time.Duration
}

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
// Right now we just return a fixed config
func readConfig() *MetricConfig {
	return &MetricConfig{
		Period: 2 * time.Second,
	}
}

type ConfigWatcher interface {
	OnInitialConfig(config *MetricConfig)
	OnUpdatedConfig(config *MetricConfig)
}

type ConfigNotifier struct {
	checkFrequency time.Duration
	config         *MetricConfig
	subscribed     map[ConfigWatcher]bool
}

// Constructor for a ConfigNotifier, also starts it running
func New(checkFrequency time.Duration, config *MetricConfig) *ConfigNotifier {
	configNotifier := &ConfigNotifier{
		checkFrequency: checkFrequency,
		config:         config,
		subscribed:     make(map[ConfigWatcher]bool),
	}

	go configNotifier.checkChanges()

	return configNotifier
}

func (notifier *ConfigNotifier) Register(watcher ConfigWatcher) {
	notifier.subscribed[watcher] = true
	watcher.OnInitialConfig(notifier.config)
}

func (notifier *ConfigNotifier) Unregister(watcher ConfigWatcher) {
	delete(notifier.subscribed, watcher)
}

func (notifier *ConfigNotifier) checkChanges() {
	for {
		time.Sleep(notifier.checkFrequency)

		newConfig := readConfig()

		if !notifier.config.equals(newConfig) {
			notifier.config = newConfig

			for watcher := range notifier.subscribed {
				watcher.OnUpdatedConfig(newConfig)
			}
		}
	}
}
