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
This tests the cimi adapter with a real CIMI server for testing.

To run this test, set up a cimi repo and set env var SLA_REPOSITORY=cimi.
If cimi is not accessible at https://localhost:10443/api, set SLA_CIMIURL=<url>
*/

package cimiadapter

import (
	"SLALite/assessment"
	"SLALite/utils"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	"SLALite/model"
	"SLALite/repositories/cimi"

	"github.com/spf13/viper"
)

var repo cimi.Repository
var a *model.Agreement
var si *cimi.ServiceInstance
var sor cimi.ServiceOperationReport
var _T utils.Timeline

/*
Tests the integration of the cimi adapter with real entities.

Creates a service instance and agreement.

Then creates service container metrics that makes availability in last 10 minutes
equal to 50%

*/
func TestIntegration(t *testing.T) {
	var err error

	if v, ok := os.LookupEnv("SLA_REPOSITORY"); !ok || v != cimi.Name {
		t.Skip("Skipping CIMI integration test")

	}

	if repo, err = createRepository(); err != nil {
		log.Fatalf("Error creating repository: %s", err.Error())
	}

	now := time.Now()
	t0 := now.Add(-10 * time.Minute)
	_T = utils.Timeline{T0: t0}

	loadSamples()
	storeSamples()

	adapter := New(repo)

	a, err = repo.GetAgreement(a.Id)
	if err != nil {
		log.Fatalf("Could not read agreement %s from repository: %s", a.Id, err.Error())
	}
	// needs a proper initialized agreement
	if a.Assessment == nil {
		a.Assessment = new(model.Assessment)
	}

	result := assessment.AssessAgreement(a, adapter, now)
	repo.UpdateAgreement(a)
	a, _ = repo.GetAgreement(a.Id)

	/*
	 * We make the assessment pass twice to catch bugs due to CIMI removing empty maps.
	 * E.g., Assessment.Guarantee[x].LastValues could be nil even if initialized in a previous
	 * execution.
	 */

	storeMeasures()
	result = assessment.AssessAgreement(a, adapter, now)
	repo.UpdateAgreement(a)

	/*
	 * The calculated availability should be around 50
	 */
	if len(result.GetViolations()) != 2 {
		t.Errorf("Error in number of violations. Expected: %d; Actual: %d",
			1, len(result.GetViolations()))
	}
	av := result.LastValues[a.Details.Guarantees[0].Name]["availability"].Value
	if math.Abs(av.(float64)-50) > 1 {
		t.Errorf("Error in availability. Expected: ~%d; Actual: %d", 50, av)
	}
}

func createRepository() (cimi.Repository, error) {

	config := viper.New()
	config.SetEnvPrefix("SLA") // Env vars start with 'SLA_'
	config.Set(cimi.InsecureProp, true)
	config.AutomaticEnv()
	repo, err := cimi.New(config)
	return repo, err
}

func loadSamples() {
	var err error
	var aV model.Agreement
	var siV cimi.ServiceInstance

	rand.Seed(time.Now().UnixNano())

	aV, err = model.ReadAgreement("testdata/integration_agreement.json")
	if err != nil {
		log.Fatal(err)
	}
	a = &aV

	siV, err = cimi.ReadServiceInstance("testdata/integration_si.json")
	if err != nil {
		log.Fatal(err)
	}
	si = &siV
	for i := range si.Agents {
		si.Agents[i].DeviceID = fmt.Sprintf("device/%d", rand.Int31())
		si.Agents[i].ContainerID = fmt.Sprintf("%08d", rand.Int31n(10000))
	}

	sor, err = cimi.ReadServiceOperationReport("testdata/integration_sor.json")
	if err != nil {
		log.Fatal(err)
	}
}

func storeSamples() {
	var err error

	a, err = repo.CreateAgreement(a)
	if err != nil {
		log.Fatal(err)
	}

	si.Agreement = a.Id

	si, err = repo.CreateServiceInstance(si)
	if err != nil {
		log.Fatal(err)
	}
}

func storeMeasures() {

	var scm cimi.ServiceContainerMetric

	scm = newServiceContainerMetric(si.Agents[0], 0, 150)
	repo.CreateServiceContainerMetric(&scm)

	scm = newServiceContainerMetric(si.Agents[0], 300, 450)
	repo.CreateServiceContainerMetric(&scm)

	sor.ServiceInstance.Href = si.Id
	_, err := repo.CreateServiceOperationReport(&sor)
	if err != nil {
		log.Fatal(err)
	}
}

func newServiceContainerMetric(agent cimi.Agent, start, stop float64) cimi.ServiceContainerMetric {

	stopTime := _T.T(stop)
	return cimi.ServiceContainerMetric{
		Device:    cimi.Href{Href: agent.DeviceID},
		Container: agent.ContainerID,
		StartTime: _T.T(start),
		StopTime:  &stopTime,
	}
}
