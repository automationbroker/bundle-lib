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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

const alphaApbTestFile = "alpha_apb.yml"

func loadTestFile(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

const PlanName = "dev"
const PlanDescription = "Mediawiki123 apb implementation"

var PlanMetadata = map[string]interface{}{
	"displayName":     "Development",
	"longDescription": "Basic development plan",
	"cost":            "$0.00",
}

const PlanFree = true
const PlanBindable = true

var PlanUpdatesTo = []string{"foo"}

var expectedPlanParameters = []ParameterDescriptor{
	ParameterDescriptor{
		Name:      "mediawiki_db_schema",
		Title:     "Mediawiki DB Schema",
		Type:      "string",
		Default:   "mediawiki",
		Updatable: false,
		Required:  true},
	ParameterDescriptor{
		Name:      "mediawiki_site_name",
		Title:     "Mediawiki Site Name",
		Type:      "string",
		Default:   "MediaWiki",
		Updatable: true,
		Required:  true},
	ParameterDescriptor{
		Name:      "mediawiki_site_lang",
		Title:     "Mediawiki Site Language",
		Type:      "string",
		Default:   "en",
		Updatable: false,
		Required:  true},
	ParameterDescriptor{
		Name:      "mediawiki_admin_user",
		Title:     "Mediawiki Admin User",
		Type:      "string",
		Default:   "admin",
		Updatable: false,
		Required:  true},
	ParameterDescriptor{
		Name:      "mediawiki_admin_pass",
		Title:     "Mediawiki Admin User Password",
		Type:      "string",
		Updatable: false,
		Required:  true},
	ParameterDescriptor{
		Name:      "mediawiki_mock_enum",
		Title:     "Mediawiki Fake Enum Param",
		Type:      "enum",
		Enum:      []string{"Yes", "No"},
		Default:   "Yes",
		Updatable: false,
		Required:  true},
	ParameterDescriptor{
		Name:  "mediawiki_conditional_show",
		Title: "Mediawiki Example Conditional Default Shown",
		Type:  "string",
		Dependencies: []Dependency{
			Dependency{
				Key:   "mediawiki_mock_enum",
				Value: "Yes",
			},
		},
		Updatable: false,
		Required:  false},
	ParameterDescriptor{
		Name:  "mediawiki_conditional_hide",
		Title: "Mediawiki Example Conditional Default Hidden",
		Type:  "string",
		Dependencies: []Dependency{
			Dependency{
				Key:   "mediawiki_mock_enum",
				Value: "No",
			},
		},
		Updatable: false,
		Required:  false},
}

var p = Plan{
	ID:          "",
	Name:        PlanName,
	Description: PlanDescription,
	Metadata:    PlanMetadata,
	Free:        PlanFree,
	Bindable:    PlanBindable,
	Parameters:  expectedPlanParameters,
	UpdatesTo:   PlanUpdatesTo,
}

const SpecVersion = "1.0"
const SpecRuntime = 1
const SpecName = "mediawiki123-apb"
const SpecImage = "ansibleplaybookbundle/mediawiki123-apb"
const SpecBindable = false
const SpecAsync = "optional"
const SpecDescription = "Mediawiki123 apb implementation"
const SpecDelete = false
const SpecPlans = `
[
   {
      "id":"",
      "name":"dev",
      "description":"Mediawiki123 apb implementation",
      "free":true,
      "bindable":true,
      "metadata":{
         "displayName":"Development",
         "longDescription":"Basic development plan",
         "cost":"$0.00"
      },
      "updates_to":[
         "foo"
      ],
      "parameters":[
         {
            "name":"mediawiki_db_schema",
            "title":"Mediawiki DB Schema",
            "type":"string",
            "default":"mediawiki",
            "updatable":false,
            "required":true
         },
         {
            "name":"mediawiki_site_name",
            "title":"Mediawiki Site Name",
            "type":"string",
            "default":"MediaWiki",
            "updatable":true,
            "required":true
         },
         {
            "name":"mediawiki_site_lang",
            "title":"Mediawiki Site Language",
            "type":"string",
            "default":"en",
            "updatable":false,
            "required":true
         },
         {
            "name":"mediawiki_admin_user",
            "title":"Mediawiki Admin User",
            "type":"string",
            "default":"admin",
            "updatable":false,
            "required":true
         },
         {
            "name":"mediawiki_admin_pass",
            "title":"Mediawiki Admin User Password",
            "type":"string",
            "updatable":false,
            "required":true
         },
         {
            "name":"mediawiki_mock_enum",
            "title":"Mediawiki Fake Enum Param",
            "type":"enum",
            "default": "Yes",
            "enum":[
               "Yes",
               "No"
            ],
            "updatable":false,
            "required":true
         },
         {
            "name":"mediawiki_conditional_show",
            "title":"Mediawiki Example Conditional Default Shown",
            "type":"string",
            "updatable":false,
            "required":false,
            "dependencies":[
               {
                  "key":"mediawiki_mock_enum",
                  "value":"Yes"
               }
            ]
         },
         {
            "name":"mediawiki_conditional_hide",
            "title":"Mediawiki Example Conditional Default Hidden",
            "type":"string",
            "updatable":false,
            "required":false,
            "dependencies":[
               {
                  "key":"mediawiki_mock_enum",
                  "value":"No"
               }
            ]
         }
      ]
   }
]
`

var SpecAlpha = map[string]interface{}{"dashboard_redirect": true}
var SpecAlphaStr = `
{
	"dashboard_redirect": true
}
`

var SpecJSON = fmt.Sprintf(`
{
	"id": "",
	"tags": null,
	"description": "%s",
	"version": "%s",
	"runtime": %d,
	"name": "%s",
	"image": "%s",
	"bindable": %t,
	"async": "%s",
	"plans": %s,
	"alpha": %s,
	"delete": %t
}
`, SpecDescription, SpecVersion, SpecRuntime, SpecName, SpecImage, SpecBindable, SpecAsync, SpecPlans, SpecAlphaStr, SpecDelete)

func TestSpecLoadJSON(t *testing.T) {
	s := Spec{}
	err := LoadJSON(SpecJSON, &s)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, s.Description, SpecDescription)
	assert.Equal(t, s.FQName, SpecName)
	assert.Equal(t, s.Version, SpecVersion)
	assert.Equal(t, s.Runtime, SpecRuntime)
	assert.Equal(t, s.Image, SpecImage)
	assert.Equal(t, s.Bindable, SpecBindable)
	assert.Equal(t, s.Async, SpecAsync)
	assert.Equal(t, s.Delete, SpecDelete)
	assert.True(t, reflect.DeepEqual(s.Plans[0].Parameters, expectedPlanParameters))
	assert.True(t, reflect.DeepEqual(s.Alpha, SpecAlpha))
}

