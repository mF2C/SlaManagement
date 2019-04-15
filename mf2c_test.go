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
