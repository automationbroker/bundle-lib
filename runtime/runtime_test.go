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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/runtime/mocks"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	rest "k8s.io/client-go/rest"
	fakerest "k8s.io/client-go/rest/fake"
	clientgotesting "k8s.io/client-go/testing"
)

type fakeClientSet struct {
	*fake.Clientset
	rest.Interface
}

type fakeCoreV1 struct {
	fakecorev1.FakeCoreV1
	rest.Interface
}

func (f fakeClientSet) CoreV1() corev1.CoreV1Interface {
	return &fakeCoreV1{
		FakeCoreV1: fakecorev1.FakeCoreV1{Fake: &f.Clientset.Fake},
		Interface:  f.Interface,
	}
}

func (f fakeCoreV1) RESTClient() rest.Interface {
	return f.Interface
}

func sandboxCreateHook(pod, ns string, targetNS []string, role string) error {
	return nil
}

func sandboxDestroyHook(pod, ns string, targetNS []string) error {
	return nil
}

func newRunBundle(ex ExecutionContext) (ExecutionContext, error) {
	return ex, nil
}

func newWatchBundle(pd, ns string, u UpdateDescriptionFn) error {
	return nil
}

func newCopySecretsToNamespace(ex ExecutionContext, cns string, targets []string) error {
	return nil
}

