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

	_, err = assessment.EvaluateAgreement(&a, adapter)
	var res amodel.Result
	if err == nil {
		a.Assessment = new(model.Assessment)
		res, err = assessment.EvaluateAgreement(&a, adapter)
	}
	if err != nil {
		t.Errorf("Error evaluating agreement: %v", err)
	}
	// Check there one violation per GT
	if nDijkstra := len(res[dijkstra].Violations); nDijkstra != 1 {
		t.Errorf("Unexpected number of dijkstra violations. Expected: %d. Actual: %d", 1, nDijkstra)
	}
	if nAll := len(res[string(catchAllName)].Violations); nAll != 1 {
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

	adapter.Initialize(&a)
	gt := a.Details.Guarantees[0]

	/* Two values should be provided */
	values := adapter.GetValues(gt, []string{ExecTimeName})
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
	}
	return a, r, nil
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
