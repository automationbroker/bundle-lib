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
	"os"
	"testing"

	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/stretchr/testify/assert"
)

func TestExecutor(t *testing.T) {
	ec := &ExtractedCredentials{
		Credentials: map[string]interface{}{"test": "testingcreds"},
	}

	sm := StatusMessage{State: StateNotYetStarted}

	// define test cases
	testCases := []*struct {
		name     string
		e        executor
		validate func(*executor) bool
	}{
		{
			name: "default executor",
			e:    executor{},
			validate: func(exec *executor) bool {
				// these next 2 calls will fail we try to access the
				// channel without checking it for nil first
				exec.actionFinishedWithError(errors.New("ignore"))
				exec.actionFinishedWithSuccess()

				if exec.extractedCredentials != nil ||
					exec.dashboardURL != "" ||
					exec.podName != "" ||
					exec.skipCreateNS {
					return false
				}

				return true
			},
		},
		{
			name: "PodName",
			e:    executor{podName: "podname"},
			validate: func(exec *executor) bool {
				return exec.PodName() == "podname"
			},
		},
		{
			name: "DashboardURL",
			e:    executor{dashboardURL: "http://some.url.com"},
			validate: func(exec *executor) bool {
				return exec.DashboardURL() == "http://some.url.com"
			},
		},
		{
			name: "ExtractedCredentials",
			e:    executor{extractedCredentials: ec},
			validate: func(exec *executor) bool {
				return exec.ExtractedCredentials() == ec
			},
		},
		{
			name: "LastStatus",
			e:    executor{lastStatus: sm},
			validate: func(exec *executor) bool {
				return exec.LastStatus() == sm
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.validate(&tc.e) {
				t.Fatalf("executor creation failed")
			}
		})
	}
}

func TestGetProxyConfig(t *testing.T) {
	testCases := []*struct {
		name     string
		setupEnv func()
		expected *runtime.ProxyConfig
	}{
		{
			name:     "no proxy vars",
			expected: nil,
			setupEnv: func() {
				// make sure variables aren't set
				os.Unsetenv("HTTP_PROXY")
				os.Unsetenv("HTTPS_PROXY")
				os.Unsetenv("NO_PROXY")
			},
		},
		{
			name:     "no proxy set, but no proxy configured",
			expected: nil,
			setupEnv: func() {
				// make sure variables aren't set
				os.Unsetenv("HTTP_PROXY")
				os.Unsetenv("HTTPS_PROXY")
				// ensure the NO_PROXY is set though
				os.Setenv("NO_PROXY", "*.aventail.com,home.com,.seanet.com")
			},
		},
		{
			name: "all configs are set",
			expected: &runtime.ProxyConfig{
				HTTPProxy:  "http://user:password@prox-server:3128",
				HTTPSProxy: "https://user:password@secure-prox-server:3128",
				NoProxy:    "*.aventail.com,home.com,.seanet.com",
			},
			setupEnv: func() {
				os.Setenv("HTTP_PROXY", "http://user:password@prox-server:3128")
				os.Setenv("HTTPS_PROXY", "https://user:password@secure-prox-server:3128")
				os.Setenv("NO_PROXY", "*.aventail.com,home.com,.seanet.com")
			},
		},
		{
			name: "only http is set",
			expected: &runtime.ProxyConfig{
				HTTPProxy:  "http://user:password@prox-server:3128",
				HTTPSProxy: "",
				NoProxy:    "",
			},
			setupEnv: func() {
				os.Setenv("HTTP_PROXY", "http://user:password@prox-server:3128")
				os.Unsetenv("HTTPS_PROXY")
				os.Unsetenv("NO_PROXY")
			},
		},
		{
			name: "only https is set",
			expected: &runtime.ProxyConfig{
				HTTPProxy:  "",
				HTTPSProxy: "https://user:password@secure-prox-server:3128",
				NoProxy:    "",
			},
			setupEnv: func() {
				os.Unsetenv("HTTP_PROXY")
				os.Setenv("HTTPS_PROXY", "https://user:password@secure-prox-server:3128")
				os.Unsetenv("NO_PROXY")
			},
		},
		{
			name:     "proxy vars set but empty",
			expected: nil,
			setupEnv: func() {
				os.Setenv("HTTP_PROXY", "")
				os.Setenv("HTTPS_PROXY", "")
				os.Unsetenv("NO_PROXY")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupEnv()
			assert.Equal(t, tc.expected, getProxyConfig())
		})
	}
}
