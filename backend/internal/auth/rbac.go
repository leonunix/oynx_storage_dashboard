package auth

import "github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"

var rolePermissions = map[string][]domain.Permission{
	"admin": {
		domain.PermissionOverviewRead,
		domain.PermissionMetricsRead,
		domain.PermissionVolumesRead,
		domain.PermissionVolumesWrite,
		domain.PermissionStorageRead,
		domain.PermissionStorageWrite,
		domain.PermissionAuditRead,
		domain.PermissionUsersManage,
	},
	"operator": {
		domain.PermissionOverviewRead,
		domain.PermissionMetricsRead,
		domain.PermissionVolumesRead,
		domain.PermissionVolumesWrite,
		domain.PermissionStorageRead,
		domain.PermissionStorageWrite,
		domain.PermissionAuditRead,
	},
	"viewer": {
		domain.PermissionOverviewRead,
		domain.PermissionMetricsRead,
		domain.PermissionVolumesRead,
		domain.PermissionStorageRead,
	},
}

func RoleDefinitions() []domain.RoleDefinition {
	out := make([]domain.RoleDefinition, 0, len(rolePermissions))
	for role, permissions := range rolePermissions {
		out = append(out, domain.RoleDefinition{
			Name:        role,
			Permissions: permissions,
		})
	}
	return out
}

func PermissionsForRole(role string) []domain.Permission {
	if permissions, ok := rolePermissions[role]; ok {
		return permissions
	}
	return rolePermissions["viewer"]
}

func HasPermission(user domain.User, permission domain.Permission) bool {
	for _, candidate := range user.Permissions {
		if candidate == permission {
			return true
		}
	}
	return false
}
