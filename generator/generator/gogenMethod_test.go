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
if err := c.requestHandler(c.baseURL+"/foo/"+projectID+query, "GET", nil, &v); err != nil {
return GetFooRespObj{}, err
}
return v, nil
}
`,
			wantTypesRepoQueueSize: 1,
			wantErr:                assert.NoError,
		},
		"method with body-less response": {
			rawSpec: []byte(`{
		"paths": {
			"/organizations/{org_id}/vpc/region/{region_id}/vpc_endpoints/{vpc_endpoint_id}": {
					"parameters": [
	                {
	                    "name": "org_id",
	                    "in": "path",
	                    "description": "The Neon organization ID",
	                    "required": true,
	                    "schema": {
	                        "type": "string",
	                        "pattern": "^[a-z0-9-]{1,60}$"
	                    }
	                },
	                {
	                    "name": "region_id",
	                    "in": "path",
	                    "description": "The Neon region ID.\nAzure regions are currently not supported.\n",
	                    "required": true,
	                    "schema": {
	                        "type": "string"
	                    }
	                },
	                {
	                    "name": "vpc_endpoint_id",
	                    "in": "path",
	                    "description": "The VPC endpoint ID",
	                    "required": true,
	                    "schema": {
	                        "type": "string"
	                    }
	                }
	            ],
				"post": {
					"description": "Assigns a VPC endpoint to a Neon organization or updates its existing assignment.\n",
					"operationId": "assignOrganizationVPCEndpoint",
					"requestBody": {
	                    "content": {
	                        "application/json": {
                              "schema": {
                                    "$ref": "#/components/schemas/VPCEndpointAssignment"
                               }
	                        }
	                    },
	                    "required": true
	                },
	                "responses": {
	                    "200": {
	                        "description": "Assigned the VPC endpoint to the specified Neon organization"
	                    },
	                    "default": {
	                        "$ref": "#/components/responses/GeneralError"
	                    }
	                }
				}
			}
		}
	}`),
			parameters: nil,
			want: `// AssignOrganizationVPCEndpoint Assigns a VPC endpoint to a Neon organization or updates its existing assignment.
func (c Client) AssignOrganizationVPCEndpoint(orgID string, regionID string, vpcEndpointID string, cfg VPCEndpointAssignment) error {
return c.requestHandler(c.baseURL+"/organizations/"+orgID+"/vpc/region/"+regionID+"/vpc_endpoints/"+vpcEndpointID, "POST", cfg, nil)
}
`,
			wantTypesRepoQueueSize: 0,
			wantErr:                assert.NoError,
		},
		"getOrganizationMembers": {
			rawSpec: []byte(`{
  "paths": {
    "/organizations/{org_id}/members": {
      "parameters": [
        {
          "name": "org_id",
          "in": "path",
          "description": "The Neon organization ID",
          "required": true,
          "schema": {
            "type": "string",
            "pattern": "^[a-z0-9-]{1,60}$"
          }
        }
      ],
      "get": {
        "summary": "List organization members",
        "tags": [
          "Organizations"
        ],
        "operationId": "getOrganizationMembers",
        "parameters": [
          {
            "name": "sort_by",
            "description": "Sort the members by the specified field. Defaults to joined_at.",
            "in": "query",
            "schema": {
              "type": "string",
              "default": "joined_at",
              "enum": [
                "email",
                "role",
                "joined_at"
              ]
            }
          },
          {
            "$ref": "#/components/parameters/CursorParam"
          },
          {
            "$ref": "#/components/parameters/SortOrderParam"
          },
          {
            "name": "limit",
            "description": "The maximum number of members to return in the response",
            "in": "query",
            "schema": {
              "type": "integer",
              "minimum": 1,
              "maximum": 500
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Returned information about organization members",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {
                      "$ref": "#/components/schemas/OrganizationMembersResponse"
                    },
                    {
                      "$ref": "#/components/schemas/CursorPaginationResponse"
                    }
                  ]
                },
                "example": {
                  "members": [
                    {
                      "member": {
                        "id": "d57833f2-d308-4ede-9d2e-468d9d013d1b",
                        "user_id": "b107d689-6dd2-4c9a-8b9e-0b25e457cf56",
                        "org_id": "my-organization-morning-bread-81040908",
                        "role": "admin",
                        "joined_at": "2024-02-23T17:42:25Z"
                      },
                      "user": {
                        "email": "user1@email.com"
                      }
                    },
                    {
                      "member": {
                        "id": "5fee13ac-957b-40cd-8de0-4d494cc28e28",
                        "user_id": "6df052ac-ca9a-4321-8963-b6507b2d7dee",
                        "org_id": "my-organization-morning-bread-81040908",
                        "role": "member",
                        "joined_at": "2024-02-21T16:42:25Z"
                      },
                      "user": {
                        "email": "user2@email.com"
                      }
                    }
                  ],
                  "pagination": {
                    "next": "eyJtZW1iZXJfaWQiOiI1ZmVlMTNhYy05NTdiLTQwY2QtOGRlMC00ZDQ5NGNjMjhlMjgiLCJzb3J0X2J5Ijoiam9pbmVkX2F0In0=",
                    "sort_by": "joined_at",
                    "sort_order": "desc"
                  }
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
    "parameters": {
      "CursorParam": {
        "name": "cursor",
        "in": "query",
        "schema": {
          "type": "string"
        }
      },
      "SortOrderParam": {
        "name": "sort_order",
        "description": "Defines the sorting order of entities.",
        "in": "query",
        "schema": {
          "type": "string",
          "default": "desc",
          "enum": [
            "asc",
            "desc"
          ]
        }
      }
    }
  }
}`),
			want: `func (c Client) GetOrganizationMembers(orgID string, sortBy *GetOrganizationMembersSortBy, cursor *string, sortOrder *SortOrderParam, limit *uint16) (GetOrganizationMembersRespObj, error) {
var (
queryElements []string
query string
)
if sortBy != nil {
queryElements = append(queryElements, "sort_by="+fmt.Sprintf("%v", *sortBy))
}
if cursor != nil {
queryElements = append(queryElements, "cursor="+*cursor)
}
if sortOrder != nil {
queryElements = append(queryElements, "sort_order="+fmt.Sprintf("%v", *sortOrder))
}
if limit != nil {
queryElements = append(queryElements, "limit="+strconv.FormatUint(uint64(*limit), 10))
}
if len(queryElements) > 0 {
query = "?" + strings.Join(queryElements, "&")
}
var v GetOrganizationMembersRespObj
if err := c.requestHandler(c.baseURL+"/organizations/"+orgID+"/members"+query, "GET", nil, &v); err != nil {
return GetOrganizationMembersRespObj{}, err
}
return v, nil
}
`,
			wantErr:                assert.NoError,
			wantTypesRepoQueueSize: 3,
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

func Test_newArgTransformationToStrFnDefinition(t *testing.T) {
	type args struct {
		name     string
		goType   string
		nillable bool
		required bool
	}
	tests := map[string]struct {
		args args
		want string
	}{
		"bool": {
			args: args{
				name:     "includeDeleted",
				goType:   "bool",
				nillable: false,
				required: false,
			},
			want: "func(v bool) string {if v {return \"true\"}; return \"false\"}(*includeDeleted)",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equalf(t, tt.want, newArgTransformationToStrFnDefinition(tt.args.name, tt.args.goType, tt.args.nillable, tt.args.required), "newArgTransformationToStrFnDefinition(%v, %v, %v, %v)", tt.args.name, tt.args.goType, tt.args.nillable, tt.args.required)
		})
	}
}
