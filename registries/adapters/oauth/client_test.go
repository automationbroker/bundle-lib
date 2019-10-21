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

package oauth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var apiV2AuthResponse = `
{
  "token": "%s",
  "expires_in": 300,
  "issued_at": "2018-03-27T19:54:19Z"
}`

var headerCases = map[string]string{
	"Bearer realm=\"http://foo/a/b/c\",service=\"bar\"":  "http://foo/a/b/c?service=bar",
	"Bearer service=\"bar\",realm=\"http://foo/a/b/c\"":  "http://foo/a/b/c?service=bar",
	"Bearer realm=\"http://foo/a/b/c/\",service=\"bar\"": "http://foo/a/b/c/?service=bar",
	"Bearer realm=\"https://foo\",service=\"bar\"":       "https://foo?service=bar",
	"Bearer realm=\"http://foo/a/b/c\"":                  "http://foo/a/b/c",
}

var headerErrorCases = map[string]string{
	"Bearer service=\"bar\"": "Could not parse www-authenticate header:",
	"Bearer realm=\"\"":      "",
}

var tokenCases = map[string]string{
	"{\"access_token\": \"abc123\"}":                        "abc123",
	"{\"token\": \"abc123\"}":                               "abc123",
	"{\"access_token\": \"abc123\", \"token\": \"def456\"}": "abc123",
	"{}": "",
}

var tokenErrorCases = map[string]string{
	"{\"token\": {}":          "unexpected end of JSON input",
	"{\"access_token\": {}":   "unexpected end of JSON input",
	"{\"token\": null":        "unexpected end of JSON input",
	"{\"access_token\": null": "unexpected end of JSON input",
}

func TestParseAuthHeader(t *testing.T) {
	for in, out := range headerCases {
		result, err := parseAuthHeader(in)
		if err != nil {
			t.Error(err.Error())
		}
		if result.String() != out {
			t.Errorf("Expected %s, got %s", out, result.String())
		}
	}
}

func TestParseAuthHeaderErrors(t *testing.T) {
	for in, out := range headerErrorCases {
		_, err := parseAuthHeader(in)
		if err == nil {
			t.Errorf("Expected an error parsing %s", in)
		} else if strings.HasPrefix(err.Error(), out) == false {
			t.Errorf("Expected prefix %s, got %s", out, err.Error())
		}
	}
}

func TestParseAuthToken(t *testing.T) {
	for in, out := range tokenCases {
		result, err := parseAuthToken([]byte(in))
		if err != nil {
			t.Error(err.Error())
		}
		if result != out {
			t.Errorf("Expected %s, got %s", out, result)
		}
	}
}

func TestParseAuthTokenErrors(t *testing.T) {
	for in, out := range tokenErrorCases {
		_, err := parseAuthToken([]byte(in))
		if err == nil {
			t.Errorf("Expected an error parsing %s", in)
		} else if strings.HasPrefix(err.Error(), out) == false {
			t.Errorf("Expected prefix %s, got %s", out, err.Error())
		}
	}
}

func TestGetTokenWithScope(t *testing.T) {
	authServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected 'GET' request, got '%s'", r.Method)
		}

		// see if we have any scopes
		u, _ := url.Parse(r.RequestURI)
		count := strings.Count(u.RawQuery, "scope")
		token := fmt.Sprintf("fake.tokenTbTRUN3VZWHEwRW9oMEM2cEd-%d-scopes", count)
		fmt.Fprintf(w, fmt.Sprintf(apiV2AuthResponse, token))
	}))
	defer authServ.Close()

	hdr := fmt.Sprintf("Bearer realm=\"%s/v2/auth\"", authServ.URL)
	u, err := url.Parse("http://automationbroker.io")
	if err != nil {
		t.Fatal("invalid url", err)
	}
	c := NewClient("", "", false, u)

	testCases := []struct {
		name          string
		imageNames    []string
		expectederr   bool
		expectedtoken string
	}{
		{
			name:          "no images",
			imageNames:    []string{},
			expectederr:   false,
			expectedtoken: "fake.tokenTbTRUN3VZWHEwRW9oMEM2cEd-0-scopes",
		},
		{
			name:          "2 images",
			imageNames:    []string{"rh-osbs/postgresql", "rh-osbs/mysql"},
			expectederr:   false,
			expectedtoken: "fake.tokenTbTRUN3VZWHEwRW9oMEM2cEd-2-scopes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := c.getTokenWithScope(hdr, tc.imageNames)
			if tc.expectederr {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error())
			} else if err != nil {
				fmt.Println(err.Error())
				t.Fatalf("unexpected error during test: %v\n", err)
			}
			assert.Equal(t, c.token, tc.expectedtoken)
		})
	}
}

