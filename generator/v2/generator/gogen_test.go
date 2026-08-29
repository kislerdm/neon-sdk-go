package generator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func Test_newGoTypeDefinition(t *testing.T) {
// 	type args struct {
// 		typeName string
// 		schema   OpenAPISchema
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    string
// 		wantErr assert.ErrorAssertionFunc
// 	}{
// 		{
// 			name: "named string",
// 			args: args{
// 				typeName: "FooEnumFooID",
// 				schema: OpenAPISchema{
// 					Type: "string",
// 				},
// 			},
// 			want:    "type FooEnumFooID string",
// 			wantErr: assert.NoError,
// 		},
// 		{
// 			name: "named float",
// 			args: args{
// 				typeName: "FooEnumFooID",
// 				schema: OpenAPISchema{
// 					Type: "number",
// 				},
// 			},
// 			want:    "type FooEnumFooID float",
// 			wantErr: assert.NoError,
// 		},
// 		{
// 			name: "named map",
// 			args: args{
// 				typeName: "FooEnumFooID",
// 				schema: OpenAPISchema{
// 					Type: "object",
// 				},
// 			},
// 			want:    "type FooEnumFooID map[string]any",
// 			wantErr: assert.NoError,
// 		},
// 		{
// 			name: "description is present",
// 			args: args{
// 				typeName: "FooEnumFooID",
// 				schema: OpenAPISchema{
// 					Description: "foo\n\nbar",
// 					Type:        "object",
// 				},
// 			},
// 			want: `// FooEnumFooID foo
//
// bar
// type FooEnumFooID map[string]any`,
// 			wantErr: assert.NoError,
// 		},
// 		{
// 			name: "type is referencing another type",
// 			args: args{
// 				typeName: "FooEnumFooID",
// 				schema:   OpenAPISchema{Ref: pointer("#/components/responses/Bar")},
// 			},
// 			want:    "type FooEnumFooID Bar",
// 			wantErr: assert.NoError,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := newGoTypesDefinition(tt.args.typeName, tt.args.schema)
// 			if !tt.wantErr(t, err, fmt.Sprintf("typeDefinitionGenerator(%typesDefinition, %typesDefinition)", tt.args.typeName, tt.args.schema)) {
// 				return
// 			}
// 			assert.Equalf(t, tt.want, got, "typeDefinitionGenerator(%typesDefinition, %typesDefinition)", tt.args.typeName, tt.args.schema)
// 		})
// 	}
// }
//
// func Test_newGoStructDefinition(t *testing.T) {
// 	type args struct {
// 		schema OpenAPISchema
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    string
// 		wantErr assert.ErrorAssertionFunc
// 	}{
// 		{
// 			name: "object with two attributes of primitive types, one required",
// 			args: args{
// 				schema: OpenAPISchema{
// 					Type:     "object",
// 					Required: []string{"id"},
// 					Properties: []OpenAPISchema{
// 						{
// 							xRefName:    "id",
// 							Type:        "string",
// 							Description: "foo",
// 						},
// 						{
// 							xRefName:    "bar_url",
// 							Type:        "string",
// 							Description: "bar",
// 						},
// 					},
// 				},
// 			},
// 			want: `struct {
// // ID foo
// ID string ` + "`json:\"id\"`" +
// 				"\n// BarURL bar\nBarURL *string `json:\"bar_url,omitempty\"`" +
// 				"\n}",
// 			wantErr: assert.NoError,
// 		},
// 		{
// 			name: "object with a single attribute: array of objects with two optional attrs",
// 			args: args{
// 				schema: OpenAPISchema{
// 					Type: "object",
// 					Properties: []OpenAPISchema{
// 						{
// 							xRefName: "foo",
// 							Type:     "array",
// 							Items: &OpenAPISchema{
// 								Type: "object",
// 								Properties: []OpenAPISchema{
// 									{
// 										xRefName: "bar",
// 										Type:     "string",
// 									},
// 									{
// 										xRefName: "baz",
// 										Type:     "object",
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want: "struct {\n" +
// 				"FooEnumFooID []struct {\n" +
// 				"Bar *string `json:\"bar,omitempty\"`\n" +
// 				"Baz map[string]any `json:\"baz,omitempty\"`\n" +
// 				"} `json:\"foo,omitempty\"`\n" +
// 				"}",
// 			wantErr: assert.NoError,
// 		},
// 		{
// 			name: "object with a single attribute of the type from schemas",
// 			args: args{
// 				schema: OpenAPISchema{
// 					Type: "object",
// 					Properties: []OpenAPISchema{
// 						{
// 							xRefName: "foo",
// 							Ref:      pointer("#/components/schemas/FooEnumFooID"),
// 						},
// 					},
// 				},
// 			},
// 			want: "struct {\n" +
// 				"FooEnumFooID *FooEnumFooID `json:\"foo,omitempty\"`\n" +
// 				"}",
// 			wantErr: assert.NoError,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := newGoStructDefinition(tt.args.schema)
// 			if !tt.wantErr(t, err, fmt.Sprintf("newGoStructDefinition(%typesDefinition)", tt.args.schema)) {
// 				return
// 			}
// 			assert.Equalf(t, tt.want, got, "newGoStructDefinition(%typesDefinition)", tt.args.schema)
// 		})
// 	}
// }

func Test_newTypesDefinition(t *testing.T) {
	tests := map[string]struct {
		raw   []byte
		want  []string
		errFn assert.ErrorAssertionFunc
	}{
		"response defined as ref to schema": {
			raw: []byte(`{
        "responses": {
            "EmptyResponse": {
                "description": "Empty response",
                "content": {
                    "application/json": {
                        "schema": {
                            "$ref": "#/components/schemas/EmptyResponse"
                        }
                    }
                }
            }
		},
		"schemas": {
			"EmptyResponse": {
                "type": "object",
                "description": "Empty response.",
                "properties": {}
            }
		}
	}`),
			want: []string{
				`// EmptyResponse Empty response.
type EmptyResponse map[string]any`,
			},
			errFn: assert.NoError,
		},
		"schema has property that references the schema itself": {
			raw: []byte(`{"schemas": {"FooEnumFooID": {
	"type": "object",
	"properties": {
		"foo": {
			"type": "array", 
			"items": {"$ref": "#/components/schemas/FooEnumFooID"}
		}
	}
}}}
`),
			want: []string{`type FooEnumFooID struct {
Foo []FooEnumFooID ` + "`json:\"foo,omitempty\"`\n" +
				"}"},
			errFn: assert.NoError,
		},
		"schema has enum property": {
			raw: []byte(`{
		"schemas": {
			"Foo": {
                "type": "object",
                "properties": {
					"foo_id": {
						"type": "string",
						"enum": ["0", "1"]
					}
				}
            }
		}
	}`),
			want: []string{
				`type Foo struct {
FooID *FooFooID ` + "`json:\"foo_id,omitempty\"`\n" +
					"}",
				`type FooFooID struct {
	v string
}

func (v FooFooID) String() string {
	return v.v
}

func (v *FooFooID) UnmarshalJSON(data []byte) error {
	o, err := NewFooFooID(string(data))
	if err != nil {
		return err
	}
	*v = o
	return nil
}

func (v FooFooID) MarshalJSON() ([]byte, error) {
	return []byte(v.v), nil
}

var (
	FooFooID0 = FooFooID{"0"}
	FooFooID1 = FooFooID{"1"}
)

func NewFooFooID(s string) (FooFooID, error) {
	m := map[string]FooFooID{
		"0": FooFooID0,
		"1": FooFooID1,
	}
	v, ok := m[s]
	if !ok {
		return FooFooID{}, fmt.Errorf("unknown value: %v", s)
	}
	return v, nil
}
`,
			},
			errFn: assert.NoError,
		},
	}

	t.Parallel()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var in Components
			assert.NoErrorf(t, json.Unmarshal(test.raw, &in), "could not deserialize raw input")
			got, err := newGoTypes(in)
			test.errFn(t, err)
			if err == nil {
				assert.Equal(t, test.want, got)
			}
		})
	}
}
