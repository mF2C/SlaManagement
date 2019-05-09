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

package main

import (
	"SLALite/model"
	"SLALite/utils"
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMf2c(t *testing.T) {
	t.Run("Generate agreement from template", testMf2cCreateAgreementFromTemplate)
}

func testMf2cCreateAgreementFromTemplate(t *testing.T) {
	var err error
	var tpl, _ = utils.ReadTemplate("repositories/cimi/testdata/template.json")

	ptpl, err := repo.CreateTemplate(&tpl)
	if err != nil {
		t.Errorf("Could not create fixture for template: %s", err.Error())
		return
	}

	ca := model.CreateAgreement{
		TemplateID: ptpl.Id,
		Parameters: map[string]interface{}{
			"user": "mf2c-user",
		},
	}
	body, err := json.Marshal(ca)
	if err != nil {
		t.Error("Unexpected marshalling error")
	}

	req, _ := http.NewRequest("POST", "/mf2c/create-agreement", bytes.NewBuffer(body))
	res := request(req)

	checkStatus(t, http.StatusCreated, res.Code)
	if res.Code != http.StatusCreated {
		var e ApiError
		_ = json.NewDecoder(res.Body).Decode(&e)
		log.Infof("Error=%#v", e)
		return
	}

	var created model.CreateAgreement
	_ = json.NewDecoder(res.Body).Decode(&created)

	a, _ := repo.GetAgreement(created.AgreementID)
	log.Infof("Generated agreement: %#v", a)
}
