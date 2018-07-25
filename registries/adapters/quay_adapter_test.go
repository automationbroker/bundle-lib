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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/stretchr/testify/assert"
)

var quayTestConfig = Configuration{
	Org: "foo",
}

const (
	quayTestCatalogResponse = `
	{
	  "repositories": [
	    {
	      "is_public": true,
	      "kind": "image",
	      "namespace": "foo",
	      "name": "bar",
	      "description": ""
	    },
	    {
	      "is_public": true,
	      "kind": "image",
	      "namespace": "foo",
	      "name": "test-apb",
	      "description": null
	    },
	    {
	      "is_public": true,
	      "kind": "image",
	      "namespace": "foo",
	      "name": "baz",
	      "description": null
	    },
	    {
	      "is_public": true,
	      "kind": "image",
	      "namespace": "foo",
	      "name": "another-apb",
	      "description": null
	    }
	  ]
	}`

	quayTestDigestResponse = `
	{
	  "trust_enabled": false,
	  "description": null,
	  "tags": {
	    "latest": {
	      "image_id": "86753098043f091a1518fd58c82e053e4184ad2dda8012673e70ab680d64f086",
	      "last_modified": "Thu, 28 Jun 2018 19:56:58 -0000",
	      "name": "latest",
	      "manifest_digest": "sha256:482e3f2c582f6facac995fff1ab70612ea41bc67788bae9e51ed21448c0fc7a2",
	      "size": 169280047
	    }
	  },
	  "tag_expiration_s": 1209600,
	  "is_public": true,
	  "is_starred": false,
	  "kind": "image",
	  "name": "test-apb",
	  "namespace": "foo",
	  "is_organization": false,
	  "can_write": false,
	  "status_token": "",
	  "can_admin": false
	}`

	quayTestManifestResponse = `
	{
		"labels": [
		  {
		    "value": "= 1.0     org.label-schema.name=CentOS Base Image     org.label-schema.vendor=CentOS     org.label-schema.license=GPLv2     org.label-schema.build-date=20180402",
		    "media_type": "text/plain",
		    "id": "a1d8a6f7-ecf8-4438-9f59-b73bc97d28b9",
		    "key": "org.label-schema.schema-version",
		    "source_type": "manifest"
		  },
		  {
		    "value": "2",
		    "media_type": "text/plain",
		    "id": "5685b7b6-840d-417b-b9d5-936a8008294d",
		    "key": "com.redhat.apb.runtime",
		    "source_type": "manifest"
		  },
		  {
		    "value": "dmVyc2lvbjogMS4wDQpuYW1lOiB0ZXN0LWFwYg0KZGVzY3JpcHRpb246IHRlc3QgYXBiIGltcGxlbWVudGF0aW9uDQpiaW5kYWJsZTogRmFsc2UNCmFzeW5jOiBvcHRpb25hbA0KbWV0YWRhdGE6DQogIGRvY3VtZW50YXRpb25Vcmw6IGh0dHBzOi8vd3d3LnRlc3Qub3JnL3dpa2kvRG9jcw0KICBsb25nRGVzY3JpcHRpb246IEFuIGFwYiB0aGF0IHRlc3RzIHlvdXIgdGVzdA0KICBkZXBlbmRlbmNpZXM6IFsncXVheS5pby90ZXN0L3Rlc3Q6bGF0ZXN0J10NCiAgZGlzcGxheU5hbWU6IFRlc3QgKEFQQikNCiAgcHJvdmlkZXJEaXNwbGF5TmFtZTogIlRlc3QgSW5jLiINCnBsYW5zOg0KICAtIG5hbWU6IGRlZmF1bHQNCiAgICBkZXNjcmlwdGlvbjogQW4gQVBCIHRoYXQgdGVzdHMNCiAgICBmcmVlOiBUcnVlDQogICAgbWV0YWRhdGE6DQogICAgICBkaXNwbGF5TmFtZTogRGVmYXVsdA0KICAgICAgbG9uZ0Rlc2NyaXB0aW9uOiBUaGlzIHBsYW4gZGVwbG95cyBhIHNpbmdsZSB0ZXN0DQogICAgICBjb3N0OiAkMC4wMA0KICAgIHBhcmFtZXRlcnM6DQogICAgICAtIG5hbWU6IHRlc3RfcGFyYW0NCiAgICAgICAgZGVmYXVsdDogdGVzdA0KICAgICAgICB0eXBlOiBzdHJpbmcNCiAgICAgICAgdGl0bGU6IFRlc3QgUGFyYW1ldGVyDQogICAgICAgIHBhdHRlcm46ICJeW2EtekEtWl9dW2EtekEtWjAtOV9dKiQiDQogICAgICAgIHJlcXVpcmVkOiBUcnVlDQo=",
		    "media_type": "text/plain",
		    "id": "ed22acaf-c68d-4361-ae3e-fda1b0105888",
		    "key": "com.redhat.apb.spec",
		    "source_type": "manifest"
		  }
		]
	}`
)

