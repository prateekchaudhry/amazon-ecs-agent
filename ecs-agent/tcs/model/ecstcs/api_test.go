//go:build unit
// +build unit

// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.
package ecstcs

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Populate a ULongStatsSet with dummy value
func getDummyULongStatsSet() *ULongStatsSet {
	dummy := aws.Int64(0)
	return &ULongStatsSet{
		Max:         dummy,
		OverflowMax: dummy,
		Min:         dummy,
		OverflowMin: dummy,
		SampleCount: dummy,
		Sum:         dummy,
		OverflowSum: dummy,
	}
}

const bytesUtilizedField = "bytesUtilized"
const maxField = "max"
const minField = "min"
const overflowMaxField = "overflowMax"
const overflowMinField = "overflowMin"
const overflowSumField = "overflowSum"
const sampleCountField = "sampleCount"
const sumField = "sum"

/*
Tests cases of ULongStatsSet that do not raise an error during validate():
 1. Non-nil StatsSet
 3. Non-nil StatsSet with "some" nil values
    a. Nil OverflowMax
    b. Nil OverflowMin
    c. Nil OverflowSum
*/
func TestULongStatsSet(t *testing.T) {
	cases := []struct {
		Name     string
		StatsSet *ULongStatsSet
	}{
		{
			Name:     "happy case",
			StatsSet: getDummyULongStatsSet(),
		},
		{
			Name: "nil OverflowMax",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.OverflowMax = nil
				return s
			}(),
		},
		{
			Name: "nil OverflowMin",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.OverflowMin = nil
				return s
			}(),
		},
		{
			Name: "nil OverflowSum",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.OverflowSum = nil
				return s
			}(),
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			// Marshal to JSON (bytes)
			bytes, err := json.Marshal(test.StatsSet)
			require.NoError(t, err)

			// convert bytes to JSON (string)
			jsonString := string(bytes)

			// test required keys are present
			assert.Contains(t, jsonString, maxField)
			assert.Contains(t, jsonString, minField)
			assert.Contains(t, jsonString, sampleCountField)
			assert.Contains(t, jsonString, sumField)
			// test optional keys if non-nil
			if test.StatsSet.OverflowMax != nil {
				assert.Contains(t, jsonString, overflowMaxField)
			}
			if test.StatsSet.OverflowMin != nil {
				assert.Contains(t, jsonString, overflowMinField)
			}
			if test.StatsSet.OverflowSum != nil {
				assert.Contains(t, jsonString, overflowSumField)
			}

			// validate no errors
			errors := test.StatsSet.Validate()
			require.NoError(t, errors)
		})
	}
}

/*
Tests cases of ULongStatsSet that do raise an error during validate():
 1. Non-nil StatsSet with "some" nil values
    a. Nil Max
    b. Nil Min
    c. Nil SampleCount
    d. Nil Sum
*/
func TestULongStatsSetNilValues(t *testing.T) {
	cases := []struct {
		Name     string
		Field    string
		StatsSet *ULongStatsSet
	}{
		{
			Name:  "nil Max",
			Field: "Max",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.Max = nil
				return s
			}(),
		},
		{
			Name:  "nil Min",
			Field: "Min",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.Min = nil
				return s
			}(),
		},
		{
			Name:  "nil SampleCount",
			Field: "SampleCount",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.SampleCount = nil
				return s
			}(),
		},
		{
			Name:  "nil Sum",
			Field: "Sum",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.Sum = nil
				return s
			}(),
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			// Build error object for comparison
			invalidParams := request.ErrInvalidParams{Context: "ULongStatsSet"}
			invalidParams.Add(request.NewErrParamRequired(test.Field))

			// validate specific error
			errors := test.StatsSet.Validate()
			assert.Equal(t, invalidParams, errors)
		})
	}
}

/*
Tests cases of EphemeralStorageMetrics that do not raise an error during validate():
 1. Non-nil BytesUtilized
 2. Nil BytesUtilized
*/
func TestEphemeralStorageMetrics(t *testing.T) {
	cases := []struct {
		Name     string
		StatsSet *ULongStatsSet
	}{
		{
			Name:     "happy case",
			StatsSet: getDummyULongStatsSet(),
		},
		{
			Name:     "nil StatsSet",
			StatsSet: nil,
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			// Construct EphemeralStorageMetrics
			metrics := EphemeralStorageMetrics{
				BytesUtilized: test.StatsSet,
			}
			// Marshal to JSON (bytes)
			bytes, err := json.Marshal(&metrics)
			require.NoError(t, err)

			// convert bytes to JSON (string)
			jsonString := string(bytes)

			// test optional keys if non-nil
			if test.StatsSet != nil {
				assert.Contains(t, jsonString, bytesUtilizedField)
			}

			// validate no errors
			errors := metrics.Validate()
			require.NoError(t, errors)
		})
	}
}

