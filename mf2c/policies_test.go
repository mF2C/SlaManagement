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

To run this test, set up a policies component and set env var SLA_POLICIES=<url>
(e.g. SLA_POLICIES=https://localhost:46050/api)
*/

package mf2c

import (
	"SLALite/utils"
	"os"
	"testing"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

var policies Policies

func TestMain(m *testing.M) {
	var err error
	var ok bool
	var policiesURL string
	result := -1

	if policiesURL, ok = os.LookupEnv("SLA_POLICIES"); !ok {
		log.Info("Skipping Policies integration test")
		os.Exit(0)
	}

	if policies, err = createPolicies(policiesURL); err != nil {
		log.Fatalf("Error creating repository: %s", err.Error())
	}

	result = m.Run()

	os.Exit(result)
}

func createPolicies(policiesURL string) (Policies, error) {

	config := viper.New()
	config.SetEnvPrefix(utils.ConfigPrefix) // Env vars start with 'SLA_'
	config.Set(policiesURLProp, policiesURL)
	config.AutomaticEnv()
	policies, err := NewPoliciesClient(config)
	return *policies, err
}

func TestIsLeader(t *testing.T) {
	result, err := policies.IsLeader()
	if err != nil {
		t.Errorf("Error getting leader: %#v", err)
	}
	log.Debugf("Policies.IsLeader = %v", result)
}
