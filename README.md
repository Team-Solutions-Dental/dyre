# \[Dy\]namic \[Re\]quests

DyRe is a interpreted request builder for a middleware service. The intent of DyRe is to help automate selections of multiple fields without removing any functionality for handling specific queries. No two APIs are the same and it is impossible to know what changes are to come, so flexibility favored over ease of use. 

Example API request
`curl http://localhost:8080/Customers?Query=CustomerID:Active:==True`  
- Fields are any independent SQL field.


## Setting up JSON config

First you will need JSON file to build with.
The name of the request is the thing being called.
Fields can be a string or an object with multiple params. 
DyRe is type aware and will reject certain unmatching typed requests.
If a type is not declared make sure to setup a default type that works for you. 
Defaults to nullable = true and type = string

```json
[
  {
    "name": "Customers",
    "tableName": "Customers",
    "schemaName": "dbo",
    "fields": [
      {
        "name": "CustomerID"
        "nullable" : false
      },
      "Name",
      {
        "name": "Active"
        "type": "bool"
      },
      {
        "name": "CreateDate"
        "type": "date"
      },
      {
        "name": "CustomerNumber"
        "type": "int"
      },
      "Phone"
    ]
  }
]

```
## Writing a query
Writing a query statement calls column names then expressions.

### Fields \ Columns

A field is called by its name with a colon following the name.

```bash
fieldName:
```

Multiple fields can be called in sequence. DyRe will respect the order in which fields were called. 
If the same field is called twice, it's ordered in the last position it was called.
```bash
CustomerID:Name:Active:
```


### Expressions
Columns can use additional expressions for filtering. 
A boolean style expression can be given with a semicolon for the construction of a where statement.
```bash
CustomerNumber: > 100;

Active: == FALSE;

CustomerNumber: > 100 AND < 200;
```
Boolean expressions include:

- ==
- !=
- \>
- <
- \>=
- <=
- AND
- OR

When a conditional expression is given as prefix DyRe assumes you are referencing the column as the other part of the expression. 
If you want to format your expression with the prefix you can declare the '@' for reference to the column name.
```bash
CustomersNumber: > 200;
# Is the same as 
CustomersNumber: @ > 200;
# Is the same as
CustomersNumber: 200 < @;
```

Expressions can include builtin function calls for specific handling of fields. 
For example `exclude(@)` will exclude a field from the top level of the query omitting it from the returned statement.
```bash
Active: exclude(@)
```

Expressions and can be given in a row for the use of multiple expressions including function calls.
```bash

Name: exclude(@); != NULL;
```

### Putting it all together

Multiple fields and expression can be called. Expressions index of the most recent field called for inference.
```bash
CustomerID:CreateDate: > date('2025/04/03');Active: exclude(@); == FASLE;
```

## Setting up middleware

Starting an example server. You can opt for a global variable or pass it in through functions.

```go

// Global Var for fetching request info
var Re map[string]dyre.DyRe_Request

func main() {
	var dyre_err error
	Re, dyre_err = dyre.Init("./dyre.json")
	if dyre_err != nil {
		log.Panicf("dyre init failed: %v", dyre_err)
	}
}
```

### Making a handler 
get all you params then check the values against the response. 
Once a response has been validated for fields and groups its pretty easy to handle the rest.

```go
func getCustomers(c *ex.Context) {
	query_string, ok := c.GetQuery("Query")
    if !ok {
      query_string = "CustomerID:Active:==True"
    }

    q, err := Re.Request('Customers', query_string)
	if err != nil {
		g.String(500, "Failed to initialize request")
		return
	}

    sql, err := q.EvaluateQuery()
	if err != nil {
		g.String(500, "Failed to build request")
		return
	}

	table, err := read_db(sql)
	if err != nil {
		g.String(500, "Failed to make request")
		return
	}

	output := make(map[string]any)
	headers := valid.FieldNames()

	output["Headers"] = headers
	output["Table"] = table

	c.JSON(200, output)
	return
}
```

## Joining tables
Joining tables as requests is possible in DyRe allowing for powerful queries from the front end.
Each tables query is made separately so they can either be query parameters or post parameters if preferred.
If a field is declared in the joined table it will be included in the result unless excluded.
The top level tables fields take precedence then joined tables fields are added after.


```go
func getCustomersWithBiling(c *ex.Context) {
	query_string, ok := c.GetQuery("Customers")
    if !ok {
      query_string = "CustomerID:Active:==True"
    }

	billing_string, ok := c.GetQuery("Billing")
    if !ok {
      query_string = "CustomerID:Balance: > 0"
    }

    q, err := Re.Request('Customers', query_string)
	if err != nil {
		g.String(500, "Failed to initialize request")
		return
	}

    _, err := q.INNERJOIN("Billing").ON("CustomerID","CustomerID").Query(billing_string)
	if err != nil {
		g.String(500, "Failed to initialize request")
		return
	}

    sql, err := q.EvaluateQuery()
	if err != nil {
		g.String(500, "Failed to build request")
		return
	}

	table, err := read_db(sql)
	if err != nil {
		g.String(500, "Failed to make request")
		return
	}

	output := make(map[string]any)
	headers := valid.FieldNames()

	output["Headers"] = headers
	output["Table"] = table

	c.JSON(200, output)
	return
}
```