/*
Tests cases of EphemeralStorageMetrics that do raise an error during validate():
 1. Non-nil StatsSet with "some" nil values
    a. Nil Max
    b. Nil Min
    c. Nil SampleCount
    d. Nil Sum
*/
func TestEphemeralStorageMetricsNilValues(t *testing.T) {
	cases := []struct {
		Name     string
		StatsSet *ULongStatsSet
	}{
		{
			Name: "nil Max",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.Max = nil
				return s
			}(),
		},
		{
			Name: "nil Min",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.Min = nil
				return s
			}(),
		},
		{
			Name: "nil SampleCount",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.SampleCount = nil
				return s
			}(),
		},
		{
			Name: "nil Sum",
			StatsSet: func() *ULongStatsSet {
				s := getDummyULongStatsSet()
				s.Sum = nil
				return s
			}(),
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			// Construct EphemeralStorageMetrics
			metrics := EphemeralStorageMetrics{
				BytesUtilized: test.StatsSet,
			}
			// Build error object for comparison
			err := test.StatsSet.Validate()
			invalidParams := request.ErrInvalidParams{Context: "EphemeralStorageMetrics"}
			invalidParams.AddNested("BytesUtilized", err.(request.ErrInvalidParams))

			// validate specific error
			errors := metrics.Validate()
			assert.Equal(t, invalidParams, errors)
		})
	}
}


// Feature: tacs-payload-gpu-metrics, Property 1: GeneralMetric JSON round-trip preserves all fields.
// Validates: Requirements 1.5, 1.6, 1.7, 1.8
func TestGeneralMetricJSONRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name   string
		Metric GeneralMetric
	}{
		{
			Name:   "all fields nil",
			Metric: GeneralMetric{},
		},
		{
			Name: "only old fields populated",
			Metric: GeneralMetric{
				MetricName:   aws.String("RequestCount"),
				MetricValues: []*float64{aws.Float64(1.0), aws.Float64(2.0)},
				MetricCounts: []*int64{aws.Int64(10), aws.Int64(20)},
			},
		},
		{
			Name: "only new fields populated",
			Metric: GeneralMetric{
				MetricName:         aws.String("GPUUtilization"),
				MetricValueDouble:  aws.Float64(45.5),
				MetricValueInteger: aws.Int64(100),
				Unit:               aws.String("Percent"),
			},
		},
		{
			Name: "all fields populated",
			Metric: GeneralMetric{
				MetricName:         aws.String("GPUMemoryUsed"),
				MetricValues:       []*float64{aws.Float64(3.14)},
				MetricCounts:       []*int64{aws.Int64(1)},
				MetricValueDouble:  aws.Float64(99.9),
				MetricValueInteger: aws.Int64(8589934592),
				Unit:               aws.String("Bytes"),
			},
		},
		{
			Name: "zero values for new fields",
			Metric: GeneralMetric{
				MetricName:         aws.String("GPUPower"),
				MetricValueDouble:  aws.Float64(0.0),
				MetricValueInteger: aws.Int64(0),
				Unit:               aws.String(""),
			},
		},
		{
			Name: "large double value",
			Metric: GeneralMetric{
				MetricName:        aws.String("GPUTemp"),
				MetricValueDouble: aws.Float64(1.7976931348623157e+308),
				Unit:              aws.String("None"),
			},
		},
		{
			Name: "negative double value",
			Metric: GeneralMetric{
				MetricName:        aws.String("GPUThrottleOffset"),
				MetricValueDouble: aws.Float64(-42.5),
			},
		},
		{
			Name: "only MetricValueInteger populated",
			Metric: GeneralMetric{
				MetricName:         aws.String("GPUCount"),
				MetricValueInteger: aws.Int64(4),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			// Serialize to JSON.
			data, err := json.Marshal(&tc.Metric)
			require.NoError(t, err)

			// Deserialize back.
			var roundTripped GeneralMetric
			err = json.Unmarshal(data, &roundTripped)
			require.NoError(t, err)

			// Compare all fields.
			assert.Equal(t, tc.Metric.MetricName, roundTripped.MetricName)
			assert.Equal(t, tc.Metric.MetricValues, roundTripped.MetricValues)
			assert.Equal(t, tc.Metric.MetricCounts, roundTripped.MetricCounts)
			assert.Equal(t, tc.Metric.MetricValueDouble, roundTripped.MetricValueDouble)
			assert.Equal(t, tc.Metric.MetricValueInteger, roundTripped.MetricValueInteger)
			assert.Equal(t, tc.Metric.Unit, roundTripped.Unit)
		})
	}
}

