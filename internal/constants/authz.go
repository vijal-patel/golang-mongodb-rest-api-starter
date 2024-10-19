package constants

const (
	CreateOrganizationPermission string = "organization:create"
	ReadOrganizationPermission   string = "organization:read"
	UpdateOrganizationPermission string = "organization:update"
	DeleteOrganizationPermission string = "organization:delete"

	OrganizationAdminRole string = "organizationAdmin"
)

const (
	CreateUserPermission string = "user:create"
	ReadUserPermission   string = "user:read"
	UpdateUserPermission string = "user:update"
	DeleteUserPermission string = "user:delete"

	BaseUserRole string = "baseUser"
)

const (
	SuperUserRole string = "superUser"
)

func GetAllowedRoles() []string {
	return []string{BaseUserRole, OrganizationAdminRole}
}
