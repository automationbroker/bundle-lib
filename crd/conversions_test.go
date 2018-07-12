//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package crd

import (
	"testing"

	"github.com/automationbroker/broker-client-go/pkg/apis/automationbroker/v1alpha1"
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertJobMethodToCRD(t *testing.T) {
	testCases := []struct {
		name     string
		input    bundle.JobMethod
		expected v1alpha1.JobMethod
	}{
		{
			name:     "bundle provision job method",
			input:    bundle.JobMethodProvision,
			expected: v1alpha1.JobMethodProvision,
		},
		{
			name:     "bundle deprovision job method",
			input:    bundle.JobMethodDeprovision,
			expected: v1alpha1.JobMethodDeprovision,
		},
		{
			name:     "bundle bind job method",
			input:    bundle.JobMethodBind,
			expected: v1alpha1.JobMethodBind,
		},
		{
			name:     "bundle unbind job method",
			input:    bundle.JobMethodUnbind,
			expected: v1alpha1.JobMethodUnbind,
		},
		{
			name:     "bundle update job method",
			input:    bundle.JobMethodUpdate,
			expected: v1alpha1.JobMethodUpdate,
		},
		{
			name:     "bundle empty job method",
			input:    "",
			expected: v1alpha1.JobMethodProvision,
		},
		{
			name:     "bundle unknown job method",
			input:    "unknown",
			expected: v1alpha1.JobMethodProvision,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConvertJobMethodToCRD(tc.input))
		})
	}
}

func TestConvertJobMethodToAPB(t *testing.T) {
	testCases := []struct {
		name     string
		input    v1alpha1.JobMethod
		expected bundle.JobMethod
	}{
		{
			name:     "crd provision job method",
			input:    v1alpha1.JobMethodProvision,
			expected: bundle.JobMethodProvision,
		},
		{
			name:     "crd deprovision job method",
			input:    v1alpha1.JobMethodDeprovision,
			expected: bundle.JobMethodDeprovision,
		},
		{
			name:     "crd bind job method",
			input:    v1alpha1.JobMethodBind,
			expected: bundle.JobMethodBind,
		},
		{
			name:     "crd unbind job method",
			input:    v1alpha1.JobMethodUnbind,
			expected: bundle.JobMethodUnbind,
		},
		{
			name:     "crd update job method",
			input:    v1alpha1.JobMethodUpdate,
			expected: bundle.JobMethodUpdate,
		},
		{
			name:     "empty job method",
			input:    "",
			expected: bundle.JobMethodProvision,
		},
		{
			name:     "unknown job method",
			input:    "unknown",
			expected: bundle.JobMethodProvision,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConvertJobMethodToAPB(tc.input))
		})
	}
}

func TestConvertStateToAPB(t *testing.T) {
	testCases := []struct {
		name     string
		input    v1alpha1.State
		expected bundle.State
	}{
		{
			name:     "crd not yet started state",
			input:    v1alpha1.StateNotYetStarted,
			expected: bundle.StateNotYetStarted,
		},
		{
			name:     "crd in progress state",
			input:    v1alpha1.StateInProgress,
			expected: bundle.StateInProgress,
		},
		{
			name:     "crd succeeded state",
			input:    v1alpha1.StateSucceeded,
			expected: bundle.StateSucceeded,
		},
		{
			name:     "crd failed state",
			input:    v1alpha1.StateFailed,
			expected: bundle.StateFailed,
		},
		{
			name:     "empty state",
			input:    "",
			expected: bundle.StateFailed,
		},
		{
			name:     "unknown state",
			input:    "unknown",
			expected: bundle.StateFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConvertStateToAPB(tc.input))
		})
	}
}

