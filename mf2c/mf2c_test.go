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
package mf2c

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {

	result := m.Run()

	os.Exit(result)
}

func TestNew(t *testing.T) {
	config := viper.New()
	config.AutomaticEnv()
	mf2c, err := New(config)
	if err != nil {
		t.Errorf("Unexpected error. Error: %s. mf2c: %#v", err, mf2c)
	}
}

func TestNewMock(t *testing.T) {

	config := viper.New()
	config.AutomaticEnv()
	config.Set(isLeaderProp, true)
	mf2c, err := New(config)
	if err != nil {
		t.Errorf("Unexpected error. Error: %s. mf2c: %#v", err, mf2c)
	}
}
