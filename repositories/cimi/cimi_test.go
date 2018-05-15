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

package cimi

import (
	"SLALite/model"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {

	// do not run this test
	//os.Exit(m.Run())
}

func getRepository(config *viper.Viper) (Repository, error) {
	if config == nil {
		config = viper.New()
		config.Set(userProp, anonUser)
		config.Set(pwdProp, "super ADMIN")
		//config.Set(userProp, defaultUser)
		//config.Set(pwdProp, defaultPwd)
		config.Set(insecureProp, true)
		config.Set(urlProp, "https://dashboard.mf2c-project.eu/api")
	}
	return New(config)
}

func TestGetSession(t *testing.T) {
	config := viper.New()
	config.Set(urlProp, "https://dashboard.mf2c-project.eu/api")
	config.Set(userProp, anonUser)
	r, err := New(config)

	if err != nil {
		t.Errorf("Could not get session %v", r)
		return
	}

	var p []userProfile

	p, err = r.getUserProfiles()
	fmt.Printf("%v %v\n", p, err)
}

func TestGetAgreements(t *testing.T) {
	r, err := getRepository(nil)
	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	var a model.Agreements
	a, err = r.GetAllAgreements()
	fmt.Printf("%v %v\n", a, err)
}

func TestGetAgreement(t *testing.T) {
	r, err := getRepository(nil)
	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	var a *model.Agreement
	a, err = r.GetAgreement("ea0d7739-4d65-4d38-b42b-aa2704ccd598")
	jsonValue, _ := json.Marshal(a)
	fmt.Printf("%v %v\n", string(jsonValue), err)
}

func TestCreateAgreement(t *testing.T) {
	//
}
func TestCreateViolation(t *testing.T) {
	r, err := getRepository(nil)
	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	var v = &model.Violation{
		AgreementId: "agreement/ea0d7739-4d65-4d38-b42b-aa2704ccd598",
		Datetime:    time.Now(),
		Guarantee:   "gt01",
	}
	v, err = r.CreateViolation(v)
	if err != nil {
		t.Errorf("Error creating violation: %v", err)
		return
	}
	fmt.Printf("%v", v)
}

func TestCreateExecutionLog(t *testing.T) {
	r, err := getRepository(nil)
	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	var e = &ServiceOperationReport{
		ServiceInstance: Href{
			"service-instance/0ff35277-866d-4ff2-9887-cfe3272c10d0",
		},
		Operation:     "dijkstra",
		ExecutionTime: rand.Float64() * 200,
	}
	e, err = r.CreateServiceOperationReport(e)
	if err != nil {
		t.Errorf("Error creating execution log: %v", err)
		return
	}
	fmt.Printf("%v", e)
}

func TestGetExecutionLog(t *testing.T) {
	var e []ServiceOperationReport

	r, err := getRepository(nil)

	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	now := time.Now()
	hourAgo := now.Add(time.Hour * -2)
	e, err = r.GetServiceOperationReportsByDate("service-instance/b08ee389-36c0-45f0-8684-61baf6e03da8", hourAgo)
	fmt.Printf("%v err=%v", e, err)
}

func TestStartAgreement(t *testing.T) {
	r, err := getRepository(nil)
	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	id := "ea0d7739-4d65-4d38-b42b-aa2704ccd598"
	err = r.StartAgreement(id)
	if err != nil {
		t.Errorf("Error starting agreement: %v", err)
		return
	}

}

func TestStopAgreement(t *testing.T) {
	r, err := getRepository(nil)
	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	id := "ea0d7739-4d65-4d38-b42b-aa2704ccd598"
	err = r.StopAgreement(id)
	if err != nil {
		t.Errorf("Error stopping agreement: %v", err)
		return
	}
}

func TestSubpath(t *testing.T) {
	r, err := getRepository(nil)
	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	fmt.Println(r.subpath("agreement", "agreement/blabla"))
	fmt.Println(r.subpath("agreement", "blabla"))
}

func TestGetServiceInstance(t *testing.T) {
	r, err := getRepository(nil)
	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	si, err := r.getServiceInstance("service-instance/aeced891-2e16-4537-ae7b-600457addfba")
	fmt.Printf("%v err=%v", si, err)
}

func TestUpdateServiceInstance(t *testing.T) {
	r, err := getRepository(nil)
	if err != nil {
		t.Errorf("Could not get repository: %s", err)
		return
	}
	si, err := r.getServiceInstance("service-instance/0ff35277-866d-4ff2-9887-cfe3272c10d0")
	si.Agreement = "agreement/ea0d7739-4d65-4d38-b42b-aa2704ccd598"
	si, err = r.updateServiceInstance(si)
	fmt.Printf("%v", err)
}
