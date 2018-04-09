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

package cimi

import (
	"SLALite/model"
)

type userProfileCollection struct {
	Count        int           `json:"count"`
	UserProfiles []userProfile `json:"userProfiles"`
}

type userProfile struct {
	Id              string `json:"id"`
	ServiceConsumer bool   `json:"service_consumer"`
}

type agreementCollection struct {
	Count      int               `json:"count"`
	Agreements []model.Agreement `json:"agreements"`
}

// Violation is the repr. of a CIMI violation
type Violation struct {
	model.Violation
	ACL ACL `json:"acl"`
}

// ACL is the ACL field of any CIMI entity
type ACL struct {
	Owner Principal `json:"owner"`
	Rules []Rule    `json:"rules"`
}

// Principal represents the Principal of a ACL
type Principal struct {
	Principal string `json:"principal"`
	Type      string `json:"type"`
}

// Rule represents a permission on a CIMI entity
type Rule struct {
	Principal string `json:"principal"`
	Type      string `json:"type"`
	Right     string `json:"right"`
}

var adminOwner = Principal{
	Principal: "ADMIN",
	Type:      "ROLE",
}

var userRule = Rule{
	Principal: "USER",
	Type:      "ROLE",
	Right:     "MODIFY",
}

var anonRule = Rule{
	Principal: "ANON",
	Type:      "ROLE",
	Right:     "ALL",
}

var userACL = ACL{
	Owner: adminOwner,
	Rules: []Rule{userRule},
}

var anonACL = ACL{
	Owner: adminOwner,
	Rules: []Rule{anonRule},
}
