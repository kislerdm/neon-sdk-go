package generator

import (
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func Test_extractSpecs(t *testing.T) {
	t.Parallel()

	parseSpecExample := func(s string) (openAPISpec, error) {
		var spec openAPISpec
		if err := spec.UnmarshalJSON([]byte(s)); err != nil {
			return openAPISpec{}, err
		}
		return spec, nil
	}

	enrichSpec := func(spec openAPISpec) openAPISpec {
		spec.Info = &openapi3.Info{
			Title:       "foo",
			Description: "bar",
		}
		spec.Servers = openapi3.Servers{
			{
				URL: "https://console.neon.tech/api/v2",
			},
		}
		return spec
	}

	t.Run(
		"listProjects", func(t *testing.T) {
			// GIVEN
			spec, err := parseSpecExample(
				`{
  "openapi": "3.0.3",
  "paths": {
    "/projects": {
      "get": {
        "summary": "Get a list of projects",
        "description": "Retrieves a list of projects for the Neon account.\nA project is the top-level object in the Neon object hierarchy.\nFor more information, see [Manage projects](https://neon.tech/docs/manage/projects/).\n",
        "tags": [
          "Project"
        ],
        "operationId": "listProjects",
        "parameters": [
          {
            "name": "cursor",
            "description": "Specify the cursor value from the previous response to get the next batch of projects.",
            "in": "query",
            "schema": {
              "type": "string"
            }
          },
          {
            "name": "limit",
            "description": "Specify a value from 1 to 100 to limit number of projects in the response.",
            "in": "query",
            "schema": {
              "type": "integer",
              "minimum": 1,
              "default": 10,
              "maximum": 100
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Returned a list of projects for the Neon account",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {
                      "$ref": "#/components/schemas/ProjectsResponse"
                    },
                    {
                      "$ref": "#/components/schemas/PaginationResponse"
                    }
                  ]
                },
                "example": {
                  "projects": [
                    {
                      "id": "shiny-wind-028834",
                      "platform_id": "aws",
                      "region_id": "aws-us-east-2",
                      "name": "shiny-wind-028834",
                      "provisioner": "k8s-pod",
                      "pg_version": 15,
                      "created_at": "2022-11-23T17:42:25Z",
                      "updated_at": "2022-11-23T17:42:25Z",
                      "proxy_host": "us-east-2.aws.neon.tech",
                      "cpu_used_sec": 0,
                      "branch_logical_size_limit": 0,
                      "owner_id": "1232111",
                      "creation_source": "console",
                      "store_passwords": true,
                      "branch_logical_size_limit_bytes": 10800,
                      "active_time": 100
                    },
                    {
                      "id": "winter-boat-259881",
                      "platform_id": "aws",
                      "region_id": "aws-us-east-2",
                      "name": "winter-boat-259881",
                      "provisioner": "k8s-pod",
                      "pg_version": 15,
                      "created_at": "2022-11-23T17:52:25Z",
                      "updated_at": "2022-11-23T17:52:25Z",
                      "proxy_host": "us-east-2.aws.neon.tech",
                      "cpu_used_sec": 0,
                      "branch_logical_size_limit": 0,
                      "owner_id": "1232111",
                      "creation_source": "console",
                      "store_passwords": true,
                      "branch_logical_size_limit_bytes": 10800,
                      "active_time": 100
                    }
                  ]
                }
              }
            }
          },
          "default": {
            "$ref": "#/components/responses/GeneralError"
          }
        }
      }
    }
  }
}
`,
			)
			if err != nil {
				t.Fatal(err)
			}
			spec = enrichSpec(spec)

			want := templateInput{
				ServerURL: "https://console.neon.tech/api/v2",
				EndpointsInterfaceMethods: []string{
					`// ListProjects Retrieves a list of projects for the Neon account.
// A project is the top-level object in the Neon object hierarchy.
// For more information, see [Manage projects](https://neon.tech/docs/manage/projects/).
ListProjects(cursor *string, limit *int) (struct {
ProjectsResponse
PaginationResponse
}, error)`,
				},
				EndpointsImplementation:     nil,
				EndpointsImplementationTest: nil,
				Types:                       nil,
				EndpointsResponseExample:    nil,
			}

			// WHEN
			got := extractSpecs(spec)

			// THEN
			if !reflect.DeepEqual(got.EndpointsInterfaceMethods, want.EndpointsInterfaceMethods) {
				t.Fatalf(
					"Faulty EndpointsInterfaceMethods: want: %#v, got: %#v",
					want.EndpointsInterfaceMethods,
					got.EndpointsInterfaceMethods,
				)
			}
		},
	)
}