func TestCreateSandbox(t *testing.T) {
	testCases := []struct {
		name      string
		podName   string
		client    *fake.Clientset
		namespace string
		targets   []string
		apbRole   string
		metadata  map[string]string
	}{
		{
			name:      "Test Create Sandbox with namespace in target",
			podName:   "pod-name",
			client:    fake.NewSimpleClientset(),
			namespace: "foo-ns",
			targets:   []string{"foo-ns"},
			apbRole:   "edit",
		},
		{
			name:      "Test Create Sandbox with namespace not in target",
			podName:   "pod-name",
			client:    fake.NewSimpleClientset(),
			namespace: "bar-ns",
			targets:   []string{"satoshi-ns", "nakamoto-ns"},
			apbRole:   "edit",
		},
	}
	k, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("test panic unexpectedly: %#+v", r)
				}
			}()

			tc.client.PrependReactor("create", "namespaces", func(action clientgotesting.Action) (handled bool, ret k8sruntime.Object, err error) {
				ca, ok := action.(clientgotesting.CreateActionImpl)
				if !ok {
					return false, nil, fmt.Errorf("can not get create action")
				}
				ns, ok := ca.Object.(*apicorev1.Namespace)
				if !ok {
					return false, nil, fmt.Errorf("can not get namespace")
				}
				// runtime.go only sets generateName so we need to explicitly set name
				if ns.Name == "" && ns.GenerateName != "" {
					ns.Name = ns.GenerateName
				}
				return false, ns, nil
			})

			k.Client = &fakeClientSet{
				tc.client,
				&fakerest.RESTClient{
					Resp: &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major": "3", "minor": "2"}`))),
					},
					NegotiatedSerializer: scheme.Codecs,
				},
			}
			// Create target namespaces for client
			for _, target := range tc.targets {
				ns := &apicorev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name:         target,
						GenerateName: target,
					},
				}
				_, err := k.Client.CoreV1().Namespaces().Create(ns)
				if err != nil {
					t.Fatalf("Failed to create ns: %v", err)
				}
			}
			NewRuntime(Configuration{})
			p := Provider.(*provider)
			_, _, err = p.CreateSandbox(tc.podName, tc.namespace, tc.targets, tc.apbRole, tc.metadata)
			if err != nil {
				t.Fatalf("Failed to create sandbox: %v", err)
			}
			list, err := k.Client.NetworkingV1().NetworkPolicies(tc.targets[0]).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Failed to get list of network policies: %v", err)
			}
			// Test case where we are deploying to target
			if isNamespaceInTargets(tc.namespace, tc.targets) && len(list.Items) > 0 {
				t.Fatalf("Found a network policy when namespace is in target")
			}
			// Test case where we are not deploying to target
			if !isNamespaceInTargets(tc.namespace, tc.targets) && len(list.Items) == 0 {
				t.Fatalf("Namespace is not in target and found no network policies")
			}
		})
	}
}
func TestNewRuntime(t *testing.T) {
	stateManager := state{nsTarget: defaultNamespace, mountLocation: defaultMountLocation}
	testCases := []struct {
		name             string
		config           Configuration
		client           *fake.Clientset
		response         *http.Response
		expectedProvider *provider
		shouldPanic      bool
	}{
		{
			name:   "New Default Openshift Runtime",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				state:                  stateManager,
				coe:                    newOpenshift(),
				ExtractedCredential:    defaultExtractedCredential{},
				watchBundle:            defaultWatchRunningBundle,
				runBundle:              defaultRunBundle,
				copySecretsToNamespace: defaultCopySecretsToNamespace,
			},
		},
		{
			name:   "New Default Kubernetes Runtime not found",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			expectedProvider: &provider{
				state:                  stateManager,
				coe:                    newKubernetes(),
				ExtractedCredential:    defaultExtractedCredential{},
				watchBundle:            defaultWatchRunningBundle,
				runBundle:              defaultRunBundle,
				copySecretsToNamespace: defaultCopySecretsToNamespace,
			},
		},
		{
			name:   "New Default Kubernetes Runtime unauth",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			expectedProvider: &provider{
				state:                  stateManager,
				coe:                    newKubernetes(),
				ExtractedCredential:    defaultExtractedCredential{},
				watchBundle:            defaultWatchRunningBundle,
				runBundle:              defaultRunBundle,
				copySecretsToNamespace: defaultCopySecretsToNamespace,
			},
		},
		{
			name:   "New Default Kubernetes Runtime forbidden",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			expectedProvider: &provider{
				state:                  stateManager,
				coe:                    newKubernetes(),
				ExtractedCredential:    defaultExtractedCredential{},
				watchBundle:            defaultWatchRunningBundle,
				runBundle:              defaultRunBundle,
				copySecretsToNamespace: defaultCopySecretsToNamespace,
			},
		},
		{
			name:   "Panic on finding cluster error",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			shouldPanic: true,
			expectedProvider: &provider{
				state:                  stateManager,
				coe:                    newKubernetes(),
				ExtractedCredential:    defaultExtractedCredential{},
				watchBundle:            defaultWatchRunningBundle,
				runBundle:              defaultRunBundle,
				copySecretsToNamespace: defaultCopySecretsToNamespace,
			},
		},
		{
			name: "New Default Openshift Runtime with mock extracted credentials",
			config: Configuration{
				ExtractedCredential: &mocks.ExtractedCredential{},
			},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				state:                  stateManager,
				coe:                    newOpenshift(),
				ExtractedCredential:    &mocks.ExtractedCredential{},
				watchBundle:            defaultWatchRunningBundle,
				runBundle:              defaultRunBundle,
				copySecretsToNamespace: defaultCopySecretsToNamespace,
			},
		},
		{
			name: "New Default Openshift Runtime with pre sandbox hooks",
			config: Configuration{
				PreCreateSandboxHooks:  []PreSandboxCreate{sandboxCreateHook},
				PreDestroySandboxHooks: []PreSandboxDestroy{sandboxDestroyHook},
			},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				state:                  stateManager,
				coe:                    newOpenshift(),
				ExtractedCredential:    defaultExtractedCredential{},
				preSandboxCreate:       []PreSandboxCreate{sandboxCreateHook},
				preSandboxDestroy:      []PreSandboxDestroy{sandboxDestroyHook},
				watchBundle:            defaultWatchRunningBundle,
				runBundle:              defaultRunBundle,
				copySecretsToNamespace: defaultCopySecretsToNamespace,
			},
		},
		{
			name: "New Default Openshift Runtime with pre sandbox hooks",
			config: Configuration{
				PostCreateSandboxHooks:  []PostSandboxCreate{sandboxCreateHook},
				PostDestroySandboxHooks: []PostSandboxDestroy{sandboxDestroyHook},
			},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				state:                  stateManager,
				coe:                    newOpenshift(),
				ExtractedCredential:    defaultExtractedCredential{},
				postSandboxCreate:      []PostSandboxCreate{sandboxCreateHook},
				postSandboxDestroy:     []PostSandboxDestroy{sandboxDestroyHook},
				watchBundle:            defaultWatchRunningBundle,
				runBundle:              defaultRunBundle,
				copySecretsToNamespace: defaultCopySecretsToNamespace,
			},
		},
		{
			name: "New Default Openshift Runtime with different run bundle",
			config: Configuration{
				RunBundle: newRunBundle,
			},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				coe:                 newOpenshift(),
				ExtractedCredential: defaultExtractedCredential{},
			},
		},
		{
			name: "New Default Openshift Runtime with different watch bundle",
			config: Configuration{
				WatchBundle: newWatchBundle,
			},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				coe:                 newOpenshift(),
				ExtractedCredential: defaultExtractedCredential{},
			},
		},
		{
			name: "New Default Openshift Runtime with different copy secrets",
			config: Configuration{
				CopySecretsToNamespace: newCopySecretsToNamespace,
			},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				coe:                 newOpenshift(),
				ExtractedCredential: defaultExtractedCredential{},
			},
		},
	}
	k, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Fatalf("test panic unexpectedly: %#+v", r)
				}
			}()
			k.Client = &fakeClientSet{
				tc.client,
				&fakerest.RESTClient{
					Resp:                 tc.response,
					NegotiatedSerializer: scheme.Codecs,
				},
			}
			NewRuntime(tc.config)
			p := Provider.(*provider)
			if p.watchBundle == nil {
				t.Fatalf("expected a watchBundle function to be defined but it was nil ")
			}
			if len(p.preSandboxCreate) != len(tc.expectedProvider.preSandboxCreate) {
				t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
			}
			if len(p.preSandboxDestroy) != len(tc.expectedProvider.preSandboxDestroy) {
				t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
			}
			if len(p.postSandboxDestroy) != len(tc.expectedProvider.postSandboxDestroy) {
				t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
			}
			if len(p.postSandboxCreate) != len(tc.expectedProvider.postSandboxCreate) {
				t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
			}
			if !reflect.DeepEqual(tc.expectedProvider.coe, p.coe) {
				t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
			}
			if !reflect.DeepEqual(tc.expectedProvider.ExtractedCredential, p.ExtractedCredential) {
				t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
			}

		})
	}
}