func TestQuayAdaptorName(t *testing.T) {
	a := QuayAdapter{}
	assert.Equal(t, "quay.io", a.RegistryName(), "registry adaptor name does not match")
}

func TestNewQuayAdapter(t *testing.T) {
	a := NewQuayAdapter(quayTestConfig)

	b := QuayAdapter{}
	b.config.Org = "foo"
	b.config.Tag = "latest"

	assert.Equal(t, b, a, "adaptor returned is not valid")
}

func TestQuayGetImageNames(t *testing.T) {
	testCases := []struct {
		name        string
		c           Configuration
		expected    []string
		expectederr bool
		handlerFunc http.HandlerFunc
	}{
		{
			name: "should return 4 images",
			c:    Configuration{Org: "foo"},
			expected: []string{
				"bar",
				"test-apb",
				"baz",
				"another-apb",
			},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected `GET` request, got `%s`", r.Method)
				}
				if strings.Contains(r.URL.String(), "namespace") {
					fmt.Fprintf(w, quayTestCatalogResponse)
				}
			},
		},
		{
			name: "config images should also be returned with repo images",
			c: Configuration{
				Org:    "foo",
				Images: []string{"additional"},
			},
			expected: []string{
				"additional",
				"bar",
				"test-apb",
				"baz",
				"another-apb",
			},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected `GET` request, got `%s`", r.Method)
				}
				if strings.Contains(r.URL.String(), "namespace") {
					fmt.Fprintf(w, quayTestCatalogResponse)
				}
			},
		},
		{
			name:        "invalid catalog response should return error",
			c:           Configuration{Org: "foo"},
			expected:    []string{},
			expectederr: true,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected `GET` request, got `%s`", r.Method)
				}
				if strings.Contains(r.URL.String(), "namespace") {
					fmt.Fprintf(w, "invalid response, should fail")
				}
			},
		},
		{
			name:        "empty list should return no error",
			c:           Configuration{Org: "foo"},
			expected:    []string{},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected `GET` request, got `%s`", r.Method)
				}
				if strings.Contains(r.URL.String(), "namespace") {
					fmt.Fprintf(w, `{"repositories": [] }`)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serv := httptest.NewServer(tc.handlerFunc)
			defer serv.Close()

			tc.c.URL = getQuayURL(t, serv)

			// create the adapter we want to test
			qa := NewQuayAdapter(tc.c)

			// test the GetImageNames method
			output, err := qa.GetImageNames()
			if tc.expectederr {
				if !assert.Error(t, err) {
					t.Fatal(err)
				}
				assert.NotEmpty(t, err.Error())
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			errmsg := fmt.Sprintf("%s returned the wrong value", tc.name)
			assert.ElementsMatch(t, tc.expected, output, errmsg)
		})
	}
}

