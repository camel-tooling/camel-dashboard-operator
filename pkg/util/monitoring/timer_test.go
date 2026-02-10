/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package monitoring

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// compile-time check
var _ prometheus.Observer = (*fakeObserver)(nil)

type fakeObserver struct {
	calls  int
	values []float64
}

func (f *fakeObserver) Observe(v float64) {
	f.calls++
	f.values = append(f.values, v)
}

func TestNewTimer(t *testing.T) {
	before := time.Now()
	timer := NewTimer()
	after := time.Now()

	if timer == nil {
		t.Fatalf("expected timer, got nil")
	}

	if timer.begin.Before(before) || timer.begin.After(after) {
		t.Errorf("timer.begin not initialized to current time, got %v", timer.begin)
	}
}

func TestObserveDurationInSeconds(t *testing.T) {
	timer := NewTimer()
	observer := &fakeObserver{}

	// ensure measurable duration
	time.Sleep(10 * time.Millisecond)

	d := timer.ObserveDurationInSeconds(observer)

	if observer.calls != 1 {
		t.Fatalf("expected Observe to be called once, got %d", observer.calls)
	}

	if len(observer.values) != 1 {
		t.Fatalf("expected 1 observed value, got %d", len(observer.values))
	}

	observedSeconds := observer.values[0]

	if observedSeconds <= 0 {
		t.Errorf("expected observed duration > 0, got %f", observedSeconds)
	}

	if d <= 0 {
		t.Errorf("expected returned duration > 0, got %v", d)
	}

	// sanity check: returned duration matches observed value (within tolerance)
	diff := d.Seconds() - observedSeconds
	if diff < 0 {
		diff = -diff
	}

	if diff > 0.001 {
		t.Errorf(
			"returned duration (%f) and observed duration (%f) differ too much",
			d.Seconds(),
			observedSeconds,
		)
	}
}

func TestObserveDurationInSeconds_MultipleCalls(t *testing.T) {
	timer := NewTimer()
	observer := &fakeObserver{}

	time.Sleep(5 * time.Millisecond)
	d1 := timer.ObserveDurationInSeconds(observer)

	time.Sleep(5 * time.Millisecond)
	d2 := timer.ObserveDurationInSeconds(observer)

	if observer.calls != 2 {
		t.Fatalf("expected 2 Observe calls, got %d", observer.calls)
	}

	if d2 <= d1 {
		t.Errorf("expected second duration > first duration, got %v <= %v", d2, d1)
	}
}
