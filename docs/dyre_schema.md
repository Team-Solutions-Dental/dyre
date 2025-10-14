# DyRe JSON Schema Documentation

This document provides comprehensive documentation on how to create the JSON configuration file for DyRe (Dynamic Requests).

## Overview

The DyRe JSON configuration file defines the structure of your endpoints, their fields, and relationships between them. This configuration is used by the DyRe system to validate and process requests.

The configuration file is an array of endpoint definitions, each representing a database table or view that can be queried.

## JSON Structure

The JSON file should contain an array of endpoint objects:

```json
[
  {
    "name": "EndpointName",
    "tableName": "DatabaseTableName",
    "schemaName": "DatabaseSchemaName",
    "fields": [
      "Field1",
      "Field2"
    ],
    "joins": [
      {
        "endpoint": "OtherEndpoint",
        "on": "CommonField"
      }
    ]
  },
  {
    "name": "OtherEndpoint",
    "tableName": "OtherTable",
    "schemaName": "DatabaseSchemaName",
    "fields": [
      "CommonField",
      "Field3"
    ]
  }
]
```

## Endpoint Definition

Each endpoint object represents a queryable resource and has the following properties:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | The name of the endpoint used in API requests |
| `tableName` | string | Yes | The name of the database table or view |
| `schemaName` | string | No | The database schema name (defaults to "dbo") |
| `fields` | array | Yes | An array of field definitions |
| `joins` | array | No | An array of join definitions |
| `security` | string or array | No | Optional permission identifiers required to access the endpoint. Accepts a single string or an array of strings. |

### Example

```json
{
  "name": "Customers",
  "tableName": "Customers",
  "schemaName": "dbo",
  "fields": [
    "FirstName",
    "LastName"
  ],
  "joins": [
    {
      "endpoint": "Invoices",
      "on": "CustomerID"
    }
  ]
}
```

## Field Definition

Fields represent columns in your database table. Each field can be defined in one of two ways:

### Simple Format (String)

For fields with default settings (type = string, nullable = true):

```json
"FieldName"
```

### Object Format

For fields with custom settings:

```json
{
  "name": "FieldName",
  "type": "TypeName",
  "nullable": true
}
```

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `name` | string | Yes | - | The name of the field |
| `type` | string | No | "STRING" | The data type of the field |
| `nullable` | boolean | No | true | Whether the field can be null |
| `security` | string or array | No | - | Optional permission identifiers required to select this field. Accepts a single string or an array of strings. |

Endpoint- and field-level `security` entries are interpreted as identifiers for your authorization system. Providing a single string is equivalent to supplying an array with one element. Omit the key when no additional permissions are required.

### Supported Data Types

The following data types are supported:

| Type | Aliases | Description |
|------|---------|-------------|
| `STRING` | - | Text data |
| `BOOLEAN` | `BOOL` | Boolean (true/false) data |
| `INTEGER` | `INT` | Integer numeric data |
| `FLOAT` | - | Floating-point numeric data |
| `DATE` | - | Date data (without time) |
| `DATETIME` | - | Date and time data |

### Field Examples

```json
[
  "FirstName",
  {
    "name": "CustomerID",
    "type": "STRING",
    "nullable": false,
    "security": "field.customers.customerid.view"
  },
  {
    "name": "Balance",
    "type": "FLOAT",
    "nullable": true,
    "security": [
      "field.customers.balance.view",
      "field.customers.balance.edit"
    ]
  },
  {
    "name": "CreateDate",
    "type": "DATE",
    "nullable": true
  }
]
```

## Join Definition

Joins define relationships between endpoints. Each join has the following properties:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `endpoint` | string | Yes | The name of the endpoint to join with |
| `on` | string or array | Yes | The field(s) to join on |

The `on` property can be specified in two ways:

1. As a string: When the field name is the same in both endpoints
2. As an array: When the field names are different, specified as [parentField, childField]

### Join Examples

```json
[
  {
    "endpoint": "Invoices",
    "on": "CustomerID"
  },
  {
    "endpoint": "Invoices",
    "on": ["CustomerID", "InvoiceCustomerID"]
  }
]
```

## Complete Example

Here's a complete example of a DyRe JSON configuration file with two endpoints:

```json
[
  {
    "name": "Customers",
    "tableName": "Customers",
    "schemaName": "dbo",
    "joins": [
      {
        "endpoint": "Invoices",
        "on": ["CustomerID", "CustomerID"]
      }
    ],
    "fields": [
      {
        "name": "CustomerID",
        "type": "STRING",
        "nullable": false
      },
      {
        "name": "Zip",
        "type": "INTEGER",
        "nullable": true
      },
      "FirstName",
      "LastName",
      {
        "name": "CreateDate",
        "type": "DATE",
        "nullable": true
      },
      {
        "name": "Active",
        "type": "BOOLEAN",
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
        "on": ["CustomerID", "CustomerID"]
      }
    ],
    "fields": [
      {
        "name": "CustomerID",
        "type": "STRING",
        "nullable": false
      },
      {
        "name": "Balance",
        "type": "FLOAT",
        "nullable": true
      },
      {
        "name": "InvoiceNumber",
        "type": "INTEGER",
        "nullable": true
      },
      {
        "name": "CreateDate",
        "type": "DATE",
        "nullable": true
      }
    ]
  }
]
```

## Best Practices

1. **Use Consistent Naming**: Use consistent naming conventions for endpoints and fields.
2. **Define Primary Keys as Non-Nullable**: Set `nullable: false` for primary key fields.
3. **Define Joins Properly**: Ensure that joined fields exist in both endpoints and have compatible types.
4. **Use Appropriate Types**: Choose the appropriate data type for each field to ensure proper validation and processing.
5. **Document Your Schema**: Keep documentation of your schema for reference, especially for complex relationships.
6. **Test Your Configuration**: Validate your JSON configuration before deploying it to production.

## Validation Rules

The DyRe system performs the following validations on your JSON configuration:

1. Each endpoint must have a unique name.
2. Each endpoint must have a tableName.
3. Each endpoint must have at least one field.
4. Each field within an endpoint must have a unique name.
5. Field types must be one of the supported types.
6. Join references must point to existing endpoints.
7. Join fields must exist in their respective endpoints.

## Troubleshooting

If you encounter errors when loading your JSON configuration, check for:

1. **JSON Syntax Errors**: Ensure your JSON is valid (no missing commas, brackets, etc.).
2. **Missing Required Fields**: Ensure all required properties are provided.
3. **Invalid Types**: Ensure field types are valid and properly capitalized.
4. **Duplicate Names**: Ensure endpoint and field names are unique.
5. **Invalid Join References**: Ensure join references point to existing endpoints.