func EncodedApb() string {
	apb := `bmFtZTogbWVkaWF3aWtpMTIzLWFwYgppbWFnZTogYW5zaWJsZXBsYXlib29rYnVuZGxlL21lZGlhd2lraTEyMy1hcGIKZGVzY3JpcHRpb246ICJNZWRpYXdpa2kxMjMgYXBiIGltcGxlbWVudGF0aW9uIgpiaW5kYWJsZTogZmFsc2UKYXN5bmM6IG9wdGlvbmFsCm1ldGFkYXRhOgogIGRpc3BsYXluYW1lOiAiUmVkIEhhdCBNZWRpYXdpa2kiCiAgbG9uZ0Rlc2NyaXB0aW9uOiAiQW4gYXBiIHRoYXQgZGVwbG95cyBNZWRpYXdpa2kgMS4yMyIKICBpbWFnZVVSTDogImh0dHBzOi8vdXBsb2FkLndpa2ltZWRpYS5vcmcvd2lraXBlZGlhL2NvbW1vbnMvMC8wMS9NZWRpYVdpa2ktc21hbGxlci1sb2dvLnBuZyIKICBkb2N1bWVudGF0aW9uVVJMOiAiaHR0cHM6Ly93d3cubWVkaWF3aWtpLm9yZy93aWtpL0RvY3VtZW50YXRpb24iCnBsYW5zOgogIC0gbmFtZTogZGV2CiAgICBkZXNjcmlwdGlvbjogIk1lZGlhd2lraTEyMyBhcGIgaW1wbGVtZW50YXRpb24iCiAgICBmcmVlOiB0cnVlCiAgICBiaW5kYWJsZTogdHJ1ZQogICAgbWV0YWRhdGE6CiAgICAgIGRpc3BsYXlOYW1lOiBEZXZlbG9wbWVudAogICAgICBsb25nRGVzY3JpcHRpb246IEJhc2ljIGRldmVsb3BtZW50IHBsYW4KICAgICAgY29zdDogJDAuMDAKICAgIHBhcmFtZXRlcnM6CiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX2RiX3NjaGVtYQogICAgICAgIHRpdGxlOiBNZWRpYXdpa2kgREIgU2NoZW1hCiAgICAgICAgdHlwZTogc3RyaW5nCiAgICAgICAgZGVmYXVsdDogbWVkaWF3aWtpCiAgICAgICAgcmVxdWlyZWQ6IHRydWUKICAgICAgLSBuYW1lOiBtZWRpYXdpa2lfc2l0ZV9uYW1lCiAgICAgICAgdGl0bGU6IE1lZGlhd2lraSBTaXRlIE5hbWUKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICBkZWZhdWx0OiBNZWRpYVdpa2kKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAgICAtIG5hbWU6IG1lZGlhd2lraV9zaXRlX2xhbmcKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIFNpdGUgTGFuZ3VhZ2UKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICBkZWZhdWx0OiBlbgogICAgICAgIHJlcXVpcmVkOiB0cnVlCiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX2FkbWluX3VzZXIKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICBkZWZhdWx0OiBhZG1pbgogICAgICAgIHJlcXVpcmVkOiB0cnVlCiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX2FkbWluX3Bhc3MKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIgUGFzc3dvcmQKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAgYmluZF9wYXJhbWV0ZXJzOgogICAgICAtIG5hbWU6IGJpbmRfcGFyYW1fMQogICAgICAgIHRpdGxlOiBCaW5kIFBhcmFtIDEKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAgICAtIG5hbWU6IGJpbmRfcGFyYW1fMgogICAgICAgIHRpdGxlOiBCaW5kIFBhcmFtIDIKICAgICAgICB0eXBlOiBpbnQKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAgICAtIG5hbWU6IGJpbmRfcGFyYW1fMwogICAgICAgIHRpdGxlOiBCaW5kIFBhcmFtIDMKICAgICAgICB0eXBlOiBzdHJpbmcKCg==`
	return apb
}

