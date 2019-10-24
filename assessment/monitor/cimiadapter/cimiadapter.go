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
	assessment_model "SLALite/assessment/model"
	"SLALite/assessment/monitor"
	"SLALite/model"
	"SLALite/repositories/cimi"
	"time"
)

const (
	// ExecTime is the name of execution time variable on mF2C
	ExecTime = "execution_time"
	// Availability is the name of the Availability variable on mF2C
	Availability = "availability"

	//compssType = cimi.CompssType

	// catchAllName is the name of the CatchAll guarantee term (i.e., term that applies to operations)
	catchAllName = operationName("*")
)

type operationName string

type adapter struct {
	repository AdapterRepository
	agreement  *model.Agreement
	metrics    map[operationName]assessment_model.GuaranteeData
}

// AdapterRepository is the interface of any repository used to get
// the needed info for assessment
type AdapterRepository interface {
	GetServiceOperationReportsByDate(serviceInstance string, from time.Time) ([]cimi.ServiceOperationReport, error)
	GetServiceInstancesByAgreement(aID string) ([]cimi.ServiceInstance, error)
	GetServiceContainerMetrics(device string, container string, begin time.Time, end time.Time) ([]cimi.ServiceContainerMetric, error)
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

func (ma *adapter) Initialize(a *model.Agreement) monitor.MonitoringAdapter {
	result := *ma
	result.agreement = a

	return &result
}

func (ma *adapter) GetValues(gt model.Guarantee, vars []string, to time.Time) assessment_model.GuaranteeData {
	a := ma.agreement

	var defaultFrom = getDefaultFrom(a, gt)

	values := map[string][]model.MetricValue{}
	for _, name := range vars {
		v, _ := a.Details.GetVariable(name)
		from := getFromForVariable(v, defaultFrom, to)
		if name == ExecTime {
			values[ExecTime] = ma.retrieveExecTime(gt, from)
		} else {
			if from.After(a.Details.Creation) {
				values[Availability] = ma.retrieveAvailability(gt, from, to)
			}
		}
	}
	result := buildExpressionData(values)
	return result
}

func buildExpressionData(valuesmap map[string][]model.MetricValue) assessment_model.GuaranteeData {

	/* XXX: Assume just one variable for the moment */
	result := assessment_model.GuaranteeData{}
	for key, values := range valuesmap {

		for _, value := range values {
			item := assessment_model.ExpressionData{
				key: value,
			}
			result = append(result, item)
		}
		return result /* second and onwards will be skipped */
	}
	return result
}

func (ma *adapter) retrieveExecTime(gt model.Guarantee, from time.Time) []model.MetricValue {
	a := ma.agreement

	result := []model.MetricValue{}

	reports := make([]cimi.ServiceOperationReport, 0, 5)
	sis, err := ma.repository.GetServiceInstancesByAgreement(a.Id)
	if err != nil {
		return nil
	}
	for _, si := range sis {
		siReports, err := ma.repository.GetServiceOperationReportsByDate(si.Id, from)
		if err != nil {
			return nil
		}
		reports = append(reports, siReports...)
	}

	for _, r := range reports {

		if r.Operation != gt.Name && operationName(gt.Name) != catchAllName {
			continue
		}
		mv := model.MetricValue{
			Key:      ExecTime,
			Value:    r.ExecutionTime,
			DateTime: r.Created,
		}
		result = append(result, mv)
	}
	return result
}

func (ma *adapter) retrieveAvailability(gt model.Guarantee, from, to time.Time) []model.MetricValue {
	sis, err := ma.repository.GetServiceInstancesByAgreement(ma.agreement.Id)
	if err != nil {
		return nil
	}

	scms := []cimi.ServiceContainerMetric{}
	for _, si := range sis {
		for _, container := range ma.containers(si) {

			aux, err := ma.repository.GetServiceContainerMetrics("", container, from, to)
			if err != nil {
				return nil
			}
			scms = append(scms, aux...)
		}
	}
	av := calculateAvailability(scms, from, to)

	return []model.MetricValue{
		model.MetricValue{
			Key:      Availability,
			Value:    av,
			DateTime: to,
		},
	}
}

func (ma *adapter) containers(si cimi.ServiceInstance) []string {
	result := make([]string, 0, len(si.Agents))

	for _, a := range si.Agents {
		if si.ServiceType != cimi.CompssType || a.MasterCompss {
			result = append(result, a.ContainerID)
		}
	}
	return result
}

func getDefaultFrom(a *model.Agreement, gt model.Guarantee) time.Time {
	if a.Assessment == nil {
		return a.Details.Creation
	}
	var defaultFrom = a.Assessment.GetGuarantee(gt.Name).LastExecution
	if defaultFrom.IsZero() {
		defaultFrom = a.Assessment.LastExecution
	}
	if defaultFrom.IsZero() {
		defaultFrom = a.Details.Creation
	}
	return defaultFrom
}

func getFromForVariable(v model.Variable, defaultFrom, to time.Time) time.Time {
	if v.Aggregation != nil && v.Aggregation.Window != 0 {
		return to.Add(-time.Duration(v.Aggregation.Window) * time.Second)
	}
	return defaultFrom
}
