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

/*
Package cimi contains the implementation of a repository using a CIMI server as backend.

See New() for usage.
*/
package cimi

import (
	"SLALite/model"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type method string
type path string

const (
	// Name is the repository identifier
	Name = "cimi"

	urlProp           = "cimiurl"
	defaultURL string = "https://localhost:10443/api"

	userProp    = "cimiuser"
	defaultUser = anonUser

	pwdProp    = "cimipwd"
	defaultPwd = "testpassword"

	InsecureProp    = "cimiinsecure"
	defaultInsecure = false

	failfastProp    = "cimifailfast"
	defaultFailfast = true

	anonUser string = "anon"

	pathSession          path = "session"
	pathOperations       path = "service-operation-report"
	pathAgreements       path = "agreement"
	pathTemplates        path = "sla-template"
	pathViolations       path = "sla-violation"
	pathUserProfiles     path = "user-profile"
	pathServiceInstances path = "service-instance"
	pathContainerMetric  path = "service-container-metric"

	authHeader = "slipstream-authn-info"

	POST   = "POST"
	GET    = "GET"
	PUT    = "PUT"
	DELETE = "DELETE"
)

// Repository implements the model.Repository interface for a CIMI repository.
type Repository struct {
	baseurl   string
	client    *http.Client
	username  string
	password  string
	failfast  bool
	logged    bool
	providers map[string]model.Provider // Not CIMI supported; added just for testing
}

// New creates a Repository according to a configuration, establishing the connection
// to a CIMI server.
//
// The configuration may have the following values:
// - "cimiurl": URL of the CIMI server
// - "cimiuser": user to connect to CIMI server. Anonymous access is used if user is "anon"
// - "cimipwd": user password. Not used on anonymous access.
// - "cimiinsecure": for debugging purposes only!. Do not check certificate.
// - "cimifailfast": Only applicable on not anonymous access. If true, it only tries to login
//   on New(); failing should make the program exit.
//
// If any of these values is not provided, a default value will be used.
// For IT-1, the value of the password will be passed as the value of slipstream-authn-info
// header for requests to CIMI server.
//
// It returns the Repository struct and an error. The possible errors:
// - nil if no error
// - config parameter is nil
// - could not create http client
// - could not connect to CIMI server
// - could not login to CIMI server
func New(config *viper.Viper) (Repository, error) {
	if config == nil {
		return Repository{}, errors.New("Must provide config to cimi.repository.New()")
	}
	setDefaults(config)
	logConfig(config)
	repo := new(Repository)

	baseurl := config.GetString(urlProp)
	username := config.GetString(userProp)
	password := config.GetString(pwdProp)
	insecure := config.GetBool(InsecureProp)
	failfast := config.GetBool(failfastProp)

	client, err := getClient(repo.baseurl, insecure)
	if err == nil {
		repo.baseurl = baseurl
		repo.client = client
		repo.username = username
		repo.password = password
		repo.failfast = failfast
		if username != anonUser {
			err = login(repo)
		} else {
			repo.logged = true
		}
	}
	repo.providers = make(map[string]model.Provider)

	return *repo, err
}

func setDefaults(config *viper.Viper) {
	config.SetDefault(urlProp, defaultURL)
	config.SetDefault(userProp, defaultUser)
	config.SetDefault(pwdProp, defaultPwd)
	config.SetDefault(failfastProp, defaultFailfast)
}

func logConfig(config *viper.Viper) {
	log.Printf("CIMI Repository configuration\n"+
		"\tURL: %v\n"+
		"\tuser: %v\n"+
		"\tinsecure: %v\n"+
		"\tfailfast: %v\n",
		config.GetString(urlProp),
		config.GetString(userProp),
		config.GetBool(InsecureProp),
		config.GetBool(failfastProp))
}

func getClient(url string, insecure bool) (*http.Client, error) {

	jar, err := cookiejar.New(nil)

	if err != nil {
		return nil, err
	}

	var netClient = &http.Client{
		Timeout: time.Second * 10,
		Jar:     jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}

	return netClient, err
}

// login tries to login to CIMI server.
// If failfast is false, if server is unavailable or a 502 will return err = nil.
func login(repo *Repository) error {

	template := map[string]string{
		"href":     "session-template/internal",
		"username": repo.username,
		"password": repo.password,
	}

	values := map[string]map[string]string{
		"sessionTemplate": template,
	}
	jsonValue, _ := json.Marshal(values)

	resp, err := repo.client.Post(repo.path(pathSession), "application/json", bytes.NewBuffer(jsonValue))

	var msg string

	if err == nil {
		/* The msg will not be used on Success or not failfast */
		msg = fmt.Sprintf("%s %s", resp.Status, http.StatusText(resp.StatusCode))
		defer resp.Body.Close()
	} else {
		msg = err.Error()
	}

	if !repo.failfast && (err != nil || resp.StatusCode == http.StatusBadGateway) {
		log.Printf("Could not login to CIMI server: %s. Login deferred.", msg)
		err = nil
	} else {

		if resp != nil && resp.StatusCode == http.StatusCreated {
			repo.logged = true
		} else {
			err = errors.New(msg)
		}
	}
	return err
}

func (r Repository) path(resource path) string {
	return r.baseurl + "/" + string(resource)
}

func (r Repository) stripID(id string) string {
	parts := strings.Split(id, "/")
	// just return the last element
	return parts[len(parts)-1]
}

func (r Repository) subpath(resource path, id string) path {
	if id == "" {
		return resource
	}
	return resource + "/" + path(r.stripID(id))
}

func (r Repository) request(method method, url string, content interface{}, target interface{}) error {

	var err error
	var resp *http.Response
	var jsonValue []byte
	var reader io.Reader

	if content != nil {
		jsonValue, err = json.Marshal(content)
		if err != nil {
			return err
		}
		reader = bytes.NewBuffer(jsonValue)
	}

	req, _ := http.NewRequest(string(method), url, reader)
	req.Header.Set(authHeader, r.password)
	if reader != nil {
		req.Header.Set("Content-type", "application/json")
	}
	resp, err = r.client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return model.ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Unexpected status: %v %s", resp.Status, string(body))
	}

	if target != nil {
		err = json.NewDecoder(resp.Body).Decode(target)
	}

	return err
}

