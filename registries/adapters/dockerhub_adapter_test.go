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

package adapters

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/stretchr/testify/assert"
)

func TestRegistryName(t *testing.T) {
	dha := DockerHubAdapter{}
	assert.Equal(t, dha.RegistryName(), "docker.io", "dockerhub name does not match docker.io")
}

func TestGetImageNames(t *testing.T) {
	testCases := []struct {
		name        string
		c           Configuration
		expected    []string
		expectederr bool
		handlerFunc http.HandlerFunc
	}{
		{
			name:        "unable to generate token should return an error",
			c:           Configuration{},
			expected:    nil,
			expectederr: true,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, r.URL.Path, "/v2/users/login/")
				w.Write([]byte("invalid response, fail token"))
			},
		},
		{
			name: "error in getNextImages should return an error",
			c: Configuration{
				Org: "testorg",
			},
			expected:    nil,
			expectederr: true,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost &&
					r.URL.Path == "/v2/users/login/" {
					// return a testtoken for login
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"token":"testtoken"}`))
				} else {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, r.URL.Path, "/v2/repositories/testorg/")
					w.Write([]byte("get images, invalid response"))
				}
			},
		},
		{
			name: "returning 0 images should return nil",
			c: Configuration{
				Org: "testorg",
			},
			expected:    nil,
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost &&
					r.URL.Path == "/v2/users/login/" {
					// return a testtoken for login
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"token":"testtoken"}`))
				} else {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, r.URL.Path, "/v2/repositories/testorg/")
					w.Write([]byte(`{"count": 0, "next":"", "results":[] }`))
				}
			},
		},
		{
			name: "returning 0 images should return nil",
			c: Configuration{
				Org: "testorg",
			},
			expected:    []string{"target/test-image-1"},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost &&
					r.URL.Path == "/v2/users/login/" {
					// return a testtoken for login
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"token":"testtoken"}`))
				} else {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, r.URL.Path, "/v2/repositories/testorg/")
					w.Write([]byte(`{"count": 1, "next":"", "results":[{"name":"test-image-1", "namespace":"target"}] }`))
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// get test server
			serv := GetServer(t, tc.handlerFunc)
			defer serv.Close()

			// use the test server's url
			dockerHubLoginURL = join(serv.URL, "/v2/users/login/")
			dockerHubRepoImages = join(serv.URL, "/v2/repositories/%v/?page_size=100")
			dockerHubManifestURL = join(serv.URL, "/v2/%v/manifests/%v")

			// create the adapter we  want to test
			dha := DockerHubAdapter{Config: tc.c}

			// test the GetImageNames method
			output, err := dha.GetImageNames()

			if tc.expectederr {
				if !assert.Error(t, err) {
					t.Fatal(err)
				}
				assert.NotEmpty(t, err.Error())
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			errmsg := fmt.Sprintf("%s returned the wrong value", tc.name)
			assert.Equal(t, tc.expected, output, errmsg)
		})
	}
}

func TestFetchSpecs(t *testing.T) {
	testCases := []struct {
		name        string
		c           Configuration
		input       []string
		expected    []*bundle.Spec
		expectederr bool
		handlerFunc http.HandlerFunc
	}{
		{
			name:        "no images returns no error",
			c:           Configuration{},
			input:       []string{},
			expected:    []*bundle.Spec{},
			expectederr: false,
			handlerFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// get test server
			// serv := GetServer(t, tc.handlerFunc)
			// defer serv.Close()

			// use the test server's url
			// dockerHubLoginURL = join(serv.URL, "/v2/users/login/")
			// dockerHubRepoImages = join(serv.URL, "/v2/repositories/%v/?page_size=100")
			// dockerHubManifestURL = join(serv.URL, "/v2/%v/manifests/%v")

			// create the adapter we  want to test
			dha := DockerHubAdapter{Config: tc.c}

			// test the GetImageNames method
			output, err := dha.FetchSpecs(tc.input)

			if tc.expectederr {
				if !assert.Error(t, err) {
					t.Fatal(err)
				}
				assert.NotEmpty(t, err.Error())
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			errmsg := fmt.Sprintf("%s returned the wrong value", tc.name)
			assert.Equal(t, tc.expected, output, errmsg)
		})
	}
}

// GetServer returns a test http server which will run whatever HandlerFunc we
// pass in.
func GetServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// stupid utility to avoid creating []string in the main test function
func join(prefix string, suffix string) string {
	var tojoin []string
	tojoin = append(tojoin, prefix)
	tojoin = append(tojoin, suffix)
	return strings.Join(tojoin, "")
}
