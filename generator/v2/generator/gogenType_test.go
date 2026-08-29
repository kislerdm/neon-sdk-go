package generator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newGoTypes(t *testing.T) {
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
			want: []string{
				`type ListOperations struct {
OperationsResponse
PaginationResponse
}`,
			},
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
			want: []string{
				`type ListOperations struct {
OperationsResponse
PaginationResponse` +
					"\n// Bar Bar\n" +
					"Bar *float `json:\"bar,omitempty\"`\n" +
					"// Foo Foo\n" +
					"Foo *string `json:\"foo,omitempty\"`\n" +
					"}",
			},
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
			want: nil,
			errFn: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorContains(t, err, "unsupported type for allOf clause")
			},
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

func Test_newGoEnumDefinition(t *testing.T) {
	raw := []byte(`{
				"type": "string",
				"enum": ["0", "1", "2"]
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
	Foo0 = Foo{"0"}
	Foo1 = Foo{"1"}
	Foo2 = Foo{"2"}
)

func NewFoo(s string) (Foo, error) {
	m := map[string]Foo{
		"0": Foo0,
		"1": Foo1,
		"2": Foo2,
	}
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