func (r Repository) get(resource path, filter string, target interface{}) error {
	if !r.logged {
		err := login(&r)
		if err != nil {
			return err
		}
	}

	url := r.path(resource)
	if filter != "" {
		url = fmt.Sprintf("%s?$filter=%s", url, filter)
	}
	log.Printf("CimiRepository.read() url=%s", url)

	err := r.request(GET, url, nil, target)

	if entity, ok := target.(model.Identity); ok {
		if entity.GetId() == "" {
			return model.ErrNotFound
		}
	}
	return err
}

func (r Repository) post(resource path, entity interface{}) (string, error) {

	if !r.logged {
		err := login(&r)
		if err != nil {
			return "", err
		}
	}

	url := r.path(resource)
	target := new(createResult)
	err := r.request(POST, url, entity, target)
	if err != nil {
		return "", err
	}
	return target.ResourceId, nil
}

func (r Repository) put(resource path, entity interface{}) error {

	if !r.logged {
		err := login(&r)
		if err != nil {
			return err
		}
	}

	url := r.path(resource)

	err := r.request(PUT, url, entity, nil)

	return err
}

func (r Repository) delete(resource path) error {

	if !r.logged {
		err := login(&r)
		if err != nil {
			return err
		}
	}

	url := r.path(resource)

	err := r.request(DELETE, url, nil, nil)

	return err
}

// GetUserProfiles returns all the user profiles
func (r Repository) getUserProfiles() ([]userProfile, error) {

	target := new(userProfileCollection)
	err := r.get(pathUserProfiles, "", target)

	return target.UserProfiles, err
}

// GetAllProviders (see model.Repository)
func (r Repository) GetAllProviders() (model.Providers, error) {
	result := make(model.Providers, 0, len(r.providers))

	for _, value := range r.providers {
		result = append(result, value)
	}
	return result, nil
}

// GetProvider (see model.Repository)
func (r Repository) GetProvider(id string) (*model.Provider, error) {
	var err error

	item, ok := r.providers[id]

	if ok {
		err = nil
	} else {
		err = model.ErrNotFound
	}
	return &item, err
}

// CreateProvider (see model.Repository)
func (r Repository) CreateProvider(provider *model.Provider) (*model.Provider, error) {
	var err error

	id := provider.Id
	_, ok := r.providers[id]

	if ok {
		err = model.ErrAlreadyExist
	} else {
		r.providers[id] = *provider
		err = nil
	}
	return provider, err
}

// DeleteProvider (see model.Repository)
func (r Repository) DeleteProvider(provider *model.Provider) error {
	var err error

	id := provider.Id

	_, ok := r.providers[id]
	if ok {
		delete(r.providers, id)
		err = nil
	} else {
		err = model.ErrNotFound
	}
	return err
}

// GetAllAgreements (see model.Repository)
func (r Repository) GetAllAgreements() (model.Agreements, error) {
	target := new(agreementCollection)
	err := r.get(pathAgreements, "", target)

	return target.Agreements, err
}

