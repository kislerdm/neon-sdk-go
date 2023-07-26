package generator

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// extractOrderedEndpointRoutes extracts routes of the endpoint following the order of the OpenAPI spec.
func extractOrderedEndpointRoutes(specBytes []byte) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(specBytes))
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t.(type) {
		case string:
			if t.(string) == "paths" {
				return objectKeys(dec)
			}
		default:
			continue
		}
	}
	return nil, errors.New("no paths found")
}

func objectKeys(dec *json.Decoder) ([]string, error) {
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if t != json.Delim('{') {
		return nil, errors.New("expected start of object")
	}

	var keys []string
	for {
		t, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if t == json.Delim('}') {
			return keys, nil
		}
		keys = append(keys, t.(string))
		if err := skipValue(dec); err != nil {
			return nil, err
		}
	}
}

var end = errors.New("invalid end of array or object")

func skipValue(d *json.Decoder) error {
	t, err := d.Token()
	if err != nil {
		return err
	}

	switch t {
	case json.Delim('['), json.Delim('{'):
		for {
			if err := skipValue(d); err != nil {
				if err == end {
					break
				}
				return err
			}
		}
	case json.Delim(']'), json.Delim('}'):
		return end
	}

	return nil
}
