package generator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenAPIDefinition_UnmarshalJSONShallStoreSortedPaths(t *testing.T) {
	in := []byte(`{
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
