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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	controllerTest "go.opentelemetry.io/otel/sdk/metric/controller/test"
)

type testWatcher struct {
	testDummy int
}

func (w *testWatcher) OnInitialConfig(config *MetricConfig) {
	w.testDummy = 1
}

func (w *testWatcher) OnUpdatedConfig(config *MetricConfig) {
	w.testDummy = 2
}

// Test config updates
func TestDynamicConfigNotifier(t *testing.T) {
	t.Log("first\n")
	watcher := testWatcher{
		testDummy: 0,
	}
	mock := controllerTest.NewMockClock()

	configNotifier := New(time.Minute, nil, "localhost:1234")
	require.Equal(t, watcher.testDummy, 0)
	configNotifier.SetClock(mock)
	configNotifier.Start()

	configNotifier.Register(&watcher)
	require.Equal(t, watcher.testDummy, 1)

	mock.Add(time.Minute)

	require.Equal(t, watcher.testDummy, 2)
	configNotifier.Stop()
}

// Test config doesn't update
func TestNonDynamicConfigNotifier(t *testing.T) {
	watcher := testWatcher{
		testDummy: 0,
	}
	mock := controllerTest.NewMockClock()
	config := &MetricConfig{
		Period: time.Minute,
	}

	configNotifier := New(time.Minute, config, "")
	require.Equal(t, watcher.testDummy, 0)
	configNotifier.SetClock(mock)
	configNotifier.Start()

	configNotifier.Register(&watcher)
	require.Equal(t, watcher.testDummy, 1)

	mock.Add(time.Minute)

	require.Equal(t, watcher.testDummy, 1)
	configNotifier.Stop()
}

func TestDoubleStop(t *testing.T) {
	configNotifier := New(time.Minute, nil, "localhost:1234")
	configNotifier.Start()
	configNotifier.Stop()
	configNotifier.Stop()
}

func TestPushDoubleStart(t *testing.T) {
	configNotifier := New(time.Minute, nil, "localhost:1234")
	configNotifier.Start()
	configNotifier.Start()
	configNotifier.Stop()
}