func TestQuayFetchSpecs(t *testing.T) {
	testCases := []struct {
		name        string
		c           Configuration
		input       []string
		expected    []*bundle.Spec
		expectederr bool
		handlerFunc http.HandlerFunc
	}{
		{
			name:  "expected one spec",
			c:     Configuration{Org: "foo"},
			input: []string{"test-apb"},
			expected: []*bundle.Spec{
				{
					Runtime: 2,
					Version: "1.0",
					FQName:  "test-apb",
					Metadata: map[string]interface{}{
						"dependencies":        []interface{}{"quay.io/test/test:latest"},
						"displayName":         "Test (APB)",
						"documentationUrl":    "https://www.test.org/wiki/Docs",
						"longDescription":     "An apb that tests your test",
						"providerDisplayName": "Test Inc.",
					},
					Image:       "%s/foo/test-apb:latest",
					Description: "test apb implementation",
					Async:       "optional",
					Plans: []bundle.Plan{
						{
							Name: "default",
							Metadata: map[string]interface{}{
								"cost":            "$0.00",
								"displayName":     "Default",
								"longDescription": "This plan deploys a single test",
							},
							Description: "An APB that tests",
							Free:        true,
							Parameters: []bundle.ParameterDescriptor{
								{
									Name:     "test_param",
									Title:    "Test Parameter",
									Type:     "string",
									Default:  "test",
									Pattern:  "^[a-zA-Z_][a-zA-Z0-9_]*$",
									Required: true,
								},
							},
						},
					},
				},
			},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "/manifest") {
					fmt.Fprintf(w, quayTestManifestResponse)
				}
				if !strings.Contains(r.URL.String(), "namespace") &&
					!strings.Contains(r.URL.String(), "/manifest") {
					fmt.Fprintf(w, quayTestDigestResponse)
				}
			},
		},
		{
			name:        "no images in, should return no specs",
			c:           Configuration{Org: "foo"},
			input:       []string{},
			expected:    []*bundle.Spec{},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("No request should be made")
			},
		},
		{
			name:        "invalid digest should return empty specs",
			c:           Configuration{Org: "foo"},
			input:       []string{"test-apb"},
			expected:    []*bundle.Spec{},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.String(), "namespace") &&
					!strings.Contains(r.URL.String(), "/manifest") {
					fmt.Fprintf(w, `{"invalid":"response"`)
				}
			},
		},
		{
			name:        "empty digest should return empty specs",
			c:           Configuration{Org: "foo"},
			input:       []string{"test-apb"},
			expected:    []*bundle.Spec{},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.String(), "namespace") &&
					!strings.Contains(r.URL.String(), "/manifest") {
					fmt.Fprintf(w, "{}")
				}
			},
		},
		{
			name:        "invalid manifest response should log error, but pass",
			c:           Configuration{Org: "foo"},
			input:       []string{"test-apb"},
			expected:    []*bundle.Spec{},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "/manifest") {
					fmt.Fprintf(w, `{"invalid":"response"`)
				}
				if !strings.Contains(r.URL.String(), "namespace") &&
					!strings.Contains(r.URL.String(), "/manifest") {
					fmt.Fprintf(w, quayTestDigestResponse)
				}
			},
		},
		{
			name:        "incorrect encoded spec should simply return empty spec",
			c:           Configuration{Org: "foo"},
			input:       []string{"test-apb"},
			expected:    []*bundle.Spec{},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "/manifest") {
					resp := map[string][]map[string]string{
						"labels": {
							{
								// pass in invalid base64 encoding
								"value":       "aW52YWxpZCByZXNwb25z==",
								"media_type":  "text/plain",
								"id":          "ed22acaf-c68d-4361-ae3e-fda1b0105888",
								"key":         "com.redhat.apb.spec",
								"source_type": "manifest",
							},
						},
					}
					respdata, _ := json.Marshal(resp)
					fmt.Fprintf(w, string(respdata))
				}
				if !strings.Contains(r.URL.String(), "namespace") &&
					!strings.Contains(r.URL.String(), "/manifest") {
					fmt.Fprintf(w, quayTestDigestResponse)
				}
			},
		},
		{
			name:        "empty encoded spec should simply return empty spec",
			c:           Configuration{Org: "foo"},
			input:       []string{"test-apb"},
			expected:    []*bundle.Spec{},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "/manifest") {
					resp := map[string][]map[string]string{
						"labels": {
							{
								"value":       "",
								"media_type":  "text/plain",
								"id":          "ed22acaf-c68d-4361-ae3e-fda1b0105888",
								"key":         "com.redhat.apb.spec",
								"source_type": "manifest",
							},
						},
					}
					respdata, _ := json.Marshal(resp)
					fmt.Fprintf(w, string(respdata))
				}
				if !strings.Contains(r.URL.String(), "namespace") &&
					!strings.Contains(r.URL.String(), "/manifest") {
					fmt.Fprintf(w, quayTestDigestResponse)
				}
			},
		},
		{
			name:        "invalid yaml properly encoded should return empty spec",
			c:           Configuration{Org: "foo"},
			input:       []string{"test-apb"},
			expected:    []*bundle.Spec{},
			expectederr: false,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "/manifest") {
					resp := map[string][]map[string]string{
						"labels": {
							{
								// encoded the string "invalid response"
								"value":       "aW52YWxpZCByZXNwb25zZQ==",
								"media_type":  "text/plain",
								"id":          "ed22acaf-c68d-4361-ae3e-fda1b0105888",
								"key":         "com.redhat.apb.spec",
								"source_type": "manifest",
							},
						},
					}
					respdata, _ := json.Marshal(resp)
					fmt.Fprintf(w, string(respdata))
				}
				if !strings.Contains(r.URL.String(), "namespace") &&
					!strings.Contains(r.URL.String(), "/manifest") {
					fmt.Fprintf(w, quayTestDigestResponse)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serv := httptest.NewServer(tc.handlerFunc)
			defer serv.Close()

			tc.c.URL = getQuayURL(t, serv)

			// Fix the expected URL
			for _, s := range tc.expected {
				s.Image = strings.Replace(fmt.Sprintf(s.Image, serv.URL), "http://", "", 1)
			}

			// create the adapter we want to test
			qa := NewQuayAdapter(tc.c)

			// test the FetchSpecs method
			output, err := qa.FetchSpecs(tc.input)
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

func getQuayServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected `GET` request, got `%s`", r.Method)
		}
		if strings.Contains(r.URL.String(), "namespace") {
			fmt.Fprintf(w, quayTestCatalogResponse)
		}
		if strings.Contains(r.URL.String(), "/manifest") {
			fmt.Fprintf(w, quayTestManifestResponse)
		}
		if !strings.Contains(r.URL.String(), "namespace") &&
			!strings.Contains(r.URL.String(), "/manifest") {
			fmt.Fprintf(w, quayTestDigestResponse)
		}
	}))
}

func getQuayURL(t *testing.T, s *httptest.Server) *url.URL {
	url, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal("Error: ", err)
	}
	return url
}
