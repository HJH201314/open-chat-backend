package constants

import (
	"time"
)

const RefreshTokenExpire = time.Hour * 24 * 7

var AuthIgnoredKey = "auth_middleware_ignore"
var PermissionSuperAdminKey = "permission_is_super_admin"
