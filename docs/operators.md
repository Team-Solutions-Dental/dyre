# DyRe Operators

This document provides a comprehensive list of all operators available in DyRe.

## Comparison Operators

These operators are used for comparing values:

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal to | `CustomerID: == 100;` |
| `!=` | Not equal to | `Active: != TRUE;` |
| `>` | Greater than | `CustomerNumber: > 200;` |
| `<` | Less than | `Balance: < 1000;` |
| `>=` | Greater than or equal to | `CustomerNumber: >= 100;` |
| `<=` | Less than or equal to | `Balance: <= 500;` |

## Logical Operators

These operators are used to combine conditions:

| Operator | Description | Example |
|----------|-------------|---------|
| `AND` or `&&` | Logical AND | `CustomerNumber: > 100 AND < 200;` |
| `OR` or `\|\|` | Logical OR | `Active: == TRUE OR Balance: > 0;` |

## Prefix Operators

These operators are used before a value:

| Operator | Description | Example |
|----------|-------------|---------|
| `!` | Logical NOT | `!Active` |
| `-` | Negative | `-Balance` |

## NULL Operators

These operators are used for NULL comparisons:

| Operator | Description | Example |
|----------|-------------|---------|
| `IS NULL` | Equal to NULL | `CustomerID: == NULL;` |
| `IS NOT NULL` | Not equal to NULL | `CustomerID: != NULL;` |

## Column Functions

These functions are used to manipulate columns:

| Function | Description | Example |
|----------|-------------|---------|
| `AS(name, expression)` | Renames a column or expression | `AS('year', datepart('year', @('CreateDate'))):` |
| `EXCLUDE(name)` | Excludes a field from the query results | `EXCLUDE('Password'):` |

## Group Functions

These functions are used for grouping and aggregation:

| Function | Description | Example |
|----------|-------------|---------|
| `GROUP(column)` | Groups results by a column | `GROUP('Active'):` |
| `GROUP(alias, expression)` | Groups results by an expression | `GROUP('year', datepart('year', @('createDate'))):` |
| `COUNT(alias, expression)` | Counts rows in a group | `COUNT('CustomerCount', @('CustomerID')):` |
| `SUM(alias, expression)` | Sums values in a group | `SUM('TotalBalance', @('Balance')):` |
| `AVG(alias, expression)` | Averages values in a group | `AVG('AverageBalance', @('Balance')):` |
| `MIN(alias, expression)` | Finds minimum value in a group | `MIN('MinBalance', @('Balance')):` |
| `MAX(alias, expression)` | Finds maximum value in a group | `MAX('MaxBalance', @('Balance')):` |

## Order By Operators

These operators are used for sorting results:

| Operator | Description | Example |
|----------|-------------|---------|
| `ASC` | Sorts in ascending order | `CreateDate: ASC;` |
| `DESC` | Sorts in descending order | `CreateDate: DESC;` |

## Built-in Functions

These functions provide additional functionality:

| Function | Description | Example |
|----------|-------------|---------|
| `len(string)` | Gets the length of a string | `len(@('Name')) > 3;` |
| `cast(expression, type)` | Casts an expression to a different type | `cast(@('CustomerNumber'), 'varchar')` |
| `timezone(date, zone)` | Applies a timezone to a date | `timezone(@('CreateDate'), 'UTC')` |
| `datepart(part, date)` | Extracts a part from a date | `datepart('year', @('CreateDate'))` |
| `dateadd(interval, number, date)` | Adds a time interval to a date | `dateadd('day', 7, @('CreateDate'))` |
| `convert(type, expression, [style])` | Converts a value to a different type | `convert('date', @('CreateDate'), 23)` |
| `date(string)` | Converts a string to a date | `date('2025/04/03')` |
| `datetime(string)` | Converts a string to a datetime | `datetime('2025-04-03T14:30:00')` |
| `like(column, pattern)` | Performs a SQL LIKE comparison | `like(@('Name'), '%Smith%')` |