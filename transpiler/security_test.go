package transpiler

import (
	"strings"
	"testing"

	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/object/objectType"
)

// Helper to create a test endpoint with security
func createTestEndpointWithSecurity() *endpoint.Endpoint {
	service := &endpoint.Service{
		Settings: endpoint.Settings{BracketedColumns: true},
	}

	ep := &endpoint.Endpoint{
		Service:    service,
		Name:       "Customers",
		TableName:  "Customers",
		SchemaName: "dbo",
		Security: &endpoint.SecurityPolicy{
			Permissions: []string{"customers.read"},
			OnDeny:      "error",
		},
		Fields: map[string]endpoint.Field{
			"CustomerID": {
				Name:      "CustomerID",
				FieldType: objectType.INTEGER,
				Nullable:  false,
				Security: &endpoint.SecurityPolicy{
					Permissions: []string{"customers.customerid.view"},
					OnDeny:      "error",
				},
			},
			"Email": {
				Name:      "Email",
				FieldType: objectType.STRING,
				Nullable:  true,
				Security: &endpoint.SecurityPolicy{
					Permissions: []string{"customers.email.view"},
					OnDeny:      "omit",
				},
			},
			"Name": {
				Name:      "Name",
				FieldType: objectType.STRING,
				Nullable:  true,
				// No security - inherits from endpoint
			},
			"Notes": {
				Name:      "Notes",
				FieldType: objectType.STRING,
				Nullable:  true,
				Security: &endpoint.SecurityPolicy{
					Permissions: []string{"*"},
					OnDeny:      "error",
				},
			},
		},
		FieldNames: []string{"CustomerID", "Email", "Name", "Notes"},
	}

	// Set endpoint reference in fields
	for name, field := range ep.Fields {
		field.Endpoint = ep
		ep.Fields[name] = field
	}

	return ep
}

