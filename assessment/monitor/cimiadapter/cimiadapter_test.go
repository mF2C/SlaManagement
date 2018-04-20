package cimiadapter

import (
	"SLALite/assessment"
	"SLALite/model"
	"SLALite/repositories/cimi"
	"encoding/json"
	"os"
	"testing"
	"time"
)

var tl = Timeline{
	T0: time.Now(),
}

type repository struct {
	values []cimi.ServiceOperationReport
}

func (r repository) GetServiceOperationReportsByDate(serviceInstance string, from time.Time) ([]cimi.ServiceOperationReport, error) {
	return r.values, nil
}

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

func TestNextValues(t *testing.T) {

	a, err := readAgreement("testdata/a.json")
	if err != nil {
		t.Errorf("Error reading agreement: %v", err)
	}
	op := "dijkstra"
	si := "service-instance"
	r := repository{
		[]cimi.ServiceOperationReport{
			cimi.ServiceOperationReport{
				Created:         tl.T(0),
				ServiceInstance: cimi.Href{Href: si},
				Operation:       op,
				ExecutionTime:   100,
			},
			cimi.ServiceOperationReport{
				Created:         tl.T(1),
				ServiceInstance: cimi.Href{Href: si},
				Operation:       op,
				ExecutionTime:   99,
			},
		},
	}
	adapter := New(r)

	_, err = assessment.EvaluateAgreement(&a, adapter)
	if err != nil {
		t.Errorf("Error evaluating agreement: %v", err)
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
