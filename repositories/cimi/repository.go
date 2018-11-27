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

const (
	urlProp           = "cimiurl"
	defaultURL string = "https://localhost:10443/api"

	userProp    = "cimiuser"
	defaultUser = anonUser

	pwdProp    = "cimipwd"
	defaultPwd = "testpassword"

	insecureProp    = "cimiinsecure"
	defaultInsecure = false

	failfastProp    = "cimifailfast"
	defaultFailfast = true

	anonUser string = "anon"

	pathOperations       = "service-operation-report"
	pathAgreements       = "agreement"
	pathViolations       = "sla-violation"
	pathUserProfiles     = "user-profile"
	pathServiceInstances = "service-instance"

	authHeader = "slipstream-authn-info"

	POST   = "POST"
	GET    = "GET"
	PUT    = "PUT"
	DELETE = "DELETE"
)

// Repository implements the model.Repository interface for a CIMI repository.
type Repository struct {
	baseurl  string
	client   *http.Client
	username string
	password string
	failfast bool
	logged   bool
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
	insecure := config.GetBool(insecureProp)
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
		config.GetBool(insecureProp),
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

	resp, err := repo.client.Post(repo.path("session"), "application/json", bytes.NewBuffer(jsonValue))

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

func (r Repository) path(resource string) string {
	return r.baseurl + "/" + resource
}

func (r Repository) stripID(id string) string {
	parts := strings.Split(id, "/")
	// just return the last element
	return parts[len(parts)-1]
}

func (r Repository) subpath(resource, id string) string {
	if id == "" {
		return resource
	}
	return resource + "/" + r.stripID(id)
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Unexpected status: %v %s", resp.Status, string(body))
	}

	if target != nil {
		err = json.NewDecoder(resp.Body).Decode(target)
	}

	return err
}

func (r Repository) get(resource string, filter string, target interface{}) error {
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

	return err
}

func (r Repository) post(resource string, entity interface{}) error {

	if !r.logged {
		err := login(&r)
		if err != nil {
			return err
		}
	}

	url := r.path(resource)

	err := r.request(POST, url, entity, nil)

	return err
}

func (r Repository) put(resource string, entity interface{}) error {

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

func (r Repository) delete(resource string) error {

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
	return nil, errors.New("Not implemented")
}

// GetProvider (see model.Repository)
func (r Repository) GetProvider(id string) (*model.Provider, error) {
	return nil, errors.New("Not implemented")
}

// CreateProvider (see model.Repository)
func (r Repository) CreateProvider(provider *model.Provider) (*model.Provider, error) {
	return nil, errors.New("Not implemented")
}

// DeleteProvider (see model.Repository)
func (r Repository) DeleteProvider(provider *model.Provider) error {
	return errors.New("Not implemented")
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
	return nil, errors.New("Not implemented")
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
	a, err := r.GetAgreement(subpath)
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
	a, err := r.GetAgreement(subpath)
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
	a.State = model.STOPPED
	err = r.put(subpath, a)
	return &a.Agreement, err
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
	err := r.post(pathViolations, cimiv)
	return v, err
}

// GetViolation gets a violation from the CIMI server by its ID
func (r Repository) GetViolation(id string) (*model.Violation, error) {
	target := new(model.Violation)
	subpath := r.subpath(pathViolations, id)
	err := r.get(subpath, "", target)

	return target, err
}

// CreateServiceOperationReport stores an execution log in the CIMI server
func (r *Repository) CreateServiceOperationReport(e *ServiceOperationReport) (*ServiceOperationReport, error) {
	var acl = r.getACL()

	e.ACL = acl
	err := r.post(pathOperations, e)
	return e, err
}

// GetServiceOperationReportsByDate return the execution logs with creation time newer than a date
func (r Repository) GetServiceOperationReportsByDate(serviceInstance string, from time.Time) ([]ServiceOperationReport, error) {
	target := new(serviceOperationReportCollection)

	t := from.UTC().Format(time.RFC3339)
	err := r.get(pathOperations,
		fmt.Sprintf("(serviceInstance/href=\"%s\")and(created>\"%s\")", serviceInstance, t), target)
	return target.ServiceOperationReports, err
}

// GetServiceInstancesByAgreement returns the ServiceInstances with a given agreement id.
func (r Repository) GetServiceInstancesByAgreement(aID string) ([]ServiceInstance, error) {
	target := new(serviceInstanceCollection)

	filter := fmt.Sprintf("agreement=\"%s\"", aID)
	err := r.get(pathServiceInstances, filter, target)
	return target.ServiceInstances, err
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