func TestNewRequest(t *testing.T) {
	u, err := url.Parse("http://automationbroker.io")
	if err != nil {
		t.Fatal("invalid url", err)
	}
	c := NewClient("foo", "bar", false, u)

	testCases := []struct {
		name        string
		input       string
		token       string
		expectederr bool
	}{
		{
			name:  "relative path",
			input: "/v2/",
			token: "letmein",
		},
		{
			name:  "relative path without trailing slash",
			input: "/v2",
		},
		{
			name:  "relative path with multiple paths",
			input: "/v2/foobar/baz",
		},
		{
			name:  "relative path with hook",
			input: "v2/_catalog/?n=5&last=mediawiki-apb",
		},
		{
			name:  "fully qualified url with hook",
			input: "https://example.com/v2/_catalog/?n=5&last=mediawiki-apb",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// set the token before calling NewRequest
			c.token = tc.token

			output, err := c.NewRequest(tc.input)
			if tc.expectederr {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error())
			} else if err != nil {
				fmt.Println(err.Error())
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			assert.Equal(t, "application/json", output.Header.Get("Accept"))
			if tc.token != "" {
				assert.Equal(t, fmt.Sprintf("Bearer %s", tc.token), output.Header.Get("Authorization"))
			} else {
				assert.Equal(t, "", output.Header.Get("Authorization"))
			}

			expectedurl, err := url.Parse(tc.input)
			if err != nil {
				t.Fatalf("Invalid input url %s; %v\n", tc.input, err)
			}

			assert.Equal(t, c.url.Scheme, output.URL.Scheme)
			assert.Equal(t, c.url.Host, output.URL.Host)
			assert.Equal(t, expectedurl.Path, output.URL.Path)
			assert.Equal(t, expectedurl.Query(), output.URL.Query())
		})
	}
}

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name       string
		username   string
		password   string
		skipVerify bool
		url        func() *url.URL
	}{
		{
			name:       "fully qualified url",
			username:   "foo",
			password:   "bar",
			skipVerify: false,
			url: func() *url.URL {
				daurl, err := url.Parse("http://automationbroker.io")
				if err != nil {
					t.Fatal(err)
				}
				return daurl
			},
		},
		{
			name:       "nil url",
			username:   "foo",
			password:   "bar",
			skipVerify: false,
			url: func() *url.URL {
				return nil
			},
		},
		{
			name:       "empty username and password",
			username:   "",
			password:   "",
			skipVerify: false,
			url: func() *url.URL {
				daurl, err := url.Parse("http://automationbroker.io")
				if err != nil {
					t.Fatal(err)
				}
				return daurl
			},
		},
		{
			name:       "skip verify true",
			username:   "user",
			password:   "pass",
			skipVerify: true,
			url: func() *url.URL {
				daurl, err := url.Parse("http://automationbroker.io")
				if err != nil {
					t.Fatal(err)
				}
				return daurl
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fullURL := tc.url()

			output := NewClient(tc.username, tc.password, tc.skipVerify, fullURL)

			assert.NotNil(t, output)
			assert.Equal(t, tc.username, output.user)
			assert.Equal(t, tc.password, output.pass)
			assert.Equal(t, fullURL, output.url)
			assert.NotNil(t, output.client)
			// token should always be empty after NewClient is called
			assert.Equal(t, "", output.token)
		})
	}
}
