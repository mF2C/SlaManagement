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
	"time"
)

// IRepository expose the interface to be fulfilled by implementations of CIMI repositories.
type IRepository interface {
	CreateViolation(v *model.Violation) (*model.Violation, error)
}

// Href is the entity that represents a resource link to other entity
type Href struct {
	Href string `json:"href"`
}

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

type createResult struct {
	Status     int    `json:"status"`
	Message    string `json:"message"`
	ResourceId string `json:"resource-id"`
}

// Agreement is the repr. of a CIMI agreement
type Agreement struct {
	model.Agreement
	ACL ACL `json:"acl"`
}

// Violation is the repr. of a CIMI violation
type Violation struct {
	Id          string                 `json:"id"`
	AgreementId Href                   `json:"agreement_id"`
	Guarantee   string                 `json:"guarantee"`
	Datetime    time.Time              `json:"datetime"`
	Constraint  string                 `json:"constraint"`
	Values      map[string]interface{} `json:"values"`
	ACL         ACL                    `json:"acl"`
}

// ServiceOperationReport represents the execution time of a service operation in DER
type ServiceOperationReport struct {
	Id              string    `json:"id"`
	ServiceInstance Href      `json:"serviceInstance"`
	Operation       string    `json:"operation"`
	Created         time.Time `json:"created"`
	Updated         time.Time `json:"updated"`
	ExecutionTime   float64   `json:"execution_time"`
	ACL             ACL       `json:"acl"`
}

type serviceOperationReportCollection struct {
	Count                   int                      `json:"count"`
	ServiceOperationReports []ServiceOperationReport `json:"serviceOperationReports"`
}

// ServiceInstance is the entity that represents the execution of a service
type ServiceInstance struct {
	Id        string      `json:"id"`
	ACL       ACL         `json:"acl"`
	User      string      `json:"user"`
	Service   string      `json:"service"`
	Agreement string      `json:"agreement"`
	Status    string      `json:"status"`
	Created   time.Time   `json:"created"`
	Updated   time.Time   `json:"updated"`
	Agents    interface{} `json:"agents"`
}

type serviceInstanceCollection struct {
	Count            int               `json:"count"`
	ServiceInstances []ServiceInstance `json:"serviceInstances"`
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
