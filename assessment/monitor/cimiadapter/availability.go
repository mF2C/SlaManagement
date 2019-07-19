/*
Copyright 2018 Atos

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
	"time"
)

// Be careful if you run the code here after 73069258126-09-25!
var _INF = time.Unix(1<<61, 0)

func calculateAvailability(scms []cimi.ServiceContainerMetric, from, to time.Time) (availability float64) {

	delta := to.Sub(from).Seconds()
	window := int(delta) + 1

	/*
	 * Make a mask of window size.
	 * A true in the mask means that that second the service was available
	 * TODO: use a bit array
	 */
	mask := make([]byte, window)

	for _, scm := range scms {

		start := max(from, scm.StartTime)
		end := min(to, timeOrInf(scm.StopTime))

		startIdx := int(start.Sub(from).Seconds())
		endIdx := int(end.Sub(from).Seconds())

		for i := startIdx; i <= endIdx; i++ {
			mask[i] = 1
		}
	}

	uptime := 0
	for _, v := range mask {
		uptime += int(v)
	}

	availability = 100 * float64(uptime) / float64(window)
	return availability
}

func max(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t2
	}
	return t1
}

func min(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t2
	}
	return t1
}

func timeOrInf(t *time.Time) time.Time {
	if t == nil {
		return _INF
	}
	return *t
}
