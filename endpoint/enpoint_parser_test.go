package endpoint

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseJSON(t *testing.T) {

	const inputJSON string = `
[
  {
    "name": "Customers",
    "tableName": "Customers",
    "schemaName": "dbo",
    "joins": [
      {
        "endpoint": "Invoices",
        "on": "CustomerID"
      }
    ],
    "fields": [
      {
        "name": "CustomerID",
        "nullable": false
      },
      {
        "name": "Zip",
        "type": "int"
      },
      "FirstName",
      "LastName",
      {
        "name": "CreateDate",
        "type": "date"
      },
      {
        "name": "Active",
        "type": "bool"
      }
    ]
  },
  {
    "name": "Invoices",
    "tableName": "Invoices",
    "schemaName": "dbo",
    "joins": [
      {
        "endpoint": "Sales",
        "on": [
          "SaleID",
          "SaleID"
        ]
      }
    ],
    "fields": [
      {
        "name": "SaleID",
        "nullable": false
      },
      {
        "name": "Balance",
        "type": "float"
      },
      {
        "name": "InvoiceNumber",
        "type": "int"
      },
      {
        "name": "CreateDate",
        "type": "date"
      }
    ]
  },
  {
    "name": "Sales",
    "tableName": "Invoices",
    "schemaName": "dbo",
    "joins": [
      {
        "endpoint": "Customers",
        "on": "CustomerID"
      },
      {
        "endpoint": "Invoices",
        "on": "SaleID"
      }
    ],
    "fields": [
      {
        "name": "CustomerID",
        "nullable": false
      },
      {
        "name": "SaleID",
        "type": "float"
      },
      {
        "name": "CreateDate",
        "type": "date"
      },
      {
        "name": "Charge",
        "type": "float"
      }
    ]
  }

]
`

	const expectedJSON string = `
[
  {
    "name": "Customers",
    "tableName": "Customers",
    "schemaName": "dbo",
    "joins": [
      {
        "endpoint": "Invoices",
        "on": [
          "CustomerID",
          "CustomerID"
        ]
      }
    ],
    "fields": [
      {
        "name": "Active",
        "type": "BOOLEAN",
        "nullable": true
      },
      {
        "name": "CreateDate",
        "type": "DATE",
        "nullable": true
      },
      {
        "name": "CustomerID",
        "type": "STRING",
        "nullable": false
      },
      {
        "name": "FirstName",
        "type": "STRING",
        "nullable": true
      },
      {
        "name": "LastName",
        "type": "STRING",
        "nullable": true
      },
      {
        "name": "Zip",
        "type": "INTEGER",
        "nullable": true
      }
    ]
  },
  {
    "name": "Invoices",
    "tableName": "Invoices",
    "schemaName": "dbo",
    "joins": [
      {
        "endpoint": "Customers",
        "on": [
          "CustomerID",
          "CustomerID"
        ]
      }
    ],
    "fields": [
      {
        "name": "Balance",
        "type": "FLOAT",
        "nullable": true
      },
      {
        "name": "CreateDate",
        "type": "DATE",
        "nullable": true
      },
      {
        "name": "CustomerID",
        "type": "STRING",
        "nullable": false
      },
      {
        "name": "InvoiceNumber",
        "type": "INTEGER",
        "nullable": true
      }
    ]
  }
]
`

	service, err := ParseJSON([]byte(inputJSON))
	if err != nil {
		t.Fatalf(`error: %v`, err)
	}

	// paths := service.AllEndpointsPaths(2)
	// for _, p := range paths {
	// 	fmt.Println(strings.Join(p, "/"))
	// }

	expected_endpoints := []string{"Customers", "Invoices"}
	for _, expected_endpoint := range expected_endpoints {
		_, ok := service.Endpoints[expected_endpoint]
		if !ok {
			t.Errorf("Enpoint error for %s", expected_endpoint)
		}
	}

	comparableJSON := strings.Replace(expectedJSON, "\n", "", -1)
	comparableJSON = strings.Replace(comparableJSON, "\t", "", -1)
	comparableJSON = strings.Replace(comparableJSON, " ", "", -1)

	evaluatedJSON := strings.Replace(service.JSON(), " ", "", -1)

	if comparableJSON != evaluatedJSON {
		t.Errorf("evaluated JSON did not match expected JSON\n\nexpected:\n%s\n\nevaluated:\n%s\n\n", comparableJSON, evaluatedJSON)
	}

}

func diffStrings(str1, str2 string) {
	var matchLen int
	if len(str1) > len(str2) {
		matchLen = len(str2)
	} else {
		matchLen = len(str1)
	}
	matchBreak := true
	matched := ""
	diff1 := ""
	diff2 := ""

	for i := range matchLen {
		if matchBreak && str1[i] == str2[i] {
			matched = matched + string(str1[i])
		} else {
			matchBreak = false
			diff1 = diff1 + string(str1[i])
			diff2 = diff2 + string(str2[i])
		}
	}

	fmt.Println("Matched JSON")
	fmt.Println(matched)
	fmt.Println("Expected JSON Diff")
	fmt.Println(diff1)
	fmt.Println("Evaluated JSON Diff")
	fmt.Println(diff2)
}
