package transpiler

import (
	"testing"

	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/object"
)

func TestSingleJoins(t *testing.T) {
	tests := []struct {
		input_parent string
		input_join   string
		parent_on    string
		on           string
		expected     string
	}{
		{"intx:string:", "inty:bool:", "intx", "inty",
			"SELECT Parent.[intx], Parent.[string], Join.[bool] FROM dbo.Parent INNER JOIN ( SELECT JoinTable.[inty], JoinTable.[bool] FROM dbo.JoinTable ) AS Join ON Parent.[intx] = Join.[inty]",
		},
		{"intx:string:", "bool:", "intx", "inty",
			"SELECT Parent.[intx], Parent.[string], Join.[bool] FROM dbo.Parent INNER JOIN ( SELECT JoinTable.[bool], JoinTable.[inty] FROM dbo.JoinTable ) AS Join ON Parent.[intx] = Join.[inty]",
		},
		{"string:", "inty:bool:", "intx", "inty",
			"SELECT Parent.[string], Join.[bool] FROM dbo.Parent INNER JOIN ( SELECT JoinTable.[inty], JoinTable.[bool] FROM dbo.JoinTable ) AS Join ON Parent.[intx] = Join.[inty]",
		},
		{"string:", "inty:bool:", "tacos", "inty",
			"SELECT Parent.[string], Join.[bool] FROM dbo.Parent INNER JOIN ( SELECT JoinTable.[inty], JoinTable.[bool] FROM dbo.JoinTable ) AS Join ON Parent.[intx] = Join.[inty]",
		},
	}

	for _, tt := range tests {
		parent_ir, err := testNewParent(tt.input_parent)
		if err != nil {
			t.Errorf("Query test error. %s\n", err)
		}
		_, err = parent_ir.INNERJOIN("Join").ON(tt.parent_on, tt.on).Query(tt.input_join)
		if err != nil {
			t.Errorf("Query test error. [%s] %v\n", tt.input_parent, err)
		}

		evaluated, err := parent_ir.EvaluateQuery()

		if err != nil {
			t.Errorf("Test Single Join Error. %s\n", err.Error())
		}

		if evaluated != tt.expected {
			t.Errorf("Test Single Join Failed. [%s] [%s]\n%s \n%s\n ", tt.input_parent, tt.input_join, evaluated, tt.expected)
		}

	}
}

func testNewParent(input string) (*PrimaryIR, error) {
	var service *endpoint.Service = &endpoint.Service{Endpoints: map[string]*endpoint.Endpoint{}}

	parent := &endpoint.Endpoint{
		Service:    service,
		Name:       "Parent",
		TableName:  "Parent",
		SchemaName: "dbo",
		FieldNames: []string{"intx", "string"},
		Fields:     map[string]endpoint.Field{},
	}

	parent.Fields["intx"] = endpoint.Field{Endpoint: parent, Name: "intx", FieldType: object.INTEGER_OBJ}
	parent.Fields["string"] = endpoint.Field{Endpoint: parent, Name: "string", FieldType: object.STRING_OBJ}

	join := &endpoint.Endpoint{
		Service:    service,
		Name:       "Join",
		TableName:  "JoinTable",
		SchemaName: "dbo",
		FieldNames: []string{"inty", "bool"},
		Fields:     map[string]endpoint.Field{},
	}

	join.Fields["inty"] = endpoint.Field{Endpoint: join, Name: "inty", FieldType: object.INTEGER_OBJ}
	join.Fields["bool"] = endpoint.Field{Endpoint: join, Name: "bool", FieldType: object.BOOLEAN_OBJ}

	service.Endpoints["Parent"] = parent
	service.Endpoints["Join"] = join

	return New(input, service.Endpoints["Parent"])
}

func TestDoubleJoins(t *testing.T) {
	test := struct {
		input_x  string
		input_xy string
		input_yz string
		expected string
	}{
		"x:",
		"x:y:",
		"y:z:",
		"SELECT X.[x], XY.[y], XY.[z] FROM X INNER JOIN ( SELECT XY.[x], XY.[y], YZ.[z] FROM XY INNER JOIN ( SELECT YZ.[y], YZ.[z] FROM YZ ) AS YZ ON XY.[y] = YZ.[y] ) AS XY ON X.[x] = XY.[x]",
	}

	x, err := testNewXYZ(test.input_x)
	if err != nil {
		t.Fatalf("testNewXYZ. %s\n", err.Error())
	}

	xy, err := x.INNERJOIN("XY").ON("x", "x").Query(test.input_xy)
	if err != nil {
		t.Fatalf("INNERJOIN XY. %s\n", err.Error())
	}

	_, err = xy.INNERJOIN("YZ").ON("y", "y").Query(test.input_yz)
	if err != nil {
		t.Fatalf("INNERJOIN YZ. %s\n", err.Error())
	}

	sql, err := x.EvaluateQuery()
	if err != nil {
		t.Errorf("Double Join Query Error. %s", err.Error())
	}
	if sql != test.expected {
		t.Errorf("Double Join Failed\n %s\n %s\n", sql, test.expected)
	}

}

func testNewXYZ(input string) (*PrimaryIR, error) {
	var service *endpoint.Service = &endpoint.Service{Endpoints: map[string]*endpoint.Endpoint{}}

	x := endpoint.Endpoint{
		Service:    service,
		Name:       "X",
		TableName:  "X",
		FieldNames: []string{"x"},
		Fields: map[string]endpoint.Field{
			"x": {Name: "x", FieldType: object.STRING_OBJ},
		},
	}

	xy := endpoint.Endpoint{
		Service:    service,
		Name:       "XY",
		TableName:  "XY",
		FieldNames: []string{"x", "y"},
		Fields: map[string]endpoint.Field{
			"x": {Name: "x", FieldType: object.STRING_OBJ},
			"y": {Name: "y", FieldType: object.STRING_OBJ},
		},
	}
	yz := endpoint.Endpoint{
		Service:    service,
		Name:       "YZ",
		TableName:  "YZ",
		FieldNames: []string{"y", "z"},
		Fields: map[string]endpoint.Field{
			"y": {Name: "y", FieldType: object.STRING_OBJ},
			"z": {Name: "z", FieldType: object.STRING_OBJ},
		},
	}

	service.Endpoints["X"] = &x
	service.Endpoints["XY"] = &xy
	service.Endpoints["YZ"] = &yz

	return New(input, service.Endpoints["X"])
}
