package endpoint

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/vamuscari/dyre/utils"
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
        "endpoint": "Customers",
        "on": [
          "CustomerID",
          "CustomerID"
        ]
      }
    ],
    "fields": [
      {
        "name": "CustomerID",
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

	var matchLen int
	if len(comparableJSON) > len(evaluatedJSON) {
		matchLen = len(evaluatedJSON)
	} else {
		matchLen = len(comparableJSON)
	}
	matchBreak := true
	matched := ""
	diff1 := ""
	diff2 := ""

	for i := range matchLen {
		if matchBreak && comparableJSON[i] == evaluatedJSON[i] {
			matched = matched + string(comparableJSON[i])
		} else {
			matchBreak = false
			diff1 = diff1 + string(comparableJSON[i])
			diff2 = diff2 + string(evaluatedJSON[i])
		}
	}

	fmt.Println(" ")
	fmt.Println(matched)
	fmt.Println(" ")
	fmt.Println(diff1)
	fmt.Println(" ")
	fmt.Println(diff2)
}

func deepEqualStringArray(arr1 []string, arr2 []string) []error {
	errors := []error{}
	for _, k := range arr1 {
		if !utils.Array_Contains(arr2, k) {
			errors = append(errors, fmt.Errorf("%s not found in %v\n", k, arr2))
		}
	}
	for _, k := range arr2 {
		if !utils.Array_Contains(arr1, k) {
			errors = append(errors, fmt.Errorf("%s not found in %v\n", k, arr1))
		}
	}
	return errors
}

func openDyreJSON(path string) []byte {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Panic(err)
		return nil
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Panic(err)
		return nil
	}

	return byteValue

}
