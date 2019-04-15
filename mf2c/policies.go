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
	"SLALite/utils/rest"
	"net/url"
)

const (
	pathAPI            = "api/v2"
	policiesDefaultURL = "http://localhost:46050" + "/" + pathAPI

	pathIsLeader = "resource-management/policies/leaderinfo"
)

// PoliciesConnecter defines the methods that a connector to the Policies component
// must declare.
type PoliciesConnecter interface {
	IsLeader() (bool, error)
}

// Policies is the struct to connect to a Policies component
type Policies struct {
	client *rest.Client
}

// PoliciesMock is the struct that returns predefined answers instead of connecting
// to a Policies component
type PoliciesMock struct {
	isLeader bool
}

type isLeader struct {
	ImBackup bool `json:"imBackup"`
	ImLeader bool `json:"imLeader"`
}

// NewPolicies returns a Policies component client
func NewPolicies(baseurl string) (*Policies, error) {

	url, err := url.Parse(baseurl)
	if err != nil {
		return nil, err
	}
	policies := Policies{
		client: rest.New(url, nil),
	}
	return &policies, nil
}

// NewPoliciesMock constructs a new PoliciesConnector that returns the values
// passed as parameter on construction
// (e.g., IsLeader() returns the parameter isLeader)
func NewPoliciesMock(isLeader bool) PoliciesConnecter {
	return PoliciesMock{
		isLeader: isLeader,
	}
}

// IsLeader returns if the current agent is leader or not
func (o Policies) IsLeader() (bool, error) {
	target := new(isLeader)
	err := o.client.Get(pathIsLeader, &target)
	if err != nil {
		return false, err
	}
	return target.ImLeader, nil
}

// IsLeader returns if the current agent is leader or not
func (o PoliciesMock) IsLeader() (bool, error) {
	return o.isLeader, nil
}
