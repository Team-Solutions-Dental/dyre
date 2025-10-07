# \[Dy\]namic \[Re\]quests

DyRe is a interpreted request builder for a middleware service. The intent of DyRe is to help automate selections of multiple fields without removing any functionality for handling specific queries. DyRe is not a full query language and is not trying to be one. No two APIs are the same and it is impossible to know what changes are to come, so flexibility favored over ease of use.

Example API request
`curl http://localhost:8080/Customers?Query=CustomerID:Active:==True`  


## Setting up JSON config

First, you will need a JSON file to build with.
The name of the request is the thing being called.
Fields are any independent SQL field on a table.
Fields can be a string or an object with multiple params. 
DyRe is type-aware and will reject certain unmatching typed requests.
If a type is not declared, make sure to set up a default type that works for you. 
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
- AND or && (keep in mind & is a reserved character for query params)
- OR or ||

When a conditional expression is given as prefix DyRe assumes you are referencing the column as the other part of the expression. 
If you want to format your expression with the prefix you can declare the '@' for reference to the column name.
```bash
CustomersNumber: > 200;
# Is the same as 
CustomersNumber: @ > 200;
# Is the same as
CustomersNumber: 200 < @;
```

In any expression a column can be provided by using the `@()` function
In the example 
`
@('CustomerNumber') > 200;
`
would provide the same where condition on the SQL statement but exclude it from the result. Note, that columns are not required for an expression to be provided.
Also, If a SQL statement has multiple column conditions then independent SQL calls will be needed. 

```bash
@('CustomerNumber') > 200 OR @('Balance') > 0;
```
Expressions can include builtin function calls for specific handling of fields. 
For example `datepart('year',@)` will return the year of a date. 
```bash
CreateDate: datepart('year',@) > 2024
```

Expressions and can be given in a row for the use of multiple expressions including function calls.  
```bash
Name: len(@) > 3; != NULL;
```

### Column Functions

Columns can be represented as functions by using a `:` at the end similar to how column names are called.
The most common column call is `AS():`. This represents the alias call AS that is often used in select statements.

```bash
AS('allTrue', true)):
```


Expression functions can be used inside to provide a modified output.
In the example below `datepart()` is being called similar to how it's used in SQL to modify the output of the expression statement

```bash
AS('year', datepart('year', @('CreateDate'))): > 2024
```

Warning: `AS(): expression;` will wrap alias select statement to make a where statement possible. Avoid this kind of expression when possible.


### Putting it all together

Multiple fields and expression can be called. Expressions index of the most recent field called for inference.
```bash
CustomerID:CreateDate: > date('2025/04/03');Active: != NULL; == FASLE;
```


### Group By  

Grouped table results are similar to how column functions are called. In the example below grouping on the Active column returns the column as expected. 

```bash
GROUP('Active'):
```
Grouping using an expression is possible when GROUP is provided with two arguments. The expression will be replicated in SQL to the Having statement to provided conditional outputs. 

```bash
GROUP('year', datepart('year', @('createDate'))): > 2024;
```

Standard grouping functions such as SUM, MIN, MAX, etc.. 

```bash
GROUP('ID'):SUM('SumBalance', @('Balance')):
```

Warn: Grouping Functions cannot be mixed with regular column functions or expressions. 

### Operators Reference

For a comprehensive list of all available operators and functions in DyRe, please refer to the [operators documentation](docs/operators.md).  


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

## Additional Expressions & Options

### Order By

Order by shares a similar syntax to query expressions where column names are declared followed by some expression.  
Only accessible through the top level query.

Order by allows the call of any defined column / alias or a top level table column

```bash
 CreateDate: DESC; CustomerID: ASC;
```

### LIMIT

This is just representation of TOP to restrict the amount returned.  
Only accessible through the top level query.

```go
func getCustomersWithBiling(c *ex.Context) {

    ...

    q, err := Re.Request('Customers', query_string)
	if err != nil {
		c.String(500, "Failed to initialize request")
		return
	}

    q.LIMIT(100)

    ...
}
```


