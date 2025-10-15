# Security Enforcement in Dyre

This document describes the security enforcement system in Dyre, which allows fine-grained permission control over endpoint and field access with flexible denial behaviors.

## Overview

The security enforcement system provides:
- **Backward compatibility** with existing code
- **Field-level permissions** with inheritance from endpoints
- **Flexible denial behavior**: error on deny or silent omission
- **Wildcard support** for universal access
- **Pluggable permission checking** via host application integration
- **Consistent metadata** across SQL, TypeScript, and field lists

## Goals

- Preserve backwards compatibility with existing `security` entries
- Allow schema authors to opt into field omission (`onDeny: "omit"`) instead of hard failures
- Give host applications a clear hook for wiring their permission systems into the transpiler
- Ensure generated SQL, TypeScript metadata, and runtime results stay internally consistent when columns are omitted

## Metadata Schema

`security` accepts three formats: string shorthand, array, or object with explicit behavior.

| Form | Example | Meaning |
| --- | --- | --- |
| String | `"customers.read"` | Require one permission; error on deny. |
| Array | `["customers.view", "customers.edit"]` | Require *all* listed permissions; error on deny. |
| Object | `{"permissions": ["customers.email.view"], "onDeny": "omit"}` | Require all permissions; omit on deny. |

The values provided should originate from the host application's role or permission catalogue; Dyre does not impose additional namespacing or prefixes.

### Full Example

```json
{
  "name": "Customers",
  "tableName": "dbo.Customers",
  "security": {
    "permissions": ["customers.read"],
    "onDeny": "error"
  },
  "fields": [
    {
      "name": "CustomerID",
      "nullable": false,
      "security": "customers.customerid.read"
    },
    {
      "name": "Email",
      "security": {
        "permissions": ["customers.email.view", "customers.email.manage"],
        "onDeny": "omit"
      }
    },
    {
      "name": "Notes",
      "security": {
        "permissions": ["*"],
        "onDeny": "omit"
      }
    },
    "CreatedAt"
  ]
}
```

### Rules

- `permissions` is a non-empty array of host-defined role or permission identifiers (e.g., `"customers.read"`). Avoid redundant prefixes; reuse the exact tokens enforced by your auth layer.
- The literal `"*"` acts as a catch-all and always evaluates to allowed without involving the checker. Use it for fields that inherit access from broader roles while keeping consistent metadata.
- `onDeny` defaults to `"error"`; setting `"omit"` causes unauthorized columns to be skipped where possible.
- String and array shorthand are internally normalised to `{ permissions: [...], onDeny: "error" }`.

### Shorthand Examples

**String Shorthand (Single Permission)**
```json
{
  "security": "customers.read"
}
```

**Array Format (Multiple Permissions, Error on Deny)**
```json
{
  "security": ["customers.read", "customers.audit"]
}
```

**Object Format (Omit on Deny)**
```json
{
  "security": {
    "permissions": ["customers.email.view"],
    "onDeny": "omit"
  }
}
```

**Wildcard (Always Allowed)**
```json
{
  "security": {
    "permissions": ["*"],
    "onDeny": "omit"
  }
}
```

## Runtime Implementation

### SecurityChecker Interface

The transpiler accepts an optional checker supplied by the host service:

```go
type SecurityChecker interface {
    Allow(required []string) (bool, error)
}

func NewWithSecurity(query string, ep *endpoint.Endpoint, checker SecurityChecker) (*PrimaryIR, error)
```

- `Allow` returns `(true, nil)` when all `required` permissions are granted.
- Returning `(false, nil)` signals missing permissions.
- Returning an error aborts evaluation immediately (e.g., upstream auth failure).
- A `nil` checker preserves the current permissive behaviour.

### Built-in Checkers

**StaticChecker**: Checks against a fixed set of permissions
```go
checker := endpoint.NewStaticChecker(map[string]struct{}{
    "customers.read": {},
    "customers.customerid.view": {},
})
```

**RoleChecker**: Custom logic via callback (e.g., admin role expansion)
```go
checker := endpoint.NewRoleChecker(func(required []string) (bool, error) {
    if userHasRole("admin") {
        return true, nil // Admin bypasses all checks
    }
    return userHasPermissions(required), nil
})
```

**PermissiveChecker**: Allows all permissions (testing/migration)
```go
checker := endpoint.NewPermissiveChecker()
```

### Endpoint-Level Flow

1. Normalise endpoint security metadata to a `SecurityPolicy` struct.
2. If the policy contains `"*"`, treat it as satisfied and skip the checker.
3. Otherwise, before parsing SQL, probe the checker with the endpoint's permissions.
4. If denied and `onDeny == "error"`, return a descriptive authorization error.
5. If denied and `onDeny == "omit"`, return an empty result placeholder (caller decides how to surface).

### Field-Level Flow

1. Each time a column (or expression alias) is about to be added to the select list, consult the field policy. The presence of `"*"` marks the policy as satisfied without a checker call.
2. If a column lacks its own policy, inherit the endpoint policy.
3. When denied:
   - If `onDeny == "omit"`: Skip appending the select statement and record the omission so `FieldNames()` stays consistent.
   - If `onDeny == "error"`: Bubble up an authorization error immediately.

### Admin or Aggregated Permissions

- The `SecurityChecker` is responsible for expanding higher-level roles (e.g., `admin`) into the granular identifiers referenced by endpoints and fields.
- You can return `true` from `Allow` when a caller holds an aggregated permission that covers the requested identifiers. Example: treat `admin` as satisfying every permission under `customers.*`.
- Use the `"*"` policy entry when you want the schema itself to mark a resource as universally accessible (or already handled upstream).

