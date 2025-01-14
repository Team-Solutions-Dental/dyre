package dyre

import (
	"encoding/json"
	"fmt"
	"testing"
)

// 󰆧 (*DyRe_Request).ValidateRequest
// 󰆧 (*DyRe_Request).FieldNames
// 󰆧 (*DyRe_Request).GroupNames
// 󰆧 (*DyRe_Validated).SQLFields
// 󰆧 (*DyRe_Validated).GenerateArray
// 󰆧 (*DyRe_Validated).Headers
// 󰆧 (*DyRe_Request).TableName
// 󰆧 (*DyRe_Request).SQLFile

func TestStandardRequest(t *testing.T) {

	const standardRequstString string = `
	[
		{
			"name": "StandardRequest",
			"tableName": "Table",
			"fields": [
				{
					"name": "field1",
					"required": true
				},
				{
					"name": "field2",
					"required": true,
					"sqlSelect": "fieldX AS field2"
				},
				"field3"
			],
			"groups": [
				{
					"name": "group1",
					"fields": [
						{
							"name": "fieldg1-1",
							"sqlSelect": "field11 AS fieldg1-1"
						},
						{
							"name": "fieldg1-2"
						},
						"fieldg1-3"
					]
				},
				{
					"name": "group2",
					"fields": [
						"fieldg2-1"
					]
				},
				{
					"name": "group3",
					"fields": [
						"fieldg3-1"
					]
				},
				{
					"name": "group4",
					"fields": [
						"fieldg4-1"
					]
				}
			]
		}
	]
	`

	re, err := dyre_test_parse(standardRequstString, "StandardRequest")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("CheckFields", func(t *testing.T) {
		want := []string{"field1", "field2", "field3"}
		got := re.FieldNames()
		errors := deepEqualStringArray(want, got)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Error(err)
			}
		}
	})
	t.Run("CheckGroups", func(t *testing.T) {
		want := []string{"group1", "group2", "group3", "group4"}
		got := re.GroupNames()
		errors := deepEqualStringArray(want, got)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Error(err)
			}
		}
	})

	t.Run("DyRe_Request.ValidateRequest() Partial", func(t *testing.T) {
		// has real field, has fake field, missing field
		given_fields := []string{"field1", "field100", "fieldg4-1"}
		given_groups := []string{"group1", "group100"}

		valid, valid_err := re.ValidateRequest(given_fields, given_groups)
		if valid_err != nil {
			t.Error(valid_err)
		}

		valid_field_names := []string{}
		for _, f := range valid._fields {
			valid_field_names = append(valid_field_names, f.name)
		}

		valid_group_names := []string{}
		for _, g := range valid._groups {
			valid_group_names = append(valid_group_names, g.name)
		}

		expected_fields_names := []string{"field1", "field2", "fieldg4-1"}
		expected_groups_names := []string{"group1"}

		field_errors := deepEqualStringArray(expected_fields_names, valid_field_names)
		if len(field_errors) > 0 {
			for _, err := range field_errors {
				t.Error(err)
			}
		}

		group_errors := deepEqualStringArray(expected_groups_names, valid_group_names)
		if len(group_errors) > 0 {
			for _, err := range group_errors {
				t.Error(err)
			}
		}

		expected_sql_field := []string{"field1", "fieldX AS field2", "field11 AS fieldg1-1", "fieldg1-2", "fieldg1-3", "fieldg4-1"}
		sql_errors := deepEqualStringArray(expected_sql_field, valid._sqlFields)
		if len(sql_errors) > 0 {
			for _, err := range sql_errors {
				t.Error(err)
			}
		}

	})

	t.Run("DyRe_Request.ValidateRequest() All", func(t *testing.T) {
		// has real field, has fake field, missing field
		given_fields := []string{"field1", "field2", "field3"}
		given_groups := []string{"group1", "group2"}

		valid, valid_err := re.ValidateRequest(given_fields, given_groups)
		if valid_err != nil {
			t.Error(valid_err)
		}

		valid_field_names := []string{}
		for _, f := range valid._fields {
			valid_field_names = append(valid_field_names, f.name)
		}

		valid_group_names := []string{}
		for _, g := range valid._groups {
			valid_group_names = append(valid_group_names, g.name)
		}

		expected_fields_names := []string{"field1", "field2", "field3"}
		expected_groups_names := []string{"group1", "group2"}

		field_errors := deepEqualStringArray(expected_fields_names, valid_field_names)
		if len(field_errors) > 0 {
			for _, err := range field_errors {
				t.Error(err)
			}
		}

		group_errors := deepEqualStringArray(expected_groups_names, valid_group_names)
		if len(group_errors) > 0 {
			for _, err := range group_errors {
				t.Error(err)
			}
		}

		expected_sql_field := []string{"field1", "fieldX AS field2", "field3", "field11 AS fieldg1-1", "fieldg1-2", "fieldg1-3", "fieldg2-1"}

		sql_errors := deepEqualStringArray(expected_sql_field, valid._sqlFields)
		if len(sql_errors) > 0 {
			for _, err := range sql_errors {
				t.Error(err)
			}
		}
	})

	t.Run("CheckValidationNone", func(t *testing.T) {
		empty := []string{}
		_, err := re.ValidateRequest(empty, empty)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestNoFields(t *testing.T) {
	const noFieldsString string = `
	[{	"name": "NoFields",
			"tableName": "Table",
			"groups": [
				{
					"name": "group1",
					"fields": [
						"field2"
					]
				},
				{
					"name": "group2",
					"fields": [
						"field3"
					]
				}
			]
		}
	]`

	re, err := dyre_test_parse(noFieldsString, "NoFields")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("CheckFields", func(t *testing.T) {
		want := []string{}
		got := re.FieldNames()
		errors := deepEqualStringArray(want, got)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Error(err)
			}
		}
	})
	t.Run("CheckGroups", func(t *testing.T) {
		want := []string{"group1", "group2"}
		got := re.GroupNames()
		errors := deepEqualStringArray(want, got)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Error(err)
			}
		}
	})
}

