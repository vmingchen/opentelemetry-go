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

package dynamicconfigloader

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/sdk/metric/controller/push"
)


// This allows us to dynamically restart the Controller
type DynamicConfigLoader struct {
	ch chan struct{}
	ticker *time.Ticker
}

// ch is used to stop the config loader when the pusher stops
// ticker is used to control how often we check the config for changes
func New(ch chan struct{}, refreshFrequency time.Duration) *DynamicConfigLoader {
	return &DynamicConfigLoader{
		ch: ch,
		ticker: time.NewTicker(refreshFrequency),
	}
}

func (loader *DynamicConfigLoader) Run(controller *push.Controller) {
	originalConfigString := ""

	for {
		select {
		// The controller shuts down the config loader when it stops
		case <-loader.ch:
			loader.ticker.Stop()
			return
		case <- loader.ticker.C:
			newConfig := readConfig()
			newConfigString := fmt.Sprintf("%+v", newConfig)

			if newConfigString != originalConfigString {
				originalConfigString = newConfigString

				controller.RestartTicker(newConfig.samplingPeriod)
			}
		}
	}
}

// TODO: Create actual config structure
type Config struct {
	samplingPeriod time.Duration
}


// TODO: Read actual config from IP address
// Right now we just return a samplingPeriod of 2 seconds
func readConfig() *Config {
	return &Config{
		samplingPeriod: 2 * time.Second,
	}
}