## Usage Examples

### Basic Usage

```go
// Create a checker with granted permissions
checker := endpoint.NewStaticChecker(map[string]struct{}{
    "customers.read": {},
    "customers.customerid.view": {},
})

// Create IR with security enforcement
ir, err := transpiler.NewWithSecurity("CustomerID:Email:", customersEndpoint, checker)
if err != nil {
    // Handle permission denied error
}

sql, err := ir.EvaluateQuery()
// SQL only includes columns the user has permission to access
```

### Successful Request

```go
checker := endpoint.NewStaticChecker(map[string]struct{}{
    "customers.read": {},
    "customers.customerid.read": {},
})

ir, err := transpiler.NewWithSecurity("CustomerID:", customersEndpoint, checker)
sql, err := ir.EvaluateQuery()
// SELECT Customers.[CustomerID] FROM dbo.Customers
```

### Omitted Column

```go
checker := endpoint.NewStaticChecker(map[string]struct{}{}) // caller has no grants

ir, _ := transpiler.NewWithSecurity("Email:, CreatedAt:", customersEndpoint, checker)
sql, _ := ir.EvaluateQuery()
// SELECT Customers.[CreatedAt] FROM dbo.Customers
// Email was annotated with onDeny == "omit" and is dropped.
```

### Authorization Error

```go
checker := endpoint.NewStaticChecker(map[string]struct{}{})

_, err := transpiler.NewWithSecurity("CustomerID:", customersEndpoint, checker)
// err => "permission denied: requires [customers.read]"
```

### Admin Role Expansion

```go
checker := endpoint.NewRoleChecker(func(required []string) (bool, error) {
    if userHasRole("admin") {
        return true, nil // admin satisfies every requirement
    }
    return userHasPermissions(required), nil
})

ir, err := transpiler.NewWithSecurity(query, ep, checker)
```

- `Allow` returns `true` for any requested identifiers when the caller has the `admin` role.
- Omitted columns still honour their `onDeny` setting when the checker declines a request.

### Backward Compatibility

```go
// Nil checker maintains pre-security behavior (allows everything)
ir, err := transpiler.New(query, endpoint)

// Or explicitly:
ir, err := transpiler.NewWithSecurity(query, endpoint, nil)
```

## Implementation Details

### Security Metadata (endpoint/security.go)

- **SecurityPolicy struct**: Normalizes security metadata with `Permissions []string` and `OnDeny string` fields
- **NormalizeSecurityValue()**: Converts string, array, or object security values to SecurityPolicy
- **SecurityChecker interface**: Defines `Allow(required []string) (bool, error)` for runtime permission checks

### Schema Updates

- **Endpoint.Security**: Uses `*SecurityPolicy` instead of `[]string`
- **Field.Security**: Uses `*SecurityPolicy` instead of `[]string`
- **JSON parsing**: Handles all three formats (string, array, object)
- **JSON output**: Maintains backward-compatible array format for `onDeny="error"`, uses object format for `onDeny="omit"`

### Transpiler Integration

**Endpoint-Level Security**
- `NewWithSecurity()` creates IR with security checker
- Checks endpoint permissions before parsing query
- Returns error or empty result based on `onDeny` setting
- Handles wildcard `"*"` permission (always allowed)

**Field-Level Security**
- `checkFieldSecurity()` validates field access during column evaluation
- Fields inherit endpoint security if they lack their own policy
- Omitted fields are tracked in `IR.omittedFields` map
- Fields with `onDeny="omit"` are silently excluded from SQL
- Fields with `onDeny="error"` cause authorization errors

### Metadata Consistency

- **FieldNames()**: Automatically reflects omissions (returns actual SQL select list)
- **TypeScript generation**: Shows full schema (represents available fields, not user-specific view)
- **Join propagation**: Security checker propagates to SubIR for joined tables

### Expressions and References

- Expressions referencing unauthorized fields should error, even under `onDeny == "omit"`, to avoid emitting partially valid SQL. This behaviour may be revisited if downstream needs differ.

## Migration Path

1. **Phase 1**: Deploy with nil checker (no behavior change)
2. **Phase 2**: Add security metadata to endpoint JSON
3. **Phase 3**: Implement SecurityChecker in host application
4. **Phase 4**: Pass checker to NewWithSecurity()

## Testing

### Parser Tests (endpoint/security_test.go)
- String format normalization
- Array format normalization
- Object format with error/omit behaviors
- Default onDeny value
- Wildcard detection
- Invalid input handling
- All three checker implementations

### Transpiler Tests (transpiler/security_test.go)
- Endpoint-level denial with error
- Endpoint-level denial with omit
- Field omission (onDeny="omit")
- Field denial (onDeny="error")
- Security inheritance from endpoint
- Wildcard always allowed
- Nil checker (permissive backward compatibility)
- PermissiveChecker and RoleChecker
- FieldNames() reflects omissions

### Test Results

All tests passing:
- endpoint package: 15 tests (including 12 new security tests)
- transpiler package: 25 tests (including 10 new security tests)
- parser, lexer packages: All existing tests pass
- No breaking changes to existing functionality

## Known Behaviors

1. **Endpoint-level omit behavior**: Returns empty IR, which generates minimal SQL. Confirmed as acceptable.

2. **ORDER BY / GROUP BY with omitted columns**: Currently allowed to propagate, may reference unavailable columns. Consider adding validation in future if needed.
