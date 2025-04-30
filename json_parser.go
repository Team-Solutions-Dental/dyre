package dyre

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
)

// TODO: Add a check for unkown fields in the json file
// check for unknow types in type fields
func Init(path string) (map[string]DyRe_Request, error) {
	m, err := readDyreJSON(path)
	if err != nil {
		return nil, err
	}

	re, err := parseDyreJSON(m)
	if err != nil {
		return nil, err
	}

	return re, nil
}

func readDyreJSON(path string) ([]map[string]interface{}, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	err = json.Unmarshal([]byte(byteValue), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseDyreJSON(m []map[string]interface{}) (map[string]DyRe_Request, error) {
	// required fields
	requests := make(map[string]DyRe_Request)
	for i, js_request := range m {
		if _, ok := js_request["name"]; !ok {
			return nil, errors.New(fmt.Sprintf("No field <name> on request index %d\n", i))
		}

		dy_request := DyRe_Request{
			name: js_request["name"].(string),
		}

		expected_keys := []string{"name", "fields", "tableName"}
		for i := range js_request {
			if !contains(expected_keys, i) {
				fmt.Printf("WARN: Unexpected Key %s on Request %s\n", i, dy_request.name)
			}
		}

		if fields, ok := js_request["fields"]; ok {
			dy_request.fields = parseDryeJSONFields(fields.([]interface{}), dy_request.name)
		}

		if dy_request.fields == nil {
			return nil, errors.New(fmt.Sprintf("No field <fields> on request  %s\n", dy_request.name))
		}

		if tableName, ok := js_request["tableName"]; ok {
			dy_request.sql = DyRe_SQL{tableName: tableName.(string)}
		}

		fieldList := []string{}
		for _, field := range dy_request.fields {
			fieldList = append(fieldList, field.name)
		}

		dy_request.fieldNames = fieldList

		requests[dy_request.name] = dy_request
	}

	return requests, nil
}

// check for string type,
// check for map/object type
// map [ name of field ] Dyre_Field.
// map is used for faster lookup times in large arrays
// TODO: check jsonMap to make sure field exists
func parseDryeJSONFields(a []any, req string) map[string]DyRe_Field {

	dyre_fields := map[string]DyRe_Field{}

	for _, v := range a {

		field_type := reflect.TypeOf(v).String()

		if field_type == "string" {
			_, check := dyre_fields[v.(string)]
			if check {
				fmt.Printf("WARN: Request {%s}, Duplicate Field '%s'\n", req, v.(string))
				continue
			}

			dyre_fields[v.(string)] = DyRe_Field{
				name:      v.(string),
				required:  false,
				sqlSelect: v.(string),
			}
		}

		if field_type == "map[string]interface {}" {

			field_map := v.(map[string]interface{})

			new_field := DyRe_Field{}

			if name, ok := field_map["name"]; ok {
				if nameString, ok := name.(string); ok {
					new_field.name = nameString
				} else {
					log.Printf("ERROR: Request %s, Type <name> not string: %v,\n", req, name)
					continue
				}
			} else {
				log.Printf("ERROR: Request %s, <name> not found: %v,\n", req, name)
				continue
			}

			_, check := dyre_fields[new_field.name]
			if check {
				fmt.Printf("WARN: Request {%s}, Duplicate Field '%s' \n", req, new_field.name)
				continue
			}

			if required, ok := field_map["required"]; ok {
				if requiredBool, ok := required.(bool); ok {
					new_field.required = requiredBool
				} else {
					log.Printf("ERROR: Request %s, Type <required> not bool on field %s\n", req, new_field.name)
					new_field.required = false
				}
			} else {
				new_field.required = false
			}

			if typeName, ok := field_map["type"]; ok {
				if typeNameString, ok := typeName.(string); ok {
					new_field.typeName = typeNameString
				} else {
					log.Printf("ERROR: Request %s, Type <typeName> not string on field %s\n", req, new_field.name)
					new_field.typeName = DefaultType
				}
			} else {
				new_field.typeName = DefaultType
			}

			if querySelect, ok := field_map["sqlSelect"]; ok {
				if querySelectString, ok := querySelect.(string); ok {
					new_field.sqlSelect = querySelectString
				} else {
					log.Printf("ERROR: Request %s, Type <querySelect> not string on field %s\n", req, new_field.name)
					new_field.sqlSelect = new_field.name
				}
			} else {
				new_field.sqlSelect = new_field.name
			}

			expected_keys := []string{"name", "required", "type", "sqlSelect"}
			for i := range field_map {
				if !contains(expected_keys, i) {
					fmt.Printf("WARN: Request %s, Unexpected Key %s on Field %s\n", req, i, new_field.name)
				}
			}

			dyre_fields[new_field.name] = new_field
		}
	}

	return dyre_fields
}

func getMapKeys(m map[string]any) []string {
	count := len(m)
	keys := make([]string, count)
	i := 0
	for key := range m {
		keys[i] = key
		i += 1
	}
	return keys
}
