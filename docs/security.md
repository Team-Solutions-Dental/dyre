# Security Metadata

DyRe now supports annotating endpoints and fields with security requirements. This
document explains how the metadata is represented in `dyre.json`, how the parser
interprets it, and what downstream code should do with the information.

## Configuration Format

Both endpoints and fields accept an optional `security` key. The value can be
either a single string or an array of strings. The parser normalises the input
to a `[]string`.

```json
[
  {
    "name": "Customers",
    "tableName": "Customers",
    "security": "endpoint.customers.read",
    "fields": [
      {
        "name": "CustomerID",
        "nullable": false,
        "security": [
          "field.customers.customerid.view"
        ]
      },
      "FirstName",
      {
        "name": "CreateDate",
        "type": "date",
        "security": [
          "field.customers.createdate.view",
          "field.customers.createdate.edit"
        ]
      }
    ]
  }
]
```

- Omit the key when no additional permission is required.
- Empty strings are rejected during parsing.
- Arrays must contain strings; mixed or empty arrays raise parser errors.

## Parser Behaviour

- `endpoint.Endpoint.Security` and `endpoint.Field.Security` store the parsed
  identifier list.
- `endpoint/endpoint_parser.go` calls `parseSecurityList` for both endpoints and
  fields; any validation errors are surfaced alongside other schema issues.
- The colours of the base schema are unchanged—existing configurations that do
  not use the `security` key parse exactly as before.

## Downstream Usage

Security lists are descriptive metadata. Enforcement happens outside the parser:

1. The request pipeline should collect the caller’s granted permissions.
2. Before adding a select statement, ensure the caller satisfies
   `endpoint.Security`. For fields, prefer the field list; fall back to the
   endpoint list when the field list is empty.
3. Deny access (or strip the column) when the requirement is not met.

Sample pseudocode inside the transpiler:

```go
func ensureAllowed(perms map[string]struct{}, required []string) error {
    for _, id := range required {
        if _, ok := perms[id]; !ok {
            return fmt.Errorf("permission %s required", id)
        }
    }
    return nil
}
```

Consumers are free to interpret the identifiers according to their own role or
policy engine (RBAC, ABAC, etc.).

## Testing

`endpoint/endpoint_parser_test.go` includes fixtures that verify both the string
and array forms. Run `go test ./...` to ensure the schema updates continue to
round-trip correctly.

## Next Steps

- Integrate permission checks in the transpiler before adding select statements.
- Extend the configuration schema if more complex logic is required (for
  example, distinguishing between “all of” vs “any of” permissions).
- Document the enforcement location in API-facing documentation once the checks
  are implemented.
