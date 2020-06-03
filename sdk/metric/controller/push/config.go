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

package push

import (
	"time"

	notifier "go.opentelemetry.io/otel/exporters/dynamicconfig"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Config contains configuration for a push Controller.
type Config struct {
	// Resource is the OpenTelemetry resource associated with all Meters
	// created by the Controller.
	Resource *resource.Resource

	// Stateful causes the controller to maintain state across
	// collection events, so that records in the exported
	// checkpoint set are cumulative.
	Stateful bool

	// Period is the interval between calls to Collect a checkpoint.
	Period time.Duration

	// Timeout is the duration a collection (i.e. collect, accumulate,
	// integrate, and export) can last before it is canceled. Defaults to
	// the controller push period.
	Timeout time.Duration

	// ConfigNotifier supplies configurations for the push controller
	// from a remote config service
	ConfigNotifier *notifier.ConfigNotifier
}

// Option is the interface that applies the value to a configuration option.
type Option interface {
	// Apply sets the Option value of a Config.
	Apply(*Config)
}

// WithResource sets the Resource configuration option of a Config.
func WithResource(r *resource.Resource) Option {
	return resourceOption{r}
}

type resourceOption struct{ *resource.Resource }

func (o resourceOption) Apply(config *Config) {
	config.Resource = o.Resource
}

// WithStateful sets the Stateful configuration option of a Config.
func WithStateful(stateful bool) Option {
	return statefulOption(stateful)
}

type statefulOption bool

func (o statefulOption) Apply(config *Config) {
	config.Stateful = bool(o)
}

// WithPeriod sets the Period configuration option of a Config.
func WithPeriod(period time.Duration) Option {
	return periodOption(period)
}

type periodOption time.Duration

func (o periodOption) Apply(config *Config) {
	config.Period = time.Duration(o)
}

// WithTimeout sets the Timeout configuration option of a Config.
func WithTimeout(timeout time.Duration) Option {
	return timeoutOption(timeout)
}

type timeoutOption time.Duration

func (o timeoutOption) Apply(config *Config) {
	config.Timeout = time.Duration(o)
}

// WithConfigNotifier sets the ConfigNotifier configuration option of a Config
func WithConfigNotifier(notifier *notifier.ConfigNotifier) Option {
	return configNotifierOption{notifier}
}

type configNotifierOption struct{ *notifier.ConfigNotifier }

func (o configNotifierOption) Apply(config *Config) {
	config.ConfigNotifier = o.ConfigNotifier
}