// GetAgreement (see model.Repository)
func (r Repository) GetAgreement(id string) (*model.Agreement, error) {
	target := new(model.Agreement)
	subpath := r.subpath(pathAgreements, id)
	err := r.get(subpath, "", target)

	return target, err
}

// GetAgreementsByState (see model.Repository)
func (r Repository) GetAgreementsByState(states ...model.State) (model.Agreements, error) {
	return nil, errors.New("Not implemented")
}

// CreateAgreement (see model.Repository)
func (r Repository) CreateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	var acl = r.getACL()

	cimia := &Agreement{
		*agreement,
		acl,
	}
	newId, err := r.post(pathAgreements, cimia)
	if err == nil {
		agreement.Id = newId
	}
	return agreement, err
}

// DeleteAgreement (see model.Repository)
func (r Repository) DeleteAgreement(agreement *model.Agreement) error {
	subpath := r.subpath(pathAgreements, agreement.Id)
	err := r.delete(subpath)

	return err
}

// StartAgreement (see model.Repository)
func (r Repository) StartAgreement(id string) error {
	subpath := r.subpath(pathAgreements, id)

	a, err := r.GetAgreement(id)
	if err != nil {
		return err
	}
	a.State = model.STARTED
	var acl = r.getACL()

	cimia := &Agreement{
		*a,
		acl,
	}

	err = r.put(subpath, cimia)
	return err
}

// StopAgreement (see model.Repository)
func (r Repository) StopAgreement(id string) error {
	subpath := r.subpath(pathAgreements, id)
	a, err := r.GetAgreement(id)
	if err != nil {
		return err
	}

	a.State = model.STOPPED
	var acl = r.getACL()

	cimia := &Agreement{
		*a,
		acl,
	}

	err = r.put(subpath, cimia)
	return err
}

// UpdateAgreement (see model.Repository)
func (r Repository) UpdateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	var acl = r.getACL()

	cimia := &Agreement{
		*agreement,
		acl,
	}

	subpath := r.subpath(pathAgreements, agreement.Id)
	err := r.put(subpath, cimia)
	return agreement, err
}

// UpdateAgreementState (see model.Repository)
func (r Repository) UpdateAgreementState(id string, newState model.State) (*model.Agreement, error) {
	a := new(Agreement)

	subpath := r.subpath(pathAgreements, id)
	err := r.get(subpath, "", a)
	if err != nil {
		return nil, err
	}
	a.State = newState
	err = r.put(subpath, a)
	return &a.Agreement, err
}

// GetAllTemplates implements model.IRepository.GetAllTemplates
func (r Repository) GetAllTemplates() (model.Templates, error) {
	target := new(templateCollection)
	err := r.get(pathTemplates, "", target)

	return target.Templates, err

}

// GetTemplate implements model.IRepository.GetTemplate
func (r Repository) GetTemplate(id string) (*model.Template, error) {
	target := new(model.Template)
	subpath := r.subpath(pathTemplates, id)
	err := r.get(subpath, "", target)

	return target, err
}

// CreateTemplate implements model.IRepository.CreateTemplate
func (r Repository) CreateTemplate(template *model.Template) (*model.Template, error) {
	var acl = r.getACL()

	cimit := &Template{
		*template,
		acl,
	}
	newID, err := r.post(pathTemplates, cimit)
	if err == nil {
		template.Id = newID
	}
	return template, err
}

// CreateViolation stores a violation in the CIMI server
func (r Repository) CreateViolation(v *model.Violation) (*model.Violation, error) {
	var acl = r.getACL()

	values := make(map[string]interface{})
	for _, m := range v.Values {
		values[m.Key] = m.Value
	}
	cimiv := &Violation{
		AgreementId: Href{v.AgreementId},
		Datetime:    v.Datetime,
		Guarantee:   v.Guarantee,
		Constraint:  v.Constraint,
		Values:      values,
		ACL:         acl,
	}
	newId, err := r.post(pathViolations, cimiv)
	if err == nil {
		v.Id = newId
	}
	fmt.Printf("cimiv: %#v\n", cimiv)
	fmt.Printf("v: %#v\n", v)
	return v, err
}

// GetViolation gets a violation from the CIMI server by its ID
func (r Repository) GetViolation(id string) (*model.Violation, error) {
	target := new(Violation)
	subpath := r.subpath(pathViolations, id)
	err := r.get(subpath, "", target)
	values := make([]model.MetricValue, 0, 1)
	for k, v := range target.Values {
		m := model.MetricValue{
			Key:      k,
			Value:    v,
			DateTime: target.Datetime,
		}
		values = append(values, m)
	}
	v := &model.Violation{
		Id:          target.Id,
		AgreementId: target.AgreementId.Href,
		Datetime:    target.Datetime,
		Guarantee:   target.Guarantee,
		Constraint:  target.Constraint,
		Values:      values,
	}
	return v, err
}

