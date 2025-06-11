package endpoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/utils"
)

func ParseJSON(b []byte) (*Service, error) {
	// required fields
	var m []map[string]any
	err := json.Unmarshal([]byte(b), &m)
	if err != nil {
		return nil, err
	}

	var service Service
	endpoints := make(map[string]*Endpoint)

	var errs []error

	for i, ep := range m {
		newEndpoint, err := parseEndpoint(ep, &service, i)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		endpoints[newEndpoint.Name] = newEndpoint
	}

	for _, ep := range endpoints {
		if len(ep.Joins) > 0 {
			for _, join := range ep.Joins {
				childEndpoint, ok := endpoints[join.childEndpointName]
				if !ok {
					errs = append(errs,
						fmt.Errorf("Endpoint '%s', Join '%s'. join.childEndpointName is not found in other endpoints",
							ep.Name,
							join.childEndpointName))
					continue
				}

				join.childEndpoint = childEndpoint
				ep.Joins[join.childEndpointName] = join

			}
		}
	}

	service.Endpoints = endpoints

	return &service, errors.Join(errs...)
}

func parseEndpoint(m map[string]any, s *Service, index int) (*Endpoint, error) {
	var errs []error
	var err error

	if _, ok := m["name"]; !ok {
		return nil, fmt.Errorf("Endpoint at index %d has no name.", index)
	}

	tableName, ok := m["tableName"]
	if !ok {
		return nil, fmt.Errorf("Endpoint %s at index %d has no tableName.", m["name"].(string), index)
	}

	request := Endpoint{
		Name:      m["name"].(string),
		TableName: tableName.(string),
		Service:   s,
	}

	if schemaName, ok := m["schemaName"]; ok {
		request.SchemaName = schemaName.(string)
	}

	if fields, ok := m["fields"]; ok {
		request.Fields, err = parseEndpointFields(fields.([]any), &request)
		if err != nil {
			errs = append(errs, fmt.Errorf("Fields: %w", err))
		}
	}

	if joins, ok := m["joins"]; ok {
		request.Joins, err = parseEndpointJoins(joins.([]any), &request)
		if err != nil {
			errs = append(errs, fmt.Errorf("Joins: %w", err))
		}
	}

	if request.Fields == nil {
		return nil, errors.New(fmt.Sprintf("No field <fields> on request  %s\n", request.Name))
	}

	fieldList := []string{}
	for _, field := range request.Fields {
		fieldList = append(fieldList, field.Name)
	}

	request.FieldNames = fieldList

	expected_keys := []string{"name", "fields", "tableName", "schemaName", "joins"}
	for i := range m {
		if !utils.Array_Contains(expected_keys, i) {
			errs = append(errs, fmt.Errorf("Unexpected key %s", i))
		}
	}

	if errs != nil {
		err = fmt.Errorf("Endpoint %s at index %d, %w", request.Name, index, errors.Join(errs...))
		return &request, err
	}

	return &request, nil
}

func parseEndpointFields(a []any, endpoint *Endpoint) (map[string]Field, error) {
	fields := map[string]Field{}
	var errs []error

	for _, jsonField := range a {
		field, err := parseField(jsonField, endpoint)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		_, check := fields[field.Name]
		if check {
			errs = append(errs, fmt.Errorf("Duplicate field %s", field.Name))
			continue
		}

		fields[field.Name] = field
	}

	return fields, errors.Join(errs...)
}

func parseField(f any, e *Endpoint) (Field, error) {

	newField := Field{
		endpoint:  e,
		FieldType: default_type,
		Nullable:  true,
	}

	switch f := f.(type) {
	case string:
		newField.Name = f
		return newField, nil
	case map[string]any:
		var err error
		newField.Name, err = parseString(f, "name")
		if err != nil {
			return newField, err
		}

		var errs []error

		errs = append(errs, err)
		newField.FieldType, err = parseFieldType(f, default_type)
		errs = append(errs, err)

		newField.Nullable, err = parseNullable(f, true)
		errs = append(errs, err)

		expected_keys := []string{"name", "type", "nullable"}
		for i := range f {
			if !utils.Array_Contains(expected_keys, i) {
				errs = append(errs,
					fmt.Errorf("Unexpected key %s", i),
				)
			}
		}

		err = errors.Join(errs...)

		if err != nil {
			err = fmt.Errorf("Field %s, %w", newField.Name, err)
			return newField, err
		}

		return newField, nil

	default:
		return newField, fmt.Errorf("Field JSON type invalid. got = %T", f)
	}

}

