package generator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
