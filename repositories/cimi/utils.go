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

package cimi

import (
	"encoding/json"
	"os"
)

// ReadServiceInstance returns a ServiceInstance read from file
func ReadServiceInstance(path string) (ServiceInstance, error) {
	res, err := readEntity(path, new(ServiceInstance))
	o := res.(*ServiceInstance)

	return *o, err
}

func readEntity(path string, result interface{}) (interface{}, error) {

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return result, err
	}
	json.NewDecoder(f).Decode(&result)
	return result, err
}
