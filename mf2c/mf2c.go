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

// Package mf2c contains code to connect to rest of components of the mF2C stack
package mf2c

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

const (
	// IsLeaderProp is the env var name that contains the value to build the PoliciesMock
	isLeaderProp = "isleader"

	// analyticsURLProp is the env var name that contains the URL of Analytics component
	analyticsURLProp = "analytics"

	// policiesURLProp is the env var name that contains the URL of Policies component
	policiesURLProp = "policies"
)

/*
Mf2c contains the clients/mocks to the rest of mF2C components
*/
type Mf2c struct {
	Analytics Analytics
	Policies  PoliciesConnecter
}

/*
New constructs an Mf2c struct that contains the clients to mf2c components
*/
func New(config *viper.Viper) (Mf2c, error) {
	if config == nil {
		return Mf2c{}, errors.New("Must provide config to mf2c.New()")
	}
	setDefaults(config)
	logConfig(config)

	analytics, err := NewAnalytics(config.GetString(analyticsURLProp))
	if err != nil {
		return Mf2c{}, err
	}

	policies, err := newPolicies(config)
	if err != nil {
		return Mf2c{}, err
	}

	mf2c := Mf2c{
		Analytics: *analytics,
		Policies:  policies,
	}
	return mf2c, nil
}

// newPolicies is a facade that uses the 'isLeaderProp' configuration
// parameter to return a Policies client or a PoliciesMock
func newPolicies(config *viper.Viper) (PoliciesConnecter, error) {
	if config == nil {
		return nil, errors.New("Must provide config to mf2c.newPolicies()")
	}

	if config.GetString(isLeaderProp) != "" {
		return NewPoliciesMock(config.GetBool(isLeaderProp)), nil
	}
	return NewPolicies(config.GetString(policiesURLProp))
}

func setDefaults(config *viper.Viper) {
	config.SetDefault(policiesURLProp, policiesDefaultURL)
	config.SetDefault(analyticsURLProp, analyticsDefaultURL)
}

func logConfig(config *viper.Viper) {
	leader := ""

	if config.GetString(isLeaderProp) != "" {
		leader = fmt.Sprint(config.GetBool(isLeaderProp))
	}
	log.Printf("mF2C configuration\n"+
		"\tPolicies.isLeader: %v\n"+
		"\tPolicies.URL: %v\n"+
		"\tAnalytics.URL: %v\n",
		leader,
		config.GetString(policiesURLProp),
		config.GetString(analyticsURLProp))
}