// CreateServiceOperationReport stores an execution log in the CIMI server
func (r *Repository) CreateServiceOperationReport(e *ServiceOperationReport) (*ServiceOperationReport, error) {
	var acl = r.getACL()

	e.ACL = acl
	newId, err := r.post(pathOperations, e)
	if err != nil {
		return nil, err
	}
	e.Id = newId
	return e, err
}

// GetServiceOperationReportsByDate return the execution logs with creation time newer than a date
func (r Repository) GetServiceOperationReportsByDate(serviceInstance string, from time.Time) ([]ServiceOperationReport, error) {
	target := new(serviceOperationReportCollection)

	t := from.UTC().Format(time.RFC3339)
	err := r.get(pathOperations,
		fmt.Sprintf("(requesting_application_id/href=\"%s\")and(updated>\"%s\")and(execution_length>0)", serviceInstance, t), target)
	return target.ServiceOperationReports, err
}

// DeleteServiceOperationReport deletes a ServiceOperationReport
func (r Repository) DeleteServiceOperationReport(e *ServiceOperationReport) error {
	subpath := r.subpath(pathOperations, e.Id)
	err := r.delete(subpath)

	return err
}

// CreateServiceInstance creates a ServiceInstance
func (r Repository) CreateServiceInstance(si *ServiceInstance) (*ServiceInstance, error) {
	var acl = r.getACL()
	si.ACL = acl

	newID, err := r.post(pathServiceInstances, si)
	if err != nil {
		return nil, err
	}
	si.Id = newID
	return si, err
}

// GetServiceInstancesByAgreement returns the ServiceInstances with a given agreement id.
func (r Repository) GetServiceInstancesByAgreement(aID string) ([]ServiceInstance, error) {
	target := new(serviceInstanceCollection)

	filter := fmt.Sprintf("agreement=\"%s\"", aID)
	err := r.get(pathServiceInstances, filter, target)
	return target.ServiceInstances, err
}

// CreateServiceContainerMetric creates a ServiceContainerMetric
func (r Repository) CreateServiceContainerMetric(e *ServiceContainerMetric) (*ServiceContainerMetric, error) {

	var acl = r.getACL()

	e.ACL = acl
	newID, err := r.post(pathContainerMetric, e)
	if err != nil {
		return nil, err
	}
	e.Id = newID
	return e, err
}

// DeleteServiceContainerMetric deletes a ServiceContainerMetric
func (r Repository) DeleteServiceContainerMetric(e *ServiceContainerMetric) error {
	subpath := r.subpath(pathContainerMetric, e.Id)
	err := r.delete(subpath)

	return err
}

/*
GetServiceContainerMetrics retrieves container metrics from specific device and container
(if set) where the container is up within the interval (begin, end].
*/
func (r Repository) GetServiceContainerMetrics(
	device string, container string, begin time.Time, end time.Time) ([]ServiceContainerMetric, error) {

	var parts = make([]string, 0, 4)
	if device != "" {
		parts = append(parts, fmt.Sprintf("(device_id/href=\"%s\")", device))
	}
	if container != "" {
		parts = append(parts, fmt.Sprintf("(container_id=\"%s\")", container))
	}
	parts = append(parts, fmt.Sprintf("(start_time<\"%s\")", end.UTC().Format(time.RFC3339)))
	parts = append(parts, fmt.Sprintf("((stop_time=null)or(stop_time>\"%s\"))",
		begin.UTC().Format(time.RFC3339)))
	filter := strings.Join(parts, "and")
	target := new(serviceContainerMetricCollection)
	err := r.get(pathContainerMetric, filter, target)

	return target.ServiceContainerMetrics, err
}

func (r *Repository) getACL() ACL {
	if r.username == anonUser {
		return userACL
	}
	return userACL
}

func (r Repository) getServiceInstance(siID string) (*ServiceInstance, error) {
	target := new(ServiceInstance)

	subpath := r.subpath(pathServiceInstances, siID)
	err := r.get(subpath, "", target)
	return target, err
}

func (r Repository) updateServiceInstance(si *ServiceInstance) (*ServiceInstance, error) {
	subpath := r.subpath(pathServiceInstances, si.Id)
	err := r.put(subpath, si)
	return si, err
}
