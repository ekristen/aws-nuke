package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/datazone"
)

func Test_Mock_DataZone_Remove(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		resource func(*datazone.DataZone) interface{ Remove(context.Context) error }
	}{
		{
			name:   "domain",
			method: http.MethodDelete,
			path:   "/v2/domains/domain-123",
			resource: func(svc *datazone.DataZone) interface{ Remove(context.Context) error } {
				return &DataZoneDomain{svc: svc, ID: aws.String("domain-123")}
			},
		},
		{
			name:   "project",
			method: http.MethodDelete,
			path:   "/v2/domains/domain-123/projects/project-456",
			resource: func(svc *datazone.DataZone) interface{ Remove(context.Context) error } {
				return &DataZoneProject{svc: svc, DomainID: aws.String("domain-123"), ID: aws.String("project-456")}
			},
		},
		{
			name:   "subscription grant",
			method: http.MethodDelete,
			path:   "/v2/domains/domain-123/subscription-grants/grant-456",
			resource: func(svc *datazone.DataZone) interface{ Remove(context.Context) error } {
				return &DataZoneSubscriptionGrant{svc: svc, DomainID: aws.String("domain-123"), ID: aws.String("grant-456")}
			},
		},
		{
			name:   "subscription target",
			method: http.MethodDelete,
			path:   "/v2/domains/domain-123/environments/environment-789/subscription-targets/target-456",
			resource: func(svc *datazone.DataZone) interface{ Remove(context.Context) error } {
				return &DataZoneSubscriptionTarget{
					svc:           svc,
					DomainID:      aws.String("domain-123"),
					EnvironmentID: aws.String("environment-789"),
					ID:            aws.String("target-456"),
				}
			},
		},
		{
			name:   "subscription",
			method: http.MethodPut,
			path:   "/v2/domains/domain-123/subscriptions/subscription-456/cancel",
			resource: func(svc *datazone.DataZone) interface{ Remove(context.Context) error } {
				return &DataZoneSubscription{svc: svc, DomainID: aws.String("domain-123"), ID: aws.String("subscription-456")}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method {
					t.Errorf("request method = %s, want %s", r.Method, tt.method)
				}
				if r.URL.Path != tt.path {
					t.Errorf("request path = %s, want %s", r.URL.Path, tt.path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("{}"))
			}))
			defer server.Close()

			sess := session.Must(session.NewSession(&aws.Config{
				Region:      aws.String("us-east-1"),
				Endpoint:    aws.String(server.URL),
				Credentials: credentials.NewStaticCredentials("test-key", "test-secret", ""),
			}))

			if err := tt.resource(datazone.New(sess)).Remove(context.Background()); err != nil {
				t.Fatalf("Remove() returned error: %v", err)
			}
		})
	}
}