func TestSpecDumpJSON(t *testing.T) {
	s := Spec{
		Description: SpecDescription,
		Runtime:     SpecRuntime,
		Version:     SpecVersion,
		FQName:      SpecName,
		Image:       SpecImage,
		Bindable:    SpecBindable,
		Async:       SpecAsync,
		Plans:       []Plan{p},
		Alpha:       SpecAlpha,
	}

	var knownMap interface{}
	var subjectMap interface{}

	raw, err := DumpJSON(&s)
	if err != nil {
		panic(err)
	}

	json.Unmarshal([]byte(SpecJSON), &knownMap)
	json.Unmarshal([]byte(raw), &subjectMap)
	assert.True(t, reflect.DeepEqual(knownMap, subjectMap))
}

func TestEncodedParameters(t *testing.T) {
	decodedyaml, err := base64.StdEncoding.DecodeString(EncodedApb())
	if err != nil {
		t.Fatal(err)
	}

	spec := &Spec{}
	if err = yaml.Unmarshal(decodedyaml, spec); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", spec)
	assert.Equal(t, spec.FQName, "mediawiki123-apb")
	assert.Equal(t, len(spec.Plans[0].Parameters), 5)

	// picking something other than the first one
	sitelang := spec.Plans[0].Parameters[2] // mediawiki_site_lang

	assert.Equal(t, sitelang.Name, "mediawiki_site_lang")
	assert.Equal(t, sitelang.Title, "Mediawiki Site Language")
	assert.Equal(t, sitelang.Type, "string")
	assert.Equal(t, sitelang.Description, "")
	assert.Equal(t, sitelang.Default, "en")
	assert.Equal(t, sitelang.DeprecatedMaxlength, 0)
	assert.Equal(t, sitelang.Pattern, "")
	assert.Equal(t, len(sitelang.Enum), 0)
}

func TestBindInstanceUserParamsNil(t *testing.T) {
	a := BindInstance{
		ID:        uuid.NewUUID(),
		ServiceID: uuid.NewUUID(),
	}
	up := a.UserParameters()
	assert.True(t, up == nil)
}

func TestBindInstanceUserParams(t *testing.T) {
	a := BindInstance{
		ID:        uuid.NewUUID(),
		ServiceID: uuid.NewUUID(),
	}
	a.Parameters = &Parameters{
		"foo":                  "bar",
		"cluster":              "mycluster",
		"namespace":            "mynamespace",
		"_apb_provision_creds": "letmein",
	}

	up := a.UserParameters()

	// Make sure the "foo" key is still included
	assert.True(t, up["foo"] == "bar")

	// Make sure all of these got filtered out
	for _, key := range []string{"cluster", "namespace", "_apb_provision_creds"} {
		_, ok := up[key]
		assert.False(t, ok)
	}

}

