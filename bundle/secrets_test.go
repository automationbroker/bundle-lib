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
	"fmt"
	"sync"
	"testing"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/automationbroker/bundle-lib/clients"
	"github.com/stretchr/testify/assert"
)

func TestSecretsConfig(t *testing.T) {
	testCases := []struct {
		name     string
		config   SecretsConfig
		expected bool
	}{
		{
			name:     "default empty config",
			config:   SecretsConfig{},
			expected: false,
		},
		{
			name:     "partially filled out config",
			config:   SecretsConfig{Name: "testconfig"},
			expected: false,
		},
		{
			name: "fully filled out config",
			config: SecretsConfig{
				Name:    "full config",
				ApbName: "marc/anthony-apb",
				Secret:  "latin-artist",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.config.Validate())
		})
	}
}

func TestSecretsCache(t *testing.T) {

	assert := assert.New(t)

	testCases := []struct {
		name        string
		spec        []*Spec
		rules       []AssociationRule
		expectation interface{}
		validate    func(interface{}, []AssociationRule, []*Spec) bool
	}{
		{
			name: "initialize cache",
			expectation: secretsCache{
				mapping: make(map[string]map[string]bool),
				rwSync:  sync.RWMutex{},
				rules: []AssociationRule{
					{
						BundleName: "initialize",
						Secret:     "initial secret",
					},
				},
			},
			rules: []AssociationRule{
				{
					BundleName: "initialize",
					Secret:     "initial secret",
				},
			},
			validate: func(exp interface{}, r []AssociationRule, s []*Spec) bool {
				InitializeSecretsCache(r)
				return assert.Equal(exp, secrets)
			},
		},
		{
			name: "get secrets",
			spec: []*Spec{
				{
					ID:     "10",
					FQName: "puertorican/marc-anthony-apb",
				},
				{
					ID:     "20",
					FQName: "colombian/shakira-apb",
				},
			},
			expectation: true,
			rules: []AssociationRule{
				{
					BundleName: "puertorican/marc-anthony-apb",
					Secret:     "initial secret",
				},
				{
					BundleName: "dockerhub/not-added-apb",
					Secret:     "secret",
				},
			},
			validate: func(exp interface{}, r []AssociationRule, s []*Spec) bool {
				InitializeSecretsCache(r)
				AddSecrets(s)

				there := getSecrets(s[0])
				notthere := getSecrets(s[1])
				return assert.Len(there, 1) && assert.Len(notthere, 0)
			},
		},
		{
			name: "add a bunch of secrets",
			spec: []*Spec{
				{
					ID:     "10",
					FQName: "puertorican/marc-anthony-apb",
				},
				{
					ID:     "20",
					FQName: "colombian/shakira-apb",
				},
			},
			expectation: true,
			rules: []AssociationRule{
				{
					BundleName: "puertorican/marc-anthony-apb",
					Secret:     "initial secret",
				},
				{
					BundleName: "dockerhub/not-added-apb",
					Secret:     "secret",
				},
			},
			validate: func(exp interface{}, r []AssociationRule, s []*Spec) bool {
				InitializeSecretsCache(r)

				AddSecrets(s)

				// only the first spec should be added since it's the only one
				// that matches.
				return assert.Len(secrets.mapping, 1) &&
					assert.Equal(exp, secrets.mapping[s[0].FQName][r[0].Secret])
			},
		},
		{
			name: "add secrets for",
			spec: []*Spec{
				{
					ID:     "10",
					FQName: "puertorican/marc-anthony-apb",
				},
			},
			expectation: true,
			rules: []AssociationRule{
				{
					BundleName: "puertorican/marc-anthony-apb",
					Secret:     "initial secret",
				},
				{
					BundleName: "dockerhub/not-added-apb",
					Secret:     "secret",
				},
			},
			validate: func(exp interface{}, r []AssociationRule, s []*Spec) bool {
				InitializeSecretsCache(r)

				AddSecretsFor(s[0])

				return assert.Len(secrets.mapping, 1) &&
					assert.Equal(exp, secrets.mapping[s[0].FQName][r[0].Secret])
			},
		},
		{
			name: "add a secret",
			spec: []*Spec{
				{
					ID:     "10",
					FQName: "puertorican/marc-anthony-apb",
				},
			},
			expectation: true,
			rules: []AssociationRule{
				{
					BundleName: "initialize",
					Secret:     "initial secret",
				},
			},
			validate: func(exp interface{}, r []AssociationRule, s []*Spec) bool {
				InitializeSecretsCache(r)

				addSecret(s[0], r[0])

				return assert.Equal(exp, secrets.mapping[s[0].FQName][r[0].Secret])
			},
		},
		{
			name: "match",
			spec: []*Spec{
				{
					ID:     "10",
					FQName: "colombian/shakira-apb",
				},
			},
			expectation: true,
			rules: []AssociationRule{
				{
					BundleName: "colombian/shakira-apb",
					Secret:     "match secret",
				},
			},
			validate: func(exp interface{}, r []AssociationRule, s []*Spec) bool {
				return assert.Equal(exp, match(s[0], r[0]))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := fmt.Sprintf("%s failed", tc.name)
			assert.True(tc.validate(tc.expectation, tc.rules, tc.spec), msg)
		})
	}
}

func TestSecretsFooCache(t *testing.T) {
	t.Skip()
	k, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	testCases := []struct {
		name   string
		client *fake.Clientset
	}{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k.Client = tc.client
		})
	}
}
