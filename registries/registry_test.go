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

package registries

import (
	"fmt"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/registries/adapters"
	"github.com/automationbroker/bundle-lib/registries/adapters/adaptertest"
	"github.com/stretchr/testify/assert"
)

var SpecTags = []string{"latest", "old-release"}

const SpecID = "ab094014-b740-495e-b178-946d5aa97ebf"
const SpecBadVersion = "2.0.0"
const SpecVersion = "1.0.0"
const SpecRuntime = 1
const SpecBadRuntime = 0
const SpecName = "etherpad-bundle"
const SpecImage = "fusor/etherpad-bundle"
const SpecBindable = false
const SpecAsync = "optional"
const SpecDescription = "A note taking webapp"
const SpecRegistryName = "test"

const PlanName = "dev"
const PlanDescription = "Basic development plan"

var PlanMetadata = map[string]interface{}{
	"displayName":     "Development",
	"longDescription": PlanDescription,
	"cost":            "$0.00",
}

const PlanFree = true
const PlanBindable = true

var expectedPlanParameters = []bundle.ParameterDescriptor{
	{
		Name:    "postgresql_database",
		Default: "admin",
		Type:    "string",
		Title:   "PostgreSQL Database Name",
	},
	{
		Name:        "postgresql_password",
		Default:     "admin",
		Type:        "string",
		Description: "A random alphanumeric string if left blank",
		Title:       "PostgreSQL Password",
	},
	{
		Name:                "postgresql_user",
		Default:             "admin",
		Title:               "PostgreSQL User",
		Type:                "string",
		DeprecatedMaxlength: 63,
	},
	{
		Name:    "postgresql_version",
		Default: 9.5,
		Enum:    []string{"9.5", "9.4"},
		Type:    "enum",
		Title:   "PostgreSQL Version",
	},
	{
		Name:        "postgresql_email",
		Pattern:     "\u201c^\\\\S+@\\\\S+$\u201d",
		Type:        "string",
		Description: "email address",
		Title:       "email",
	},
}

var p = bundle.Plan{
	Name:        PlanName,
	Description: PlanDescription,
	Metadata:    PlanMetadata,
	Free:        PlanFree,
	Bindable:    PlanBindable,
	Parameters:  expectedPlanParameters,
}

var s = bundle.Spec{
	Version:     SpecVersion,
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []bundle.Plan{p},
}

var dupePlansSpec = bundle.Spec{
	Version:     SpecVersion,
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []bundle.Plan{p, p},
}

var noPlansSpec = bundle.Spec{
	Version:     SpecVersion,
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
}

var noVersionSpec = bundle.Spec{
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []bundle.Plan{p},
}

var badVersionSpec = bundle.Spec{
	Version:     SpecBadVersion,
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []bundle.Plan{p},
}

var badRuntimeSpec = bundle.Spec{
	Version:     SpecVersion,
	Runtime:     SpecBadRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []bundle.Plan{p},
}

type errorAdapter struct {
	errGetImageNames bool
	errFetchSpecs    bool
}

func (e errorAdapter) GetImageNames() ([]string, error) {
	if e.errGetImageNames {
		return []string{}, fmt.Errorf("always return an error")
	}
	return []string{}, nil
}

func (e errorAdapter) FetchSpecs(names []string) ([]*bundle.Spec, error) {
	if e.errFetchSpecs {
		return []*bundle.Spec{}, fmt.Errorf("always return an error")
	}
	return []*bundle.Spec{}, nil
}

func (e errorAdapter) RegistryName() string {
	return ""
}

type TestingAdapter struct {
	Name   string
	Images []string
	Specs  []*bundle.Spec
	Called map[string]bool
}

func (t TestingAdapter) GetImageNames() ([]string, error) {
	t.Called["GetImageNames"] = true
	return t.Images, nil
}

func (t TestingAdapter) FetchSpecs(images []string) ([]*bundle.Spec, error) {
	t.Called["FetchSpecs"] = true
	return t.Specs, nil
}

