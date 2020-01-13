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

const (
	// CompssType is the type of a COMPSS serviceInstance
	CompssType ServiceType = "compss"
)

// ServiceType is the type a ServiceInstance can be
type ServiceType string

// Interval indicates an interval between two points of time
type Interval struct {
	Start time.Time
	End   time.Time
}

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

type templateCollection struct {
	Count     int              `json:"count"`
	Templates []model.Template `json:"templates"`
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

// Template is the repr. of a CIMI template
type Template struct {
	model.Template
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

// GetId implements model.Identity
func (v *Violation) GetId() string {
	return v.Id
}

// ServiceOperationReport represents the execution time of a service operation in DER
// A ServiceOperationReport is created when an operation is executed, and it is
// updated periodically until the operation has finished. ExecutionTime
// holds zero until that time.
type ServiceOperationReport struct {
	Id              string    `json:"id"`
	ServiceInstance Href      `json:"requesting_application_id"`
	Invocation      string    `json:"operation_id"`
	Created         time.Time `json:"created"`
	Updated         time.Time `json:"updated"`
	ExecutionTime   float64   `json:"execution_length"`
	ComputeNodeID   string    `json:"compute_node_id"`
	ExpectedEndTime time.Time `json:"expected_end_time"`
	Operation       string    `json:"operation_name"`
	Result          string    `json:"result"`
	StartTime       time.Time `json:"start_time"`
	ACL             ACL       `json:"acl"`
}

// GetId implements model.Identity
func (sor *ServiceOperationReport) GetId() string {
	return sor.Id
}

type serviceOperationReportCollection struct {
	Count                   int                      `json:"count"`
	ServiceOperationReports []ServiceOperationReport `json:"serviceOperationReports"`
}

// ServiceContainerMetric stores start and stop times of containers running on a device
type ServiceContainerMetric struct {
	Id        string     `json:"id"`
	Device    Href       `json:"device_id"`
	Container string     `json:"container_id"`
	StartTime time.Time  `json:"start_time"`
	StopTime  *time.Time `json:"stop_time,omitempty"`
	ACL       ACL        `json:"acl"`
}

// GetId implements model.Identity
func (scm *ServiceContainerMetric) GetId() string {
	return scm.Id
}

type serviceContainerMetricCollection struct {
	Count                   int                      `json:"count"`
	ServiceContainerMetrics []ServiceContainerMetric `json:"serviceContainerMetrics"`
}

// ServiceInstance is the entity that represents the execution of a service
type ServiceInstance struct {
	Id             string      `json:"id"`
	ACL            ACL         `json:"acl"`
	User           string      `json:"user"`
	DeviceID       string      `json:"device_id"`
	DeviceIP       string      `json:"device_ip"`
	ParentDeviceID string      `json:"parent_device_id"`
	ParentDeviceIP string      `json:"parent_device_ip"`
	Service        string      `json:"service"`
	Agreement      string      `json:"agreement"`
	Status         string      `json:"status"`
	ServiceType    ServiceType `json:"service_type"`
	Created        time.Time   `json:"created"`
	Updated        time.Time   `json:"updated"`
	Agents         []Agent     `json:"agents"`
}

// Agent represents the list of agents running a service instance
type Agent struct {
	AppType      string      `json:"app_type"`
	URL          string      `json:"url"`
	DeviceID     string      `json:"device_id"`
	Ports        interface{} `json:"ports"`
	Status       string      `json:"status"`
	ContainerID  string      `json:"container_id"`
	Allow        bool        `json:"allow"`
	MasterCompss bool        `json:"master_compss"`
}

// GetId implements model.Identity
func (si *ServiceInstance) GetId() string {
	return si.Id
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
