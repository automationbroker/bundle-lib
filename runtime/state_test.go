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

package runtime

import (
	"reflect"
	"testing"

	"github.com/automationbroker/bundle-lib/clients"
	"k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestState(t *testing.T) {
	k, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}
	testCases := []struct {
		name     string
		s        state
		client   *fake.Clientset
		validate func(state) bool
	}{
		{
			name:   "mastername",
			s:      state{},
			client: fake.NewSimpleClientset(),
			validate: func(s state) bool {
				return "foo-state" == s.MasterName("foo")
			},
		},
		{
			name:   "master namespace",
			s:      state{nsTarget: "master-ns"},
			client: fake.NewSimpleClientset(),
			validate: func(s state) bool {
				return "master-ns" == s.MasterNamespace()
			},
		},
		{
			name:   "mount location",
			s:      state{mountLocation: "/tmp/foo"},
			client: fake.NewSimpleClientset(),
			validate: func(s state) bool {
				return "/tmp/foo" == s.MountLocation()
			},
		},
		{
			name: "state is present",
			s: state{
				nsTarget:      "nsTarget",
				mountLocation: "/tmp/foo",
			},
			client: fake.NewSimpleClientset(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "nsTarget",
				},
				Data: map[string]string{"fields": `{"db": "name"}`},
			}),
			validate: func(s state) bool {
				present, err := s.StateIsPresent("foo")
				if err != nil {
					t.Fatal(err)
					return false
				}
				return present
			},
		},
		{
			name: "state is not present should not return error on not found",
			s: state{
				nsTarget:      "nsTarget",
				mountLocation: "/tmp/foo",
			},
			client: fake.NewSimpleClientset(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "nsTarget",
				},
				Data: map[string]string{"fields": `{"db": "name"}`},
			}),
			validate: func(s state) bool {
				present, err := s.StateIsPresent("not_there")
				if err != nil {
					t.Fatal(err)
					return false
				}
				return !present
			},
		},
		{
			name: "copy state",
			s: state{
				nsTarget:      "nsTarget",
				mountLocation: "/tmp/foo",
			},
			client: fake.NewSimpleClientset(
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "from",
						Namespace: "fromNS",
					},
					Data: map[string]string{"fields": `{"db": "from"}`},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "to",
						Namespace: "toNS",
					},
					Data: map[string]string{"fields": `{"db": "to"}`},
				},
			),
			validate: func(s state) bool {
				err := s.CopyState("from", "to", "fromNS", "toNS")
				if err != nil {
					t.Fatal(err)
					return false
				}

				tomap, err := k.Client.CoreV1().ConfigMaps("toNS").Get(
					"to", metav1.GetOptions{})
				if err != nil {
					t.Fatalf("state is not present: %v", err)
					return false
				}
				expectedMap := map[string]string{"fields": `{"db": "from"}`}
				return reflect.DeepEqual(expectedMap, tomap.Data)
			},
		},
		{
			name: "copy state from not found",
			s: state{
				nsTarget:      "nsTarget",
				mountLocation: "/tmp/foo",
			},
			client: fake.NewSimpleClientset(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "nsTarget",
				},
				Data: map[string]string{"fields": `{"db": "name"}`},
			}),
			validate: func(s state) bool {
				err := s.CopyState("from", "to", "fromNS", "toNS")
				if err != nil {
					t.Fatal(err)
					return false
				}
				return true
			},
		},
		{
			name: "delete state successful",
			s: state{
				nsTarget:      "nsTarget",
				mountLocation: "/tmp/foo",
			},
			client: fake.NewSimpleClientset(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "nsTarget",
				},
				Data: map[string]string{"fields": `{"db": "name"}`},
			}),
			validate: func(s state) bool {
				err := s.DeleteState("foo")
				if err != nil {
					t.Fatal(err)
					return false
				}
				_, err = k.Client.CoreV1().ConfigMaps("nsTarget").Get(
					"foo", metav1.GetOptions{})
				if err != nil {
					if kerror.IsNotFound(err) {
						return true
					}
				}
				t.Fatalf("state is present, it should have been deleted")
				return false
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k.Client = tc.client
			if !tc.validate(tc.s) {
				t.Fatal("validation failed")
			}
		})
	}
}
