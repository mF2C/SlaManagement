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
	"SLALite/assessment"
	amodel "SLALite/assessment/model"
	"SLALite/model"
	"SLALite/repositories/cimi"
	"encoding/json"
	"os"
	"testing"
	"time"
)

const (
	dijkstra = "dijkstra"
)

var tl = Timeline{
	T0: time.Now(),
}

type repository struct {
	values           []cimi.ServiceOperationReport
	serviceInstances []cimi.ServiceInstance
	containerMetrics []cimi.ServiceContainerMetric
}

func (r repository) GetServiceOperationReportsByDate(serviceInstance string, from time.Time) ([]cimi.ServiceOperationReport, error) {
	result := make([]cimi.ServiceOperationReport, 0, 1)
	for _, log := range r.values {
		if log.ServiceInstance.Href == serviceInstance {
			result = append(result, log)
		}
	}
	return result, nil
}

func (r repository) GetServiceInstancesByAgreement(aID string) ([]cimi.ServiceInstance, error) {
	result := make([]cimi.ServiceInstance, 0, 2)
	for _, si := range r.serviceInstances {
		if si.Agreement == aID {
			result = append(result, si)
		}
	}
	return result, nil
}

func (r repository) GetServiceContainerMetrics(device string, container string, startTime time.Time, stopTime time.Time) ([]cimi.ServiceContainerMetric, error) {
	result := make([]cimi.ServiceContainerMetric, 0, 2)
	for _, scm := range r.containerMetrics {
		if scm.Container == container {
			result = append(result, scm)
		}
	}
	return result, nil
}

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

func TestEvaluate(t *testing.T) {

	a, r, err := initVars()
	if err != nil {
		t.Errorf("Error reading agreement: %v", err)
		return
	}
	adapter := New(r)

	_, err = assessment.EvaluateAgreement(&a, adapter, time.Now())
	var res amodel.Result
	if err == nil {
		a.Assessment = new(model.Assessment)
		res, err = assessment.EvaluateAgreement(&a, adapter, time.Now())
	}
	if err != nil {
		t.Errorf("Error evaluating agreement: %v", err)
	}
	// Check there one violation per GT
	if nDijkstra := len(res.Violated[dijkstra].Violations); nDijkstra != 1 {
		t.Errorf("Unexpected number of dijkstra violations. Expected: %d. Actual: %d", 1, nDijkstra)
	}
	if nAll := len(res.Violated[string(catchAllName)].Violations); nAll != 1 {
		t.Errorf("Unexpected number of * violations. Expected: %d. Actual: %d", 1, nAll)
	}
}

func TestGetValues(t *testing.T) {
	a, r, err := initVars()
	if err != nil {
		t.Errorf("Error reading agreement: %v", err)
		return
	}
	adapter := New(r)

	adapter = adapter.Initialize(&a)
	gt := a.Details.Guarantees[0]

	/* Two values should be provided */
	values := adapter.GetValues(gt, []string{ExecTime}, time.Now())
	if len(values) != 2 {
		t.Errorf("Unexpected GetValues result: %v", values)
	}
}

func initVars() (model.Agreement, repository, error) {
	var a model.Agreement
	var r repository

	a, err := readAgreement("testdata/a.json")
	if err != nil {
		return a, r, err
	}
	op := dijkstra
	si1 := "service-instance1"
	si2 := "service-instance2"
	r = repository{
		[]cimi.ServiceOperationReport{
			cimi.ServiceOperationReport{
				Created:         tl.T(0),
				ServiceInstance: cimi.Href{Href: si1},
				Operation:       op,
				ExecutionTime:   100,
			},
			cimi.ServiceOperationReport{
				Created:         tl.T(1),
				ServiceInstance: cimi.Href{Href: si2},
				Operation:       op,
				ExecutionTime:   99,
			},
		},
		[]cimi.ServiceInstance{
			cimi.ServiceInstance{
				Id:        si1,
				Agreement: a.Id,
			},
			cimi.ServiceInstance{
				Id:        si2,
				Agreement: a.Id,
			},
		},
		[]cimi.ServiceContainerMetric{},
	}
	return a, r, nil
}

