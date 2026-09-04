package generator

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenAPIDefinition_UnmarshalJSONShallStoreSortedPaths(t *testing.T) {
	in := []byte(`{
"servers": [
    {
        "url": "https://console.neon.tech/api/v2"
    }
],
"paths": {
	"/b/{foo}/c": {},
	"/x/": {},
	"/a/{foo}/baz": {},
	"/a/{foo}/bar": {},
	"/a/{foo}": {}
}}`)
	want := []OpenAPIPath{
		{
			URLPath: "/a/{foo}",
		},
		{
			URLPath: "/a/{foo}/bar",
		},
		{
			URLPath: "/a/{foo}/baz",
		},
		{
			URLPath: "/b/{foo}/c",
		},
		{
			URLPath: "/x/",
		},
	}

	var v OpenAPIDefinition
	assert.NoError(t, json.Unmarshal(in, &v))
	assert.Equal(t, want, v.Paths)
}

func TestOpenAPISchema_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in   []byte
		want OpenAPISchema
	}{
		"shall store sorted parameters": {
			in: []byte(`{
    "type": "object",
    "properties": {
        "x": {
            "type": "string"
        },
		"b": {
            "type": "string"
        },
		"a": {
            "type": "string"
        }
    }
}`),
			want: OpenAPISchema{
				Type: "object",
				Properties: []OpenAPISchema{
					{xRefName: "a", Type: "string"},
					{xRefName: "b", Type: "string"},
					{xRefName: "x", Type: "string"},
				},
			},
		},
		"shall extract allOf": {
			in: []byte(`{
				        "allOf": [
				            {
				                "$ref": "#/components/schemas/ApiKeyCreateRequest"
				            },
				            {
				                "type": "object",
				                "properties": {
				                    "project_id": {
				                        "description": "If set, the API key can access only this project",
				                        "type": "string",
				                        "pattern": "^[a-z0-9-]{1,60}$"
				                    }
				                }
				            }
						]
			        }`),
			want: OpenAPISchema{
				AllOf: []OpenAPISchema{
					{Ref: pointer("#/components/schemas/ApiKeyCreateRequest")},
					{
						Type: "object",
						Properties: []OpenAPISchema{
							{
								xRefName:    "project_id",
								Description: "If set, the API key can access only this project",
								Type:        "string",
							},
						},
					},
				},
			},
		},
		"shall extract enum": {
			in: []byte(`{
                "description": "The action performed by the operation",
                "type": "string",
                "enum": [
                    "create_compute",
                    "create_timeline",
                    "start_compute",
                    "suspend_compute",
                    "apply_config",
                    "check_availability",
                    "delete_timeline",
                    "create_branch",
                    "import_data",
                    "tenant_ignore",
                    "tenant_attach",
                    "tenant_detach",
                    "tenant_detach_safekeepers",
                    "tenant_attach_safekeepers",
                    "tenant_reattach",
                    "replace_safekeeper",
                    "disable_maintenance",
                    "apply_storage_config",
                    "prepare_secondary_pageserver",
                    "switch_pageserver",
                    "detach_parent_branch",
                    "timeline_archive",
                    "timeline_unarchive",
                    "start_reserved_compute",
                    "sync_dbs_and_roles_from_compute",
                    "apply_schema_from_branch",
                    "timeline_mark_invisible",
                    "timeline_update_protected_config",
                    "prewarm_replica",
                    "promote_replica",
                    "set_storage_non_dirty",
                    "swap_binding_id",
                    "finalize_migration",
                    "mark_migration_prepared",
                    "update_catalog"
                ]
            }`),
			want: OpenAPISchema{
				Type:        "string",
				Description: "The action performed by the operation",
				Enum: []string{
					"create_compute",
					"create_timeline",
					"start_compute",
					"suspend_compute",
					"apply_config",
					"check_availability",
					"delete_timeline",
					"create_branch",
					"import_data",
					"tenant_ignore",
					"tenant_attach",
					"tenant_detach",
					"tenant_detach_safekeepers",
					"tenant_attach_safekeepers",
					"tenant_reattach",
					"replace_safekeeper",
					"disable_maintenance",
					"apply_storage_config",
					"prepare_secondary_pageserver",
					"switch_pageserver",
					"detach_parent_branch",
					"timeline_archive",
					"timeline_unarchive",
					"start_reserved_compute",
					"sync_dbs_and_roles_from_compute",
					"apply_schema_from_branch",
					"timeline_mark_invisible",
					"timeline_update_protected_config",
					"prewarm_replica",
					"promote_replica",
					"set_storage_non_dirty",
					"swap_binding_id",
					"finalize_migration",
					"mark_migration_prepared",
					"update_catalog",
				},
			},
		},
		"shall extract items for array property": {
			in: []byte(`{
                "type": "object",
                "properties": {
                    "periods": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/ConsumptionHistoryPerPeriodV2"
						}
                    }
                },
                "required": [
                    "periods"
                ]
            }`),
			want: OpenAPISchema{
				Type: "object",
				Required: []string{
					"periods",
				},
				Properties: []OpenAPISchema{
					{
						xRefName: "periods",
						Type:     "array",
						Items:    &OpenAPISchema{Ref: pointer("#/components/schemas/ConsumptionHistoryPerPeriodV2")},
					},
				},
			},
		},
		"shall extract format and minimum/maximum": {
			in: []byte(`{
                "type": "integer",
                "format": "int64",
                "minimum": -1,
                "maximum": 604800
            }`),
			want: OpenAPISchema{
				Type:    "integer",
				Format:  pointer("int64"),
				Minimum: pointer(-1.),
				Maximum: pointer(604800.),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got OpenAPISchema
			assert.NoError(t, json.Unmarshal(test.in, &got))
			assert.Equal(t, test.want, got)
		})
	}
}