func parseFieldType(field_map map[string]any, default_type object.ObjectType) (object.ObjectType, error) {
	fieldType, ok := field_map["type"]
	if !ok {
		return default_type, nil
	}
	str, ok := fieldType.(string)
	if !ok {
		return default_type, errors.New(fmt.Sprintf("'type' not string. got=%T", fieldType))
	}
	convertedType, err := typeConvert(str)
	if err != nil {
		return default_type, err
	}
	return convertedType, nil
}

func parseNullable(field_map map[string]any, def bool) (bool, error) {
	fieldType, ok := field_map["nullable"]
	if !ok {
		return def, nil
	}
	bool, ok := fieldType.(bool)
	if !ok {
		return def, errors.New(fmt.Sprintf("'nullable' not boolean. got=%T", fieldType))
	}

	return bool, nil
}

func parseEndpointJoins(a []any, e *Endpoint) (map[string]Join, error) {

	joins := map[string]Join{}
	var errs []error

	for _, join := range a {
		newJoin, err := parseJoin(join, e)

		if err != nil {
			errs = append(errs, err)
			continue
		}

		_, check := joins[newJoin.childEndpointName]
		if check {
			errs = append(errs, fmt.Errorf("Duplicate join %s", newJoin.childEndpointName))
			continue
		}

		joins[newJoin.childEndpointName] = newJoin
	}

	return joins, errors.Join(errs...)
}

func parseJoin(a any, e *Endpoint) (Join, error) {
	newJoin := Join{parentEndpoint: e}

	m, ok := a.(map[string]any)
	if !ok {
		return newJoin, fmt.Errorf("Invalid 'join' JSON type %T", a)
	}

	var err error
	newJoin.childEndpointName, err = parseString(m, "endpoint")
	if err != nil {
		return newJoin, err
	}

	var errs []error
	ons, err := parseJoinOn(m)
	errs = append(errs, err)
	newJoin.Parent_ON = ons[0]
	newJoin.Child_ON = ons[1]

	expected_keys := []string{"endpoint", "on"}
	for i := range m {
		if !utils.Array_Contains(expected_keys, i) {
			errs = append(errs, fmt.Errorf("Unexpected key %s", i))
		}
	}

	err = errors.Join(errs...)
	if err != nil {
		err = fmt.Errorf("Join %s, %w", newJoin.childEndpointName, err)
	}

	return newJoin, err
}

func parseJoinOn(m map[string]any) ([2]string, error) {
	var output [2]string
	o, ok := m["on"]
	if !ok {
		return output, errors.New("Missing 'on'")
	}

	switch o := o.(type) {
	case string:
		output[0] = o
		output[1] = o
	case []any:
		if len(o) != 2 {
			return output, fmt.Errorf("On array length is not two. [ParentON, ChildOn]. got %d", len(o))
		}
		var ok bool
		if output[0], ok = o[0].(string); !ok {
			return output, fmt.Errorf("On array[%d] not string. got %T", 0, o[0])
		}
		if output[1], ok = o[1].(string); !ok {
			return output, fmt.Errorf("On array[%d] not string. got %T", 1, o[1])
		}
	default:
		return output, fmt.Errorf("Invalid 'on' JSON type %T", o)
	}
	return output, nil
}

func parseString(m map[string]any, index string) (string, error) {
	n, ok := m[index]
	if !ok {
		return "", fmt.Errorf("Missing %s", index)
	}
	str, ok := n.(string)
	if !ok {
		return "", fmt.Errorf("'%s' not string. got=%T", index, n)
	}

	return str, nil
}

func typeConvert(input string) (object.ObjectType, error) {
	var output object.ObjectType
	compare := strings.ToUpper(input)

	switch compare {
	case "STRING":
		output = object.STRING_OBJ
	case "BOOL":
		output = object.BOOLEAN_OBJ
	case "BOOLEAN":
		output = object.BOOLEAN_OBJ
	case "INT":
		output = object.INTEGER_OBJ
	case "INTEGER":
		output = object.INTEGER_OBJ
	case "FLOAT":
		output = object.FLOAT_OBJ
	case "DATE":
		output = object.DATE_OBJ
	case "DATETIME":
		output = object.DATETIME_OBJ
	default:
		return output, errors.New("Unknown Type " + input)
	}

	return output, nil
}
