package k8s

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/automationbroker/bundle-lib/authorization"
	"k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

func TestNewAuthorizer(t *testing.T) {
	a, err := NewAuthorizer("group", "resource", "verb")
	if err != nil {
		t.Fatalf("unable to get new k8s authorizer - %v", err)
	}
	auth, ok := a.(k8sAuthorization)
	if !ok {
		t.Fatal("unable to get new k8sauthorizer type")
	}
	resource := authorizationv1.ResourceAttributes{
		Group:    "group",
		Resource: "resource",
		Verb:     "verb",
	}
	if !reflect.DeepEqual(resource, auth.resource) {
		t.Fatalf("invalid resource attribute\nexpected: %#+v\nactual: %#+v", resource, auth.resource)
	}
}

type fakeSubjectAccessReview struct {
	SubjectAccessReview *authorizationv1.SubjectAccessReview
}

func (fsar fakeSubjectAccessReview) Create(sar *authorizationv1.SubjectAccessReview) (*authorizationv1.SubjectAccessReview, error) {
	if !reflect.DeepEqual(fsar.SubjectAccessReview.Spec, sar.Spec) {
		return nil, fmt.Errorf("unknown subject access review")
	}
	return fsar.SubjectAccessReview, nil
}

type FakeAuthUser struct {
	v1.UserInfo
}

func (f FakeAuthUser) Username() string {
	return f.UserInfo.Username
}

func TestSARUserInfoAuthorizer(t *testing.T) {
	a, err := NewAuthorizer("group", "resource", "verb")
	if err != nil {
		t.Fatalf("unable to get new k8s authorizer - %v", err)
	}
	auth, ok := a.(k8sAuthorization)
	if !ok {
		t.Fatal("unable to get new k8sauthorizer type")
	}
	testCases := []struct {
		name             string
		sar              authv1.SubjectAccessReviewExpansion
		user             v1.UserInfo
		expectedDecision authorization.Decision
		shouldError      bool
		useFakeAuthUser  bool
	}{
		{
			name: "allowed request",
			sar: fakeSubjectAccessReview{
				SubjectAccessReview: &authorizationv1.SubjectAccessReview{
					Spec: authorizationv1.SubjectAccessReviewSpec{
						User:   "foo",
						Groups: []string{},
						Extra:  nil,
						ResourceAttributes: &authorizationv1.ResourceAttributes{
							Group:     "group",
							Resource:  "resource",
							Verb:      "verb",
							Namespace: "location",
						},
					},
					Status: authorizationv1.SubjectAccessReviewStatus{
						Allowed: true,
					},
				},
			},
			user: v1.UserInfo{
				Username: "foo",
				Groups:   []string{},
				Extra:    nil,
			},
			expectedDecision: authorization.DecisionAllowed,
		},
		{
			name: "no opinion request",
			sar: fakeSubjectAccessReview{
				SubjectAccessReview: &authorizationv1.SubjectAccessReview{
					Spec: authorizationv1.SubjectAccessReviewSpec{
						User:   "foo",
						Groups: []string{},
						Extra:  nil,
						ResourceAttributes: &authorizationv1.ResourceAttributes{
							Group:     "group",
							Resource:  "resource",
							Verb:      "verb",
							Namespace: "location",
						},
					},
					Status: authorizationv1.SubjectAccessReviewStatus{},
				},
			},
			user: v1.UserInfo{
				Username: "foo",
				Groups:   []string{},
				Extra:    nil,
			},
			expectedDecision: authorization.DecisionNoOpinion,
		},
		{
			name: "denied request",
			sar: fakeSubjectAccessReview{
				SubjectAccessReview: &authorizationv1.SubjectAccessReview{
					Spec: authorizationv1.SubjectAccessReviewSpec{
						User:   "foo",
						Groups: []string{},
						Extra:  nil,
						ResourceAttributes: &authorizationv1.ResourceAttributes{
							Group:     "group",
							Resource:  "resource",
							Verb:      "verb",
							Namespace: "location",
						},
					},
					Status: authorizationv1.SubjectAccessReviewStatus{
						Denied: true,
					},
				},
			},
			user: v1.UserInfo{
				Username: "foo",
				Groups:   []string{},
				Extra:    map[string]v1.ExtraValue{"scope": []string{"hello"}},
			},
			expectedDecision: authorization.DecisionDeny,
		},
		{
			name: "allowed and denied",
			sar: fakeSubjectAccessReview{
				SubjectAccessReview: &authorizationv1.SubjectAccessReview{
					Spec: authorizationv1.SubjectAccessReviewSpec{
						User:   "foo",
						Groups: []string{},
						Extra:  nil,
						ResourceAttributes: &authorizationv1.ResourceAttributes{
							Group:     "group",
							Resource:  "resource",
							Verb:      "verb",
							Namespace: "location",
						},
					},
					Status: authorizationv1.SubjectAccessReviewStatus{
						Denied:  true,
						Allowed: true,
					},
				},
			},
			user: v1.UserInfo{
				Username: "foo",
				Groups:   []string{},
				Extra:    nil,
			},
			expectedDecision: authorization.DecisionDeny,
			shouldError:      true,
		},
		{
			name: "errored on creation",
			sar: fakeSubjectAccessReview{
				SubjectAccessReview: &authorizationv1.SubjectAccessReview{
					Spec: authorizationv1.SubjectAccessReviewSpec{
						User:   "unknown",
						Groups: []string{},
						Extra:  nil,
						ResourceAttributes: &authorizationv1.ResourceAttributes{
							Group:     "group",
							Resource:  "resource",
							Verb:      "verb",
							Namespace: "location",
						},
					},
					Status: authorizationv1.SubjectAccessReviewStatus{
						Denied:  true,
						Allowed: true,
					},
				},
			},
			user: v1.UserInfo{
				Username: "foo",
				Groups:   []string{},
				Extra:    nil,
			},
			expectedDecision: authorization.DecisionDeny,
			shouldError:      true,
		},
		{
			name: "error unkown user.,",
			sar: fakeSubjectAccessReview{
				SubjectAccessReview: &authorizationv1.SubjectAccessReview{
					Spec: authorizationv1.SubjectAccessReviewSpec{
						User:   "unknown",
						Groups: []string{},
						Extra:  nil,
						ResourceAttributes: &authorizationv1.ResourceAttributes{
							Group:     "group",
							Resource:  "resource",
							Verb:      "verb",
							Namespace: "location",
						},
					},
					Status: authorizationv1.SubjectAccessReviewStatus{
						Denied:  true,
						Allowed: true,
					},
				},
			},
			user: v1.UserInfo{
				Username: "foo",
				Groups:   []string{},
				Extra:    nil,
			},
			expectedDecision: authorization.DecisionDeny,
			shouldError:      true,
			useFakeAuthUser:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var au authorization.AuthorizeUser
			if tc.useFakeAuthUser {
				au = &FakeAuthUser{
					UserInfo: tc.user,
				}
			} else {
				au = &AuthorizationUser{
					UserInfo: tc.user,
				}
			}
			if au.Username() != tc.user.Username {
				t.Fatalf("username should be passed through")
			}
			auth.client = tc.sar
			dec, err := auth.Authorize(au, "location")
			if err != nil {
				if tc.shouldError {
					return
				}
				t.Fatalf("unknown error occured: %v", err)
			}
			if dec != tc.expectedDecision {
				t.Fatalf("expected: %v decision got: %v", tc.expectedDecision, dec)
			}
		})
	}
}
