package authorization

import "errors"

// Role represents a user role in the system
type Role struct {
	Name        string
	Permissions []string
}

// RBAC handles role-based access control
type RBAC struct {
	roles map[string]Role
}

// NewRBAC creates a new RBAC instance
func NewRBAC() *RBAC {
	return &RBAC{
		roles: make(map[string]Role),
	}
}

// AddRole adds a new role with permissions
func (r *RBAC) AddRole(name string, permissions []string) error {
	if _, exists := r.roles[name]; exists {
		return errors.New("role already exists")
	}
	r.roles[name] = Role{
		Name:        name,
		Permissions: permissions,
	}
	return nil
}

// HasPermission checks if a role has a specific permission
func (r *RBAC) HasPermission(role, permission string) bool {
	if r, exists := r.roles[role]; exists {
		for _, p := range r.Permissions {
			if p == permission {
				return true
			}
		}
	}
	return false
}

// GetRolePermissions returns all permissions for a role
func (r *RBAC) GetRolePermissions(role string) ([]string, error) {
	if r, exists := r.roles[role]; exists {
		return r.Permissions, nil
	}
	return nil, errors.New("role not found")
} 