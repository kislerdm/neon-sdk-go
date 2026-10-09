package sdk

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func Test_convertErrorResponse(t *testing.T) {
	tests := map[string]struct {
		res  *http.Response
		want Error
	}{
		"default": {
			res: &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader(`{"code":"foo", "message":"bar", "request_id":"test"}`)),
			},
			want: Error{
				HTTPCode:  404,
				RequestID: "test",
				Details: []ErrorDetails{
					{
						Code:    "foo",
						Message: "bar",
					},
				},
			},
		},
		"AcceptProjectTransferRequestSatisfiesPlanError": {
			res: &http.Response{
				StatusCode: 406,
				Body: io.NopCloser(strings.NewReader(
					`{"reasons":[{"code":"foo", "message":"bar"},{"code":"baz", "message":"qux"}]}`)),
			},
			want: Error{
				HTTPCode: 406,
				Details: []ErrorDetails{
					{
						Code:    "foo",
						Message: "bar",
					},
					{
						Code:    "baz",
						Message: "qux",
					},
				},
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := convertErrorResponse(tt.res)
			if !errors.As(err, &Error{}) {
				t.Error("wrong error type")
				return
			}
			if !reflect.DeepEqual(tt.want, err.(Error)) {
				t.Error("unexpected error content")
				return
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	type fields struct {
		HTTPCode  int
		RequestID string
		Details   []ErrorDetails
	}
	tests := map[string]struct {
		fields fields
		want   string
	}{
		"default": {
			fields: fields{
				HTTPCode:  404,
				RequestID: "test",
				Details:   []ErrorDetails{{Code: "foo", Message: "bar"}},
			},
			want: "[HTTP Code: 404][Error Code: foo][Request ID: test] bar",
		},
		"AcceptProjectTransferRequestSatisfiesPlanError": {
			fields: fields{
				HTTPCode:  406,
				RequestID: "test",
				Details: []ErrorDetails{
					{
						Code:    "foo",
						Message: "bar",
					},
					{
						Code:    "baz",
						Message: "qux",
					},
				},
			},
			want: "[HTTP Code: 406][Error Code: foo][Request ID: test] bar\n" +
				"[HTTP Code: 406][Error Code: baz][Request ID: test] qux",
		},
		"default-no-requestID": {
			fields: fields{
				HTTPCode: 404,
				Details:  []ErrorDetails{{Code: "foo", Message: "bar"}},
			},
			want: "[HTTP Code: 404][Error Code: foo] bar",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := Error{
				HTTPCode:  tt.fields.HTTPCode,
				RequestID: tt.fields.RequestID,
				Details:   tt.fields.Details,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
