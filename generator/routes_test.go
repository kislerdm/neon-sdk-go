package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractOrderedEndpointRoutes(t *testing.T) {
	type args struct {
		specBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "dummy case with three paths",
			args: args{
				[]byte(`{
	"paths": {
		"foo": {},
		"bar": {},
		"qux": {}
	}
}`),
			},
			want:    []string{"foo", "bar", "qux"},
			wantErr: false,
		},
		{
			name: "real case",
			args: args{openAPIFixture},
			want: []string{
				"/projects/{project_id}/applications/vercel",
				"/applications/oauth/{client_id}",
				"/api_keys",
				"/api_keys/{key_id}",
				"/projects/{project_id}/operations/{operation_id}",
				"/projects",
				"/projects/{project_id}",
				"/projects/{project_id}/operations",
				"/projects/{project_id}/saved_queries",
				"/saved_queries/{saved_query_id}",
				"/projects/{project_id}/branches",
				"/projects/{project_id}/branches/{branch_id}",
				"/projects/{project_id}/branches/{branch_id}/endpoints",
				"/projects/{project_id}/branches/{branch_id}/databases",
				"/projects/{project_id}/branches/{branch_id}/databases/{database_name}",
				"/projects/{project_id}/branches/{branch_id}/roles",
				"/projects/{project_id}/branches/{branch_id}/roles/{role_name}",
				"/projects/{project_id}/branches/{branch_id}/roles/{role_name}/reset_password",
				"/projects/{project_id}/endpoints",
				"/projects/{project_id}/endpoints/{endpoint_id}",
				"/projects/{project_id}/endpoints/{endpoint_id}/start",
				"/projects/{project_id}/endpoints/{endpoint_id}/suspend",
				"/projects/{project_id}/endpoints/{endpoint_id}/passwordless_auth",
			},
			wantErr: false,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := extractOrderedEndpointRoutes(tt.args.specBytes)
				if tt.wantErr && err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				assert.Equalf(t, tt.want, got, "extractOrderedEndpointRoutes(%v)", tt.args.specBytes)
			},
		)
	}
}
