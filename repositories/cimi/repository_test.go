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

/*
This tests cimi repository, making use of the repository_testbase file.

To run this test, set up a cimi repo and set env var SLA_REPOSITORY=cimi.
If cimi is not accessible at https://localhost:10443/api, set SLA_CIMIURL=<url>
*/

package cimi

import (
	"SLALite/model"
	"SLALite/repositories"
	"os"
	"testing"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

var repo model.IRepository

func TestMain(m *testing.M) {
	var err error
	result := -1

	if v, ok := os.LookupEnv("SLA_REPOSITORY"); !ok || v != Name {
		log.Info("Skipping CIMI integration test")
		os.Exit(0)
	}

	if repo, err = createRepository(); err != nil {
		log.Fatalf("Error creating repository: %s", err.Error())
	}
	// if err = repositories.CheckSetup(repo); err != nil {
	// 	log.Fatalf("Cannot run test: %s", err.Error())
	// }

	loadSamples()
	setup()
	result = m.Run()
	tearDown()

	os.Exit(result)
}

func createRepository() (model.IRepository, error) {

	config := viper.New()
	config.SetEnvPrefix("SLA") // Env vars start with 'SLA_'
	config.Set(InsecureProp, true)
	config.AutomaticEnv()
	repo, err := New(config)
	return repo, err
}

func loadSamples() {
	var err error

	repositories.Data.A01, err = model.ReadAgreement("testdata/a01.json")
	if err != nil {
		log.Fatal(err)
	}
	repositories.Data.A02, err = model.ReadAgreement("testdata/a02.json")
	if err != nil {
		log.Fatal(err)
	}
	repositories.Data.A03, err = model.ReadAgreement("testdata/a03.json")
	if err != nil {
		log.Fatal(err)
	}
	repositories.Data.T01, err = model.ReadTemplate("testdata/t01.json")
	if err != nil {
		log.Fatal(err)
	}
}

func setup() {
	agreements, _ := repo.GetAllAgreements()
	for _, a := range agreements {
		repo.DeleteAgreement(&a)
	}
}

func tearDown() {
	repo.DeleteAgreement(&repositories.Data.A01)
	repo.DeleteAgreement(&repositories.Data.A02)
	repo.DeleteAgreement(&repositories.Data.A03)
}

func TestRepository(t *testing.T) {
	ctx := repositories.TestContext{Repo: repo}
	/* Agreements */
	t.Run("CreateAgreement", ctx.TestCreateAgreement)
	// N/A in CIMI t.Run("CreateAgreementExists", ctx.TestCreateAgreementExists)
	t.Run("GetAllAgreements", ctx.TestGetAllAgreements)
	t.Run("GetAgreement", ctx.TestGetAgreement)
	t.Run("GetAgreementNotExists", ctx.TestGetAgreementNotExists)
	t.Run("UpdateAgreementState", ctx.TestUpdateAgreementState)
	t.Run("UpdateAgreementStateNotExists", ctx.TestUpdateAgreementStateNotExists)
	// Not implemented t.Run("GetAgreementsByState", testGetAgreementsByState)
	t.Run("UpdateAgreement", ctx.TestUpdateAgreement)
	t.Run("UpdateAgreementNotExists", ctx.TestUpdateAgreementNotExists)
	t.Run("DeleteAgreement", ctx.TestDeleteAgreement)
	// Commented out until CIMI is fixed t.Run("DeleteAgreementNotExists", ctx.TestDeleteAgreementNotExists)

	/* Violations */
	/* Commented out until CIMI is fixed
	t.Run("CreateViolation", ctx.TestCreateViolation)
	// N/A in CIMI t.Run("CreateViolationExists", ctx.TestCreateViolationExists)
	t.Run("GetViolation", ctx.TestGetViolation)
	*/
	t.Run("GetViolationNotExists", ctx.TestGetViolationNotExists)

	/* Templates */
	t.Run("CreateTemplate", ctx.TestCreateTemplate)
	// N/A in CIMI t.Run("CreateTemplateExists", ctx.TestCreateTemplateExists)
	t.Run("GetAllTemplates", ctx.TestGetAllTemplates)
	t.Run("GetTemplate", ctx.TestGetTemplate)
	t.Run("GetTemplateNotExists", ctx.TestGetTemplateNotExists)

	//
	// TODO tests on ServiceOperationReport and ServiceInstance
	//
}
