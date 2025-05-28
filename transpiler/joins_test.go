package transpiler

import (
	"testing"

	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/lexer"
	"github.com/vamuscari/dyre/parser"
)

func TestSingleJoins(t *testing.T) {
	tests := []struct {
		input_parent string
		input_join   string
		parent_on    string
		on           string
		expected     string
	}{
		{"intx:,string:,", "inty:,bool:,", "intx", "inty",
			"SELECT Parent.intx, Parent.string, Join.bool FROM Parent INNER JOIN ( SELECT Join.inty, Join.bool FROM Join ) AS Join ON Parent.intx = Join.inty",
		},
	}

	for _, tt := range tests {
		parent_ir := testNewParent(tt.input_parent)

		_, eerrs := parent_ir.INNERJOIN("Join").ON(tt.parent_on, tt.on).Query(tt.input_join)
		if eerrs != nil {
			for _, e := range eerrs {
				t.Errorf("Query test error. [%s] %v\n", tt.input_parent, e)
			}
		}

		evaluated := parent_ir.EvaluateQuery()

		if evaluated != tt.expected {
			t.Errorf("Query failed. [%s] [%s]\n%s \n%s\n ", tt.input_parent, tt.input_join, evaluated, tt.expected)
		}

	}
}

func testNewParent(input string) *IR {
	l := lexer.New(input)
	p := parser.New(l)
	q := p.ParseQuery()
	var service *endpoint.Service = &endpoint.Service{Endpoints: map[string]*endpoint.Endpoint{}}

	parent := endpoint.Endpoint{
		Service:    service,
		Name:       "Parent",
		TableName:  "Parent",
		FieldNames: []string{"intx", "string"},
		Fields: map[string]endpoint.Field{
			"intx":   {Name: "intx", DefaultField: false},
			"string": {Name: "string", DefaultField: false},
		},
	}

	join := endpoint.Endpoint{
		Service:    service,
		Name:       "Join",
		TableName:  "Join",
		FieldNames: []string{"inty", "bool"},
		Fields: map[string]endpoint.Field{
			"inty": {Name: "inty", DefaultField: false},
			"bool": {Name: "bool", DefaultField: false},
		},
	}

	service.Endpoints["Parent"] = &parent
	service.Endpoints["Join"] = &join

	ir := &IR{
		endpoint: service.Endpoints["Parent"],
		ast:      q,
	}

	return ir
}

func TestDoubleJoins(t *testing.T) {
	test := struct {
		input_x  string
		input_xy string
		input_yz string
		expected string
	}{
		"x:,",
		"x:,y:,",
		"y:,z:,",
		"SELECT X.x, XY.y, XY.z FROM X INNER JOIN ( SELECT XY.x, XY.y, YZ.z FROM XY INNER JOIN ( SELECT YZ.y, YZ.z FROM YZ ) AS YZ ON XY.y = YZ.y ) AS XY ON X.x = XY.x",
	}

	x := testNewXYZ(test.input_x)

	xy, errs := x.INNERJOIN("XY").ON("x", "x").Query(test.input_xy)
	if errs != nil {
		for _, e := range errs {
			t.Errorf("Query test error. %s\n", e)
		}
	}

	_, errs = xy.INNERJOIN("YZ").ON("y", "y").Query(test.input_yz)
	if errs != nil {
		for _, e := range errs {
			t.Errorf("Query test error. %s\n", e)
		}
	}

	sql := x.EvaluateQuery()
	if sql != test.expected {
		t.Errorf("Double Join Failed\n %s\n %s\n", sql, test.expected)
	}

}

func testNewXYZ(input string) *IR {
	l := lexer.New(input)
	p := parser.New(l)
	q := p.ParseQuery()
	var service *endpoint.Service = &endpoint.Service{Endpoints: map[string]*endpoint.Endpoint{}}

	x := endpoint.Endpoint{
		Service:    service,
		Name:       "X",
		TableName:  "X",
		FieldNames: []string{"x"},
		Fields: map[string]endpoint.Field{
			"x": {Name: "x", DefaultField: false},
		},
	}

	xy := endpoint.Endpoint{
		Service:    service,
		Name:       "XY",
		TableName:  "XY",
		FieldNames: []string{"x", "y"},
		Fields: map[string]endpoint.Field{
			"x": {Name: "x", DefaultField: false},
			"y": {Name: "y", DefaultField: false},
		},
	}
	yz := endpoint.Endpoint{
		Service:    service,
		Name:       "YZ",
		TableName:  "YZ",
		FieldNames: []string{"y", "z"},
		Fields: map[string]endpoint.Field{
			"y": {Name: "y", DefaultField: false},
			"z": {Name: "z", DefaultField: false},
		},
	}

	service.Endpoints["X"] = &x
	service.Endpoints["XY"] = &xy
	service.Endpoints["YZ"] = &yz

	ir := &IR{
		endpoint: service.Endpoints["X"],
		ast:      q,
	}

	return ir
}