func TestGetAvailabilityValues(t *testing.T) {
	a, err := readAgreement("testdata/b.json")
	if err != nil {
		t.Fatalf("Error loading agreement")
	}
	if v, _ := a.Details.GetVariable("availability"); v.Aggregation.Window != 600 {
		t.Fatalf("Error in agreement schema")
	}

	si1 := "service-instance1"
	r := repository{
		[]cimi.ServiceOperationReport{},
		[]cimi.ServiceInstance{
			cimi.ServiceInstance{
				Id:        si1,
				Agreement: a.Id,
				Agents: []cimi.Agent{
					cimi.Agent{
						ContainerID: "C01",
					},
				},
			},
		},
		[]cimi.ServiceContainerMetric{
			cimi.ServiceContainerMetric{
				Container: "C01",
				StartTime: tstart,
				StopTime:  &tend,
			},
		},
	}
	adapter := New(r)

	adapter = adapter.Initialize(&a)
	gt := a.Details.Guarantees[0]

	values := adapter.GetValues(gt, []string{Availability}, time.Now())
	if len(values) != 1 || d(values[0][Availability].Value.(float64), 100.0) > _MaxDelta {
		t.Errorf("Error calculating availability. values = %v", values)
	}
}

func TestFromAfterCreation(t *testing.T) {
	a, err := readAgreement("testdata/b.json")
	if err != nil {
		t.Fatalf("Error loading agreement")
	}
	if v, _ := a.Details.GetVariable("availability"); v.Aggregation.Window != 600 {
		t.Fatalf("Error in agreement schema")
	}
	// Creation was 100 seconds ago
	a.Details.Creation = tend.Add(-100 * time.Second)

	si1 := "service-instance1"
	r := repository{
		[]cimi.ServiceOperationReport{},
		[]cimi.ServiceInstance{
			cimi.ServiceInstance{
				Id:        si1,
				Agreement: a.Id,
				Agents: []cimi.Agent{
					cimi.Agent{
						ContainerID: "C01",
					},
				},
			},
		},
		[]cimi.ServiceContainerMetric{
			cimi.ServiceContainerMetric{
				Container: "C01",
				StartTime: tstart,
				StopTime:  &tend,
			},
		},
	}
	adapter := New(r)

	adapter = adapter.Initialize(&a)
	gt := a.Details.Guarantees[0]

	values := adapter.GetValues(gt, []string{Availability}, time.Now())
	if len(values) != 0 {
		t.Errorf("No values expected (a.Details.Creation > now-time.Window). values = %v", values)
	}

}
func TestGetCompssAvailabilityValues(t *testing.T) {
	a, err := readAgreement("testdata/b.json")
	if err != nil {
		t.Fatalf("Error loading agreement")
	}
	if v, _ := a.Details.GetVariable("availability"); v.Aggregation.Window != 600 {
		t.Fatalf("Error in agreement schema")
	}

	si1 := "service-instance1"
	r := repository{
		[]cimi.ServiceOperationReport{},
		[]cimi.ServiceInstance{
			cimi.ServiceInstance{
				Id:          si1,
				Agreement:   a.Id,
				ServiceType: cimi.CompssType,
				Agents: []cimi.Agent{
					cimi.Agent{
						ContainerID:  "C01",
						MasterCompss: true,
					},
					cimi.Agent{
						ContainerID:  "C02",
						MasterCompss: false,
					},
				},
			},
		},
		[]cimi.ServiceContainerMetric{
			cimi.ServiceContainerMetric{
				Container: "C02",
				StartTime: tstart,
				StopTime:  &tend,
			},
			cimi.ServiceContainerMetric{
				Container: "C01",
				StartTime: tend.Add(-300 * time.Second),
				StopTime:  &tend,
			},
		},
	}
	adapter := New(r)

	adapter = adapter.Initialize(&a)
	gt := a.Details.Guarantees[0]

	values := adapter.GetValues(gt, []string{Availability}, time.Now())
	if len(values) != 1 || d(values[0][Availability].Value.(float64), 50) > _MaxDelta {
		t.Errorf("Error calculating availability. values = %v", values)
	}
}

// Timeline calculates delta times from a time origin
// Inialize the struct with t0 as your desired time origin
// Ex.:
//    t = Timeline { T0: time.Now() }
type Timeline struct {
	T0 time.Time
}

// T calculates the delta in seconds with respect to the origin
// Ex:
//     t.T(2)
//     t.T(-1)
func (t *Timeline) T(second time.Duration) time.Time {
	return t.T0.Add(time.Second * second)
}

func readAgreement(path string) (model.Agreement, error) {
	var a model.Agreement

	f, err := os.Open(path)
	if err != nil {
		return a, err
	}
	json.NewDecoder(f).Decode(&a)
	f.Close()
	return a, nil
}
