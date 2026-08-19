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

	t.Run("shall store sorted parameters", func(t *testing.T) {
		in := []byte(`{
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
}`)

		want := OpenAPISchema{
			Type: "object",
			Properties: []OpenAPISchema{
				{xRefName: "a", Type: "string"},
				{xRefName: "b", Type: "string"},
				{xRefName: "x", Type: "string"},
			},
		}
		var got OpenAPISchema
		assert.NoError(t, json.Unmarshal(in, &got))
		assert.Equal(t, want, got)
	})

	t.Run("shall extract allOf", func(t *testing.T) {
		in := []byte(`{
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
    }`)
		want := OpenAPISchema{
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
		}
		var got OpenAPISchema
		assert.NoError(t, json.Unmarshal(in, &got))
		assert.Equal(t, want, got)
	})
}

func pointer[V string](v V) *V {
	return &v
}
