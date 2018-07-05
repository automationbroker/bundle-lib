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

package bundle

import (
	"errors"
	"reflect"
	"testing"

	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
)

// mockExecuteApb mocks out the methods called by executeApb
func mockExecuteApb(rt *runtime.MockRuntime, e Executor, instanceID string) {

	rt.On("CopySecretsToNamespace",
		mock.Anything, mock.Anything, mock.Anything,
	).Return(nil)

	rt.On("GetRuntime").Return("kubernetes")
	rt.On("MasterName", instanceID).Return("new-master-name")
	rt.On("MasterNamespace").Return("new-masternamespace")
	rt.On("StateIsPresent", "new-master-name").Return(false, nil)
	rt.On("RunBundle", mock.Anything).Return(runtime.ExecutionContext{}, nil)
}

func mockCommonBind(rt *runtime.MockRuntime, e Executor) {
	rt.On("CopyState",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(nil)
	rt.On("WatchRunningBundle",
		mock.Anything, mock.Anything, mock.Anything,
	).Return(nil)
	rt.On("DestroySandbox",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything)
	rt.On("ExtractCredentials",
		mock.Anything, mock.Anything, mock.Anything,
	).Return([]byte(`{"test": "testingcreds"}`), nil)
}

func TestBind(t *testing.T) {
	// common variables for majority of the testcases
	bID := uuid.NewUUID()
	u := uuid.NewUUID()

	ctx := &Context{
		Namespace: "target",
		Platform:  "kubernetes",
	}

	spec := &Spec{
		ID:       "new-spec-id",
		Image:    "new-image",
		FQName:   "new-fq-name",
		Runtime:  2,
		Bindable: true,
	}

	// define test cases
	testCases := []*struct {
		name            string
		config          ExecutorConfig
		rt              runtime.MockRuntime
		si              ServiceInstance
		bindingID       string
		params          *Parameters
		extractedCreds  *ExtractedCredentials
		addExpectations func(rt *runtime.MockRuntime, e Executor)
		validateMessage func([]StatusMessage) bool
	}{
		{
			name:   "bind successfully",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID:         u,
				Spec:       spec,
				Context:    ctx,
				Parameters: &Parameters{"test-param": true},
			},
			bindingID: bID.String(),
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				mockExecuteApb(rt, e, u.String())
				mockCommonBind(rt, e)
				rt.On("CreateSandbox",
					mock.Anything, mock.Anything, []string{"target"},
					mock.Anything, mock.Anything,
				).Return("service-account-1", "location", nil)

				rt.On("CreateExtractedCredential", bID.String(), mock.Anything,
					map[string]interface{}{"test": "testingcreds"},
					map[string]string{
						"bundleAction": "bind",
						"bundleName":   "new-fq-name",
					},
				).Return(nil)
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateSucceeded {
					return false
				}
				return true
			},
			extractedCreds: &ExtractedCredentials{
				Credentials: map[string]interface{}{"test": "testingcreds"},
			},
		},
		{
			name: "bind successfully skip ns",
			config: ExecutorConfig{
				SkipCreateNS: true,
			},
			rt: *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID:         u,
				Spec:       spec,
				Context:    ctx,
				Parameters: &Parameters{"test-param": true},
			},
			bindingID: bID.String(),
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				mockExecuteApb(rt, e, u.String())
				mockCommonBind(rt, e)

				rt.On("CreateSandbox",
					mock.Anything, "target", []string{"target"},
					mock.Anything, mock.Anything,
				).Return("service-account-1", "location", nil)

				rt.On("CreateExtractedCredential", bID.String(), mock.Anything,
					map[string]interface{}{"test": "testingcreds"},
					map[string]string{
						"bundleAction": "bind",
						"bundleName":   "new-fq-name",
					},
				).Return(nil)
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateSucceeded {
					return false
				}
				return true
			},
			extractedCreds: &ExtractedCredentials{
				Credentials: map[string]interface{}{"test": "testingcreds"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtime.Provider = &tc.rt
			e := NewExecutor(tc.config)
			if tc.addExpectations != nil {
				tc.addExpectations(&tc.rt, e)
			}
			s := e.Bind(&tc.si, tc.si.Parameters, bID.String())
			m := []StatusMessage{}
			for msg := range s {
				m = append(m, msg)
			}
			if !tc.validateMessage(m) {
				t.Fatalf("invalid messages - %#v", m)
			}

			// verify credentials
			if tc.extractedCreds != nil {
				if !reflect.DeepEqual(e.ExtractedCredentials(), tc.extractedCreds) {
					t.Fatalf("Invalid extracted credentials\nexpected: %#+v\n\nactual: %#+v",
						tc.extractedCreds, e.ExtractedCredentials())
				}
			}

		})
	}
}

