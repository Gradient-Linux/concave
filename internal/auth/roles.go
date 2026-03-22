package auth

import (
	"encoding/json"
	"fmt"
	"os/user"
	"sort"
)

// Role represents a permission tier.
type Role int

const (
	RoleViewer Role = iota
	RoleDeveloper
	RoleOperator
	RoleAdmin
)

var (
	groupToRole = map[string]Role{
		"gradient-admin":     RoleAdmin,
		"gradient-operator":  RoleOperator,
		"gradient-developer": RoleDeveloper,
		"gradient-viewer":    RoleViewer,
	}
	roleToGroup = map[Role]string{
		RoleViewer:    "gradient-viewer",
		RoleDeveloper: "gradient-developer",
		RoleOperator:  "gradient-operator",
		RoleAdmin:     "gradient-admin",
	}
	lookupUser      = user.Lookup
	lookupGroupID   = user.LookupGroupId
	lookupGroupIDs  = func(u *user.User) ([]string, error) { return u.GroupIds() }
	userCurrentFunc = user.Current
)

func (r Role) String() string {
	switch r {
	case RoleViewer:
		return "viewer"
	case RoleDeveloper:
		return "developer"
	case RoleOperator:
		return "operator"
	case RoleAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

// MarshalJSON keeps API responses stable and human-readable.
func (r Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON accepts both string and integer encodings for compatibility.
func (r *Role) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		parsed, err := ParseRole(asString)
		if err != nil {
			return err
		}
		*r = parsed
		return nil
	}

	var asInt int
	if err := json.Unmarshal(data, &asInt); err != nil {
		return err
	}
	*r = Role(asInt)
	return nil
}

// ParseRole converts an API/session string into a Role.
func ParseRole(value string) (Role, error) {
	switch value {
	case "viewer":
		return RoleViewer, nil
	case "developer":
		return RoleDeveloper, nil
	case "operator":
		return RoleOperator, nil
	case "admin":
		return RoleAdmin, nil
	default:
		return 0, fmt.Errorf("unknown role %q", value)
	}
}

// ResolveRole returns the highest Gradient role for a Unix username.
func ResolveRole(username string) (Role, error) {
	u, err := lookupUser(username)
	if err != nil {
		return 0, fmt.Errorf("user not found: %s", username)
	}

	gids, err := lookupGroupIDs(u)
	if err != nil {
		return 0, fmt.Errorf("could not read groups for user: %s", username)
	}

	highest := Role(-1)
	for _, gid := range gids {
		group, err := lookupGroupID(gid)
		if err != nil {
			continue
		}
		role, ok := groupToRole[group.Name]
		if ok && role > highest {
			highest = role
		}
	}
	if highest < 0 {
		return 0, fmt.Errorf(
			"user %s is not in any gradient-* group. Add them to gradient-admin, gradient-operator, gradient-developer, or gradient-viewer",
			username,
		)
	}
	return highest, nil
}

// GradientGroups returns the matching gradient-* groups for a user.
func GradientGroups(username string) ([]string, error) {
	u, err := lookupUser(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	gids, err := lookupGroupIDs(u)
	if err != nil {
		return nil, fmt.Errorf("could not read groups for user: %s", username)
	}
	groups := make([]string, 0, 4)
	for _, gid := range gids {
		group, err := lookupGroupID(gid)
		if err != nil {
			continue
		}
		if _, ok := groupToRole[group.Name]; ok {
			groups = append(groups, group.Name)
		}
	}
	sort.Strings(groups)
	return groups, nil
}

// RoleGroup returns the canonical Unix group name for a role.
func RoleGroup(role Role) string {
	return roleToGroup[role]
}
