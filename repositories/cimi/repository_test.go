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
	"SLALite/utils"
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
		log.Fatal("Error creating repository: %s", err.Error())
	}
	// if err = model.CheckSetup(repo); err != nil {
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
	config.SetEnvPrefix(utils.ConfigPrefix) // Env vars start with 'SLA_'
	config.Set(insecureProp, true)
	config.Set(userProp, anonUser)
	config.Set(pwdProp, "super ADMIN")
	config.Set(urlProp, "https://213.205.14.16:4433/api")
	config.AutomaticEnv()
	repo, err := New(config)
	return repo, err
}

func loadSamples() {
	var err error

	model.Data.A01, err = utils.ReadAgreement("testdata/a01.json")
	if err != nil {
		log.Fatal(err)
	}
	model.Data.A02, err = utils.ReadAgreement("testdata/a02.json")
	if err != nil {
		log.Fatal(err)
	}
	model.Data.A03, err = utils.ReadAgreement("testdata/a03.json")
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
	repo.DeleteAgreement(&model.Data.A01)
	repo.DeleteAgreement(&model.Data.A02)
	repo.DeleteAgreement(&model.Data.A03)
}

func TestRepository(t *testing.T) {
	/* Agreements */
	t.Run("CreateAgreement", testCreateAgreement)
	t.Run("GetAllAgreements", testGetAllAgreements)
	t.Run("GetAgreement", testGetAgreement)
	t.Run("GetAgreementNotExists", testGetAgreementNotExists)
	t.Run("UpdateAgreementState", testUpdateAgreementState)
	t.Run("UpdateAgreementStateNotExists", testUpdateAgreementStateNotExists)
	// t.Run("GetAgreementsByState", testGetAgreementsByState)
	t.Run("UpdateAgreement", testUpdateAgreement)
	t.Run("UpdateAgreementNotExists", testUpdateAgreementNotExists)
	t.Run("DeleteAgreement", testDeleteAgreement)
	t.Run("DeleteAgreementNotExists", testDeleteAgreementNotExists)

	/* Violations */
	t.Run("CreateViolation", testCreateViolation)

	t.Run("GetViolation", testGetViolation)
	t.Run("GetViolationNotExists", testGetViolationNotExists)
}

func testCreateAgreement(t *testing.T) {
	model.TestCreateAgreement(t, repo)
}

func testGetAllAgreements(t *testing.T) {
	model.TestGetAllAgreements(t, repo)
}

func testGetAgreement(t *testing.T) {
	model.TestGetAgreement(t, repo)
}

func testGetAgreementNotExists(t *testing.T) {
	model.TestGetAgreementNotExists(t, repo)
}

func testUpdateAgreementState(t *testing.T) {
	model.TestUpdateAgreementState(t, repo)
}

func testUpdateAgreementStateNotExists(t *testing.T) {
	model.TestUpdateAgreementStateNotExists(t, repo)
}

func testGetAgreementsByState(t *testing.T) {
	model.TestGetAgreementsByState(t, repo)
}

func testUpdateAgreement(t *testing.T) {
	model.TestUpdateAgreement(t, repo)
}

func testUpdateAgreementNotExists(t *testing.T) {
	model.TestUpdateAgreementNotExists(t, repo)
}

func testDeleteAgreement(t *testing.T) {
	model.TestDeleteAgreement(t, repo)
}

func testDeleteAgreementNotExists(t *testing.T) {
	model.TestDeleteAgreementNotExists(t, repo)
}

func testCreateViolation(t *testing.T) {
	model.TestCreateViolation(t, repo)
}

func testCreateViolationExists(t *testing.T) {
	model.TestCreateViolationExists(t, repo)
}

func testGetViolation(t *testing.T) {
	model.TestGetViolation(t, repo)
}

func testGetViolationNotExists(t *testing.T) {
	model.TestGetViolationNotExists(t, repo)
}