func TestEndpointSecurity_Success(t *testing.T) {
	ep := createTestEndpointWithSecurity()
	checker := endpoint.NewStaticChecker(map[string]struct{}{
		"customers.read":            {},
		"customers.customerid.view": {},
	})

	ir, err := NewWithSecurity("CustomerID:", ep, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(sql, "CustomerID") {
		t.Errorf("expected CustomerID in query, got: %s", sql)
	}
}

func TestEndpointSecurity_DeniedWithError(t *testing.T) {
	ep := createTestEndpointWithSecurity()
	checker := endpoint.NewStaticChecker(map[string]struct{}{}) // no permissions

	_, err := NewWithSecurity("CustomerID:", ep, checker)
	if err == nil {
		t.Fatal("expected permission denied error")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected 'permission denied' error, got: %v", err)
	}
}

func TestEndpointSecurity_DeniedWithOmit(t *testing.T) {
	ep := createTestEndpointWithSecurity()
	ep.Security.OnDeny = "omit"

	checker := endpoint.NewStaticChecker(map[string]struct{}{}) // no permissions

	ir, err := NewWithSecurity("CustomerID:", ep, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error during evaluation: %v", err)
	}

	// Should return empty or minimal query since endpoint access is denied
	if sql == "" {
		t.Log("Empty SQL as expected for omitted endpoint")
	}
}

func TestFieldSecurity_OmittedColumn(t *testing.T) {
	ep := createTestEndpointWithSecurity()
	checker := endpoint.NewStaticChecker(map[string]struct{}{
		"customers.read":            {},
		"customers.customerid.view": {},
		// Note: customers.email.view is NOT granted
	})

	ir, err := NewWithSecurity("CustomerID:Email:", ep, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have CustomerID but not Email (omitted due to lack of permission)
	if !strings.Contains(sql, "CustomerID") {
		t.Error("expected CustomerID in query")
	}
	if strings.Contains(sql, "Email") {
		t.Error("expected Email to be omitted from query")
	}

	// Check FieldNames reflects the omission
	fields := ir.FieldNames()
	if len(fields) != 1 || fields[0] != "CustomerID" {
		t.Errorf("expected FieldNames to be [CustomerID], got %v", fields)
	}
}

func TestFieldSecurity_ErrorOnDeny(t *testing.T) {
	ep := createTestEndpointWithSecurity()
	checker := endpoint.NewStaticChecker(map[string]struct{}{
		"customers.read": {},
		// Note: customers.customerid.view is NOT granted, and it has onDeny="error"
	})

	ir, err := NewWithSecurity("CustomerID:", ep, checker)
	if err != nil {
		t.Fatalf("unexpected error during creation: %v", err)
	}

	_, err = ir.EvaluateQuery()
	if err == nil {
		t.Fatal("expected permission denied error for CustomerID")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected 'permission denied' error, got: %v", err)
	}
}

func TestFieldSecurity_InheritsFromEndpoint(t *testing.T) {
	ep := createTestEndpointWithSecurity()
	checker := endpoint.NewStaticChecker(map[string]struct{}{
		"customers.read": {}, // endpoint permission granted
		// Name field has no security, so it inherits from endpoint
	})

	ir, err := NewWithSecurity("Name:", ep, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(sql, "Name") {
		t.Errorf("expected Name in query (should inherit endpoint permission), got: %s", sql)
	}
}

func TestFieldSecurity_WildcardAlwaysAllowed(t *testing.T) {
	ep := createTestEndpointWithSecurity()
	checker := endpoint.NewStaticChecker(map[string]struct{}{
		"customers.read": {},
		// Notes field has wildcard "*", so it should be allowed without explicit grant
	})

	ir, err := NewWithSecurity("Notes:", ep, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(sql, "Notes") {
		t.Errorf("expected Notes in query (wildcard should allow), got: %s", sql)
	}
}

func TestSecurity_NilChecker(t *testing.T) {
	ep := createTestEndpointWithSecurity()

	// Nil checker should allow everything (backward compatible)
	ir, err := NewWithSecurity("CustomerID:Email:Name:Notes:", ep, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All fields should be present
	requiredFields := []string{"CustomerID", "Email", "Name", "Notes"}
	for _, field := range requiredFields {
		if !strings.Contains(sql, field) {
			t.Errorf("expected %s in query with nil checker, got: %s", field, sql)
		}
	}
}

func TestSecurity_PermissiveChecker(t *testing.T) {
	ep := createTestEndpointWithSecurity()
	checker := endpoint.NewPermissiveChecker()

	ir, err := NewWithSecurity("CustomerID:Email:Name:Notes:", ep, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All fields should be present with permissive checker
	requiredFields := []string{"CustomerID", "Email", "Name", "Notes"}
	for _, field := range requiredFields {
		if !strings.Contains(sql, field) {
			t.Errorf("expected %s in query with permissive checker, got: %s", field, sql)
		}
	}
}

func TestSecurity_RoleChecker(t *testing.T) {
	ep := createTestEndpointWithSecurity()

	// Admin role grants all permissions
	checker := endpoint.NewRoleChecker(func(required []string) (bool, error) {
		// Simulate admin role that grants everything
		return true, nil
	})

	ir, err := NewWithSecurity("CustomerID:Email:Name:Notes:", ep, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All fields should be present with admin role
	requiredFields := []string{"CustomerID", "Email", "Name", "Notes"}
	for _, field := range requiredFields {
		if !strings.Contains(sql, field) {
			t.Errorf("expected %s in query with admin role, got: %s", field, sql)
		}
	}
}

// Helper to create service with joined endpoints for testing
func createServiceWithJoins() *endpoint.Service {
	service := &endpoint.Service{
		Settings: endpoint.Settings{BracketedColumns: true},
	}

	customers := &endpoint.Endpoint{
		Service:    service,
		Name:       "Customers",
		TableName:  "Customers",
		SchemaName: "dbo",
		Security: &endpoint.SecurityPolicy{
			Permissions: []string{"customers.read"},
			OnDeny:      "error",
		},
		Fields: map[string]endpoint.Field{
			"CustomerID": {
				Name:      "CustomerID",
				FieldType: objectType.INTEGER,
				Nullable:  false,
			},
			"Name": {
				Name:      "Name",
				FieldType: objectType.STRING,
				Nullable:  true,
			},
		},
		FieldNames: []string{"CustomerID", "Name"},
	}

	invoices := &endpoint.Endpoint{
		Service:    service,
		Name:       "Invoices",
		TableName:  "Invoices",
		SchemaName: "dbo",
		Security: &endpoint.SecurityPolicy{
			Permissions: []string{"invoices.read"},
			OnDeny:      "error",
		},
		Fields: map[string]endpoint.Field{
			"InvoiceID": {
				Name:      "InvoiceID",
				FieldType: objectType.INTEGER,
				Nullable:  false,
			},
			"CustomerID": {
				Name:      "CustomerID",
				FieldType: objectType.INTEGER,
				Nullable:  false,
			},
			"Amount": {
				Name:      "Amount",
				FieldType: objectType.FLOAT,
				Nullable:  true,
				Security: &endpoint.SecurityPolicy{
					Permissions: []string{"invoices.amount.view"},
					OnDeny:      "omit",
				},
			},
		},
		FieldNames: []string{"InvoiceID", "CustomerID", "Amount"},
	}

	// Set endpoint references in fields
	for name, field := range customers.Fields {
		field.Endpoint = customers
		customers.Fields[name] = field
	}
	for name, field := range invoices.Fields {
		field.Endpoint = invoices
		invoices.Fields[name] = field
	}

	// Setup join from Customers to Invoices
	join := endpoint.Join{}
	join.Parent_ON = "CustomerID"
	join.Child_ON = "CustomerID"
	customers.Joins = map[string]endpoint.Join{
		"Invoices": join,
	}
	customers.JoinNames = []string{"Invoices"}

	service.Endpoints = map[string]*endpoint.Endpoint{
		"Customers": customers,
		"Invoices":  invoices,
	}
	service.EndpointNames = []string{"Customers", "Invoices"}

	return service
}

func TestJoinSecurity_PropagatesChecker(t *testing.T) {
	service := createServiceWithJoins()
	customersEp := service.Endpoints["Customers"]

	// Grant access to customers but NOT invoices
	checker := endpoint.NewStaticChecker(map[string]struct{}{
		"customers.read": {},
	})

	ir, err := NewWithSecurity("CustomerID:", customersEp, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to join Invoices without permission
	joinIR := ir.INNERJOIN("Invoices").ON("CustomerID", "CustomerID")
	_, err = joinIR.Query("InvoiceID:")

	if err == nil {
		t.Fatal("expected permission denied error for joined endpoint")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected 'permission denied' error, got: %v", err)
	}
}

func TestJoinSecurity_AllowedJoin(t *testing.T) {
	service := createServiceWithJoins()
	customersEp := service.Endpoints["Customers"]

	// Grant access to both customers and invoices
	checker := endpoint.NewStaticChecker(map[string]struct{}{
		"customers.read": {},
		"invoices.read":  {},
	})

	ir, err := NewWithSecurity("CustomerID:", customersEp, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joinIR := ir.INNERJOIN("Invoices").ON("CustomerID", "CustomerID")
	_, err = joinIR.Query("InvoiceID:")
	if err != nil {
		t.Fatalf("unexpected error on allowed join: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(sql, "INNER JOIN") {
		t.Error("expected INNER JOIN in query")
	}
	if !strings.Contains(sql, "Invoices") {
		t.Error("expected Invoices table in query")
	}
}

func TestJoinSecurity_FieldOmissionInJoin(t *testing.T) {
	service := createServiceWithJoins()
	customersEp := service.Endpoints["Customers"]

	// Grant access to customers and invoices, but NOT to invoices.amount.view
	checker := endpoint.NewStaticChecker(map[string]struct{}{
		"customers.read": {},
		"invoices.read":  {},
		// Note: invoices.amount.view is NOT granted
	})

	ir, err := NewWithSecurity("CustomerID:", customersEp, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joinIR := ir.INNERJOIN("Invoices").ON("CustomerID", "CustomerID")
	_, err = joinIR.Query("InvoiceID:Amount:")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have InvoiceID but not Amount (omitted)
	if !strings.Contains(sql, "InvoiceID") {
		t.Error("expected InvoiceID in query")
	}
	if strings.Contains(sql, "Amount") {
		t.Error("expected Amount to be omitted from query")
	}
}

func TestJoinSecurity_NilCheckerInJoin(t *testing.T) {
	service := createServiceWithJoins()
	customersEp := service.Endpoints["Customers"]

	// Nil checker should allow everything in joins too
	ir, err := NewWithSecurity("CustomerID:", customersEp, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joinIR := ir.INNERJOIN("Invoices").ON("CustomerID", "CustomerID")
	_, err = joinIR.Query("InvoiceID:Amount:")
	if err != nil {
		t.Fatalf("unexpected error with nil checker: %v", err)
	}

	sql, err := ir.EvaluateQuery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both fields should be present
	if !strings.Contains(sql, "InvoiceID") {
		t.Error("expected InvoiceID in query")
	}
	if !strings.Contains(sql, "Amount") {
		t.Error("expected Amount in query with nil checker")
	}
}