// Feature: tacs-payload-gpu-metrics, Property 2: Nil GeneralMetricsPayload is omitted from JSON.
// Validates: Requirements 2.3, 3.3
func TestNilGeneralMetricsPayloadOmittedFromJSON(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name       string
		StructType string
		Marshal    func() ([]byte, error)
	}{
		{
			Name:       "TaskMetric with nil GeneralMetricsPayload",
			StructType: "TaskMetric",
			Marshal: func() ([]byte, error) {
				tm := TaskMetric{
					TaskArn:               aws.String("arn:aws:ecs:us-east-1:123456789012:task/test"),
					GeneralMetricsPayload: nil,
				}
				return json.Marshal(&tm)
			},
		},
		{
			Name:       "TaskMetric with populated GeneralMetricsPayload",
			StructType: "TaskMetric",
			Marshal: func() ([]byte, error) {
				tm := TaskMetric{
					TaskArn: aws.String("arn:aws:ecs:us-east-1:123456789012:task/test"),
					GeneralMetricsPayload: []*GeneralMetricsWrapper{
						{
							MetricType: aws.String("GPU"),
							GeneralMetrics: []*GeneralMetric{
								{MetricName: aws.String("GPUUtilization"), MetricValueDouble: aws.Float64(50.0), Unit: aws.String("Percent")},
							},
						},
					},
				}
				return json.Marshal(&tm)
			},
		},
		{
			Name:       "InstanceMetrics with nil GeneralMetricsPayload",
			StructType: "InstanceMetrics",
			Marshal: func() ([]byte, error) {
				im := InstanceMetrics{
					Storage:               &InstanceStorageMetrics{RootFilesystem: aws.Float64(80.0)},
					GeneralMetricsPayload: nil,
				}
				return json.Marshal(&im)
			},
		},
		{
			Name:       "InstanceMetrics with populated GeneralMetricsPayload",
			StructType: "InstanceMetrics",
			Marshal: func() ([]byte, error) {
				im := InstanceMetrics{
					Storage: &InstanceStorageMetrics{RootFilesystem: aws.Float64(80.0)},
					GeneralMetricsPayload: []*GeneralMetricsWrapper{
						{
							MetricType: aws.String("GPU"),
							GeneralMetrics: []*GeneralMetric{
								{MetricName: aws.String("GPULimit"), MetricValueInteger: aws.Int64(4), Unit: aws.String("Count")},
							},
						},
					},
				}
				return json.Marshal(&im)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			data, err := tc.Marshal()
			require.NoError(t, err)

			jsonString := string(data)
			var raw map[string]json.RawMessage
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)

			_, hasKey := raw["generalMetricsPayload"]
			if tc.Name == "TaskMetric with nil GeneralMetricsPayload" || tc.Name == "InstanceMetrics with nil GeneralMetricsPayload" {
				assert.False(t, hasKey, "generalMetricsPayload should be omitted when nil, got: %s", jsonString)
			} else {
				assert.True(t, hasKey, "generalMetricsPayload should be present when populated, got: %s", jsonString)
			}
		})
	}
}

// Feature: tacs-payload-gpu-metrics, Property 3: TaskMetric Validate passes with GeneralMetricsPayload.
// Validates: Requirements 2.4
func TestTaskMetricValidateWithGeneralMetricsPayload(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name                  string
		GeneralMetricsPayload []*GeneralMetricsWrapper
	}{
		{
			Name:                  "nil GeneralMetricsPayload",
			GeneralMetricsPayload: nil,
		},
		{
			Name:                  "empty GeneralMetricsPayload",
			GeneralMetricsPayload: []*GeneralMetricsWrapper{},
		},
		{
			Name: "single wrapper with single metric",
			GeneralMetricsPayload: []*GeneralMetricsWrapper{
				{
					MetricType: aws.String("GPU"),
					GeneralMetrics: []*GeneralMetric{
						{MetricName: aws.String("GPUUtilization"), MetricValueDouble: aws.Float64(75.0), Unit: aws.String("Percent")},
					},
				},
			},
		},
		{
			Name: "multiple wrappers with dimensions",
			GeneralMetricsPayload: []*GeneralMetricsWrapper{
				{
					MetricType: aws.String("GPU"),
					Dimensions: []*Dimension{{Key: aws.String("GpuDevice"), Value: aws.String("0")}},
					GeneralMetrics: []*GeneralMetric{
						{MetricName: aws.String("GPUUtilization"), MetricValueDouble: aws.Float64(45.0), Unit: aws.String("Percent")},
						{MetricName: aws.String("GPUMemoryUsed"), MetricValueInteger: aws.Int64(4096), Unit: aws.String("Megabytes")},
					},
				},
				{
					MetricType: aws.String("GPU"),
					Dimensions: []*Dimension{{Key: aws.String("GpuDevice"), Value: aws.String("1")}},
					GeneralMetrics: []*GeneralMetric{
						{MetricName: aws.String("GPUUtilization"), MetricValueDouble: aws.Float64(90.0), Unit: aws.String("Percent")},
					},
				},
			},
		},
		{
			Name: "wrapper with nil GeneralMetrics inside",
			GeneralMetricsPayload: []*GeneralMetricsWrapper{
				{
					MetricType:     aws.String("GPU"),
					GeneralMetrics: nil,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tm := TaskMetric{
				TaskArn:               aws.String("arn:aws:ecs:us-east-1:123456789012:task/test"),
				GeneralMetricsPayload: tc.GeneralMetricsPayload,
			}
			err := tm.Validate()
			assert.NoError(t, err)
		})
	}
}
