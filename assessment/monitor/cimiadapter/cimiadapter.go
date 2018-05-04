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

// Package cimiadapter provides the monitoring adapter that get values
// from a CIMI repository.
// TODO: Allow more than one variable per guarantee
package cimiadapter

import (
	"SLALite/assessment/monitor"
	"SLALite/model"
	"SLALite/repositories/cimi"
	"log"
	"time"
)

const (
	// ExecTimeName Name of execution time metric on mF2C
	ExecTimeName = "execution_time"
)

type operationName string

type adapter struct {
	repository AdapterRepository
	agreement  *model.Agreement
	metrics    map[operationName][]monitor.MetricValue
	i          map[operationName]int // index over array map[operationName][]MetricValue
}

// AdapterRepository is the interface of any repository used to get
// the needed info for assessment
type AdapterRepository interface {
	GetServiceOperationReportsByDate(serviceInstance string, from time.Time) ([]cimi.ServiceOperationReport, error)
	GetServiceInstancesByAgreement(aID string) ([]cimi.ServiceInstance, error)
}

// New returns a CIMI monitoring adapter
// Usage:
//   ma = cimiadapter.New()
//   ma.Initialize(agreement)
//   assessment.AssessAgreement(&a, ma, time.Now())
func New(repo AdapterRepository) monitor.MonitoringAdapter {
	return &adapter{
		repository: repo,
		agreement:  nil,
	}
}

func (ma *adapter) Initialize(a *model.Agreement) {
	var from time.Time
	if a.Assessment == nil {
		from = a.Details.Creation
	} else {
		from = a.Assessment.LastExecution
	}
	sis, err := ma.repository.GetServiceInstancesByAgreement(a.Id)
	reports := make([]cimi.ServiceOperationReport, 0, 5)
	for _, si := range sis {
		siReports, err := ma.repository.GetServiceOperationReportsByDate(si.Id, from)
		if err != nil {
			log.Printf("Error initializing adapter: %v", err)
			return
		}
		reports = append(reports, siReports...)
	}
	if err != nil {
		log.Printf("Error initializing adapter: %v", err)
		return
	}
	log.Printf("cimiadapter.Initialize(): reports=%#v", reports)

	ma.agreement = a
	ma.metrics = make(map[operationName][]monitor.MetricValue)
	ma.i = make(map[operationName]int)
	for _, r := range reports {
		mv := monitor.MetricValue{
			Key:      ExecTimeName,
			Value:    r.ExecutionTime,
			DateTime: r.Created,
		}

		var op = operationName(r.Operation)
		if _, ok := ma.metrics[op]; !ok {
			ma.metrics[op] = make([]monitor.MetricValue, 0)
			ma.i[op] = 0
		}
		ma.metrics[op] = append(ma.metrics[op], mv)
	}
}

func (ma *adapter) NextValues(gt model.Guarantee) map[string]monitor.MetricValue {
	// XXX We are assuming for IT-1 only one var per constraint

	var op = operationName(gt.Name)
	i := ma.i[op]
	l := ma.metrics[op]

	if i == len(l) {
		return nil
	}
	m := l[i]
	result := make(map[string]monitor.MetricValue)
	result[m.Key] = m
	ma.i[op]++

	return result
}
