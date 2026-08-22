package generator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newGoTypeDefinition(t *testing.T) {
	type args struct {
		typeName string
		schema   OpenAPISchema
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "named string",
			args: args{
				typeName: "Foo",
				schema: OpenAPISchema{
					Type: "string",
				},
			},
			want:    "type Foo string",
			wantErr: assert.NoError,
		},
		{
			name: "named float",
			args: args{
				typeName: "Foo",
				schema: OpenAPISchema{
					Type: "number",
				},
			},
			want:    "type Foo float",
			wantErr: assert.NoError,
		},
		{
			name: "named map",
			args: args{
				typeName: "Foo",
				schema: OpenAPISchema{
					Type: "object",
				},
			},
			want:    "type Foo map[string]any",
			wantErr: assert.NoError,
		},
		{
			name: "description is present",
			args: args{
				typeName: "Foo",
				schema: OpenAPISchema{
					Description: "foo\n\nbar",
					Type:        "object",
				},
			},
			want: `// Foo foo

bar
type Foo map[string]any`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newGoTypeDefinition(tt.args.typeName, tt.args.schema)
			if !tt.wantErr(t, err, fmt.Sprintf("newGoTypeDefinition(%v, %v)", tt.args.typeName, tt.args.schema)) {
				return
			}
			assert.Equalf(t, tt.want, got, "newGoTypeDefinition(%v, %v)", tt.args.typeName, tt.args.schema)
		})
	}
}

func Test_newGoStructDefinition(t *testing.T) {
	type args struct {
		schema OpenAPISchema
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "object with two attributes, one required",
			args: args{
				schema: OpenAPISchema{
					Type:     "object",
					Required: []string{"id"},
					Properties: []OpenAPISchema{
						{
							xRefName:    "id",
							Type:        "string",
							Description: "foo",
						},
						{
							xRefName:    "bar_url",
							Type:        "string",
							Description: "bar",
						},
					},
				},
			},
			want: `struct {
// ID foo
ID string ` + "`json:\"id\"`" +
				"\n// BarURL bar\nBarURL *string `json:\"bar_url,omitempty\"`" +
				"\n}",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newGoStructDefinition(tt.args.schema)
			if !tt.wantErr(t, err, fmt.Sprintf("newGoStructDefinition(%v)", tt.args.schema)) {
				return
			}
			assert.Equalf(t, tt.want, got, "newGoStructDefinition(%v)", tt.args.schema)
		})
	}
}
