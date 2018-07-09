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
	"testing"
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
