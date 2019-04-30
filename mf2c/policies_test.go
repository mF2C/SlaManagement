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

/*
This tests access to Policies component.

To run the integration test, set up a policies component and set env var SLA_POLICIES=<url>
(e.g. SLA_POLICIES=https://localhost:46050/api/v2
*/

package mf2c

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

var policiesSrv *httptest.Server

func TestIntegrationPolicies(t *testing.T) {
	var err error
	var policies *Policies
	var ok bool
	var policiesURL string

	if policiesURL, ok = os.LookupEnv("SLA_POLICIES"); !ok {
		t.Skip("Skipping Policies integration test")
	}

	if policies, err = NewPolicies(policiesURL); err != nil {
		log.Fatalf("Error creating Policies client: %s", err.Error())
	}

	result, err := policies.IsLeader()
	if err != nil {
		t.Errorf("Error getting leader: %#v", err)
	}
	log.Debugf("Policies.IsLeader = %v", result)
}

func TestPoliciesMock(t *testing.T) {
	expected := true
	policies := NewPoliciesMock(expected)
	actual, _ := policies.IsLeader()

	if expected != actual {
		t.Errorf("TestPoliciesMock. Expected:%v. Actual:%v", expected, actual)
	}
}

func TestPolicies(t *testing.T) {

	/* create mock server */
	policiesSrv = httptest.NewServer(http.HandlerFunc(handler))
	defer policiesSrv.Close()

	t.Run("IsLeader", testIsLeader)
}

func testIsLeader(t *testing.T) {
	var actual bool
	var err error
	var policies *Policies

	policies, err = NewPolicies(policiesSrv.URL + "/" + pathAPI)
	if err == nil {
		actual, err = policies.IsLeader()
	}
	if err != nil {
		t.Errorf("testIsLeader. Unexpected error: %s", err.Error())
		return
	}
	if !actual {
		t.Errorf("testIsLeader. Expected:%v. Actual:%v", true, actual)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.EscapedPath() {
	case "/" + pathAPI + "/" + pathIsLeader:
		var data isLeader
		f, err := os.Open("testdata/isleader.json")
		defer f.Close()
		json.NewDecoder(f).Decode(&data)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		enc.Encode(data)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
