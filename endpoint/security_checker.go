package endpoint

// StaticChecker is a simple implementation that checks permissions against
// a fixed set. Useful for testing and simple use cases.
type StaticChecker struct {
	grants map[string]struct{}
}

// NewStaticChecker creates a SecurityChecker with a fixed set of granted permissions
func NewStaticChecker(grants map[string]struct{}) *StaticChecker {
	return &StaticChecker{grants: grants}
}

// Allow returns true if all required permissions are in the grants set
func (sc *StaticChecker) Allow(required []string) (bool, error) {
	if len(required) == 0 {
		return true, nil
	}

	for _, perm := range required {
		if _, ok := sc.grants[perm]; !ok {
			return false, nil
		}
	}
	return true, nil
}

// RoleChecker is an implementation that supports role expansion via a callback.
// The callback can implement custom logic like treating "admin" as granting all permissions.
type RoleChecker struct {
	checkFunc func(required []string) (bool, error)
}

// NewRoleChecker creates a SecurityChecker that uses a custom function
// to evaluate permissions. The function should return:
//   - (true, nil) if all permissions are granted
//   - (false, nil) if permissions are missing
//   - (false, error) if the check itself fails
func NewRoleChecker(checkFunc func(required []string) (bool, error)) *RoleChecker {
	return &RoleChecker{checkFunc: checkFunc}
}

// Allow delegates to the custom check function
func (rc *RoleChecker) Allow(required []string) (bool, error) {
	return rc.checkFunc(required)
}

// PermissiveChecker is a checker that allows all permissions.
// Useful for testing or when security is handled elsewhere.
type PermissiveChecker struct{}

// NewPermissiveChecker creates a checker that allows everything
func NewPermissiveChecker() *PermissiveChecker {
	return &PermissiveChecker{}
}

// Allow always returns true
func (pc *PermissiveChecker) Allow(required []string) (bool, error) {
	return true, nil
}
