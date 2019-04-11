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

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
)

// Method is the type for allowed REST verbs
type Method string

// Path is the type for REST subpaths from the base URL
type Path string

const (
	// POST method
	POST = "POST"
	// GET method
	GET = "GET"
	// PUT method
	PUT = "PUT"
	// DELETE method
	DELETE = "DELETE"
)

// Error represents REST errors raised when requests return 4xx or 5xx code
type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

/*
Client is a type to build a basic REST client.

It is intended to be used for application/json content type requests.

The struct contains a base URL and an http.Client. It can be built directly
or using NewRestClient. The former allows passing a nil http.Client so the
RestClient is initialized with a default http.Client.

Request methods (Request, Post, Get...) return errors on fail.

* Requests with 4xx or 5xx status are return as Error type

* Other errors coming from http.Client or serialization may be returned

	url := &url.URL{Path: baseurl}
	client := rest.New(url, nil)
*/
type Client struct {
	BaseURL *url.URL
	Client  *http.Client
}

/*
New builds a REST client.

If httpClient is nil, it is intialized with a default http.Client.
Provide httpClient to support cookies, additional TLS configuration...
*/
func New(baseurl *url.URL, httpClient *http.Client) *Client {
	return &Client{
		BaseURL: baseurl,
		Client:  buildHTTPClient(httpClient),
	}
}

/*
Request makes a REST request to the URL derived from the BaseURL and the subpath 'path',
using the 'method' verb.

The content parameter is passed as the request body, and will be sent with
application/json content-type. The response body will be parsed as JSON and returned in target.
*/
func (r *Client) Request(method Method, subpath Path, content interface{}, target interface{}) error {

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

	/*
	 * ResolveReference should be used to calculate the derived url, but
	 * if BaseURL specifies a directory, the directory is discarded
	 * (i.e. ResolveReference("http://host/api", "users") -> "http://host/users")
	 *
	 * This is why path.Join is used (which is somehow dangerous)
	 */

	//url := r.BaseURL.ResolveReference(&url.URL{Path: string(subpath)})
	url := *r.BaseURL
	url.Path = path.Join(r.BaseURL.Path, string(subpath))
	req, _ := http.NewRequest(string(method), url.String(), reader)
	if reader != nil {
		req.Header.Set("Content-type", "application/json")
	}
	resp, err = r.Client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return Error{404, fmt.Sprintf("Path '%s' not found", subpath)}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return Error{resp.StatusCode, string(body)}
	}

	if target != nil {
		err = json.NewDecoder(resp.Body).Decode(target)
	}

	return err
}

// Get makes a GET request. See Request.
func (r *Client) Get(resource Path, target interface{}) error {

	err := r.Request(GET, resource, nil, target)

	return err
}

// Post makes a POST request. See Request.
func (r *Client) Post(resource Path, entity interface{}, target interface{}) error {

	err := r.Request(POST, resource, entity, target)
	if err != nil {
		return err
	}
	return nil
}

// Put makes a PUT request. See Request.
func (r *Client) Put(resource Path, entity interface{}) error {

	err := r.Request(PUT, resource, entity, nil)

	return err
}

// Delete makes a DELETE request. See Request.
func (r *Client) Delete(resource Path) error {

	err := r.Request(DELETE, resource, nil, nil)

	return err
}

func buildHTTPClient(httpClient *http.Client) *http.Client {
	if httpClient != nil {
		return httpClient
	}
	var result = &http.Client{
		Timeout: time.Second * 10,
	}
	return result
}
