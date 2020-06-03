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

// Option is the interface that applies the value to a configuration option.
type Option interface {
	// Apply sets the Option value of a Config.
	Apply(*ConfigNotifier)
}

// WithConfigHost sets the ConfigHost configuration option of a Config.
func WithConfigHost(host string) Option {
	return configHostOption(host)
}

type configHostOption string

func (o configHostOption) Apply(notifier *ConfigNotifier) {
	notifier.configHost = string(o)
}