func pointer[V string | int64 | int | float64](v V) *V {
	return &v
}

func TestOpenAPIPathMethodRequestBody_UnmarshalJSON(t *testing.T) {
	in := []byte(`{
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/ProjectCreateRequest"
                            },
                            "examples": {
                                "required_attributes_only": {
                                    "summary": "Required attributes only",
                                    "value": {
                                        "project": {
                                            "name": "myproject"
                                        }
                                    }
                                },
                                "commonly_specified_attributes": {
                                    "summary": "Commonly-specified attributes",
                                    "value": {
                                        "project": {
                                            "name": "myproject",
                                            "region_id": "aws-us-east-2",
                                            "pg_version": 15
                                        }
                                    }
                                },
                                "with_autoscaling": {
                                    "summary": "With autoscaling attributes",
                                    "value": {
                                        "project": {
                                            "name": "myproject",
                                            "region_id": "aws-us-east-2",
                                            "pg_version": 15,
                                            "autoscaling_limit_min_cu": 0.25,
                                            "autoscaling_limit_max_cu": 1,
                                            "provisioner": "k8s-neonvm"
                                        }
                                    }
                                },
                                "with_branch_attributes": {
                                    "summary": "With branch attributes",
                                    "value": {
                                        "project": {
                                            "name": "myproject",
                                            "region_id": "aws-us-east-2",
                                            "pg_version": 15,
                                            "branch": {
                                                "name": "mybranch",
                                                "role_name": "sally",
                                                "database_name": "mydb"
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }`)

	want := OpenAPIPathMethodRequestBody{
		Required: true,
		Schema: OpenAPISchema{
			Ref: pointer("#/components/schemas/ProjectCreateRequest"),
		},
	}

	var got OpenAPIPathMethodRequestBody
	assert.NoError(t, json.Unmarshal(in, &got))
	assert.Equal(t, want, got)
}

func TestOpenAPIResponse_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in   []byte
		want OpenAPIResponse
	}{
		"shall extract schema from content": {
			in: []byte(`{
                        "description": "Returned a list of shared projects for the Neon account",
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
                    }`),
			want: OpenAPIResponse{
				Description: "Returned a list of shared projects for the Neon account",
				Schema: OpenAPISchema{
					AllOf: []OpenAPISchema{
						{Ref: pointer("#/components/schemas/ProjectsResponse")},
						{Ref: pointer("#/components/schemas/PaginationResponse")},
					},
				},
			},
		},
		"shall extract ref": {
			in: []byte(`{
                    "$ref": "#/components/responses/GeneralError"
                }`),
			want: OpenAPIResponse{
				Ref: pointer("#/components/responses/GeneralError"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got OpenAPIResponse
			assert.NoError(t, json.Unmarshal(test.in, &got))
			assert.Equal(t, test.want, got)
		})
	}
}

//go:embed openAPIDefinition.json
var testSchemaNeon []byte

func TestDeserializationNeonOpenAPISpec(t *testing.T) {
	t.Run("20260904", func(t *testing.T) {
		wantPathsCnt := 120
		wantComponentsResponsesCnt := 7
		wantComponentsSchemasCnt := 275
		wantComponentsParametersCnt := 4

		var got OpenAPIDefinition
		assert.NoError(t, json.Unmarshal(testSchemaNeon, &got))
		assert.Len(t, got.Paths, wantPathsCnt)
		assert.Len(t, got.Components.Responses, wantComponentsResponsesCnt)
		assert.Len(t, got.Components.Schemas, wantComponentsSchemasCnt)
		assert.Len(t, got.Components.Parameters, wantComponentsParametersCnt)
	})
}