func TestConvertStateToCRD(t *testing.T) {
	testCases := []struct {
		name     string
		input    bundle.State
		expected v1alpha1.State
	}{
		{
			name:     "bundle not yet started state",
			input:    bundle.StateNotYetStarted,
			expected: v1alpha1.StateNotYetStarted,
		},
		{
			name:     "bundle in progress state",
			input:    bundle.StateInProgress,
			expected: v1alpha1.StateInProgress,
		},
		{
			name:     "bundle succeeded state",
			input:    bundle.StateSucceeded,
			expected: v1alpha1.StateSucceeded,
		},
		{
			name:     "bundle failed state",
			input:    bundle.StateFailed,
			expected: v1alpha1.StateFailed,
		},
		{
			name:     "empty state",
			input:    "",
			expected: v1alpha1.StateFailed,
		},
		{
			name:     "unknown state",
			input:    "unknown",
			expected: v1alpha1.StateFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConvertStateToCRD(tc.input))
		})
	}
}

func TestConvertServiceBindingToAPB(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       v1alpha1.BundleBinding
		expected    *bundle.BindInstance
		expectederr bool
	}{
		{
			name:  "BundleBinding zero value",
			input: v1alpha1.BundleBinding{},
			expected: &bundle.BindInstance{
				Parameters: &bundle.Parameters{},
			},
		},
		{
			name: "invalid json string should return error",
			input: v1alpha1.BundleBinding{
				Spec: v1alpha1.BundleBindingSpec{
					BundleInstance: v1alpha1.LocalObjectReference{
						Name: "mynameis",
					},
					// removed final curly to make it invalid json
					Parameters: `{"_apb_creds":"letmein","foo":"bar"`,
				},
			},
			expected:    &bundle.BindInstance{},
			expectederr: true,
		},
		{
			name: "parameters should get copied",
			input: v1alpha1.BundleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      uid,
					Namespace: "testing",
				},
				Spec: v1alpha1.BundleBindingSpec{
					BundleInstance: v1alpha1.LocalObjectReference{
						Name: uid,
					},
					Parameters: `{"_apb_creds":"letmein","foo":"bar"}`,
				},
			},
			expected: &bundle.BindInstance{
				ID:        uuid.Parse(uid),
				ServiceID: uuid.Parse(uid),
				Parameters: &bundle.Parameters{
					"foo":        "bar",
					"_apb_creds": "letmein",
				},
			},
			expectederr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ConvertServiceBindingToAPB(tc.input, tc.input.GetName())
			if tc.expectederr {
				assert.Error(t, err)
			}
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertServiceBindingToCRD(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       *bundle.BindInstance
		expected    v1alpha1.BundleBinding
		expectederr bool
	}{
		{
			name:     "BindInstance zero value",
			input:    &bundle.BindInstance{},
			expected: v1alpha1.BundleBinding{},
		},
		{
			name: "parameters should get copied",
			input: &bundle.BindInstance{
				ID:        uuid.Parse(uid),
				ServiceID: uuid.Parse(uid),
				Parameters: &bundle.Parameters{
					"foo":        "bar",
					"_apb_creds": "letmein",
				},
			},
			expected: v1alpha1.BundleBinding{
				Spec: v1alpha1.BundleBindingSpec{
					BundleInstance: v1alpha1.LocalObjectReference{
						Name: uid,
					},
					Parameters: `{"_apb_creds":"letmein","foo":"bar"}`,
				},
			},
			expectederr: false,
		},
		{
			name: "invalid parameters should return error",
			input: &bundle.BindInstance{
				ID:        uuid.Parse(uid),
				ServiceID: uuid.Parse(uid),
				Parameters: &bundle.Parameters{
					// force json marshal to fail
					"foo": make(chan int),
				},
			},
			expected:    v1alpha1.BundleBinding{},
			expectederr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ConvertServiceBindingToCRD(tc.input)
			if tc.expectederr {
				assert.Error(t, err)
			}
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertServiceInstanceToAPB(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       v1alpha1.BundleInstance
		spec        *bundle.Spec
		expected    *bundle.ServiceInstance
		expectederr bool
	}{
		{
			name:  "BindInstance zero value",
			input: v1alpha1.BundleInstance{},
			spec:  &bundle.Spec{},
			expected: &bundle.ServiceInstance{
				ID:         uuid.Parse(uid),
				Spec:       &bundle.Spec{},
				Context:    &bundle.Context{},
				Parameters: &bundle.Parameters{},
				BindingIDs: map[string]bool{},
			},
		},
		{
			name: "parameters should get copied",
			input: v1alpha1.BundleInstance{
				Spec: v1alpha1.BundleInstanceSpec{
					Bundle: v1alpha1.LocalObjectReference{Name: uid},
					Context: v1alpha1.Context{
						Namespace: "testnamespace",
						Platform:  "kubernetes",
					},
					Parameters:   `{"_apb_creds":"letmein","foo":"bar"}`,
					DashboardURL: "http://example.com/dashboard",
				},
				Status: v1alpha1.BundleInstanceStatus{
					Bindings: []v1alpha1.LocalObjectReference{
						{
							Name: "a binding",
						},
					},
				},
			},
			spec: &bundle.Spec{},
			expected: &bundle.ServiceInstance{
				ID:   uuid.Parse(uid),
				Spec: &bundle.Spec{},
				Context: &bundle.Context{
					Namespace: "testnamespace",
					Platform:  "kubernetes",
				},
				Parameters: &bundle.Parameters{
					"foo":        "bar",
					"_apb_creds": "letmein",
				},
				BindingIDs: map[string]bool{
					"a binding": true,
				},
				DashboardURL: "http://example.com/dashboard",
			},

			expectederr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ConvertServiceInstanceToAPB(tc.input, tc.spec, uid)
			if tc.expectederr {
				assert.Error(t, err)
			}
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertSpecToBundle(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       *bundle.Spec
		expected    v1alpha1.BundleSpec
		expectederr bool
	}{
		{
			name:  "bundle.Spec zero value",
			input: &bundle.Spec{},
			expected: v1alpha1.BundleSpec{
				Async:    convertToAsyncType("required"),
				Metadata: "null",
				Alpha:    "null",
				Plans:    []v1alpha1.Plan{},
			},
		},
		{
			name: "parameters should get copied",
			input: &bundle.Spec{
				ID:          uid,
				Runtime:     2,
				Version:     "1.2.3",
				FQName:      "chevy/camaro-apb",
				Image:       "chevy/cavalier-apb",
				Tags:        []string{"cars", "chevy"},
				Bindable:    true,
				Description: "description",
				Async:       "optional",
				Metadata: map[string]interface{}{
					"_apb_creds": "letmein",
					"foo":        "bar",
				},
				Alpha: map[string]interface{}{
					"alpha_apb_creds": "letmein",
					"alphafoo":        "bar",
				},
				Plans: []bundle.Plan{
					{
						Name:     "dev",
						Bindable: true,
						Metadata: map[string]interface{}{
							"plan_param1": "letmein",
							"plan_param2": "bar",
						},
						Parameters:     []bundle.ParameterDescriptor{},
						BindParameters: []bundle.ParameterDescriptor{},
					},
				},
			},
			expected: v1alpha1.BundleSpec{
				Runtime:     2,
				Version:     "1.2.3",
				FQName:      "chevy/camaro-apb",
				Image:       "chevy/cavalier-apb",
				Tags:        []string{"cars", "chevy"},
				Bindable:    true,
				Description: "description",
				Async:       convertToAsyncType("optional"),
				Metadata:    `{"_apb_creds":"letmein","foo":"bar"}`,
				Alpha:       `{"alpha_apb_creds":"letmein","alphafoo":"bar"}`,
				Plans: []v1alpha1.Plan{
					{
						Name:           "dev",
						Bindable:       true,
						Metadata:       `{"plan_param1":"letmein","plan_param2":"bar"}`,
						Parameters:     []v1alpha1.Parameter{},
						BindParameters: []v1alpha1.Parameter{},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			output, err := ConvertSpecToBundle(tc.input)
			if tc.expectederr {
				assert.Error(t, err)
			}

			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertBundleToSpec(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       v1alpha1.BundleSpec
		expected    *bundle.Spec
		expectederr bool
	}{
		{
			name:        "BundleSpec zero value",
			input:       v1alpha1.BundleSpec{},
			expected:    &bundle.Spec{},
			expectederr: true,
		},
		{
			name: "parameters should get copied",
			input: v1alpha1.BundleSpec{
				Runtime:     2,
				Version:     "1.2.3",
				FQName:      "chevy/camaro-apb",
				Image:       "chevy/cavalier-apb",
				Tags:        []string{"cars", "chevy"},
				Bindable:    true,
				Description: "description",
				Async:       convertToAsyncType("optional"),
				Metadata:    `{"_apb_creds":"letmein","foo":"bar"}`,
				Alpha:       `{"alpha_apb_creds":"letmein","alphafoo":"bar"}`,
				Plans: []v1alpha1.Plan{
					{
						Name:     "dev",
						Bindable: true,
						Metadata: `{"plan_param1":"letmein","plan_param2":"bar"}`,
					},
				},
			},

			expected: &bundle.Spec{
				ID:          uid,
				Runtime:     2,
				Version:     "1.2.3",
				FQName:      "chevy/camaro-apb",
				Image:       "chevy/cavalier-apb",
				Tags:        []string{"cars", "chevy"},
				Bindable:    true,
				Description: "description",
				Async:       "optional",
				Metadata: map[string]interface{}{
					"_apb_creds": "letmein",
					"foo":        "bar",
				},
				Alpha: map[string]interface{}{
					"alpha_apb_creds": "letmein",
					"alphafoo":        "bar",
				},
				Plans: []bundle.Plan{
					{
						Name:     "dev",
						Bindable: true,
						Metadata: map[string]interface{}{
							"plan_param1": "letmein",
							"plan_param2": "bar",
						},
						Parameters:     []bundle.ParameterDescriptor{},
						BindParameters: []bundle.ParameterDescriptor{},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			output, err := ConvertBundleToSpec(tc.input, tc.expected.ID)
			if tc.expectederr {
				assert.Error(t, err)
			}

			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertServiceInstanceToCRD(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       *bundle.ServiceInstance
		expected    v1alpha1.BundleInstance
		expectederr bool
		panics      bool
	}{
		{
			name:     "nil spec should cause error",
			input:    &bundle.ServiceInstance{},
			expected: v1alpha1.BundleInstance{},
			panics:   true,
		},
		{
			name:     "nil ServiceInstance should cause error",
			input:    nil,
			expected: v1alpha1.BundleInstance{},
			panics:   true,
		},
		{
			name: "BindInstance zero value",
			input: &bundle.ServiceInstance{
				Spec:    &bundle.Spec{},
				Context: &bundle.Context{},
			},
			expected: v1alpha1.BundleInstance{
				Status: v1alpha1.BundleInstanceStatus{
					Bindings: []v1alpha1.LocalObjectReference{},
				},
			},
		},
		{
			name: "invalid parameters should return error",
			input: &bundle.ServiceInstance{
				Parameters: &bundle.Parameters{
					// force json marshal to fail
					"foo": make(chan int),
				},
			},
			expected:    v1alpha1.BundleInstance{},
			expectederr: true,
		},
		{
			name: "parameters should get copied",
			input: &bundle.ServiceInstance{
				ID: uuid.Parse(uid),
				Spec: &bundle.Spec{
					ID: uid,
				},
				Context: &bundle.Context{
					Namespace: "testnamespace",
					Platform:  "kubernetes",
				},
				Parameters: &bundle.Parameters{
					"foo":        "bar",
					"_apb_creds": "letmein",
				},
				BindingIDs: map[string]bool{
					"a binding": true,
				},
				DashboardURL: "http://example.com/dashboard",
			},
			expected: v1alpha1.BundleInstance{
				Spec: v1alpha1.BundleInstanceSpec{
					Bundle: v1alpha1.LocalObjectReference{Name: uid},
					Context: v1alpha1.Context{
						Namespace: "testnamespace",
						Platform:  "kubernetes",
					},
					Parameters:   `{"_apb_creds":"letmein","foo":"bar"}`,
					DashboardURL: "http://example.com/dashboard",
				},
				Status: v1alpha1.BundleInstanceStatus{
					Bindings: []v1alpha1.LocalObjectReference{
						{
							Name: "a binding",
						},
					},
				},
			},
			expectederr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.panics {
				assert.Panics(t, func() { ConvertServiceInstanceToCRD(tc.input) })
				return
			}

			output, err := ConvertServiceInstanceToCRD(tc.input)
			if tc.expectederr {
				assert.Error(t, err)
			}

			assert.Equal(t, tc.expected, output)
		})
	}
}
