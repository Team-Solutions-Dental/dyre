package endpoint

import (
	"testing"
)

func TestNormalizeSecurityValue_String(t *testing.T) {
	policy, err := NormalizeSecurityValue("customers.read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy == nil {
		t.Fatal("expected policy, got nil")
	}
	if len(policy.Permissions) != 1 || policy.Permissions[0] != "customers.read" {
		t.Errorf("expected permissions [customers.read], got %v", policy.Permissions)
	}
	if policy.OnDeny != "error" {
		t.Errorf("expected onDeny 'error', got %s", policy.OnDeny)
	}
}

func TestNormalizeSecurityValue_Array(t *testing.T) {
	input := []any{"perm1", "perm2", "perm3"}
	policy, err := NormalizeSecurityValue(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policy.Permissions) != 3 {
		t.Errorf("expected 3 permissions, got %d", len(policy.Permissions))
	}
	if policy.OnDeny != "error" {
		t.Errorf("expected onDeny 'error', got %s", policy.OnDeny)
	}
}

func TestNormalizeSecurityValue_ObjectWithError(t *testing.T) {
	input := map[string]any{
		"permissions": []any{"perm1", "perm2"},
		"onDeny":      "error",
	}
	policy, err := NormalizeSecurityValue(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policy.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(policy.Permissions))
	}
	if policy.OnDeny != "error" {
		t.Errorf("expected onDeny 'error', got %s", policy.OnDeny)
	}
}

func TestNormalizeSecurityValue_ObjectWithOmit(t *testing.T) {
	input := map[string]any{
		"permissions": []any{"sensitive.field.view"},
		"onDeny":      "omit",
	}
	policy, err := NormalizeSecurityValue(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policy.Permissions) != 1 {
		t.Errorf("expected 1 permission, got %d", len(policy.Permissions))
	}
	if policy.OnDeny != "omit" {
		t.Errorf("expected onDeny 'omit', got %s", policy.OnDeny)
	}
}

func TestNormalizeSecurityValue_ObjectDefaultsToError(t *testing.T) {
	input := map[string]any{
		"permissions": []any{"perm1"},
	}
	policy, err := NormalizeSecurityValue(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.OnDeny != "error" {
		t.Errorf("expected onDeny to default to 'error', got %s", policy.OnDeny)
	}
}

func TestNormalizeSecurityValue_InvalidOnDeny(t *testing.T) {
	input := map[string]any{
		"permissions": []any{"perm1"},
		"onDeny":      "invalid",
	}
	_, err := NormalizeSecurityValue(input)
	if err == nil {
		t.Fatal("expected error for invalid onDeny value")
	}
}

func TestNormalizeSecurityValue_MissingPermissions(t *testing.T) {
	input := map[string]any{
		"onDeny": "error",
	}
	_, err := NormalizeSecurityValue(input)
	if err == nil {
		t.Fatal("expected error for missing permissions")
	}
}

func TestNormalizeSecurityValue_Nil(t *testing.T) {
	policy, err := NormalizeSecurityValue(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy != nil {
		t.Error("expected nil policy for nil input")
	}
}

func TestSecurityPolicy_HasWildcard(t *testing.T) {
	tests := []struct {
		name     string
		policy   *SecurityPolicy
		expected bool
	}{
		{
			name:     "nil policy",
			policy:   nil,
			expected: false,
		},
		{
			name: "has wildcard",
			policy: &SecurityPolicy{
				Permissions: []string{"*"},
				OnDeny:      "error",
			},
			expected: true,
		},
		{
			name: "wildcard among others",
			policy: &SecurityPolicy{
				Permissions: []string{"perm1", "*", "perm2"},
				OnDeny:      "error",
			},
			expected: true,
		},
		{
			name: "no wildcard",
			policy: &SecurityPolicy{
				Permissions: []string{"perm1", "perm2"},
				OnDeny:      "error",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.policy.HasWildcard(); got != tt.expected {
				t.Errorf("HasWildcard() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestStaticChecker(t *testing.T) {
	grants := map[string]struct{}{
		"read":  {},
		"write": {},
	}
	checker := NewStaticChecker(grants)

	tests := []struct {
		name     string
		required []string
		allowed  bool
	}{
		{"all granted", []string{"read", "write"}, true},
		{"one granted", []string{"read"}, true},
		{"none granted", []string{"delete"}, false},
		{"some granted", []string{"read", "delete"}, false},
		{"empty required", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := checker.Allow(tt.required)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if allowed != tt.allowed {
				t.Errorf("Allow(%v) = %v, expected %v", tt.required, allowed, tt.allowed)
			}
		})
	}
}

func TestPermissiveChecker(t *testing.T) {
	checker := NewPermissiveChecker()
	allowed, err := checker.Allow([]string{"any", "permissions"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("PermissiveChecker should allow all permissions")
	}
}
