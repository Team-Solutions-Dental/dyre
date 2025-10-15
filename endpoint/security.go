package endpoint

import (
	"errors"
	"fmt"
)

// SecurityPolicy represents normalized security metadata with permissions and denial behavior
type SecurityPolicy struct {
	Permissions []string
	OnDeny      string // "error" or "omit"
}

// HasWildcard returns true if the policy contains the wildcard permission "*"
func (sp *SecurityPolicy) HasWildcard() bool {
	if sp == nil {
		return false
	}
	for _, p := range sp.Permissions {
		if p == "*" {
			return true
		}
	}
	return false
}

// IsEmpty returns true if the policy has no permissions defined
func (sp *SecurityPolicy) IsEmpty() bool {
	return sp == nil || len(sp.Permissions) == 0
}

// SecurityChecker is the interface for checking permissions at runtime.
// Implementations should handle role expansion, aggregated permissions,
// and integration with the host application's auth system.
type SecurityChecker interface {
	// Allow returns (true, nil) when all required permissions are granted.
	// Returns (false, nil) when permissions are missing (triggers omit or error based on policy).
	// Returns error when auth check itself fails (e.g., upstream auth service failure).
	Allow(required []string) (bool, error)
}

// NormalizeSecurityValue converts string, array, or object security values
// to a SecurityPolicy struct
func NormalizeSecurityValue(value any) (*SecurityPolicy, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		// String shorthand: "customers.read" -> {permissions: ["customers.read"], onDeny: "error"}
		perms, err := parseSecurityList(v)
		if err != nil {
			return nil, err
		}
		return &SecurityPolicy{
			Permissions: perms,
			OnDeny:      "error",
		}, nil

	case []any:
		// Array shorthand: ["perm1", "perm2"] -> {permissions: ["perm1", "perm2"], onDeny: "error"}
		perms, err := parseSecurityList(v)
		if err != nil {
			return nil, err
		}
		return &SecurityPolicy{
			Permissions: perms,
			OnDeny:      "error",
		}, nil

	case map[string]any:
		// Object form: {permissions: [...], onDeny: "omit"}
		return parseSecurityObject(v)

	default:
		return nil, fmt.Errorf("security value has invalid type. got=%T", value)
	}
}

// parseSecurityObject parses the object form of security metadata
func parseSecurityObject(m map[string]any) (*SecurityPolicy, error) {
	policy := &SecurityPolicy{
		OnDeny: "error", // default
	}

	// Parse permissions (required)
	permsAny, ok := m["permissions"]
	if !ok {
		return nil, errors.New("security object missing 'permissions' field")
	}

	perms, err := parseSecurityList(permsAny)
	if err != nil {
		return nil, fmt.Errorf("security.permissions: %w", err)
	}
	policy.Permissions = perms

	// Parse onDeny (optional)
	if onDenyAny, ok := m["onDeny"]; ok {
		onDeny, ok := onDenyAny.(string)
		if !ok {
			return nil, fmt.Errorf("security.onDeny not string. got=%T", onDenyAny)
		}
		if onDeny != "error" && onDeny != "omit" {
			return nil, fmt.Errorf("security.onDeny must be 'error' or 'omit'. got=%s", onDeny)
		}
		policy.OnDeny = onDeny
	}

	// Validate no unexpected keys
	expectedKeys := []string{"permissions", "onDeny"}
	for key := range m {
		found := false
		for _, expected := range expectedKeys {
			if key == expected {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unexpected key in security object: %s", key)
		}
	}

	return policy, nil
}
