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

type DynamicConfigLoader struct {
	ch chan struct{}
	ticker *time.Ticker
}

func New(refreshFrequency time.Duration) *DynamicConfigLoader {
	return &DynamicConfigLoader{
		ch: make(chan struct{}),
		ticker: time.NewTicker(refreshFrequency),
	}
}

func (loader *DynamicConfigLoader) GetCh() chan struct{} {
	return loader.ch
}

func (loader *DynamicConfigLoader) Run(controller *push.Controller) {
	originalConfigString := ""

	for {
		select {
		case <-loader.ch:
			loader.ticker.Stop()
			return
		case <- loader.ticker.C:
			newConfig := readConfig()
			newConfigString := fmt.Sprintf("%+v", newConfig)

			if newConfigString != originalConfigString {
				originalConfigString = newConfigString

				controller.Stop()
				controller.SetPeriod(newConfig.samplingPeriod)
				controller.Start()
			}
		}

	}
}

type Config struct {
	samplingPeriod time.Duration
}

func readConfig() *Config {
	return &Config{
		samplingPeriod: 2 * time.Second,
	}
}
