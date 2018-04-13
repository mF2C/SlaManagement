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
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/spf13/viper"
)

const (
	urlProp           = "cimiurl"
	defaultURL string = "https://localhost:10443/api"

	userProp    = "cimiuser"
	defaultUser = "testuser"

	pwdProp    = "cimipwd"
	defaultPwd = "testpassword"

	insecureProp    = "cimiinsecure"
	defaultInsecure = false

	failfastProp    = "cimifailfast"
	defaultFailfast = true

	anonUser string = "anon"
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

func (r Repository) read(resource string, target interface{}) error {
	if !r.logged {
		err := login(&r)
		if err != nil {
			return err
		}
	}
	resp, err := r.client.Get(r.path(resource))

	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status: %v", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(target)

	return err
}

func (r Repository) post(resource string, entity interface{}) error {
	var err error
	var resp *http.Response
	var jsonValue []byte

	if !r.logged {
		err := login(&r)
		if err != nil {
			return err
		}
	}

	jsonValue, err = json.Marshal(entity)
	if err != nil {
		return err
	}

	reader := bytes.NewBuffer(jsonValue)
	resp, err = r.client.Post(r.path(resource), "application/json", reader)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status: %v", resp.StatusCode)
	}
	return err
}

// GetUserProfiles returns all the user profiles
func (r Repository) GetUserProfiles() ([]userProfile, error) {

	target := new(userProfileCollection)
	err := r.read("user-profile", target)

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
	err := r.read("agreement", target)

	return target.Agreements, err
}

// GetAgreement (see model.Repository)
func (r Repository) GetAgreement(id string) (*model.Agreement, error) {
	return nil, errors.New("Not implemented")
}

// GetActiveAgreements (see model.Repository)
func (r Repository) GetActiveAgreements() (model.Agreements, error) {
	return nil, errors.New("Not implemented")
}

// CreateAgreement (see model.Repository)
func (r Repository) CreateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	return nil, errors.New("Not implemented")
}

// DeleteAgreement (see model.Repository)
func (r Repository) DeleteAgreement(agreement *model.Agreement) error {
	return errors.New("Not implemented")
}

// StartAgreement (see model.Repository)
func (r Repository) StartAgreement(id string) error {
	return errors.New("Not implemented")
}

// StopAgreement (see model.Repository)
func (r Repository) StopAgreement(id string) error {
	return errors.New("Not implemented")
}

// UpdateAgreement (see model.Repository)
func (r Repository) UpdateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	return nil, errors.New("Not implemented")
}

// CreateViolation stores a violation in the CIMI server
func (r *Repository) CreateViolation(v *model.Violation) (*model.Violation, error) {
	var acl ACL

	if r.username == anonUser {
		acl = anonACL
	} else {
		acl = userACL
	}
	cimiv := &Violation{
		*v,
		acl,
	}
	err := r.post("sla-violation", cimiv)
	return v, err
}