func TestEnsureDefaults(t *testing.T) {
	cases := []struct {
		Name           string
		ProvidedParams func() Parameters
		Validate       func(t *testing.T, params Parameters)
	}{
		{
			Name: "test defaults are set",
			ProvidedParams: func() Parameters {
				p := Parameters{}
				p.EnsureDefaults()
				return p
			},
			Validate: func(t *testing.T, actual Parameters) {
				if _, ok := actual[ProvisionCredentialsKey]; !ok {
					t.Fatalf("expected the key %s to be present but it was missing", ProvisionCredentialsKey)
				}
			},
		},
		{
			Name: "test existing key not overwritten",
			ProvidedParams: func() Parameters {
				p := Parameters{ProvisionCredentialsKey: "avalue"}
				p.EnsureDefaults()
				return p
			},
			Validate: func(t *testing.T, p Parameters) {
				if v, ok := p[ProvisionCredentialsKey]; ok {
					if v != "avalue" {
						t.Fatalf("expected the value for %s to be %v but got %v", ProvisionCredentialsKey, "avalue", v)
					}
					return
				}
				t.Fatalf("missing key %v from params", ProvisionCredentialsKey)
			},
		},
		{
			Name: "test default key set if other keys present",
			ProvidedParams: func() Parameters {
				p := Parameters{"somekey": "avalue"}
				p.EnsureDefaults()
				return p
			},
			Validate: func(t *testing.T, p Parameters) {
				if v, ok := p["somekey"]; ok {
					if v != "avalue" {
						t.Fatalf("expected somekey to be set to avalue but was %s", v)
					}
				}
				if v, ok := p[ProvisionCredentialsKey]; ok {
					if v != struct{}{} {
						t.Fatalf("expected the default value for %v to be %v but got %v", ProvisionCredentialsKey, struct{}{}, v)
					}
					return
				}
				t.Fatalf("expected key somekey to be set but it wasnt")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Validate(t, tc.ProvidedParams())
		})
	}
}

func TestBindInstanceEqual(t *testing.T) {
	a := BindInstance{
		ID:         uuid.NewUUID(),
		ServiceID:  uuid.NewUUID(),
		Parameters: &Parameters{"foo": "bar"},
	}
	b := BindInstance{
		ID:         a.ID,
		ServiceID:  a.ServiceID,
		Parameters: &Parameters{"foo": "bar"},
	}
	assert.True(t, a.IsEqual(&b))
	assert.True(t, b.IsEqual(&a))
}

func TestBindInstanceNotEqual(t *testing.T) {

	a := BindInstance{
		ID:         uuid.NewUUID(),
		ServiceID:  uuid.NewUUID(),
		Parameters: &Parameters{"foo": "bar"},
	}

	data := map[string]BindInstance{
		"different parameters": BindInstance{
			ID:         a.ID,
			ServiceID:  a.ServiceID,
			Parameters: &Parameters{"foo": "notbar"},
		},
		"different ID": BindInstance{
			ID:         uuid.NewUUID(),
			ServiceID:  a.ServiceID,
			Parameters: &Parameters{"foo": "bar"},
		},
		"different ServiceID": BindInstance{
			ID:         a.ID,
			ServiceID:  uuid.NewUUID(),
			Parameters: &Parameters{"foo": "bar"},
		},
		"no parameters": BindInstance{
			ID:        a.ID,
			ServiceID: a.ServiceID,
		},
	}

	for key, binding := range data {
		if a.IsEqual(&binding) {
			t.Errorf("bindings were equal for case: %s", key)
		}
		if binding.IsEqual(&a) {
			t.Errorf("bindings were equal for case: %s", key)
		}
	}
}

func TestBuildExtractedCredentials(t *testing.T) {
	output := []byte(`{"db": "fusor_guestbook_db", "user": "duder_two", "pass" :"dog8two"}`)
	bd, _ := buildExtractedCredentials(output)
	assert.NotNil(t, bd, "credential is nil")
	assert.Equal(t, bd.Credentials["db"], "fusor_guestbook_db", "db is not fusor_guestbook_db")
	assert.Equal(t, bd.Credentials["user"], "duder_two", "user is not duder_two")
	assert.Equal(t, bd.Credentials["pass"], "dog8two", "password is not dog8two")
}