func TestBindFailure(t *testing.T) {
	// common variables for majority of the testcases
	bID := uuid.NewUUID()
	u := uuid.NewUUID()

	ctx := &Context{
		Namespace: "target",
		Platform:  "kubernetes",
	}

	spec := &Spec{
		ID:       "new-spec-id",
		Image:    "new-image",
		FQName:   "new-fq-name",
		Runtime:  2,
		Bindable: true,
	}

	// define test cases
	testCases := []*struct {
		name            string
		config          ExecutorConfig
		rt              runtime.MockRuntime
		si              ServiceInstance
		bindingID       string
		params          *Parameters
		extractedCreds  *ExtractedCredentials
		addExpectations func(rt *runtime.MockRuntime, e Executor)
		validateMessage func([]StatusMessage) bool
	}{
		{
			name:   "bind failed to copystate",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID:         u,
				Spec:       spec,
				Context:    ctx,
				Parameters: &Parameters{"test-param": true},
			},
			bindingID: bID.String(),
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				mockExecuteApb(rt, e, u.String())

				// include only what's required for failing CopyState
				rt.On("CreateSandbox",
					mock.Anything, mock.Anything, []string{"target"},
					mock.Anything, mock.Anything,
				).Return("service-account-1", "location", nil)

				rt.On("WatchRunningBundle",
					mock.Anything, mock.Anything, mock.Anything,
				).Return(nil)

				rt.On("DestroySandbox",
					mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything)

				rt.On("CopyState",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(errors.New("copy state failed"))
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				return true
			},
			extractedCreds: nil,
		},
		{
			name:   "bind failed to createsandbox",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID:         u,
				Spec:       spec,
				Context:    ctx,
				Parameters: &Parameters{"test-param": true},
			},
			bindingID: bID.String(),
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				rt.On("CreateSandbox",
					mock.Anything, mock.Anything, []string{"target"},
					mock.Anything, mock.Anything,
				).Return("", "", errors.New("create sandbox failed"))
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				return true
			},
			extractedCreds: nil,
		},
		{
			name:   "bind failed to extract credentials",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID:         u,
				Spec:       spec,
				Context:    ctx,
				Parameters: &Parameters{"test-param": true},
			},
			bindingID: bID.String(),
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				mockExecuteApb(rt, e, u.String())

				// include only what's required for failing CopyState
				rt.On("CreateSandbox",
					mock.Anything, mock.Anything, []string{"target"},
					mock.Anything, mock.Anything,
				).Return("service-account-1", "location", nil)
				rt.On("CopyState",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil)
				rt.On("WatchRunningBundle",
					mock.Anything, mock.Anything, mock.Anything,
				).Return(nil)
				rt.On("DestroySandbox",
					mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything)

				b := make([]byte, 1)
				rt.On("ExtractCredentials",
					mock.Anything, mock.Anything, mock.Anything,
				).Return(b, errors.New("extract creds failed"))
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				return true
			},
			extractedCreds: nil,
		},
		{
			name:   "bind fails to create extracted credentials",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID:         u,
				Spec:       spec,
				Context:    ctx,
				Parameters: &Parameters{"test-param": true},
			},
			bindingID: bID.String(),
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				mockExecuteApb(rt, e, u.String())
				mockCommonBind(rt, e)
				rt.On("CreateSandbox",
					mock.Anything, mock.Anything, []string{"target"},
					mock.Anything, mock.Anything,
				).Return("service-account-1", "location", nil)

				rt.On("CreateExtractedCredential", bID.String(), mock.Anything,
					mock.Anything, mock.Anything,
				).Return(errors.New("failed to create credentials"))
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				return true
			},
			extractedCreds: nil,
		},
		{
			name:   "watch pod fails",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID:         u,
				Spec:       spec,
				Context:    ctx,
				Parameters: &Parameters{"test-param": true},
			},
			bindingID: bID.String(),
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				mockExecuteApb(rt, e, u.String())

				// include only what's required for failing CopyState
				rt.On("CreateSandbox",
					mock.Anything, mock.Anything, []string{"target"},
					mock.Anything, mock.Anything,
				).Return("service-account-1", "location", nil)

				rt.On("WatchRunningBundle",
					mock.Anything, mock.Anything, mock.Anything,
				).Return(errors.New("watch pod failed"))

				rt.On("DestroySandbox",
					mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything)
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				return true
			},
			extractedCreds: nil,
		},
		{
			name:   "executeApb fails",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID:         u,
				Spec:       spec,
				Context:    ctx,
				Parameters: &Parameters{"test-param": true},
			},
			bindingID: bID.String(),
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				// this will cause executeApb to fail
				rt.On("GetRuntime").Return("kubernetes")
				rt.On("CopySecretsToNamespace",
					mock.Anything, mock.Anything, mock.Anything,
				).Return(errors.New("executeApb failed"))

				rt.On("CreateSandbox",
					mock.Anything, mock.Anything, []string{"target"},
					mock.Anything, mock.Anything,
				).Return("service-account-1", "location", nil)

				rt.On("DestroySandbox",
					mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything)
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				return true
			},
			extractedCreds: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtime.Provider = &tc.rt
			e := NewExecutor(tc.config)
			if tc.addExpectations != nil {
				tc.addExpectations(&tc.rt, e)
			}
			s := e.Bind(&tc.si, tc.si.Parameters, bID.String())
			m := []StatusMessage{}
			for msg := range s {
				m = append(m, msg)
			}
			if !tc.validateMessage(m) {
				t.Fatalf("invalid messages - %#v", m)
			}

			// verify we get an error
			if e.LastStatus().Error == nil {
				t.Fatal("we expected the executor to have an error")
			}

			// verify credentials
			if tc.extractedCreds != nil {
				if !reflect.DeepEqual(e.ExtractedCredentials(), tc.extractedCreds) {
					t.Fatalf("Invalid extracted credentials\nexpected: %#+v\n\nactual: %#+v",
						tc.extractedCreds, e.ExtractedCredentials())
				}
			}

		})
	}
}
