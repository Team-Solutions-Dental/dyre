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

		expected_keys := []string{"name", "fields", "groups", "tableName"}
		for i := range js_request {
			if !contains(expected_keys, i) {
				fmt.Printf("WARN: Unexpected Key %s on Request %s\n", i, dy_request.name)
			}
		}

		if fields, ok := js_request["fields"]; ok {
			dy_request.fields = parseDryeJSONFields(fields.([]interface{}), dy_request.name)
		}

		if groups, ok := js_request["groups"]; ok {
			if groupsMap, ok := groups.([]interface{}); ok {
				dy_request.groups = parseDryeJSONGroups(groupsMap, dy_request.name)
			} else {
				log.Printf("Field <groups> should be an array of objects on request  %s\nGivenType: %s", dy_request.name, reflect.TypeOf(groupsMap).String())
			}

		}

		if dy_request.fields == nil && dy_request.groups == nil {
			return nil, errors.New(fmt.Sprintf("No field <fields> on request  %s\n", dy_request.name))
		}

		if tableName, ok := js_request["tableName"]; ok {
			dy_request.sql = DyRe_SQL{tableName: tableName.(string)}
		}

		fieldList := []string{}
		for _, field := range dy_request.fields {
			fieldList = append(fieldList, field.name)
		}
		groupList := []string{}
		for _, group := range dy_request.groups {
			groupList = append(groupList, group.name)
		}

		groupFieldList := []string{}
		for _, group := range dy_request.groups {
			for _, field := range group.fields {
				groupFieldList = append(groupFieldList, field.name)
			}
		}

		dy_request.fieldNames = fieldList
		dy_request.groupNames = groupList
		dy_request.groupFieldNames = groupFieldList

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

// look for defined group.
// infer infer group from json map if only name is given.
// not sure I want to "fill in the blank" on the map type
// If someone doesnt put all of the fields in on the map type is should not infer. This could lead to confusion otherwise
//
//	{
//	  "name": "group1",
//	  "fields": [
//	    "field5",
//	    "field6"
//	  ]
//	},
func parseDryeJSONGroups(a []interface{}, req string) map[string]DyRe_Group {
	dyre_groups := map[string]DyRe_Group{}
	for _, v := range a {
		if _, ok := v.(map[string]interface{}); !ok {
			log.Printf("A dyre group must be a Object:\n%v", v)
			continue
		}
		group := v.(map[string]interface{})
		new_group := DyRe_Group{}

		if name, ok := group["name"]; ok {
			if nameString, ok := name.(string); ok {
				new_group.name = nameString
			} else {
				log.Printf("ERROR: Request %s, Type group <name> not string: %v,\n", req, name)
				continue
			}
		} else {
			log.Printf("ERROR: Request %s, Field <name> does not exit on group\n", req)
			continue
		}

		_, check := dyre_groups[new_group.name]
		if check {
			fmt.Printf("WARN: Request {%s}, Duplicate group '%s' \n", req, new_group.name)
			continue
		}

		if group_required, ok := group["required"]; ok {
			new_group.required = group_required.(bool)
		} else {
			new_group.required = false
		}

		if group_fields, ok := group["fields"]; ok {
			new_group.fields = parseDryeJSONFields(group_fields.([]interface{}), req)
		} else {
			continue
		}

		expected_keys := []string{"name", "fields", "required"}
		for i := range group {
			if !contains(expected_keys, i) {
				fmt.Printf("WARN: Request %s, Unexpected Key %s on Group %s\n", req, i, new_group.name)
			}
		}

		dyre_groups[new_group.name] = new_group
	}

	return dyre_groups
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
