//go:build e2etest
// +build e2etest

package sdk_test

import (
	"testing"

	sdk "github.com/kislerdm/neon-sdk-go"
)

func pointer[V string | float64 | int | sdk.ComputeUnit](v V) *V {
	return &v
}

func TestSmoke(t *testing.T) {
	cl, err := sdk.NewClient()
	if err != nil {
		t.Fatalf("cannot initialise SDK: %v", err)
	}

	t.Run(
		"shall return empty list of pre-existing projects", func(t *testing.T) {
			// GIVEN
			// No pre-existing projects

			// WHEN

			o, err := cl.ListProjects(nil, nil)
			if err != nil {
				t.Fatal(err)
			}

			// THEN

			if len(o.Projects) > 0 {
				t.Errorf("unexpected number of project found, expected 0")
			}
		},
	)

	t.Run(
		"shall create a project", func(t *testing.T) {
			// GIVEN
			// Project name and autoscalling limit

			const wantName = "foo"
			var wantAutoscallingLimit sdk.ComputeUnit = 1. / 4

			// WHEN

			o, err := cl.CreateProject(
				sdk.ProjectCreateRequest{
					Project: sdk.ProjectCreateRequestProject{
						AutoscalingLimitMaxCu: pointer(wantAutoscallingLimit),
						Name:                  pointer(wantName),
					},
				},
			)
			if err != nil {
				t.Fatal(err)
			}

			// THEN

			if o.Project.Name != wantName {
				t.Errorf("unexpected error, project name does not match expected %s", wantName)
			}

			gotAutoscallingLimit := o.Project.DefaultEndpointSettings.AutoscalingLimitMaxCu
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
			o, err := cl.ListProjects(nil, nil)
			if err != nil {
				t.Fatalf("cannot clear the test")
			}

			for _, pr := range o.Projects {
				_, err := cl.DeleteProject(pr.ID)
				if err != nil {
					t.Errorf("cannot deleted project %s", pr.ID)
				}
			}
		},
	)
}
