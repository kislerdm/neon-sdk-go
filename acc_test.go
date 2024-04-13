package sdk_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	sdk "github.com/kislerdm/neon-sdk-go"
)

func TestSmoke(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	cl, err := sdk.NewClient(sdk.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatalf("cannot initialise SDK: %v", err)
	}

	// GIVEN
	// Project name and autoscalling limit
	var wantName = fmt.Sprintf("%d", time.Now().UnixMilli())
	var wantAutoscallingLimit sdk.ComputeUnit = 1. / 4

	// WHEN
	o, err := cl.CreateProject(
		sdk.ProjectCreateRequest{
			Project: sdk.ProjectCreateRequestProject{
				Name: &wantName,
				DefaultEndpointSettings: &sdk.DefaultEndpointSettings{
					AutoscalingLimitMinCu: &wantAutoscallingLimit,
					AutoscalingLimitMaxCu: &wantAutoscallingLimit,
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	project := o.ProjectResponse.Project
	projectID := project.ID

	t.Run(
		"shall create a project", func(t *testing.T) {
			// THEN
			if project.Name != wantName {
				t.Errorf("unexpected error, project name does not match expected %s", wantName)
			}

			gotAutoscallingLimit := project.DefaultEndpointSettings.AutoscalingLimitMaxCu
			if *gotAutoscallingLimit != wantAutoscallingLimit {
				t.Errorf(
					"unexpected autoscalling limit, want: %v, got: %v", wantAutoscallingLimit,
					gotAutoscallingLimit,
				)
			}

			gotAutoscallingLimit = project.DefaultEndpointSettings.AutoscalingLimitMinCu
			if *gotAutoscallingLimit != wantAutoscallingLimit {
				t.Errorf(
					"unexpected autoscalling limit, want: %v, got: %v", wantAutoscallingLimit,
					gotAutoscallingLimit,
				)
			}

			if !reflect.DeepEqual(*project.Settings.AllowedIps.Ips, []string{}) || *project.Settings.AllowedIps.PrimaryBranchOnly {
				t.Errorf("unexpected project's allowed IPs: %v", *project.Settings.AllowedIps.Ips)
			}
		},
	)

	t.Run("shall grant and revoke permissions successfully", func(t *testing.T) {
		// disposable email on yopmail.com
		const grantee = "neon-go-sdk-test@1xp.fr"

		resp, err := cl.ListProjectPermissions(projectID)
		if err != nil {
			t.Fatal(err)
		}

		initGrants := len(resp.ProjectPermissions)
		if initGrants > 0 {
			t.Errorf("unexpected number of granted permissions: want = 0, got = %d", initGrants)
		}

		r, err := cl.GrantPermissionToProject(projectID, sdk.GrantPermissionToProjectRequest{
			Email: grantee,
		})
		if err != nil {
			t.Fatal(err)
		}

		if r.GrantedToEmail != grantee {
			t.Fatalf("unexpected grantee email was set. want = %s, got = %s", grantee, r.GrantedToEmail)
		}

		r, err = cl.RevokePermissionFromProject(projectID, r.ID)
		if err != nil {
			t.Fatal(err)
		}

		if r.RevokedAt == nil || r.GrantedAt.After(*r.RevokedAt) {
			t.Fatal("unexpected revokedAt, it must be not nil and not before the grantedAt")
		}
	})

	t.Cleanup(
		func() {
			if _, err := cl.DeleteProject(projectID); err != nil {
				t.Errorf("cannot deleted project %s", projectID)
			}
		},
	)
}
