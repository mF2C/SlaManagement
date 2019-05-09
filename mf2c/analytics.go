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
	"net/http"
	"net/url"
	"time"
)

const (
	analyticsDefaultURL string = "http://localhost:46020"
	pathOptimal                = "mf2c/optimal"
)

// Analytics is the struct to connect to a Analytics component
type Analytics struct {
	client *rest.Client
}

/*
AnalyticsItem is the type that represents an item of the result array of Analytics component

An example is:

{
	'ipaddress': '172.28.0.20',
	'mf2c_device_id': '52523f94-f454-4d3c-97b7-486c7f74e176',
	'network utilization': 0.0,
	'compute saturation': 0.0,
	'node_name': 'mf2c-leader',
	'disk utilization': 0.0,
	'compute utilization': 0.0,
	'memory saturation': 0.0,
	'type': 'machine',
	'memory utilization': 0.0,
	'network saturation': 0.0,
	'disk saturation': 0.0
}
*/
type AnalyticsItem struct {
	NodeName           string  `json:"node_name"`
	Type               string  `json:"type"`
	IPAddress          string  `json:"ipaddress"`
	Mf2cDeviceID       string  `json:"mf2c_device_id"`
	ComputeSaturation  float32 `json:"compute saturation"`
	ComputeUtilization float32 `json:"compute utilization"`
	DiskSaturation     float32 `json:"disk saturation"`
	DiskUtilization    float32 `json:"disk utilization"`
	MemorySaturation   float32 `json:"memory saturation"`
	MemoryUtilization  float32 `json:"memory utilization"`
	NetworkSaturation  float32 `json:"network saturation"`
	NetworkUtilization float32 `json:"network utilization"`
}

// OptimalRequest is the request body for the optimal operation
type OptimalRequest struct {
	Name string `json:"name"`
}

// NewAnalytics returns a Analytics component client
func NewAnalytics(baseurl string) (*Analytics, error) {

	url, err := url.Parse(baseurl)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	analytics := Analytics{
		client: rest.New(url, client),
	}
	return &analytics, nil
}

// Optimal returns output of optimal call API
func (o *Analytics) Optimal() ([]AnalyticsItem, error) {
	target := new([]AnalyticsItem)
	body := OptimalRequest{
		Name: "clearwater_ims",
	}
	err := o.client.Post(pathOptimal, body, target)
	if err != nil {
		return nil, err
	}
	return *target, nil

}
