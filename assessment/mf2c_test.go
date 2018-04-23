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
	"SLALite/model"
	"SLALite/repositories/cimi"
	"SLALite/repositories/memrepository"
	"fmt"
	"testing"
	"time"
)

var provider = model.Provider{Id: "p01", Name: "Provider01"}
var client = model.Client{Id: "c02", Name: "A client"}

type mf2cTestRepo struct {
	*memrepository.MemRepository
}

func (r mf2cTestRepo) CreateViolation(v *model.Violation) (*model.Violation, error) {
	fmt.Printf("Creating violation: %v\n", v)
	return v, nil
}

func (r mf2cTestRepo) GetServiceOperationReportsByDate(
	serviceInstance string, from time.Time) ([]cimi.ServiceOperationReport, error) {

	fmt.Println("Getting operations")
	return nil, nil
}

func TestStartedAgreement(t *testing.T) {
	var memRepo, _ = memrepository.New(nil)
	var mf2cRepo = mf2cTestRepo{&memRepo}

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
			Expiration: time.Now().Add(24 * time.Hour),
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
	AssessMf2cAgreements(mf2cRepo, mf2cRepo, ma)
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
			Expiration: time.Now().Add(24 * time.Hour),
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
	AssessMf2cAgreements(mf2cRepo, mf2cRepo, ma)
	pa, _ := mf2cRepo.GetAgreement("id")
	if pa.Assessment != nil {
		t.Errorf("Unexpected final conditions: Assessment != nil\n")
	}
}