func TestAlphaParser(t *testing.T) {
	spec := &Spec{}
	testYaml := loadTestFile(t, alphaApbTestFile)
	if err := yaml.Unmarshal(testYaml, spec); err != nil {
		t.Fatal(err)
	}

	if len(spec.Alpha) == 0 {
		t.Error("spec.Alpha should not be empty")
	}

	var val interface{}
	var dr, ok bool

	if val, ok = spec.Alpha["dashboard_redirect"]; !ok {
		t.Error("spec.Alpha should contain dashboard_redirect key")
	}

	if dr, ok = val.(bool); !ok {
		t.Error(`spec.Alpha["dashboard_redirect"] should assert to bool`)
	}

	assert.True(t, dr)
}

func TestAddRemoveBinding(t *testing.T) {
	si := &ServiceInstance{
		ID: uuid.NewUUID(),
	}

	bID := uuid.NewUUID()
	si.AddBinding(bID)
	assert.True(t, si.BindingIDs[bID.String()], "binding not added")

	si.RemoveBinding(bID)
	toDelete, ok := si.BindingIDs[bID.String()]
	assert.True(t, ok, "binding has been removed, should be marked only")
	assert.False(t, toDelete, "binding not marked as deleted")
}

func TestGetParameter(t *testing.T) {
	testCases := []struct {
		name     string
		plan     *Plan
		input    string
		expected *ParameterDescriptor
	}{
		{
			name:     "no parameters on empty plan should return nil",
			plan:     &Plan{},
			input:    "does_not_exist",
			expected: nil,
		},
		{
			name: "if name does not match should return nil",
			plan: &Plan{
				Name: "plan b",
				Parameters: []ParameterDescriptor{
					{
						Name:        "vncpass",
						Title:       "VNC Password",
						Type:        "string",
						DisplayType: "password",
						Required:    true,
					},
				},
			},
			input:    "does_not_match",
			expected: nil,
		},
		{
			name: "if name matches we should return the ParameterDecriptor",
			plan: &Plan{
				Name: "plan b",
				Parameters: []ParameterDescriptor{
					{
						Name:        "vncpass",
						Title:       "VNC Password",
						Type:        "string",
						DisplayType: "password",
						Required:    true,
					},
				},
			},
			input: "vncpass",
			expected: &ParameterDescriptor{
				Name:        "vncpass",
				Title:       "VNC Password",
				Type:        "string",
				DisplayType: "password",
				Required:    true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := tc.plan.GetParameter(tc.input)
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestSpecGetPlan(t *testing.T) {
	testCases := []struct {
		name      string
		spec      *Spec
		input     string
		expplan   Plan
		expstatus bool
	}{
		{
			name:      "no plans on empty spec should return false",
			spec:      &Spec{},
			input:     "does_not_exist",
			expplan:   Plan{},
			expstatus: false,
		},
		{
			name: "if name does not match should return false",
			spec: &Spec{
				FQName: "spec b",
				Plans: []Plan{
					{
						Name: "plan b",
					},
				},
			},
			input:     "does_not_match",
			expplan:   Plan{},
			expstatus: false,
		},
		{
			name: "if name matches we should return the Plan",
			spec: &Spec{
				FQName: "spec b",
				Plans: []Plan{
					{
						Name: "plan b",
					},
				},
			},
			input: "plan b",
			expplan: Plan{
				Name: "plan b",
			},
			expstatus: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, ok := tc.spec.GetPlan(tc.input)
			assert.Equal(t, tc.expstatus, ok)
			assert.Equal(t, tc.expplan, output)
		})
	}
}

func TestSpecGetPlanFromID(t *testing.T) {
	testCases := []struct {
		name      string
		spec      *Spec
		input     string
		expplan   Plan
		expstatus bool
	}{
		{
			name:      "no plans on empty spec should return false",
			spec:      &Spec{},
			input:     "does_not_exist",
			expplan:   Plan{},
			expstatus: false,
		},
		{
			name: "if name does not match should return false",
			spec: &Spec{
				FQName: "spec b",
				Plans: []Plan{
					{
						ID: "plan b",
					},
				},
			},
			input:     "does_not_match",
			expplan:   Plan{},
			expstatus: false,
		},
		{
			name: "if name matches we should return the Plan",
			spec: &Spec{
				FQName: "spec b",
				Plans: []Plan{
					{
						ID: "plan b",
					},
				},
			},
			input: "plan b",
			expplan: Plan{
				ID: "plan b",
			},
			expstatus: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, ok := tc.spec.GetPlanFromID(tc.input)
			assert.Equal(t, tc.expstatus, ok)
			assert.Equal(t, tc.expplan, output)
		})
	}
}

func TestSpecsLogDump(t *testing.T) {
	testCases := []struct {
		name            string
		input           []*Spec
		expectedrows    int
		expectedentries []string
	}{
		{
			name: "spec with no plans",
			input: []*Spec{
				{
					ID:          "abcdef",
					FQName:      "test-spec",
					Image:       "test-image",
					Bindable:    true,
					Description: "test description",
					Async:       "optional",
				},
			},
			expectedrows: 8,
			expectedentries: []string{
				"============================================================",
				"Spec: abcdef",
				"============================================================",
				"Name: test-spec",
				"Image: test-image",
				"Bindable: true",
				"Description: test description",
				"Async: optional",
			},
		},
		{
			name: "spec with a plan and no parameters",
			input: []*Spec{
				{
					ID:          "abcdef",
					FQName:      "test-spec",
					Image:       "test-image",
					Bindable:    true,
					Description: "test description",
					Async:       "optional",
					Plans: []Plan{
						{
							ID:   "plan b",
							Name: "plan b",
						},
					},
				},
			},
			expectedrows: 9,
			expectedentries: []string{
				"============================================================",
				"Spec: abcdef",
				"============================================================",
				"Name: test-spec",
				"Image: test-image",
				"Bindable: true",
				"Description: test description",
				"Async: optional",
				"Plan: plan b",
			},
		},
		{
			name: "spec with a plan and parameters",
			input: []*Spec{
				{
					ID:          "abcdef",
					FQName:      "test-spec",
					Image:       "test-image",
					Bindable:    true,
					Description: "test description",
					Async:       "optional",
					Plans: []Plan{
						{
							ID:   "plan b",
							Name: "plan b",
							Parameters: []ParameterDescriptor{
								{
									Name:        "vncpass",
									Title:       "VNC Password",
									Type:        "string",
									DisplayType: "password",
									Required:    true,
									Updatable:   true,
									MaxLength:   20,
									MinLength:   8,
								},
							},
						},
					},
				},
			},
			expectedrows: 25,
			expectedentries: []string{
				"============================================================",
				"Spec: abcdef",
				"============================================================",
				"Name: test-spec",
				"Image: test-image",
				"Bindable: true",
				"Description: test description",
				"Async: optional",
				"Plan: plan b",
				"  Name: vncpass",
				"  Title: VNC Password",
				"  Type: string",
				"  Description: ",
				"  Default: <nil>",
				"  DeprecatedMaxlength: 0",
				"  MaxLength: 20",
				"  MinLength: 8",
				"  Pattern: ",
				"  MultipleOf: 0.000000",
				"  Minimum: (*bundle.NilableNumber)(nil)",
				"  Maximum: (*bundle.NilableNumber)(nil)",
				"  ExclusiveMinimum: (*bundle.NilableNumber)(nil)",
				"  ExclusiveMaximum: (*bundle.NilableNumber)(nil)",
				"  Required: true",
				"  Enum: []",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// capture the logs
			logger, hook := test.NewNullLogger()

			// need to log the debug level from the SpecLogDump
			log.SetLevel(log.DebugLevel)

			// don't print it out during the run
			log.SetOutput(logger.Out)

			// capture and verify
			log.AddHook(hook)

			// test the dump
			SpecsLogDump(tc.input)

			assert.Equal(t, tc.expectedrows, len(hook.Entries))
			for i, entry := range hook.Entries {
				assert.Equal(t, log.DebugLevel, hook.LastEntry().Level)
				assert.Equal(t, tc.expectedentries[i], entry.Message)
			}

			hook.Reset()
			assert.Nil(t, hook.LastEntry())
		})
	}
}

func TestNewSpecManifest(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*Spec
		expected SpecManifest
	}{
		{
			name:     "empty spec list should return empty SpecManifest",
			input:    []*Spec{},
			expected: SpecManifest{},
		},
		{
			name:     "spec list with nils should return nil",
			input:    []*Spec{nil},
			expected: nil,
		},
		{
			name: "given a list of specs, manifest should contain them",
			input: []*Spec{
				{
					ID:     "abcdef",
					FQName: "test-spec-a",
				},
				{
					ID:     "ghijk",
					FQName: "test-spec-b",
				},
			},
			expected: SpecManifest{
				"abcdef": &Spec{
					ID:     "abcdef",
					FQName: "test-spec-a",
				},
				"ghijk": &Spec{
					ID:     "ghijk",
					FQName: "test-spec-b",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, NewSpecManifest(tc.input))
		})
	}
}