func TestNoGroups(t *testing.T) {
	const noGroupsString string = `
  [{ "name": "NoGroups",
    "fields": [
      "field1"
    ]
  }]`

	re, err := dyre_test_parse(noGroupsString, "NoGroups")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("CheckFields", func(t *testing.T) {
		want := []string{"field1"}
		got := re.FieldNames()
		errors := deepEqualStringArray(want, got)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Error(err)
			}
		}
	})
	t.Run("CheckGroups", func(t *testing.T) {
		want := []string{}
		got := re.GroupNames()
		errors := deepEqualStringArray(want, got)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Error(err)
			}
		}
	})
}

// func TestTypeSub(t *testing.T) {
// 	const typesString string = `
//   [{"name": "TypeRequest",
//     "tableName": "Table",
//     "fields": [
//       {
//         "name": "field1",
//         "type": "sql.NullInt64",
//         "required": true
//       },
//       {
//         "name": "field2",
//         "type": "sql.NullBool",
//         "sqlSelect": "fieldX AS field2"
//       },
//       "field3"
//     ],
//     "groups": [
//       {
//         "name": "group1",
//         "fields": [
//           {
//             "name": "fieldg1-1",
//             "type": "sql.NullFloat64",
//             "required": true,
//             "sqlSelect": "field11 AS fieldg1-1"
//           },
//           {
//             "name": "fieldg1-2",
//             "type": "sql.NullByte"
//           },
//           {
//             "name": "fieldg1-3"
//           },
//           "fieldg1-4"
//         ]
//       }
//     ]
//   }]`
//
// 	re, err := dyre_test_parse(typesString, "TypeRequest")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Run("DyRe_Request.ValidateRequest() All", func(t *testing.T) {
// 		// has real field, has fake field, missing field
// 		given_fields := []string{"field1", "field2", "field3"}
// 		given_groups := []string{"group1"}
//
// 		valid, valid_err := re.ValidateRequest(given_fields, given_groups)
// 		if valid_err != nil {
// 			t.Error(valid_err)
// 		}
//
// 		valid_field_names := []string{}
// 		for _, f := range valid._fields {
// 			valid_field_names = append(valid_field_names, f.name)
// 		}
//
// 		valid_group_names := []string{}
// 		for _, g := range valid._groups {
// 			valid_group_names = append(valid_group_names, g.name)
// 		}
//
// 		t.Run("FieldNames", func(t *testing.T) {
// 			expected_fields_names := []string{"field1", "field2", "field3"}
// 			field_errors := deepEqualStringArray(expected_fields_names, valid_field_names)
// 			if len(field_errors) > 0 {
// 				for _, err := range field_errors {
// 					t.Error(err)
// 				}
// 			}
// 		})
//
// 		t.Run("DyRe_Request.Fields()", func(t *testing.T) {
// 			expected_fields_names := []string{"field1", "field2", "field3"}
// 			field_info := re.Fields()
// 			field_info_names := []string{}
// 			for _, v := range field_info {
// 				field_info_names = append(field_info_names, v.name)
// 			}
// 			field_errors := deepEqualStringArray(expected_fields_names, field_info_names)
// 			if len(field_errors) > 0 {
// 				for _, err := range field_errors {
// 					t.Error(err)
// 				}
// 			}
// 		})
//
// 		t.Run("GroupNames", func(t *testing.T) {
// 			expected_groups_names := []string{"group1"}
// 			group_errors := deepEqualStringArray(expected_groups_names, valid_group_names)
// 			if len(group_errors) > 0 {
// 				for _, err := range group_errors {
// 					t.Error(err)
// 				}
// 			}
// 		})
//
// 		t.Run("DyRe_Request.Groups()", func(t *testing.T) {
// 			expected_groups_names := []string{"group1"}
// 			group_info := re.Groups()
// 			group_info_names := []string{}
// 			for _, v := range group_info {
// 				group_info_names = append(group_info_names, v.name)
// 			}
// 			group_errors := deepEqualStringArray(expected_groups_names, group_info_names)
// 			if len(group_errors) > 0 {
// 				for _, err := range group_errors {
// 					t.Error(err)
// 				}
// 			}
// 		})
//
// 		t.Run("sqlFields", func(t *testing.T) {
//
// 			expected_sql_field := []string{"field1", "fieldX AS field2", "field3", "field11 AS fieldg1-1", "fieldg1-2", "fieldg1-3", "fieldg1-4"}
//
// 			sql_errors := deepEqualStringArray(expected_sql_field, valid._sqlFields)
// 			if len(sql_errors) > 0 {
// 				for _, err := range sql_errors {
// 					t.Error(err)
// 				}
// 			}
// 		})
//
// 		t.Run("headers", func(t *testing.T) {
//
// 			expected := []string{"field1", "field2", "field3", "fieldg1-1", "fieldg1-2", "fieldg1-3", "fieldg1-4"}
//
// 			errors := deepEqualStringArray(expected, valid._headers)
// 			if len(errors) > 0 {
// 				for _, err := range errors {
// 					t.Error(err)
// 				}
// 			}
// 		})
//
// 	})
//
// }

func dyre_test_parse(data string, name string) (DyRe_Request, error) {
	var dyre_test_map []map[string]interface{}
	err := json.Unmarshal([]byte(data), &dyre_test_map)
	if err != nil {
		return DyRe_Request{}, fmt.Errorf("Failed unmarshal for: %s\nError: %v\n", name, err)
	}

	dyre_requests, err := parseDyreJSON(dyre_test_map)
	if err != nil {
		return DyRe_Request{}, fmt.Errorf("Failed to parse for: %s\nError: %v\n", name, err)
	}

	fmt.Printf("RequestName: %v\n", name)

	re, ok := dyre_requests[name]
	if !ok {
		return DyRe_Request{}, fmt.Errorf("Map key error for: %s\n", name)
	}

	return re, nil
}
