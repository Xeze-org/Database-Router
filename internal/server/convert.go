// Package server contains gRPC server implementations that translate between
// protobuf messages and the domain service layer.
package server

import (
	"fmt"

	"db-router/internal/service"

	"google.golang.org/protobuf/types/known/structpb"
)

// rowToProtoFields converts a service.Row to a proto map[string]*structpb.Value.
func rowToProtoFields(row service.Row) map[string]*structpb.Value {
	m := make(map[string]*structpb.Value, len(row))
	for k, v := range row {
		m[k] = toProtoValue(v)
	}
	return m
}

// protoFieldsToRow converts a proto map[string]*structpb.Value back to a service.Row.
func protoFieldsToRow(fields map[string]*structpb.Value) service.Row {
	row := make(service.Row, len(fields))
	for k, v := range fields {
		row[k] = fromProtoValue(v)
	}
	return row
}

func toProtoValue(v interface{}) *structpb.Value {
	switch val := v.(type) {
	case nil:
		return structpb.NewNullValue()
	case bool:
		return structpb.NewBoolValue(val)
	case float64:
		return structpb.NewNumberValue(val)
	case float32:
		return structpb.NewNumberValue(float64(val))
	case int:
		return structpb.NewNumberValue(float64(val))
	case int32:
		return structpb.NewNumberValue(float64(val))
	case int64:
		return structpb.NewNumberValue(float64(val))
	case string:
		return structpb.NewStringValue(val)
	case []byte:
		return structpb.NewStringValue(string(val))
	default:
		return structpb.NewStringValue(fmt.Sprintf("%v", val))
	}
}

func fromProtoValue(v *structpb.Value) interface{} {
	if v == nil {
		return nil
	}
	switch k := v.GetKind().(type) {
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_NumberValue:
		return k.NumberValue
	case *structpb.Value_StringValue:
		return k.StringValue
	case *structpb.Value_BoolValue:
		return k.BoolValue
	case *structpb.Value_StructValue:
		row := make(service.Row)
		for key, val := range k.StructValue.GetFields() {
			row[key] = fromProtoValue(val)
		}
		return row
	case *structpb.Value_ListValue:
		list := make([]interface{}, len(k.ListValue.GetValues()))
		for i, val := range k.ListValue.GetValues() {
			list[i] = fromProtoValue(val)
		}
		return list
	default:
		return nil
	}
}
