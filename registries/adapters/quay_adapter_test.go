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
	"net/url"
	"strings"
	"testing"

	ft "github.com/stretchr/testify/assert"
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
	ft.Equal(t, a.RegistryName(), "quay.io", "registry adaptor name does not match")
}

func TestNewQuayAdapter(t *testing.T) {
	a, err := NewQuayAdapter(quayTestConfig)
	if err != nil {
		t.Fatal("Error: ", err)
	}

	b := QuayAdapter{}
	b.config.Org = "foo"
	b.config.Tag = "latest"

	ft.Equal(t, a, b, "adaptor returned is not valid")
}

func TestQuayGetImageNames(t *testing.T) {
	serv := getQuayServer(t)
	defer serv.Close()
	quayTestConfig.URL = getQuayURL(t, serv)
	a, _ := NewQuayAdapter(quayTestConfig)

	imagesFound, err := a.GetImageNames()
	if err != nil {
		t.Fatal("Error: ", err)
	}
	ft.Equal(t, 4, len(imagesFound), "image names returned did not match expected config")
}

func TestQuayFetchSpecs(t *testing.T) {
	serv := getQuayServer(t)
	defer serv.Close()
	quayTestConfig.URL = getQuayURL(t, serv)
	a, _ := NewQuayAdapter(quayTestConfig)

	imgList := []string{"test-apb"}
	s, err := a.FetchSpecs(imgList)
	if err != nil {
		t.Fatal("Error: ", err)
	}

	ft.Equal(t, 1, len(s), "image names returned did not match expected config")
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
