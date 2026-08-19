package generator

import (
	"errors"
	"math"
)

func newGoType(schema OpenAPISchema) (t string, isStruct bool, err error) {
	switch {
	case schema.Type == "" && (schema.Ref != nil || schema.AllOf != nil):
		return "", false, nil

	case schema.Type == "" && schema.AllOf == nil && schema.Ref == nil:
		return "", false, errors.New("unknown type")

	case schema.AllOf != nil || schema.Ref != nil:
		return "struct", true, nil

	case schema.Type == "object":
		if len(schema.Properties) > 0 {
			return "struct", true, nil
		}
		return "map[string]any", false, nil

	case schema.Type == "string":
		if schema.Format != nil && *schema.Format == "date-time" {
			return "time.Time", false, nil
		}
		return "string", false, nil

	case schema.Type == "integer":
		switch {
		case schema.Format != nil && *schema.Format == "int64":
			return "int64", false, nil

		case schema.Format != nil && *schema.Format == "int32":
			return "int32", false, nil

		case schema.Minimum != nil && *schema.Minimum >= 0:
			switch {
			case schema.Maximum != nil && *schema.Maximum <= 255:
				return "uint8", false, nil
			case schema.Maximum != nil && *schema.Maximum <= 65535:
				return "uint16", false, nil
			case schema.Maximum != nil && *schema.Maximum <= 4294967295:
				return "uint32", false, nil
			default:
				return "uint", false, nil
			}

		default:
			return "int", false, nil
		}

	case schema.Type == "number":
		switch {
		case schema.Format != nil && *schema.Format == "double":
			return "float64", false, nil
		case schema.Maximum != nil && *schema.Maximum <= math.MaxFloat32:
			return "float32", false, nil
		default:
			return "float", false, nil
		}

	case schema.Type == "boolean":
		return "bool", false, nil

	case schema.Type == "array":
		if schema.Items == nil {
			return "", true, errors.New("array type must have items")
		}
		itemType, structItem, err := newGoType(*schema.Items)
		if err != nil {
			return "", false, err
		}
		if structItem {
			panic("TODO0: add code generator to handle arrays of inlined structs")
		}
		return "[]" + itemType, false, nil

	default:
		return "any", false, err
	}
}
