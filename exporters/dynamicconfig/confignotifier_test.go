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

package dynamicconfig_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	notifier "go.opentelemetry.io/otel/exporters/dynamicconfig"
	controllerTest "go.opentelemetry.io/otel/sdk/metric/controller/test"
)

// testLock is to prevent race conditions in test code
// testVar is used to verify OnInitialConfig and OnUpdatedConfig are called
type testWatcher struct {
	testLock sync.Mutex
	testVar  int
}

func (w *testWatcher) OnInitialConfig(config *notifier.MetricConfig) {
	w.testLock.Lock()
	defer w.testLock.Unlock()
	w.testVar = 1
}

func (w *testWatcher) OnUpdatedConfig(config *notifier.MetricConfig) {
	w.testLock.Lock()
	defer w.testLock.Unlock()
	w.testVar = 2
}

// Use a getter to prevent race conditions around testVar
func (w *testWatcher) getTestVar() int {
	w.testLock.Lock()
	defer w.testLock.Unlock()
	return w.testVar
}

func newExampleNotifier() *notifier.ConfigNotifier {
	return notifier.New(
		time.Minute,
		&notifier.MetricConfig{Period: time.Minute},
		notifier.WithConfigHost("localhost:1234"),
	)
}

// Test config updates
func TestDynamicConfigNotifier(t *testing.T) {
	watcher := testWatcher{
		testVar: 0,
	}
	mock := controllerTest.NewMockClock()

	configNotifier := newExampleNotifier()
	require.Equal(t, watcher.getTestVar(), 0)

	configNotifier.SetClock(mock)
	configNotifier.Start()

	configNotifier.Register(&watcher)
	require.Equal(t, watcher.getTestVar(), 1)

	mock.Add(time.Minute)

	require.Equal(t, watcher.getTestVar(), 2)
	configNotifier.Stop()
}

// Test config doesn't update
func TestNonDynamicConfigNotifier(t *testing.T) {
	watcher := testWatcher{
		testVar: 0,
	}
	mock := controllerTest.NewMockClock()
	configNotifier := notifier.New(
		time.Minute,
		&notifier.MetricConfig{Period: time.Minute},
	)
	require.Equal(t, watcher.getTestVar(), 0)

	configNotifier.SetClock(mock)
	configNotifier.Start()

	configNotifier.Register(&watcher)
	require.Equal(t, watcher.getTestVar(), 1)

	mock.Add(time.Minute)

	require.Equal(t, watcher.getTestVar(), 1)
	configNotifier.Stop()
}

func TestDoubleStop(t *testing.T) {
	configNotifier := newExampleNotifier()
	configNotifier.Start()
	configNotifier.Stop()
	configNotifier.Stop()
}

func TestPushDoubleStart(t *testing.T) {
	configNotifier := newExampleNotifier()
	configNotifier.Start()
	configNotifier.Start()
	configNotifier.Stop()
}