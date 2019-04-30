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
This tests access to Analytics component.

To run the integration test, set up a analytics component and set env var SLA_ANALYTICS=<url>
(e.g. SLA_ANALYTICS=https://localhost:46020
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

var analyticsSrv *httptest.Server

func TestIntegrationAnalytics(t *testing.T) {
	var err error
	var analytics *Analytics
	var ok bool
	var url string

	if url, ok = os.LookupEnv("SLA_ANALYTICS"); !ok {
		t.Skip("Skipping Analytics integration test")
	}

	if analytics, err = NewAnalytics(url); err != nil {
		log.Fatalf("Error creating Analytics client: %s", err.Error())
	}

	result, err := analytics.Optimal()
	if err != nil {
		t.Errorf("Error getting optimal: %s", err.Error())
	}
	log.Debugf("Analytics.IsLeader = %v", result)
}

func TestAnalytics(t *testing.T) {

	/* create mock server */
	analyticsSrv = httptest.NewServer(http.HandlerFunc(optimal))
	defer analyticsSrv.Close()

	t.Run("test Optimal", testOptimal)
	t.Run("test wrong URL", testWrongURL)
}

func testOptimal(t *testing.T) {

	a, err := NewAnalytics(analyticsSrv.URL)
	if err != nil {
		t.Errorf("Error %v not expected", err)
		return
	}
	o, err := a.Optimal()
	if err != nil {
		t.Errorf("Error %v not expected", err)
	}
	if len(o) != 1 {
		t.Errorf("Len(o). Expected: 1. Actual: %d", len(o))
		return
	}
	if o[0].NodeName != "mf2c-leader" {
		t.Errorf("NodeName. Expected: mf2c-leader. Actual: %s", o[0].NodeName)
	}
}

func testWrongURL(t *testing.T) {

	a, err := NewAnalytics(analyticsSrv.URL + "/fail")
	_, err = a.Optimal()
	if err == nil {
		t.Errorf("Error expected")
	}
}

func optimal(w http.ResponseWriter, r *http.Request) {
	if r.URL.EscapedPath() != "/mf2c/optimal" || r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var data []AnalyticsItem

	f, err := os.Open("testdata/optimal.json")
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
}
