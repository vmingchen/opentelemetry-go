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
	"reflect"

	pb "go.opentelemetry.io/otel/exporters/dynamicconfig/v1"
)

type Config struct {
	MetricConfig *pb.ConfigResponse_MetricConfig
	TraceConfig  *pb.ConfigResponse_TraceConfig
}

// TODO: Either get rid of this or replace later
// This is for convenient development/testing purposes
func GetDefaultConfig(period pb.ConfigResponse_MetricConfig_Schedule_CollectionPeriod) *Config {
	schedule := pb.ConfigResponse_MetricConfig_Schedule{Period: period}

	return &Config{
		MetricConfig: &pb.ConfigResponse_MetricConfig{
			CollectingSchedules: []*pb.ConfigResponse_MetricConfig_Schedule{&schedule},
		},
	}
}

func (config *Config) Equals(otherConfig *Config) bool {
	return reflect.DeepEqual(config, otherConfig)
}
