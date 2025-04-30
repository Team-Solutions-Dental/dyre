# \[Dy\]namic \[Re\]quests

DyRe is a request builder for a middleware service. The intent of DyRe is to help automate selections of multiple fields without removing any functionality for handling requests. No two APIs are the same and it is impossible to know what changes are to come, so flexibility favored over ease of use. 

Example API request
`curl http://localhost:8080/Customers?fields=Name,Phone&groups=Address`  
- Fields are any independent SQL field.
- Groups are a convenience for simplifying requests by returning multiple fields. (At no point have I ever just wanted City in a request lol)


## Setting up JSON config

First you will need JSON file to build with.
The name of the request is the thing being called.
Fields can be a string or an object with multiple params. 
If a type is not declared make sure to setup a default type that works for you. 
Groups are collections of fields to help simplify requests. A group must have a name and fields. 
The fields in a group are the same as in the base fields.

```json
[
  {
    "name": "Customers",
    "tableName": "Customers",
    "fields": [
      {
        "name": "CustomerID",
        "required": true
      },
      "Name",
      "Phone"
    ],
    "groups": [
      {
        "name": "Address",
        "fields": [
         "Street"
          {
            "name": "City"
          },
          {
            "name": "State",
          },
          {
            "name": "Zipcode",
          },
          {
            "name": "Zipint",
          }
        ]
      }
    ]
  }
]

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

## Making a handler 
get all you params then check the values against the response. Once a response has been validated for fields and groups its pretty easy to handle the rest.

```go
func getCustomers(g *gin.Context) {
	fields_string, _ := g.GetQuery("fields")
	fields := strings.Split(fields_string, ",")

	groups_string, _ := g.GetQuery("groups")
	groups := strings.Split(groups_string, ",")
	customers, ok := Re["Customers"]
	if !ok {
		g.String(500, "Failed to start request")
		return
	}

	valid, err := customers.ValidateRequest(fields, groups)
	if err != nil {
		g.String(400, "Failed to parse request")
		return
	}

	query := fmt.Sprintf("SELECT %s FROM %s LIMIT 10", strings.Join(valid.SQLFields(), ", "), customers.TableName())

	table, err := read_db(query, valid)
	if err != nil {
		g.String(500, "Failed to make request")
		return
	}

	output := make(map[string]any)
	headers := valid.Headers()

	output["Headers"] = headers
	output["Table"] = table

	g.JSON(200, output)
	return
}
```

