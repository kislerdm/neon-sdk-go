package generator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newGoMethodsDefinition(t *testing.T) {
	tests := map[string]struct {
		rawSpec                []byte
		parameters             []OpenAPIParameter
		want                   string
		wantTypesRepoQueueSize int
		wantErr                assert.ErrorAssertionFunc
	}{
		"method with one common required parameter, one required for path and one optional query inputs": {
			rawSpec: []byte(`{
	"paths": {
		"/foo/{project_id}": {
			"parameters": [
	            {
	                "name": "project_id",
	                "description": "Neon project ID",
	                "in": "path",
	                "required": true,
	                "schema": {
	                    "type": "string",
	                    "pattern": "^[a-z0-9-]{1,60}$"
	                }
	            }
	        ],
			"get": {
                "description": "Foo bar\n\nqux.\n",
                "operationId": "getFoo",
				"parameters": [{
                    "name": "metrics",
					"in": "query",
                    "schema": {
                        "$ref": "#/components/schemas/ConsumptionHistoryQueryMetrics"
                    }
                }],
				"responses": {
                    "200": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/ConsumptionHistoryPerProjectResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/PaginationResponse"
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "403": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GeneralError"
                                }
                            }
                        }
                    },
                    "404": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GeneralError"
                                }
                            }
                        }
                    },
                    "406": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GeneralError"
                                }
                            }
                        }
                    },
                    "429": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GeneralError"
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
	},
	"components": {
		"schemas": {
			"ConsumptionHistoryQueryMetrics": {
                "type": "array",
                "items": {
                    "type": "string"
                }
            }
		}
	}
}`),
			want: `// GetFoo Foo bar
//
// qux.
func (c Client) GetFoo(projectID string, metrics *ConsumptionHistoryQueryMetrics) (GetFooRespObj, error) {
var (
queryElements []string
query string
)
if metrics != nil {
queryElements = append(queryElements, "metrics="+fmt.Sprintf("%v", *metrics))
}
if len(queryElements) > 0 {
query = "?" + strings.Join(queryElements, "&")
}
var v GetFooRespObj
if err := c.requestHandler(c.baseURL+"/foo/"+projectID, "GET", nil, &v); err != nil {
return GetFooRespObj{}, err
}
return v, nil
}
`,
			wantTypesRepoQueueSize: 1,
			wantErr:                assert.NoError,
		},
	}

	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var spec OpenAPIDefinition
			err := json.Unmarshal(tt.rawSpec, &spec)
			assert.NoErrorf(t, err, "could not deserialize raw input")

			typesRepo := new(TypesRepo)
			got, err := newGoMethodsDefinition(spec.Paths, spec.Components.Parameters, typesRepo)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
			assert.Len(t, typesRepo.inputQueue, tt.wantTypesRepoQueueSize)
		})
	}
}

func Test_newPathCode(t *testing.T) {
	tests := map[string]struct {
		path       string
		parameters map[string]typeDescriptor
		want       string
		errFn      assert.ErrorAssertionFunc
	}{
		"single int path arg": {
			path: "/foo/{foo_id}",
			parameters: map[string]typeDescriptor{
				"foo_id": {
					FnArgumentDefinition:            "fooID int",
					TransformationToStrFnDefinition: "strconv.FormatInt(fooID, 10)",
					InQuery:                         false,
					Nillable:                        false,
					Name:                            "foo_id",
					Required:                        true,
				},
			},
			want:  `"/foo/"+strconv.FormatInt(fooID, 10)`,
			errFn: assert.NoError,
		},
		"int arg followed by a string arg": {
			path: "/foo/{foo_id}/{bar}",
			parameters: map[string]typeDescriptor{
				"foo_id": {
					FnArgumentDefinition:            "fooID int",
					TransformationToStrFnDefinition: "strconv.FormatInt(fooID, 10)",
					InQuery:                         false,
					Nillable:                        false,
					Name:                            "foo_id",
					Required:                        true,
				},
				"bar": {
					FnArgumentDefinition:            "bar string",
					TransformationToStrFnDefinition: "bar",
					InQuery:                         false,
					Nillable:                        false,
					Name:                            "bar",
					Required:                        true,
				},
			},
			want:  `"/foo/"+strconv.FormatInt(fooID, 10)+"/"+bar`,
			errFn: assert.NoError,
		},
		"empty arg": {
			path:  "{}",
			want:  "",
			errFn: assert.Error,
		},
		"no path arguments": {
			path:  "/foo/bar",
			want:  `"/foo/bar"`,
			errFn: assert.NoError,
		},
		"arg is surrounded by parts": {
			path: "/foo/{foo_id}/bar",
			parameters: map[string]typeDescriptor{
				"foo_id": {
					TransformationToStrFnDefinition: "fooID",
					InQuery:                         false,
					Nillable:                        false,
					Name:                            "foo_id",
					Required:                        true,
				},
			},
			want:  `"/foo/"+fooID+"/bar"`,
			errFn: assert.NoError,
		},
		"empty string": {
			path:  "",
			want:  "",
			errFn: assert.NoError,
		},
		`/`: {
			path:  `/`,
			want:  `"/"`,
			errFn: assert.NoError,
		},
		`///`: {
			path:  `///`,
			want:  `"///"`,
			errFn: assert.NoError,
		},
		"only args": {
			path: "/{foo_id}/{bar}",
			parameters: map[string]typeDescriptor{
				"foo_id": {
					FnArgumentDefinition:            "fooID int",
					TransformationToStrFnDefinition: "strconv.FormatInt(fooID, 10)",
					InQuery:                         false,
					Nillable:                        false,
					Name:                            "foo_id",
					Required:                        true,
				},
				"bar": {
					FnArgumentDefinition:            "bar string",
					TransformationToStrFnDefinition: "bar",
					InQuery:                         false,
					Nillable:                        false,
					Name:                            "bar",
					Required:                        true,
				},
			},
			want:  `"/"+strconv.FormatInt(fooID, 10)+"/"+bar`,
			errFn: assert.NoError,
		},
	}

	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := newPathCode(tt.path, tt.parameters)
			tt.errFn(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