func (t TestingAdapter) RegistryName() string {
	t.Called["RegistryName"] = true
	return t.Name
}

var a *TestingAdapter
var r Registry

func setUp() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&s},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpDupePlans() Registry {
	a = &TestingAdapter{
		Name:   "dupeadapter",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&dupePlansSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpNoPlans() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&noPlansSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpNoVersion() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&noVersionSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpBadVersion() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&badVersionSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpBadRuntime() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&badRuntimeSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpWithErrors(eg bool, ef bool) Registry {
	e := &errorAdapter{
		errGetImageNames: eg,
		errFetchSpecs:    ef,
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: e,
		filter:  filter}
	return r
}

func setUpValidNameFilter() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"fusor/etherpad-bundle", "image2"},
		Specs:  []*bundle.Spec{&s},
		Called: map[string]bool{},
	}
	filter := Filter{
		whitelist: []string{".*-bundle$"},
	}
	filter.Init()
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func TestRegistryLoadSpecs(t *testing.T) {
	testCases := []struct {
		name        string
		r           Registry
		validate    func([]*bundle.Spec, int, error) bool
		expectederr bool
	}{
		{
			name: "load specs no error",
			r:    setUp(),
			validate: func(specs []*bundle.Spec, images int, err error) bool {
				assert.Equal(t, images, 2)
				assert.Equal(t, len(specs), 1)
				assert.Equal(t, specs[0], &s)
				return true
			},
		},
		{
			name: "load specs with duplicate plans",
			r:    setUpDupePlans(),
			validate: func(specs []*bundle.Spec, images int, err error) bool {
				assert.Equal(t, images, 2)
				assert.Equal(t, len(specs), 0)
				return true
			},
			expectederr: false,
		},
		{
			name: "load specs no plans",
			r:    setUpNoPlans(),
			validate: func(specs []*bundle.Spec, images int, err error) bool {
				assert.Equal(t, len(specs), 0)
				return true
			},
		},
		{
			name: "load specs no version",
			r:    setUpNoVersion(),
			validate: func(specs []*bundle.Spec, images int, err error) bool {
				assert.Equal(t, len(specs), 0)
				return true
			},
		},
		{
			name: "load specs bad version",
			r:    setUpBadVersion(),
			validate: func(specs []*bundle.Spec, images int, err error) bool {
				assert.Equal(t, len(specs), 0)
				return true
			},
		},
		{
			name: "load specs bad runtime",
			r:    setUpBadRuntime(),
			validate: func(specs []*bundle.Spec, images int, err error) bool {
				assert.Equal(t, len(specs), 0)
				return true
			},
		},
		{
			name: "load specs getimagenames returns error",
			r:    setUpWithErrors(true, false),
			validate: func(specs []*bundle.Spec, images int, err error) bool {
				assert.Equal(t, len(specs), 0)
				return true
			},
			expectederr: true,
		},
		{
			name: "load specs fetchspecs returns error",
			r:    setUpWithErrors(false, true),
			validate: func(specs []*bundle.Spec, images int, err error) bool {
				assert.Equal(t, len(specs), 0)
				return true
			},
			expectederr: true,
		},
		{
			name: "load specs validnames",
			r:    setUpValidNameFilter(),
			validate: func(specs []*bundle.Spec, images int, err error) bool {
				return assert.Equal(t, len(specs), 1) &&
					assert.Equal(t, "fusor/etherpad-bundle", specs[0].Image)
			},
			expectederr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			specs, numImages, err := tc.r.LoadSpecs()

			if tc.expectederr {
				assert.Error(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			// assert.True(t, tc.r.adapter.Called["GetImageNames"])
			// assert.True(t, tc.r.adapter.Called["FetchSpecs"])
			assert.True(t, tc.validate(specs, numImages, err))
		})
	}
}

func TestFail(t *testing.T) {
	inputerr := fmt.Errorf("sample test err")

	testCases := []struct {
		name     string
		r        Registry
		expected bool
	}{
		{
			name: "fail should return true",
			r: Registry{
				config: Config{
					Fail: true,
				},
			},
			expected: true,
		},
		{
			name: "fail should return false",
			r: Registry{
				config: Config{
					Fail: false,
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.r.Fail(inputerr))
		})
	}
}

func TestNewRegistry(t *testing.T) {
	testCases := []struct {
		name        string
		c           Config
		validate    func(Registry) bool
		expectederr bool
	}{
		{
			name: "rhcc registry",
			c: Config{
				Type: "rhcc",
				Name: "rhcc",
			},
			validate: func(reg Registry) bool {
				_, ok := reg.adapter.(*adapters.RHCCAdapter)
				return ok
			},
		},
		{
			name: "dockerhub registry",
			c: Config{
				Type: "dockerhub",
				Name: "dh",
				URL:  "https://registry.hub.docker.com",
				User: "shurley",
				Org:  "shurley",
			},
			validate: func(reg Registry) bool {
				_, ok := reg.adapter.(*adapters.DockerHubAdapter)
				return ok
			},
		},
		{
			name: "mock registry",
			c: Config{
				Type: "mock",
				Name: "mock",
			},
			validate: func(reg Registry) bool {
				_, ok := reg.adapter.(*adapters.MockAdapter)
				return ok
			},
		},
		{
			name: "unknown registry",
			c: Config{
				Type: "makes_no_sense",
				Name: "dh",
			},
			validate: func(reg Registry) bool {
				return true
			},
			expectederr: true,
		},
		{
			name: "local_openshift should return a LocalOpenShiftAdapter",
			c: Config{
				Type: "local_openshift",
				Name: "localopenshift",
			},
			validate: func(reg Registry) bool {
				_, ok := reg.adapter.(*adapters.LocalOpenShiftAdapter)
				return ok
			},
		},
		{
			name: "helm should return a HelmAdapter",
			c: Config{
				Type: "helm",
				Name: "helm",
			},
			validate: func(reg Registry) bool {
				_, ok := reg.adapter.(*adapters.HelmAdapter)
				return ok
			},
		},
		{
			name: "openshift should return an OpenShiftAdapter",
			c: Config{
				Type:          "openshift",
				Name:          "openshift",
				SkipVerifyTLS: true,
				URL:           "NEEDSURL",
			},
			validate: func(reg Registry) bool {
				_, ok := reg.adapter.(adapters.OpenShiftAdapter)
				return ok
			},
		},
		{
			name: "partner_rhcc should return a PartnerRhccAdapter",
			c: Config{
				Type:          "partner_rhcc",
				Name:          "partnerrhcc",
				SkipVerifyTLS: true,
				URL:           "NEEDSURL",
			},
			validate: func(reg Registry) bool {
				_, ok := reg.adapter.(adapters.PartnerRhccAdapter)
				return ok
			},
		},
		{
			name: "apiv2 should return an APIV2Adapter",
			c: Config{
				Type:          "apiv2",
				Name:          "genericapiv2",
				SkipVerifyTLS: true,
				URL:           "NEEDSURL",
			},
			validate: func(reg Registry) bool {
				_, ok := reg.adapter.(adapters.APIV2Adapter)
				return ok
			},
		},
		{
			name: "underscores in names should fail",
			c: Config{
				Type: "helm",
				Name: "underscores_are_bad",
			},
			validate: func(reg Registry) bool {
				return true
			},
			expectederr: true,
		},
		{
			name: "apiv2 with no url should fail",
			c: Config{
				Type: "apiv2",
				Name: "nourl",
			},
			validate: func(reg Registry) bool {
				return true
			},
			expectederr: true,
		},
		{
			name: "retrieve registry auth should fail",
			c: Config{
				Type:     "dockerhub",
				Name:     "nourl",
				AuthType: "file",
				AuthName: "fakefile/tocause/error",
			},
			validate: func(reg Registry) bool {
				// should probably verify the registry, but the important part
				// is that we got an error
				return true
			},
			expectederr: true,
		},
		{
			name: "bad url should not fail",
			c: Config{
				Type: "dockerhub",
				Name: "nourl",
				URL:  "http://%41:8080/",
			},
			validate: func(reg Registry) bool {
				return true
			},
			expectederr: true,
		},
		{
			name: "invalid whitelist should not fail",
			c: Config{
				Type:      "dockerhub",
				Name:      "invalidwhitelist",
				WhiteList: []string{"[0-9]++"},
			},
			validate: func(reg Registry) bool {
				return assert.True(t, len(reg.filter.whiteRegexp) == 0) &&
					assert.True(t, len(reg.filter.failedWhiteRegexp) == 1)
			},
		},
		{
			name: "invalid blacklist should not fail",
			c: Config{
				Type:      "dockerhub",
				Name:      "invalidblackist",
				BlackList: []string{"[0-9]++"},
			},
			validate: func(reg Registry) bool {
				return assert.True(t, len(reg.filter.blackRegexp) == 0) &&
					assert.True(t, len(reg.filter.failedBlackRegexp) == 1)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// HACK: if we need a url for testing the server, then set it here
			if tc.c.URL == "NEEDSURL" {
				s := adaptertest.GetAPIV2Server(t)
				defer s.Close()
				tc.c.URL = adaptertest.GetURL(t, s).String()
			}

			reg, err := NewRegistry(tc.c, "")

			if tc.expectederr {
				assert.Error(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			// assert.True(t, tc.r.adapter.Called["GetImageNames"])
			// assert.True(t, tc.r.adapter.Called["FetchSpecs"])
			assert.True(t, tc.validate(reg))
		})
	}
}

func TestRegistryName(t *testing.T) {
	testCases := []struct {
		name     string
		r        Registry
		expected string
	}{
		{
			name: "registry name",
			r: Registry{
				config: Config{
					Name: "registryname",
				},
			},
			expected: "registryname",
		},
		{
			name: "empty name",
			r: Registry{
				config: Config{
					Name: "",
				},
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.r.RegistryName())
		})
	}
}

func TestValidate(t *testing.T) {

	testCases := []struct {
		name     string
		c        Config
		expected bool
	}{
		{
			name:     "empty name",
			c:        Config{Name: ""},
			expected: false,
		},
		{
			name: "valid name, empty authtype and authname",
			c: Config{
				Name:     "daname",
				AuthName: "",
				AuthType: "",
			},
			expected: true,
		},
		{
			name: "valid name, empty authtype, non-empty authname",
			c: Config{
				Name:     "daname",
				AuthName: "shouldfail",
				AuthType: "",
			},
			expected: false,
		},
		{
			name: "valid name, file, empty authname",
			c: Config{
				Name:     "daname",
				AuthName: "",
				AuthType: "file",
			},
			expected: false,
		},
		{
			name: "valid name, file, non-empty authname",
			c: Config{
				Name:     "daname",
				AuthName: "non-empty",
				AuthType: "file",
			},
			expected: true,
		},
		{
			name: "valid name, secret, empty authname",
			c: Config{
				Name:     "daname",
				AuthName: "",
				AuthType: "secret",
			},
			expected: false,
		},
		{
			name: "valid name, secret, non-empty authname",
			c: Config{
				Name:     "daname",
				AuthName: "non-empty",
				AuthType: "secret",
			},
			expected: true,
		},
		{
			name: "valid name, config, without user",
			c: Config{
				Name:     "daname",
				User:     "",
				AuthType: "config",
			},
			expected: false,
		},
		{
			name: "valid name, config, without pass",
			c: Config{
				Name:     "daname",
				User:     "user",
				Pass:     "",
				AuthType: "config",
			},
			expected: false,
		},
		{
			name: "valid name, config, user, pass",
			c: Config{
				Name:     "daname",
				User:     "user",
				Pass:     "$3cr3+",
				AuthType: "config",
			},
			expected: true,
		},
		{
			name: "valid name, unknown",
			c: Config{
				Name:     "daname",
				AuthType: "unknown",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.c.Validate())
		})
	}
}

type fakeAdapter struct{}

func (f fakeAdapter) GetImageNames() ([]string, error) {
	return []string{}, nil
}

func (f fakeAdapter) FetchSpecs(names []string) ([]*bundle.Spec, error) {
	return []*bundle.Spec{}, nil
}

func (f fakeAdapter) RegistryName() string {
	return ""
}

func TestAdapterWithConfiguration(t *testing.T) {
	c := Config{
		Name: "nsa",
		Type: "custom",
	}

	f := fakeAdapter{}

	reg, err := NewCustomRegistry(c, f, "")
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, reg.adapter, f, "registry uses wrong adapter")
	assert.Equal(t, reg.config, c, "registrying using wrong config")
}

