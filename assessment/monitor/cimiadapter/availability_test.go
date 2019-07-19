/*
Copyright 2019 Atos

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cimiadapter

import (
	"SLALite/repositories/cimi"
	"math"
	"testing"
	"time"
)

const _MaxDelta = 0.5

var tend = time.Now()
var tstart = tend.Add(-600 * time.Second)

type interval struct {
	start time.Time
	end   time.Time
}

func (i interval) availability(times [][2]int) float64 {

	scms := make([]cimi.ServiceContainerMetric, len(times))
	for i := range scms {
		start := tstart.Add(time.Duration(times[i][0]) * time.Second)
		end := tstart.Add(time.Duration(times[i][1]) * time.Second)
		scms[i] = cimi.ServiceContainerMetric{
			StartTime: start,
			StopTime:  &end,
		}
	}
	av := calculateAvailability(scms, tstart, tend)
	return av
}

func d(a, b float64) float64 {
	return math.Abs(a - b)
}

func TestCalculateAvailability(t *testing.T) {
	i := interval{tstart, tend}

	var times [][2]int
	times = [][2]int{
		{0, 600},
	}
	if act, exp := i.availability(times), 100.0; d(act, exp) > _MaxDelta {
		t.Errorf("Error in availability for intervals %v: expected=%v; actual=%v.", times, exp, act)
	}

	times = [][2]int{
		{-150, 600},
	}
	if act, exp := i.availability(times), 100.0; d(act, exp) > _MaxDelta {
		t.Errorf("Error in availability for intervals %v: expected=%v; actual=%v.", times, exp, act)
	}

	times = [][2]int{
		{0, 300},
		{300, 600},
	}
	if act, exp := i.availability(times), 100.0; d(act, exp) > _MaxDelta {
		t.Errorf("Error in availability for intervals %v: expected=%v; actual=%v.", times, exp, act)
	}

	times = [][2]int{
		{0, 150},
		{450, 600},
	}
	if act, exp := i.availability(times), 50.0; d(act, exp) > _MaxDelta {
		t.Errorf("Error in availability for intervals %v: expected=%v; actual=%v.", times, exp, act)
	}

	times = [][2]int{ // 0..200 + 300..450 + 500..600 = 450/600
		{-50, 50},
		{0, 100},
		{50, 150},
		{100, 200},
		{300, 400},
		{350, 450},
		{500, 600},
		{550, 650},
	}
	if act, exp := i.availability(times), 75.0; d(act, exp) > _MaxDelta {
		t.Errorf("Error in availability for intervals %v: expected=%v; actual=%v.", times, exp, act)
	}

}

var av float64

func BenchmarkAvailabilityHour(b *testing.B) {
	tend := tstart.Add(3600 * time.Second)

	scms := []cimi.ServiceContainerMetric{
		cimi.ServiceContainerMetric{
			StartTime: tstart,
			StopTime:  &tend,
		},
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		av = calculateAvailability(scms, tstart, tend)
	}

}

func TestCalculateAvailabilityWithNil(t *testing.T) {

	scms := []cimi.ServiceContainerMetric{
		cimi.ServiceContainerMetric{
			StartTime: tstart,
			StopTime:  nil,
		},
	}
	if act, exp := calculateAvailability(scms, tstart, tend), 100.0; act != exp {
		t.Errorf("Error in availability for intervals %v: expected=%v; actual=%v.", scms, exp, act)
	}
}
