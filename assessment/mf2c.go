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
	"SLALite/assessment/monitor"
	"SLALite/model"
	"SLALite/repositories/cimi"
	"log"
	"time"
)

/*
This file contains the mF2C asssessment code
*/

// AssessMf2cAgreements is the main process for the mf2c assessment
func AssessMf2cAgreements(repo model.IRepository, mf2cRepo cimi.IRepository, ma monitor.MonitoringAdapter) {
	agreements, err := repo.GetAllAgreements()
	log.Printf("Running assessment. Processing %d agreement(s)", len(agreements))
	if err != nil {
		log.Printf("Error getting agreements: %v\n", err)
		return
	}

	now := time.Now()

	for _, a := range agreements {
		log.Printf("Evaluating agreement %s", a.Id)
		if a.State == model.STARTED && a.Assessment == nil {
			a.Assessment = new(model.Assessment)
		}

		var result = AssessAgreement(&a, ma, now)
		log.Printf("Result: %v\n", result)

		for _, v := range result.GetViolations() {
			pv := &v
			pv, err = mf2cRepo.CreateViolation(pv)
			if err != nil {
				log.Printf("Error creating violation: %v", err)
			}
		}
		_, err = repo.UpdateAgreement(&a)
		if err != nil {
			log.Printf("Error updating agreement: %v", err)
		}
	}
}
