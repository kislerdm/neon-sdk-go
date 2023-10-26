//go:build acctest
// +build acctest

package sdk_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	sdk "github.com/kislerdm/neon-sdk-go"
)

func TestSmoke(t *testing.T) {
	cl, err := sdk.NewClient(sdk.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatalf("cannot initialise SDK: %v", err)
	}

	var projectID string

	t.Run(
		"shall create a project", func(t *testing.T) {
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
							AutoscalingLimitMinCu: sdk.ComputeUnit(wantAutoscallingLimit),
							AutoscalingLimitMaxCu: sdk.ComputeUnit(wantAutoscallingLimit),
						},
					},
				},
			)
			if err != nil {
				t.Fatal(err)
			}

			// THEN
			project := o.ProjectResponse.Project
			projectID = project.ID

			if project.Name != wantName {
				t.Errorf("unexpected error, project name does not match expected %s", wantName)
			}

			gotAutoscallingLimit := project.DefaultEndpointSettings.AutoscalingLimitMaxCu
			if gotAutoscallingLimit != wantAutoscallingLimit {
				t.Errorf(
					"unexpected autoscalling limit, want: %v, got: %v", wantAutoscallingLimit,
					gotAutoscallingLimit,
				)
			}

			gotAutoscallingLimit = project.DefaultEndpointSettings.AutoscalingLimitMinCu
			if gotAutoscallingLimit != wantAutoscallingLimit {
				t.Errorf(
					"unexpected autoscalling limit, want: %v, got: %v", wantAutoscallingLimit,
					gotAutoscallingLimit,
				)
			}
		},
	)

	t.Cleanup(
		func() {
			if _, err := cl.DeleteProject(projectID); err != nil {
				t.Errorf("cannot deleted project %s", projectID)
			}
		},
	)
}
