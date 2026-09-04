package generator

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed openAPIDefinition.json
var openAPIDefinitionNeon20260831Raw []byte

func TestGoTypesDefinition(t *testing.T) {
	tests := map[string]struct {
		raw   []byte
		want  string
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
			want: `// EmptyResponse Empty response.
type EmptyResponse map[string]any`,
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
			want: `type FooEnumFooID struct {
Foo []FooEnumFooID ` + "`json:\"foo,omitempty\"`\n" +
				"}",
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
			want: `type Foo struct {
FooID *FooFooID ` + "`json:\"foo_id,omitempty\"`\n" +
				"}" +
				`
type FooFooID struct {
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
	s = strings.TrimLeft(strings.TrimRight(s, "\""), "\"")
	v, ok := m[s]
	if !ok {
		return FooFooID{}, fmt.Errorf("unknown value: %v", s)
	}
	return v, nil
}
`,
			errFn: assert.NoError,
		},
		"response has allOf w/ ref to schemas only": {
			raw: []byte(`{
			"responses": {
	            "ListOperations": {
	                "description": "Returned a list of operations\n",
	                "content": {
	                    "application/json": {
	                        "schema": {
	                            "allOf": [
	                                {
	                                    "$ref": "#/components/schemas/OperationsResponse"
	                                },
	                                {
	                                    "$ref": "#/components/schemas/PaginationResponse"
	                                }
	                            ]
	                        }
	                    }
	                }
	            }
			}
		}`),
			want: `// ListOperations Returned a list of operations
type ListOperations struct {
OperationsResponse
PaginationResponse
}`,
			errFn: assert.NoError,
		},
		"response has allOf w/ ref to schemas and inline schema": {
			raw: []byte(`{
			"responses": {
	            "ListOperations": {
	                "description": "Returned a list of operations\n",
	                "content": {
	                    "application/json": {
	                        "schema": {
	                            "allOf": [
	                                {
	                                    "$ref": "#/components/schemas/OperationsResponse"
	                                },
									{
										"type": "object",
										"properties": {
											"foo": {"type": "string", "description": "Foo"},
											"bar": {"type": "number", "description": "Bar"}
										}
									},
									{
	                                    "$ref": "#/components/schemas/PaginationResponse"
	                                }
	                            ]
	                        }
	                    }
	                }
	            }
			}
		}`),
			want: `// ListOperations Returned a list of operations
type ListOperations struct {
OperationsResponse
PaginationResponse` +
				"\n// Bar Bar\n" +
				"Bar *float64 `json:\"bar,omitempty\"`\n" +
				"// Foo Foo\n" +
				"Foo *string `json:\"foo,omitempty\"`\n" +
				"}",

			errFn: assert.NoError,
		},
		"response has allOf w/ ref to schema and inline schema of the string": {
			raw: []byte(`{
			"responses": {
	            "ListOperations": {
	                "description": "Returned a list of operations\n",
	                "content": {
	                    "application/json": {
	                        "schema": {
	                            "allOf": [
	                                {
	                                    "$ref": "#/components/schemas/OperationsResponse"
	                                },
									{"type": "string"}
	                            ]
	                        }
	                    }
	                }
	            }
			}
		}`),
			want: "",
			errFn: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorContains(t, err, "unsupported type for allOf clause")
			},
		},
		"schema AllowedIps": {
			raw: []byte(`{
	"schemas": {
		"AllowedIps": {
            "description": "A list of IP addresses that are allowed to connect to the compute endpoint.\nIf the list is empty or not set, all IP addresses are allowed.\nIf protected_branches_only is true, the list will be applied only to protected branches.\n",
            "type": "object",
            "properties": {
                "ips": {
                    "description": "A list of IP addresses that are allowed to connect to the endpoint.",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "protected_branches_only": {
                    "description": "If true, the list will be applied only to protected branches.",
                    "type": "boolean"
                }
            }
        }
	}
}`),
			want: `// AllowedIps A list of IP addresses that are allowed to connect to the compute endpoint.
// If the list is empty or not set, all IP addresses are allowed.
// If protected_branches_only is true, the list will be applied only to protected branches.
type AllowedIps struct {
` +
				"// Ips A list of IP addresses that are allowed to connect to the endpoint.\n" +
				"Ips []string `json:\"ips,omitempty\"`\n" +
				"// ProtectedBranchesOnly If true, the list will be applied only to protected branches.\n" +
				"ProtectedBranchesOnly *bool `json:\"protected_branches_only,omitempty\"`\n" +
				"}",
			errFn: assert.NoError,
		},
	}

	t.Parallel()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var in Components
			assert.NoErrorf(t, json.Unmarshal(test.raw, &in), "could not deserialize raw input")

			var repo = new(TypesRepo)
			typesDefinitionInputFromComponents(repo, in)
			err := newGoTypesDefinition(repo)
			test.errFn(t, err)

			if err == nil {
				got := repo.TypesDefinition()
				assert.Equal(t, test.want, got)
			}
		})
	}

	t.Run("shall generate definitions of correct number of types based on the Neon openAPI spec",
		func(t *testing.T) {
			var spec OpenAPIDefinition
			assert.NoErrorf(t, json.Unmarshal(openAPIDefinitionNeon20260831Raw, &spec),
				"could not deserialize openAPI spec")
			var repo = new(TypesRepo)
			typesDefinitionInputFromComponents(repo, spec.Components)
			err := newGoTypesDefinition(repo)
			assert.NoError(t, err)
			if err == nil {
				got := repo.TypesDefinition()
				assert.GreaterOrEqual(t, len(got), len(spec.Components.Schemas)+len(spec.Components.Responses))
			}
		})
}

func Test_newGoEnumDefinition(t *testing.T) {
	raw := []byte(`{
				"type": "string",
				"enum": ["0.1", "1-1", "2_2"]
			}`)
	want := `struct {
	v string
}

func (v Foo) String() string {
	return v.v
}

func (v *Foo) UnmarshalJSON(data []byte) error {
	o, err := NewFoo(string(data))
	if err != nil {
		return err
	}
	*v = o
	return nil
}

func (v Foo) MarshalJSON() ([]byte, error) {
	return []byte(v.v), nil
}

var (
	Foo01 = Foo{"0.1"}
	Foo11 = Foo{"1-1"}
	Foo22 = Foo{"2_2"}
)

func NewFoo(s string) (Foo, error) {
	m := map[string]Foo{
		"0.1": Foo01,
		"1-1": Foo11,
		"2_2": Foo22,
	}
	s = strings.TrimLeft(strings.TrimRight(s, "\""), "\"")
	v, ok := m[s]
	if !ok {
		return Foo{}, fmt.Errorf("unknown value: %v", s)
	}
	return v, nil
}
`
	var in OpenAPISchema
	assert.NoErrorf(t, json.Unmarshal(raw, &in), "could not deserialize raw input")
	got, err := newGoEnumDefinition(in, "Foo")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_newGoDocString(t *testing.T) {
	name := "AllowedIps"
	description := `A list of IP addresses that are allowed to connect to the compute endpoint.


If the list is empty or not set, all IP addresses are allowed.
If protected_branches_only is true, the list will be applied only to protected branches.


`
	want := `// AllowedIps A list of IP addresses that are allowed to connect to the compute endpoint.
//
//
// If the list is empty or not set, all IP addresses are allowed.
// If protected_branches_only is true, the list will be applied only to protected branches.
`
	got := newGoDocString(name, description)
	assert.Equal(t, want, got)
}
