package clients

import (
	"reflect"
	"testing"

	authapi "github.com/openshift/api/authorization/v1"
	authfake "github.com/openshift/client-go/authorization/clientset/versioned/fake"
	authv1 "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
)

func TestOpenshiftSubjectRulesReview(t *testing.T) {
	o, err := Openshift()
	if err != nil {
		t.Fail()
	}

	testCases := []struct {
		name      string
		auth      authv1.AuthorizationV1Interface
		rules     []authapi.PolicyRule
		user      string
		groups    []string
		scopes    []string
		namespace string
		shouldErr bool
	}{
		{
			name: "get rules",
			auth: authfake.NewSimpleClientset(&authapi.SubjectRulesReview{
				Spec: authapi.SubjectRulesReviewSpec{
					User: "test-users",
				},
			}).AuthorizationV1(),
			rules:     []authapi.PolicyRule{},
			user:      "test-user",
			namespace: "ns1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o.authClient = tc.auth
			rules, err := o.SubjectRulesReview(tc.user, tc.groups, tc.scopes, tc.namespace)
			if err != nil && !tc.shouldErr {
				t.Fatalf("unknown error - %v", err)
				return
			}
			if err != nil && tc.shouldErr {
				return
			}
			if !reflect.DeepEqual(rules, tc.rules) {
				t.Fatalf("\nActual Rules: %v\n\nExpected Rules: %v\n", rules, tc.rules)
				return
			}
		})
	}
}
