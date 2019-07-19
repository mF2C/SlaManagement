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

package assessment

import (
	"SLALite/assessment/monitor/cimiadapter"
	"SLALite/mf2c"
	"SLALite/model"
	"SLALite/repositories/cimi"
	"SLALite/repositories/memrepository"
	"SLALite/utils/rest"
	"fmt"
	"testing"
	"time"
)

var provider = model.Provider{Id: "p01", Name: "Provider01"}
var client = model.Client{Id: "c02", Name: "A client"}

type mf2cTestRepo struct {
	*memrepository.MemRepository
}

type failingPolicies struct {
}

func (o failingPolicies) IsLeader() (bool, error) {
	return false, rest.Error{Code: 404, Message: fmt.Sprint("Path not found")}
}

func (r mf2cTestRepo) GetServiceOperationReportsByDate(
	serviceInstance string, from time.Time) ([]cimi.ServiceOperationReport, error) {

	fmt.Println("Getting operations")
	return nil, nil
}

func (r mf2cTestRepo) GetServiceInstancesByAgreement(aID string) ([]cimi.ServiceInstance, error) {
	fmt.Println("GetServiceInstancesByAgreement")
	return nil, nil
}

func (r mf2cTestRepo) GetServiceContainerMetrics(device string, container string, begin time.Time, end time.Time) ([]cimi.ServiceContainerMetric, error) {
	fmt.Println("GetServiceContainerMetrics")
	return nil, nil
}

func TestIsNotLeader(t *testing.T) {
	var policies = mf2c.NewPoliciesMock(false)
	AssessMf2cAgreements(nil, nil, nil, policies)
}

func TestErrorGettingIsLeader(t *testing.T) {
	var policies = failingPolicies{}
	AssessMf2cAgreements(nil, nil, nil, policies)
	// AssessMf2cAgreements should return a code or error to check behaviour
}

func TestStartedAgreement(t *testing.T) {
	var memRepo, _ = memrepository.New(nil)
	var mf2cRepo = mf2cTestRepo{&memRepo}
	var policies = mf2c.NewPoliciesMock(true)

	expiration := time.Now().Add(24 * time.Hour)

	a := model.Agreement{
		Id:    "id",
		Name:  "name",
		State: model.STARTED,
		Details: model.Details{
			Id:       "id",
			Name:     "name",
			Type:     model.AGREEMENT,
			Provider: provider, Client: client,
			Creation:   time.Now(),
			Expiration: &expiration,
			Guarantees: []model.Guarantee{
				model.Guarantee{Name: "TestGuarantee", Constraint: "test_value > 10"},
			},
		},
	}

	if a.Assessment != nil {
		t.Errorf("Unexpected initial conditions: Assessment != nil\n")
	}
	mf2cRepo.CreateAgreement(&a)

	ma := cimiadapter.New(mf2cRepo)
	AssessMf2cAgreements(mf2cRepo, mf2cRepo, ma, policies)
	pa, _ := mf2cRepo.GetAgreement("id")
	if pa.Assessment == nil {
		t.Errorf("Unexpected final conditions: Assessment == nil\n")
	} else {
		fmt.Printf("%v", pa.Assessment)
	}
}

func TestStoppedAgreement(t *testing.T) {

	var memRepo, _ = memrepository.New(nil)
	var mf2cRepo = mf2cTestRepo{&memRepo}
	var policies = mf2c.NewPoliciesMock(true)

	expiration := time.Now().Add(24 * time.Hour)

	a := model.Agreement{
		Id:    "id",
		Name:  "name",
		State: model.STOPPED,
		Details: model.Details{
			Id:       "id",
			Name:     "name",
			Type:     model.AGREEMENT,
			Provider: provider, Client: client,
			Creation:   time.Now(),
			Expiration: &expiration,
			Guarantees: []model.Guarantee{
				model.Guarantee{Name: "TestGuarantee", Constraint: "test_value > 10"},
			},
		},
	}

	if a.Assessment != nil {
		t.Errorf("Unexpected initial conditions: Assessment != nil\n")
	}
	mf2cRepo.CreateAgreement(&a)

	ma := cimiadapter.New(mf2cRepo)
	AssessMf2cAgreements(mf2cRepo, mf2cRepo, ma, policies)
	pa, _ := mf2cRepo.GetAgreement("id")
	if pa.Assessment != nil {
		t.Errorf("Unexpected final conditions: Assessment != nil\n")
	}
}
