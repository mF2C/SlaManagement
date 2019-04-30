package mf2c

import (
	"SLALite/utils/rest"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/spf13/viper"
)

/*
Mf2c contains the clients/mocks to the rest of mF2C components
*/
type Mf2c struct {
	Policies PoliciesConnecter
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

	policies, err := NewPolicies(config)
	if err != nil {
		return Mf2c{}, err
	}

	mf2c := Mf2c{
		Policies: policies,
	}
	return mf2c, nil
}

// NewPoliciesClient returns a Policies component client
func NewPoliciesClient(config *viper.Viper) (*Policies, error) {

	baseurl := config.GetString(policiesURLProp)

	url, err := url.Parse(baseurl)
	if err != nil {
		return nil, err
	}
	policies := Policies{
		client: rest.New(url, nil),
	}
	return &policies, nil
}

func setDefaults(config *viper.Viper) {
	config.SetDefault(policiesURLProp, policiesDefaultURL)
}

func logConfig(config *viper.Viper) {
	leader := ""

	if config.GetString(isLeaderProp) != "" {
		leader = fmt.Sprint(config.GetBool(isLeaderProp))
	}
	log.Printf("Policies configuration\n"+
		"\tisLeader: %v\n"+
		"\tURL: %v\n",
		leader,
		config.GetString(policiesURLProp))
}
