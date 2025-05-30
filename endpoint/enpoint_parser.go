package endpoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/utils"
	"log"
	"reflect"
)

func ParseJSON(b []byte) (*Service, error) {
	// required fields
	var m []map[string]interface{}
	err := json.Unmarshal([]byte(b), &m)
	if err != nil {
		return nil, err
	}

	var service Service
	requests := make(map[string]*Endpoint)

	for i, js_request := range m {
		if _, ok := js_request["name"]; !ok {
			return nil, errors.New(fmt.Sprintf("No field <name> on request index %d\n", i))
		}

		request := Endpoint{
			Name:    js_request["name"].(string),
			Service: &service,
		}

		expected_keys := []string{"name", "fields", "tableName", "schemaName"}
		for i := range js_request {
			if !utils.Array_Contains(expected_keys, i) {
				fmt.Printf("WARN: Unexpected Key %s on Request %s\n", i, request.Name)
			}
		}

		if tableName, ok := js_request["tableName"]; ok {
			request.TableName = tableName.(string)
		}

		if schemaName, ok := js_request["schemaName"]; ok {
			request.SchemaName = schemaName.(string)
		}

		if fields, ok := js_request["fields"]; ok {
			request.Fields = parseDryeJSONFields(fields.([]interface{}), request.Name, &request)
		}

		if request.Fields == nil {
			return nil, errors.New(fmt.Sprintf("No field <fields> on request  %s\n", request.Name))
		}

		fieldList := []string{}
		for _, field := range request.Fields {
			fieldList = append(fieldList, field.Name)
		}

		request.FieldNames = fieldList

		requests[request.Name] = &request
	}

	service.Endpoints = requests

	return &service, nil
}

// check for string type,
// check for map/object type
// map [ name of field ] Dyre_Field.
// map is used for faster lookup times in large arrays
// TODO: check jsonMap to make sure field exists
func parseDryeJSONFields(a []any, req string, endpoint *Endpoint) map[string]Field {

	dyre_fields := map[string]Field{}

	for _, v := range a {

		field_type := reflect.TypeOf(v).String()

		if field_type == "string" {
			_, check := dyre_fields[v.(string)]
			if check {
				fmt.Printf("WARN: Request {%s}, Duplicate Field '%s'\n", req, v.(string))
				continue
			}

			dyre_fields[v.(string)] = Field{
				endpoint:  endpoint,
				Name:      v.(string),
				FieldType: default_type,
				// SqlType:      "NVARCHAR(MAX)",
			}
		}

		if field_type == "map[string]interface {}" {

			field_map := v.(map[string]interface{})

			new_field := Field{endpoint: endpoint}

			if name, ok := field_map["name"]; ok {
				if nameString, ok := name.(string); ok {
					new_field.Name = nameString
				} else {
					log.Printf("ERROR: Request %s, Type <name> not string: %v,\n", req, name)
					continue
				}
			} else {
				log.Printf("ERROR: Request %s, <name> not found: %v,\n", req, name)
				continue
			}

			_, check := dyre_fields[new_field.Name]
			if check {
				fmt.Printf("WARN: Request {%s}, Duplicate Field '%s' \n", req, new_field.Name)
				continue
			}

			if fieldType, ok := field_map["type"]; ok {
				if fieldTypeString, ok := fieldType.(string); ok {
					switch fieldTypeString {
					case "string":
						new_field.FieldType = object.STRING_OBJ
					case "bool":
						new_field.FieldType = object.BOOLEAN_OBJ
					case "int":
						new_field.FieldType = object.INTEGER_OBJ
					case "float":
						new_field.FieldType = object.FLOAT_OBJ
					case "date":
						new_field.FieldType = object.DATE_OBJ
					case "datetime":
						new_field.FieldType = object.DATETIME_OBJ
					}
				} else {
					log.Printf("ERROR: Request %s, <type> not string on field %s\n", req, new_field.Name)
					new_field.FieldType = default_type
				}
			} else {
				new_field.FieldType = default_type
			}

			// if typeName, ok := field_map["type"]; ok {
			// 	if typeNameString, ok := typeName.(string); ok {
			// 		new_field.TypeName = typeNameString
			// 	} else {
			// 		log.Printf("ERROR: Request %s, Type <typeName> not string on field %s\n", req, new_field.Name)
			// 		new_field.TypeName = DefaultType
			// 	}
			// } else {
			// 	new_field.TypeName = DefaultType
			// }

			// if querySelect, ok := field_map["sqlSelect"]; ok {
			// 	if querySelectString, ok := querySelect.(string); ok {
			// 		new_field.SqlSelect = querySelectString
			// 	} else {
			// 		log.Printf("ERROR: Request %s, Type <querySelect> not string on field %s\n", req, new_field.Name)
			// 		new_field.SqlSelect = new_field.Name
			// 	}
			// } else {
			// 	new_field.SqlSelect = new_field.Name
			// }

			expected_keys := []string{"name", "defaultField"}
			for i := range field_map {
				if !utils.Array_Contains(expected_keys, i) {
					fmt.Printf("WARN: Request %s, Unexpected Key %s on Field %s\n", req, i, new_field.Name)
				}
			}

			dyre_fields[new_field.Name] = new_field
		}
	}

	return dyre_fields
}