func TestRetrieveRegistryAuth(t *testing.T) {

	testCases := []struct {
		name        string
		input       Config
		ns          string
		client      *fake.Clientset
		expected    Config
		expectederr bool
	}{
		{
			name: "secret auth type with no client should fail",
			input: Config{
				AuthName: "secret",
				AuthType: "secret",
			},
			expected:    Config{},
			expectederr: true,
		},
		{
			name: "secret auth type",
			ns:   "testing",
			input: Config{
				AuthName: "registrysecret",
				AuthType: "secret",
			},
			client: fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "registrysecret",
					Namespace: "testing",
				},
				Data: map[string][]byte{
					"username": []byte("secretusername"),
					"password": []byte("secretpassword"),
				},
			}),
			expected: Config{
				AuthName: "registrysecret",
				AuthType: "secret",
				User:     "secretusername",
				Pass:     "secretpassword",
			},
		},
		{
			name: "secret auth type with empty secret should fail",
			ns:   "testing",
			input: Config{
				AuthName: "registrysecret",
				AuthType: "secret",
			},
			client: fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "registrysecret",
					Namespace: "testing",
				},
				Data: map[string][]byte{
					"username": []byte(""),
					"password": []byte(""),
				},
			}),
			expected:    Config{},
			expectederr: true,
		},
		{
			name: "file auth type with no auth name should fail",
			input: Config{
				AuthName: "",
				AuthType: "file",
			},
			expected:    Config{},
			expectederr: true,
		},
		{
			name: "file auth type",
			input: Config{
				AuthName: "testdata/fileauthtest",
				AuthType: "file",
			},
			expected: Config{
				AuthName: "testdata/fileauthtest",
				AuthType: "file",
				User:     "fileuser",
				Pass:     "filepassword",
			},
		},
		{
			name: "file auth type with invalidfile",
			input: Config{
				AuthName: "testdata/invalidfileauthtest",
				AuthType: "file",
			},
			expected:    Config{},
			expectederr: true,
		},
		{
			name: "config auth type with empty user should fail",
			input: Config{
				AuthName: "config",
				AuthType: "config",
				User:     "",
				Pass:     "password",
			},
			expected:    Config{},
			expectederr: true,
		},
		{
			name: "config auth type with empty password should fail",
			input: Config{
				AuthName: "config",
				AuthType: "config",
				User:     "username",
				Pass:     "",
			},
			expected:    Config{},
			expectederr: true,
		},
		{
			name: "config auth type",
			input: Config{
				AuthName: "config",
				AuthType: "config",
				User:     "username",
				Pass:     "password",
			},
			expected: Config{
				AuthName: "config",
				AuthType: "config",
				User:     "username",
				Pass:     "password",
			},
		},
		{
			name: "empty auth type",
			input: Config{
				AuthName: "empty",
				AuthType: "",
			},
			expected: Config{
				AuthName: "empty",
				AuthType: "",
				User:     "",
				Pass:     "",
			},
		},
		{
			name: "unknown auth type",
			input: Config{
				AuthName: "unknown",
				AuthType: "unknown",
			},
			expected:    Config{},
			expectederr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// we need the client for the secrets portion of the test
			k, err := clients.Kubernetes()
			if err != nil {
				t.Fail()
			}

			if tc.client != nil {
				k.Client = tc.client
			}

			output, err := retrieveRegistryAuth(tc.input, tc.ns)

			if tc.expectederr {
				assert.Error(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			assert.Equal(t, tc.expected, output)
		})
	}
}
