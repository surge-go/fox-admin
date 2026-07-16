package errcode

import (
	"net/http"

	"github.com/surge-go/fox/core/errors"
)

var (
	// ErrMenuCreateReqNil 表示创建菜单请求参数为空。
	ErrMenuCreateReqNil = errors.NewWithStatus(1001, http.StatusOK, "创建菜单参数不能为空")
	// ErrMenuParentIDInvalid 表示父菜单 ID 非法。
	ErrMenuParentIDInvalid = errors.NewWithStatus(1002, http.StatusOK, "父菜单 ID 不能小于 0")
	// ErrMenuPathRequired 表示菜单路径为空。
	ErrMenuPathRequired = errors.NewWithStatus(1003, http.StatusOK, "菜单路径不能为空")
	// ErrMenuNameRequired 表示菜单名称为空。
	ErrMenuNameRequired = errors.NewWithStatus(1004, http.StatusOK, "菜单名称不能为空")
	// ErrMenuTypeRequired 表示菜单类型为空。
	ErrMenuTypeRequired = errors.NewWithStatus(1005, http.StatusOK, "菜单类型不能为空")
	// ErrMenuTitleRequired 表示菜单标题为空。
	ErrMenuTitleRequired = errors.NewWithStatus(1006, http.StatusOK, "菜单标题不能为空")
	// ErrMenuSortInvalid 表示菜单排序值非法。
	ErrMenuSortInvalid = errors.NewWithStatus(1007, http.StatusOK, "菜单排序不能小于 0")
	// ErrMenuStatusRequired 表示菜单状态为空。
	ErrMenuStatusRequired = errors.NewWithStatus(1008, http.StatusOK, "菜单状态不能为空")
	// ErrMenuTypeInvalid 表示菜单类型非法。
	ErrMenuTypeInvalid = errors.NewWithStatus(1009, http.StatusOK, "菜单类型非法")
	// ErrMenuParentNotFound 表示父菜单不存在。
	ErrMenuParentNotFound = errors.NewWithStatus(1010, http.StatusOK, "父菜单不存在")
	// ErrMenuPathExists 表示菜单路径已存在。
	ErrMenuPathExists = errors.NewWithStatus(1011, http.StatusOK, "菜单路径已存在")
	// ErrMenuNameExists 表示菜单名称已存在。
	ErrMenuNameExists = errors.NewWithStatus(1012, http.StatusOK, "菜单名称已存在")
	// ErrMenuParentQueryFailed 表示查询父菜单失败。
	ErrMenuParentQueryFailed = errors.NewWithStatus(1013, http.StatusInternalServerError, "查询父菜单失败")
	// ErrMenuPathQueryFailed 表示查询菜单路径失败。
	ErrMenuPathQueryFailed = errors.NewWithStatus(1014, http.StatusInternalServerError, "查询菜单路径失败")
	// ErrMenuNameQueryFailed 表示查询菜单名称失败。
	ErrMenuNameQueryFailed = errors.NewWithStatus(1015, http.StatusInternalServerError, "查询菜单名称失败")
	// ErrMenuCreateFailed 表示创建菜单失败。
	ErrMenuCreateFailed = errors.NewWithStatus(1016, http.StatusInternalServerError, "创建菜单失败")
	// ErrMenuDeleteReqNil 表示删除菜单请求参数为空。
	ErrMenuDeleteReqNil = errors.NewWithStatus(1017, http.StatusOK, "删除菜单参数不能为空")
	// ErrMenuIDInvalid 表示菜单 ID 非法。
	ErrMenuIDInvalid = errors.NewWithStatus(1018, http.StatusOK, "菜单 ID 必须大于 0")
	// ErrMenuNotFound 表示菜单不存在。
	ErrMenuNotFound = errors.NewWithStatus(1019, http.StatusOK, "菜单不存在")
	// ErrMenuHasChildren 表示菜单存在子菜单。
	ErrMenuHasChildren = errors.NewWithStatus(1020, http.StatusOK, "菜单存在子菜单，不能删除")
	// ErrMenuHasRoleBinding 表示菜单已绑定角色。
	ErrMenuHasRoleBinding = errors.NewWithStatus(1021, http.StatusOK, "菜单已绑定角色，不能删除")
	// ErrMenuQueryFailed 表示查询菜单失败。
	ErrMenuQueryFailed = errors.NewWithStatus(1022, http.StatusInternalServerError, "查询菜单失败")
	// ErrMenuChildrenQueryFailed 表示查询子菜单失败。
	ErrMenuChildrenQueryFailed = errors.NewWithStatus(1023, http.StatusInternalServerError, "查询子菜单失败")
	// ErrMenuRoleBindingQueryFailed 表示查询菜单角色绑定失败。
	ErrMenuRoleBindingQueryFailed = errors.NewWithStatus(1024, http.StatusInternalServerError, "查询菜单角色绑定失败")
	// ErrMenuDeleteFailed 表示删除菜单失败。
	ErrMenuDeleteFailed = errors.NewWithStatus(1025, http.StatusInternalServerError, "删除菜单失败")
	// ErrMenuUpdateReqNil 表示更新菜单请求参数为空。
	ErrMenuUpdateReqNil = errors.NewWithStatus(1026, http.StatusOK, "更新菜单参数不能为空")
	// ErrMenuParentSelf 表示父菜单不能是自己。
	ErrMenuParentSelf = errors.NewWithStatus(1027, http.StatusOK, "父菜单不能是当前菜单")
	// ErrMenuParentDescendant 表示父菜单不能是当前菜单的子孙菜单。
	ErrMenuParentDescendant = errors.NewWithStatus(1028, http.StatusOK, "父菜单不能是当前菜单的子孙菜单")
	// ErrMenuUpdateFailed 表示更新菜单失败。
	ErrMenuUpdateFailed = errors.NewWithStatus(1029, http.StatusInternalServerError, "更新菜单失败")
	// ErrMenuTreeQueryFailed 表示查询菜单树失败。
	ErrMenuTreeQueryFailed = errors.NewWithStatus(1030, http.StatusInternalServerError, "查询菜单树失败")
	// ErrMenuDetailReqNil 表示菜单详情请求参数为空。
	ErrMenuDetailReqNil = errors.NewWithStatus(1031, http.StatusOK, "菜单详情参数不能为空")
	// ErrRoleCreateReqNil 表示创建角色请求参数为空。
	ErrRoleCreateReqNil = errors.NewWithStatus(1032, http.StatusOK, "创建角色参数不能为空")
	// ErrRoleNameRequired 表示角色名称为空。
	ErrRoleNameRequired = errors.NewWithStatus(1033, http.StatusOK, "角色名称不能为空")
	// ErrRoleCodeRequired 表示角色编码为空。
	ErrRoleCodeRequired = errors.NewWithStatus(1034, http.StatusOK, "角色编码不能为空")
	// ErrRoleDataScopeRequired 表示角色数据权限范围为空。
	ErrRoleDataScopeRequired = errors.NewWithStatus(1035, http.StatusOK, "角色数据权限范围不能为空")
	// ErrRoleDataScopeInvalid 表示角色数据权限范围非法。
	ErrRoleDataScopeInvalid = errors.NewWithStatus(1036, http.StatusOK, "角色数据权限范围非法")
	// ErrRoleSortInvalid 表示角色排序值非法。
	ErrRoleSortInvalid = errors.NewWithStatus(1037, http.StatusOK, "角色排序不能小于 0")
	// ErrRoleStatusRequired 表示角色状态为空。
	ErrRoleStatusRequired = errors.NewWithStatus(1038, http.StatusOK, "角色状态不能为空")
	// ErrRoleDeptIDInvalid 表示角色部门 ID 非法。
	ErrRoleDeptIDInvalid = errors.NewWithStatus(1039, http.StatusOK, "角色部门 ID 必须大于 0")
	// ErrRoleMenuIDInvalid 表示角色菜单 ID 非法。
	ErrRoleMenuIDInvalid = errors.NewWithStatus(1040, http.StatusOK, "角色菜单 ID 必须大于 0")
	// ErrRoleNameExists 表示角色名称已存在。
	ErrRoleNameExists = errors.NewWithStatus(1041, http.StatusOK, "角色名称已存在")
	// ErrRoleCodeExists 表示角色编码已存在。
	ErrRoleCodeExists = errors.NewWithStatus(1042, http.StatusOK, "角色编码已存在")
	// ErrRoleDeptNotFound 表示角色绑定部门不存在。
	ErrRoleDeptNotFound = errors.NewWithStatus(1043, http.StatusOK, "角色绑定部门不存在")
	// ErrRoleMenuNotFound 表示角色绑定菜单不存在。
	ErrRoleMenuNotFound = errors.NewWithStatus(1044, http.StatusOK, "角色绑定菜单不存在")
	// ErrRoleNameQueryFailed 表示查询角色名称失败。
	ErrRoleNameQueryFailed = errors.NewWithStatus(1045, http.StatusInternalServerError, "查询角色名称失败")
	// ErrRoleCodeQueryFailed 表示查询角色编码失败。
	ErrRoleCodeQueryFailed = errors.NewWithStatus(1046, http.StatusInternalServerError, "查询角色编码失败")
	// ErrRoleDeptQueryFailed 表示查询角色部门失败。
	ErrRoleDeptQueryFailed = errors.NewWithStatus(1047, http.StatusInternalServerError, "查询角色部门失败")
	// ErrRoleMenuQueryFailed 表示查询角色菜单失败。
	ErrRoleMenuQueryFailed = errors.NewWithStatus(1048, http.StatusInternalServerError, "查询角色菜单失败")
	// ErrRoleCreateFailed 表示创建角色失败。
	ErrRoleCreateFailed = errors.NewWithStatus(1049, http.StatusInternalServerError, "创建角色失败")
	// ErrRoleDeleteReqNil 表示删除角色请求参数为空。
	ErrRoleDeleteReqNil = errors.NewWithStatus(1050, http.StatusOK, "删除角色参数不能为空")
	// ErrRoleIDInvalid 表示角色 ID 非法。
	ErrRoleIDInvalid = errors.NewWithStatus(1051, http.StatusOK, "角色 ID 必须大于 0")
	// ErrRoleNotFound 表示角色不存在。
	ErrRoleNotFound = errors.NewWithStatus(1052, http.StatusOK, "角色不存在")
	// ErrRoleHasUserBinding 表示角色已绑定用户。
	ErrRoleHasUserBinding = errors.NewWithStatus(1053, http.StatusOK, "角色已绑定用户，不能删除")
	// ErrRoleQueryFailed 表示查询角色失败。
	ErrRoleQueryFailed = errors.NewWithStatus(1054, http.StatusInternalServerError, "查询角色失败")
	// ErrRoleUserBindingQueryFailed 表示查询角色用户绑定失败。
	ErrRoleUserBindingQueryFailed = errors.NewWithStatus(1055, http.StatusInternalServerError, "查询角色用户绑定失败")
	// ErrRoleDeleteFailed 表示删除角色失败。
	ErrRoleDeleteFailed = errors.NewWithStatus(1056, http.StatusInternalServerError, "删除角色失败")
	// ErrRoleUpdateReqNil 表示更新角色请求参数为空。
	ErrRoleUpdateReqNil = errors.NewWithStatus(1057, http.StatusOK, "更新角色参数不能为空")
	// ErrRoleUpdateFailed 表示更新角色失败。
	ErrRoleUpdateFailed = errors.NewWithStatus(1058, http.StatusInternalServerError, "更新角色失败")
	// ErrRoleListQueryFailed 表示查询角色列表失败。
	ErrRoleListQueryFailed = errors.NewWithStatus(1059, http.StatusInternalServerError, "查询角色列表失败")
	// ErrRoleDetailReqNil 表示角色详情请求参数为空。
	ErrRoleDetailReqNil = errors.NewWithStatus(1060, http.StatusOK, "角色详情参数不能为空")
	// ErrRoleUpdateStatusReqNil 表示更新角色状态请求参数为空。
	ErrRoleUpdateStatusReqNil = errors.NewWithStatus(1063, http.StatusOK, "更新角色状态参数不能为空")
	// ErrRoleIDsRequired 表示角色 ID 集合为空。
	ErrRoleIDsRequired = errors.NewWithStatus(1064, http.StatusOK, "角色 ID 集合不能为空")
	// ErrRoleUpdateStatusFailed 表示更新角色状态失败。
	ErrRoleUpdateStatusFailed = errors.NewWithStatus(1065, http.StatusInternalServerError, "更新角色状态失败")
	// ErrUserCreateReqNil 表示创建用户请求参数为空。
	ErrUserCreateReqNil = errors.NewWithStatus(1066, http.StatusOK, "创建用户参数不能为空")
	// ErrUserUsernameRequired 表示用户账号为空。
	ErrUserUsernameRequired = errors.NewWithStatus(1067, http.StatusOK, "用户账号不能为空")
	// ErrUserPasswordRequired 表示用户密码为空。
	ErrUserPasswordRequired = errors.NewWithStatus(1068, http.StatusOK, "用户密码不能为空")
	// ErrUserStatusRequired 表示用户状态为空。
	ErrUserStatusRequired = errors.NewWithStatus(1069, http.StatusOK, "用户状态不能为空")
	// ErrUserDeptIDInvalid 表示用户部门 ID 非法。
	ErrUserDeptIDInvalid = errors.NewWithStatus(1070, http.StatusOK, "用户部门 ID 必须大于 0")
	// ErrUserRoleIDInvalid 表示用户角色 ID 非法。
	ErrUserRoleIDInvalid = errors.NewWithStatus(1071, http.StatusOK, "用户角色 ID 必须大于 0")
	// ErrUserPostIDInvalid 表示用户岗位 ID 非法。
	ErrUserPostIDInvalid = errors.NewWithStatus(1072, http.StatusOK, "用户岗位 ID 必须大于 0")
	// ErrUserUsernameExists 表示用户账号已存在。
	ErrUserUsernameExists = errors.NewWithStatus(1073, http.StatusOK, "用户账号已存在")
	// ErrUserEmailExists 表示用户邮箱已存在。
	ErrUserEmailExists = errors.NewWithStatus(1074, http.StatusOK, "用户邮箱已存在")
	// ErrUserPhoneExists 表示用户手机号已存在。
	ErrUserPhoneExists = errors.NewWithStatus(1075, http.StatusOK, "用户手机号已存在")
	// ErrUserDeptNotFound 表示用户所属部门不存在。
	ErrUserDeptNotFound = errors.NewWithStatus(1076, http.StatusOK, "用户所属部门不存在")
	// ErrUserRoleNotFound 表示用户绑定角色不存在。
	ErrUserRoleNotFound = errors.NewWithStatus(1077, http.StatusOK, "用户绑定角色不存在")
	// ErrUserPostNotFound 表示用户绑定岗位不存在。
	ErrUserPostNotFound = errors.NewWithStatus(1078, http.StatusOK, "用户绑定岗位不存在")
	// ErrUserUsernameQueryFailed 表示查询用户账号失败。
	ErrUserUsernameQueryFailed = errors.NewWithStatus(1079, http.StatusInternalServerError, "查询用户账号失败")
	// ErrUserEmailQueryFailed 表示查询用户邮箱失败。
	ErrUserEmailQueryFailed = errors.NewWithStatus(1080, http.StatusInternalServerError, "查询用户邮箱失败")
	// ErrUserPhoneQueryFailed 表示查询用户手机号失败。
	ErrUserPhoneQueryFailed = errors.NewWithStatus(1081, http.StatusInternalServerError, "查询用户手机号失败")
	// ErrUserDeptQueryFailed 表示查询用户部门失败。
	ErrUserDeptQueryFailed = errors.NewWithStatus(1082, http.StatusInternalServerError, "查询用户部门失败")
	// ErrUserRoleQueryFailed 表示查询用户角色失败。
	ErrUserRoleQueryFailed = errors.NewWithStatus(1083, http.StatusInternalServerError, "查询用户角色失败")
	// ErrUserPostQueryFailed 表示查询用户岗位失败。
	ErrUserPostQueryFailed = errors.NewWithStatus(1084, http.StatusInternalServerError, "查询用户岗位失败")
	// ErrUserCreateFailed 表示创建用户失败。
	ErrUserCreateFailed = errors.NewWithStatus(1085, http.StatusInternalServerError, "创建用户失败")
	// ErrUserDeleteReqNil 表示删除用户请求参数为空。
	ErrUserDeleteReqNil = errors.NewWithStatus(1086, http.StatusOK, "删除用户参数不能为空")
	// ErrUserIDInvalid 表示用户 ID 非法。
	ErrUserIDInvalid = errors.NewWithStatus(1087, http.StatusOK, "用户 ID 必须大于 0")
	// ErrUserNotFound 表示用户不存在。
	ErrUserNotFound = errors.NewWithStatus(1088, http.StatusOK, "用户不存在")
	// ErrUserQueryFailed 表示查询用户失败。
	ErrUserQueryFailed = errors.NewWithStatus(1089, http.StatusInternalServerError, "查询用户失败")
	// ErrUserDeleteFailed 表示删除用户失败。
	ErrUserDeleteFailed = errors.NewWithStatus(1090, http.StatusInternalServerError, "删除用户失败")
	// ErrUserUpdateReqNil 表示更新用户请求参数为空。
	ErrUserUpdateReqNil = errors.NewWithStatus(1091, http.StatusOK, "更新用户参数不能为空")
	// ErrUserUpdateFailed 表示更新用户失败。
	ErrUserUpdateFailed = errors.NewWithStatus(1092, http.StatusInternalServerError, "更新用户失败")
	// ErrUserListQueryFailed 表示查询用户列表失败。
	ErrUserListQueryFailed = errors.NewWithStatus(1093, http.StatusInternalServerError, "查询用户列表失败")
	// ErrUserDetailReqNil 表示用户详情请求参数为空。
	ErrUserDetailReqNil = errors.NewWithStatus(1094, http.StatusOK, "用户详情参数不能为空")
	// ErrUserUpdateStatusReqNil 表示更新用户状态请求参数为空。
	ErrUserUpdateStatusReqNil = errors.NewWithStatus(1095, http.StatusOK, "更新用户状态参数不能为空")
	// ErrUserIDsRequired 表示用户 ID 集合为空。
	ErrUserIDsRequired = errors.NewWithStatus(1096, http.StatusOK, "用户 ID 集合不能为空")
	// ErrUserUpdateStatusFailed 表示更新用户状态失败。
	ErrUserUpdateStatusFailed = errors.NewWithStatus(1097, http.StatusInternalServerError, "更新用户状态失败")
	// ErrUserResetPasswordReqNil 表示重置用户密码请求参数为空。
	ErrUserResetPasswordReqNil = errors.NewWithStatus(1098, http.StatusOK, "重置用户密码参数不能为空")
	// ErrUserResetPasswordFailed 表示重置用户密码失败。
	ErrUserResetPasswordFailed = errors.NewWithStatus(1099, http.StatusInternalServerError, "重置用户密码失败")
	// ErrUserAssignRolesReqNil 表示分配用户角色请求参数为空。
	ErrUserAssignRolesReqNil = errors.NewWithStatus(1100, http.StatusOK, "分配用户角色参数不能为空")
	// ErrUserAssignRolesFailed 表示分配用户角色失败。
	ErrUserAssignRolesFailed = errors.NewWithStatus(1101, http.StatusInternalServerError, "分配用户角色失败")
	// ErrAuthLoginReqNil 表示登录请求参数为空。
	ErrAuthLoginReqNil = errors.NewWithStatus(1102, http.StatusOK, "登录参数不能为空")
	// ErrAuthUsernameRequired 表示登录账号为空。
	ErrAuthUsernameRequired = errors.NewWithStatus(1103, http.StatusOK, "登录账号不能为空")
	// ErrAuthPasswordRequired 表示登录密码为空。
	ErrAuthPasswordRequired = errors.NewWithStatus(1104, http.StatusOK, "登录密码不能为空")
	// ErrAuthCredentialsInvalid 表示登录账号或密码错误。
	ErrAuthCredentialsInvalid = errors.NewWithStatus(1105, http.StatusOK, "账号或密码错误")
	// ErrAuthUserDisabled 表示登录用户已禁用。
	ErrAuthUserDisabled = errors.NewWithStatus(1107, http.StatusOK, "用户已禁用")
	// ErrAuthTokenInvalid 表示认证 token 非法。
	ErrAuthTokenInvalid = errors.NewWithStatus(1108, http.StatusOK, "登录状态无效")
	// ErrAuthTokenExpired 表示认证 token 已过期。
	ErrAuthTokenExpired = errors.NewWithStatus(1109, http.StatusOK, "登录状态已过期")
	// ErrAuthUserQueryFailed 表示查询认证用户失败。
	ErrAuthUserQueryFailed = errors.NewWithStatus(1110, http.StatusInternalServerError, "查询认证用户失败")
	// ErrAuthRoleQueryFailed 表示查询认证角色失败。
	ErrAuthRoleQueryFailed = errors.NewWithStatus(1111, http.StatusInternalServerError, "查询认证角色失败")
	// ErrAuthPermissionQueryFailed 表示查询认证权限失败。
	ErrAuthPermissionQueryFailed = errors.NewWithStatus(1163, http.StatusInternalServerError, "查询认证权限失败")
	// ErrAuthMenuQueryFailed 表示查询认证菜单失败。
	ErrAuthMenuQueryFailed = errors.NewWithStatus(1112, http.StatusInternalServerError, "查询认证菜单失败")
	// ErrAuthTokenSignFailed 表示签发认证 token 失败。
	ErrAuthTokenSignFailed = errors.NewWithStatus(1113, http.StatusInternalServerError, "签发登录凭证失败")
	// ErrAuthServiceUnavailable 表示认证服务暂不可用。
	ErrAuthServiceUnavailable = errors.NewWithStatus(1114, http.StatusInternalServerError, "认证服务暂不可用")
	// ErrAuthPlatformInvalid 表示登录平台非法。
	ErrAuthPlatformInvalid = errors.NewWithStatus(1115, http.StatusOK, "登录平台不支持")
	// ErrAuthPlatformDisabled 表示登录平台被禁用。
	ErrAuthPlatformDisabled = errors.NewWithStatus(1116, http.StatusOK, "登录平台已禁用")
	// ErrAuthDeviceIDRequired 表示登录设备 ID 为空。
	ErrAuthDeviceIDRequired = errors.NewWithStatus(1117, http.StatusOK, "登录设备不能为空")
	// ErrAuthLoginConflict 表示登录命中并发策略冲突。
	ErrAuthLoginConflict = errors.NewWithStatus(1118, http.StatusOK, "当前账号已在其他设备登录")
	// ErrUserGenderInvalid 表示用户性别非法。
	ErrUserGenderInvalid = errors.NewWithStatus(1119, http.StatusOK, "用户性别非法")
	// ErrRoleAssignDeptsReqNil 表示分配角色数据权限部门请求参数为空。
	ErrRoleAssignDeptsReqNil = errors.NewWithStatus(1120, http.StatusOK, "分配角色数据权限部门参数不能为空")
	// ErrRoleAssignDeptsFailed 表示分配角色数据权限部门失败。
	ErrRoleAssignDeptsFailed = errors.NewWithStatus(1121, http.StatusInternalServerError, "分配角色数据权限部门失败")
	// ErrMenuComponentRequired 表示页面菜单组件路径为空。
	ErrMenuComponentRequired = errors.NewWithStatus(1122, http.StatusOK, "页面菜单组件路径不能为空")
	// ErrMenuExternalURLRequired 表示外链菜单地址为空。
	ErrMenuExternalURLRequired = errors.NewWithStatus(1123, http.StatusOK, "外链菜单地址不能为空")
	// ErrMenuHasPermissions 表示菜单下仍存在操作权限。
	ErrMenuHasPermissions = errors.NewWithStatus(1124, http.StatusOK, "菜单下存在操作权限，不能删除")
	// ErrMenuPermissionQueryFailed 表示查询菜单操作权限失败。
	ErrMenuPermissionQueryFailed = errors.NewWithStatus(1125, http.StatusInternalServerError, "查询菜单操作权限失败")
	// ErrMenuOptionsQueryFailed 表示查询菜单选项失败。
	ErrMenuOptionsQueryFailed = errors.NewWithStatus(1126, http.StatusInternalServerError, "查询菜单选项失败")
	// ErrMenuParentExternal 表示外链菜单不能作为父菜单。
	ErrMenuParentExternal = errors.NewWithStatus(1127, http.StatusOK, "外链菜单不能作为父菜单")
	// ErrMenuExternalHasChildren 表示存在子菜单的节点不能改为外链菜单。
	ErrMenuExternalHasChildren = errors.NewWithStatus(1128, http.StatusOK, "存在子菜单的菜单不能改为外链菜单")
	// ErrRolePermissionIDInvalid 表示角色权限 ID 非法。
	ErrRolePermissionIDInvalid = errors.NewWithStatus(1129, http.StatusOK, "角色权限 ID 必须大于 0")
	// ErrRolePermissionNotFound 表示角色绑定权限不存在或已禁用。
	ErrRolePermissionNotFound = errors.NewWithStatus(1130, http.StatusOK, "角色绑定权限不存在或已禁用")
	// ErrRolePermissionQueryFailed 表示查询角色权限失败。
	ErrRolePermissionQueryFailed = errors.NewWithStatus(1131, http.StatusInternalServerError, "查询角色权限失败")
	// ErrRolePermissionMenuRequired 表示权限所属菜单未分配给角色。
	ErrRolePermissionMenuRequired = errors.NewWithStatus(1132, http.StatusOK, "权限所属菜单必须同时分配给角色")
	// ErrRoleMenuDisabled 表示角色不能绑定禁用菜单。
	ErrRoleMenuDisabled = errors.NewWithStatus(1133, http.StatusOK, "角色不能绑定禁用菜单")
	// ErrRoleAssignResourcesReqNil 表示分配角色资源请求参数为空。
	ErrRoleAssignResourcesReqNil = errors.NewWithStatus(1134, http.StatusOK, "分配角色资源参数不能为空")
	// ErrRoleAssignResourcesFailed 表示分配角色资源失败。
	ErrRoleAssignResourcesFailed = errors.NewWithStatus(1135, http.StatusInternalServerError, "分配角色资源失败")
	// ErrPermissionCreateReqNil 表示创建权限请求参数为空。
	ErrPermissionCreateReqNil = errors.NewWithStatus(1136, http.StatusOK, "创建权限参数不能为空")
	// ErrPermissionMenuIDInvalid 表示权限所属菜单 ID 非法。
	ErrPermissionMenuIDInvalid = errors.NewWithStatus(1137, http.StatusOK, "权限所属菜单 ID 必须大于 0")
	// ErrPermissionNameRequired 表示权限名称为空。
	ErrPermissionNameRequired = errors.NewWithStatus(1138, http.StatusOK, "权限名称不能为空")
	// ErrPermissionCodeRequired 表示权限标识为空。
	ErrPermissionCodeRequired = errors.NewWithStatus(1139, http.StatusOK, "权限标识不能为空")
	// ErrPermissionSortInvalid 表示权限排序值非法。
	ErrPermissionSortInvalid = errors.NewWithStatus(1140, http.StatusOK, "权限排序不能小于 0")
	// ErrPermissionStatusInvalid 表示权限状态非法。
	ErrPermissionStatusInvalid = errors.NewWithStatus(1141, http.StatusOK, "权限状态只能是 0 或 1")
	// ErrPermissionMenuNotFound 表示权限所属菜单不存在。
	ErrPermissionMenuNotFound = errors.NewWithStatus(1142, http.StatusOK, "权限所属菜单不存在")
	// ErrPermissionCodeExists 表示权限标识已存在。
	ErrPermissionCodeExists = errors.NewWithStatus(1143, http.StatusOK, "权限标识已存在")
	// ErrPermissionMenuQueryFailed 表示查询权限所属菜单失败。
	ErrPermissionMenuQueryFailed = errors.NewWithStatus(1144, http.StatusInternalServerError, "查询权限所属菜单失败")
	// ErrPermissionCodeQueryFailed 表示查询权限标识失败。
	ErrPermissionCodeQueryFailed = errors.NewWithStatus(1145, http.StatusInternalServerError, "查询权限标识失败")
	// ErrPermissionCreateFailed 表示创建权限失败。
	ErrPermissionCreateFailed = errors.NewWithStatus(1146, http.StatusInternalServerError, "创建权限失败")
	// ErrPermissionDeleteReqNil 表示删除权限请求参数为空。
	ErrPermissionDeleteReqNil = errors.NewWithStatus(1147, http.StatusOK, "删除权限参数不能为空")
	// ErrPermissionIDInvalid 表示权限 ID 非法。
	ErrPermissionIDInvalid = errors.NewWithStatus(1148, http.StatusOK, "权限 ID 必须大于 0")
	// ErrPermissionNotFound 表示权限不存在。
	ErrPermissionNotFound = errors.NewWithStatus(1149, http.StatusOK, "权限不存在")
	// ErrPermissionHasRoleBinding 表示权限已绑定角色。
	ErrPermissionHasRoleBinding = errors.NewWithStatus(1150, http.StatusOK, "权限已绑定角色，不能删除")
	// ErrPermissionQueryFailed 表示查询权限失败。
	ErrPermissionQueryFailed = errors.NewWithStatus(1151, http.StatusInternalServerError, "查询权限失败")
	// ErrPermissionRoleBindingQueryFailed 表示查询权限角色绑定失败。
	ErrPermissionRoleBindingQueryFailed = errors.NewWithStatus(1152, http.StatusInternalServerError, "查询权限角色绑定失败")
	// ErrPermissionDeleteFailed 表示删除权限失败。
	ErrPermissionDeleteFailed = errors.NewWithStatus(1153, http.StatusInternalServerError, "删除权限失败")
	// ErrPermissionUpdateReqNil 表示更新权限请求参数为空。
	ErrPermissionUpdateReqNil = errors.NewWithStatus(1154, http.StatusOK, "更新权限参数不能为空")
	// ErrPermissionMenuChangeRoleBinding 表示已绑定角色的权限不能变更所属菜单。
	ErrPermissionMenuChangeRoleBinding = errors.NewWithStatus(1155, http.StatusOK, "权限已绑定角色，不能变更所属菜单")
	// ErrPermissionUpdateFailed 表示更新权限失败。
	ErrPermissionUpdateFailed = errors.NewWithStatus(1156, http.StatusInternalServerError, "更新权限失败")
	// ErrPermissionListReqNil 表示查询权限列表请求参数为空。
	ErrPermissionListReqNil = errors.NewWithStatus(1157, http.StatusOK, "查询权限列表参数不能为空")
	// ErrPermissionListQueryFailed 表示查询权限列表失败。
	ErrPermissionListQueryFailed = errors.NewWithStatus(1158, http.StatusInternalServerError, "查询权限列表失败")
	// ErrPermissionDetailReqNil 表示查询权限详情请求参数为空。
	ErrPermissionDetailReqNil = errors.NewWithStatus(1159, http.StatusOK, "查询权限详情参数不能为空")
	// ErrPermissionUpdateStatusReqNil 表示更新权限状态请求参数为空。
	ErrPermissionUpdateStatusReqNil = errors.NewWithStatus(1160, http.StatusOK, "更新权限状态参数不能为空")
	// ErrPermissionIDsRequired 表示权限 ID 集合为空。
	ErrPermissionIDsRequired = errors.NewWithStatus(1161, http.StatusOK, "权限 ID 集合不能为空")
	// ErrPermissionUpdateStatusFailed 表示更新权限状态失败。
	ErrPermissionUpdateStatusFailed = errors.NewWithStatus(1162, http.StatusInternalServerError, "更新权限状态失败")
	// ErrRoleMenuAncestorRequired 表示角色分配子菜单时缺少父菜单。
	ErrRoleMenuAncestorRequired = errors.NewWithStatus(1164, http.StatusOK, "子菜单的父菜单必须同时分配给角色")
	// ErrUserSessionRevokeFailed 表示吊销用户登录会话失败。
	ErrUserSessionRevokeFailed = errors.NewWithStatus(1165, http.StatusInternalServerError, "吊销用户登录会话失败")
	// ErrDeptCreateReqNil 表示创建部门请求参数为空。
	ErrDeptCreateReqNil = errors.NewWithStatus(1166, http.StatusOK, "创建部门参数不能为空")
	// ErrDeptParentIDInvalid 表示父部门 ID 非法。
	ErrDeptParentIDInvalid = errors.NewWithStatus(1167, http.StatusOK, "父部门 ID 不能小于 0")
	// ErrDeptNameRequired 表示部门名称为空。
	ErrDeptNameRequired = errors.NewWithStatus(1168, http.StatusOK, "部门名称不能为空")
	// ErrDeptLeaderIDInvalid 表示部门负责人 ID 非法。
	ErrDeptLeaderIDInvalid = errors.NewWithStatus(1169, http.StatusOK, "部门负责人 ID 必须大于 0")
	// ErrDeptSortInvalid 表示部门排序值非法。
	ErrDeptSortInvalid = errors.NewWithStatus(1170, http.StatusOK, "部门排序不能小于 0")
	// ErrDeptStatusInvalid 表示部门状态非法。
	ErrDeptStatusInvalid = errors.NewWithStatus(1171, http.StatusOK, "部门状态只能是 0 或 1")
	// ErrDeptParentNotFound 表示父部门不存在。
	ErrDeptParentNotFound = errors.NewWithStatus(1172, http.StatusOK, "父部门不存在")
	// ErrDeptNameExists 表示同级部门名称已存在。
	ErrDeptNameExists = errors.NewWithStatus(1173, http.StatusOK, "同级部门名称已存在")
	// ErrDeptCodeExists 表示部门编码已存在。
	ErrDeptCodeExists = errors.NewWithStatus(1174, http.StatusOK, "部门编码已存在")
	// ErrDeptLeaderNotFound 表示部门负责人不存在。
	ErrDeptLeaderNotFound = errors.NewWithStatus(1175, http.StatusOK, "部门负责人不存在")
	// ErrDeptParentQueryFailed 表示查询父部门失败。
	ErrDeptParentQueryFailed = errors.NewWithStatus(1176, http.StatusInternalServerError, "查询父部门失败")
	// ErrDeptNameQueryFailed 表示查询部门名称失败。
	ErrDeptNameQueryFailed = errors.NewWithStatus(1177, http.StatusInternalServerError, "查询部门名称失败")
	// ErrDeptCodeQueryFailed 表示查询部门编码失败。
	ErrDeptCodeQueryFailed = errors.NewWithStatus(1178, http.StatusInternalServerError, "查询部门编码失败")
	// ErrDeptLeaderQueryFailed 表示查询部门负责人失败。
	ErrDeptLeaderQueryFailed = errors.NewWithStatus(1179, http.StatusInternalServerError, "查询部门负责人失败")
	// ErrDeptCreateFailed 表示创建部门失败。
	ErrDeptCreateFailed = errors.NewWithStatus(1180, http.StatusInternalServerError, "创建部门失败")
	// ErrDeptDeleteReqNil 表示删除部门请求参数为空。
	ErrDeptDeleteReqNil = errors.NewWithStatus(1181, http.StatusOK, "删除部门参数不能为空")
	// ErrDeptIDInvalid 表示部门 ID 非法。
	ErrDeptIDInvalid = errors.NewWithStatus(1182, http.StatusOK, "部门 ID 必须大于 0")
	// ErrDeptNotFound 表示部门不存在。
	ErrDeptNotFound = errors.NewWithStatus(1183, http.StatusOK, "部门不存在")
	// ErrDeptHasChildren 表示部门存在子部门。
	ErrDeptHasChildren = errors.NewWithStatus(1184, http.StatusOK, "部门存在子部门，不能删除")
	// ErrDeptHasUsers 表示部门存在直属用户。
	ErrDeptHasUsers = errors.NewWithStatus(1185, http.StatusOK, "部门存在直属用户，不能删除")
	// ErrDeptHasRoleBinding 表示部门已绑定角色数据权限。
	ErrDeptHasRoleBinding = errors.NewWithStatus(1186, http.StatusOK, "部门已绑定角色数据权限，不能删除")
	// ErrDeptQueryFailed 表示查询部门失败。
	ErrDeptQueryFailed = errors.NewWithStatus(1187, http.StatusInternalServerError, "查询部门失败")
	// ErrDeptChildrenQueryFailed 表示查询子部门失败。
	ErrDeptChildrenQueryFailed = errors.NewWithStatus(1188, http.StatusInternalServerError, "查询子部门失败")
	// ErrDeptUserBindingQueryFailed 表示查询部门用户绑定失败。
	ErrDeptUserBindingQueryFailed = errors.NewWithStatus(1189, http.StatusInternalServerError, "查询部门用户绑定失败")
	// ErrDeptRoleBindingQueryFailed 表示查询部门角色绑定失败。
	ErrDeptRoleBindingQueryFailed = errors.NewWithStatus(1190, http.StatusInternalServerError, "查询部门角色绑定失败")
	// ErrDeptDeleteFailed 表示删除部门失败。
	ErrDeptDeleteFailed = errors.NewWithStatus(1191, http.StatusInternalServerError, "删除部门失败")
	// ErrDeptUpdateReqNil 表示更新部门请求参数为空。
	ErrDeptUpdateReqNil = errors.NewWithStatus(1192, http.StatusOK, "更新部门参数不能为空")
	// ErrDeptParentSelf 表示父部门不能是自己。
	ErrDeptParentSelf = errors.NewWithStatus(1193, http.StatusOK, "父部门不能是当前部门")
	// ErrDeptParentDescendant 表示父部门不能是当前部门的子孙部门。
	ErrDeptParentDescendant = errors.NewWithStatus(1194, http.StatusOK, "父部门不能是当前部门的子孙部门")
	// ErrDeptUpdateFailed 表示更新部门失败。
	ErrDeptUpdateFailed = errors.NewWithStatus(1195, http.StatusInternalServerError, "更新部门失败")
	// ErrDeptTreeQueryFailed 表示查询部门树失败。
	ErrDeptTreeQueryFailed = errors.NewWithStatus(1196, http.StatusInternalServerError, "查询部门树失败")
	// ErrDeptOptionsQueryFailed 表示查询部门选项失败。
	ErrDeptOptionsQueryFailed = errors.NewWithStatus(1197, http.StatusInternalServerError, "查询部门选项失败")
	// ErrDeptDetailReqNil 表示查询部门详情请求参数为空。
	ErrDeptDetailReqNil = errors.NewWithStatus(1198, http.StatusOK, "查询部门详情参数不能为空")
	// ErrDeptUpdateStatusReqNil 表示更新部门状态请求参数为空。
	ErrDeptUpdateStatusReqNil = errors.NewWithStatus(1199, http.StatusOK, "更新部门状态参数不能为空")
	// ErrDeptIDsRequired 表示部门 ID 集合为空。
	ErrDeptIDsRequired = errors.NewWithStatus(1200, http.StatusOK, "部门 ID 集合不能为空")
	// ErrDeptUpdateStatusFailed 表示更新部门状态失败。
	ErrDeptUpdateStatusFailed = errors.NewWithStatus(1201, http.StatusInternalServerError, "更新部门状态失败")
	// ErrPostCreateReqNil 表示创建岗位请求参数为空。
	ErrPostCreateReqNil = errors.NewWithStatus(1202, http.StatusOK, "创建岗位参数不能为空")
	// ErrPostNameRequired 表示岗位名称为空。
	ErrPostNameRequired = errors.NewWithStatus(1203, http.StatusOK, "岗位名称不能为空")
	// ErrPostCodeRequired 表示岗位编码为空。
	ErrPostCodeRequired = errors.NewWithStatus(1204, http.StatusOK, "岗位编码不能为空")
	// ErrPostSortInvalid 表示岗位排序值非法。
	ErrPostSortInvalid = errors.NewWithStatus(1205, http.StatusOK, "岗位排序不能小于 0")
	// ErrPostStatusInvalid 表示岗位状态非法。
	ErrPostStatusInvalid = errors.NewWithStatus(1206, http.StatusOK, "岗位状态只能是 0 或 1")
	// ErrPostNameExists 表示岗位名称已存在。
	ErrPostNameExists = errors.NewWithStatus(1207, http.StatusOK, "岗位名称已存在")
	// ErrPostCodeExists 表示岗位编码已存在。
	ErrPostCodeExists = errors.NewWithStatus(1208, http.StatusOK, "岗位编码已存在")
	// ErrPostNameQueryFailed 表示查询岗位名称失败。
	ErrPostNameQueryFailed = errors.NewWithStatus(1209, http.StatusInternalServerError, "查询岗位名称失败")
	// ErrPostCodeQueryFailed 表示查询岗位编码失败。
	ErrPostCodeQueryFailed = errors.NewWithStatus(1210, http.StatusInternalServerError, "查询岗位编码失败")
	// ErrPostCreateFailed 表示创建岗位失败。
	ErrPostCreateFailed = errors.NewWithStatus(1211, http.StatusInternalServerError, "创建岗位失败")
	// ErrPostDeleteReqNil 表示删除岗位请求参数为空。
	ErrPostDeleteReqNil = errors.NewWithStatus(1212, http.StatusOK, "删除岗位参数不能为空")
	// ErrPostIDInvalid 表示岗位 ID 非法。
	ErrPostIDInvalid = errors.NewWithStatus(1213, http.StatusOK, "岗位 ID 必须大于 0")
	// ErrPostIDsRequired 表示岗位 ID 集合为空。
	ErrPostIDsRequired = errors.NewWithStatus(1214, http.StatusOK, "岗位 ID 集合不能为空")
	// ErrPostNotFound 表示岗位不存在。
	ErrPostNotFound = errors.NewWithStatus(1215, http.StatusOK, "岗位不存在")
	// ErrPostHasUserBinding 表示岗位已绑定用户。
	ErrPostHasUserBinding = errors.NewWithStatus(1216, http.StatusOK, "岗位已绑定用户，不能删除")
	// ErrPostQueryFailed 表示查询岗位失败。
	ErrPostQueryFailed = errors.NewWithStatus(1217, http.StatusInternalServerError, "查询岗位失败")
	// ErrPostUserBindingQueryFailed 表示查询岗位用户绑定失败。
	ErrPostUserBindingQueryFailed = errors.NewWithStatus(1218, http.StatusInternalServerError, "查询岗位用户绑定失败")
	// ErrPostDeleteFailed 表示删除岗位失败。
	ErrPostDeleteFailed = errors.NewWithStatus(1219, http.StatusInternalServerError, "删除岗位失败")
	// ErrPostUpdateReqNil 表示更新岗位请求参数为空。
	ErrPostUpdateReqNil = errors.NewWithStatus(1220, http.StatusOK, "更新岗位参数不能为空")
	// ErrPostUpdateFailed 表示更新岗位失败。
	ErrPostUpdateFailed = errors.NewWithStatus(1221, http.StatusInternalServerError, "更新岗位失败")
	// ErrPostListQueryFailed 表示查询岗位列表失败。
	ErrPostListQueryFailed = errors.NewWithStatus(1222, http.StatusInternalServerError, "查询岗位列表失败")
	// ErrPostDetailReqNil 表示查询岗位详情请求参数为空。
	ErrPostDetailReqNil = errors.NewWithStatus(1223, http.StatusOK, "查询岗位详情参数不能为空")
	// ErrPostUpdateStatusReqNil 表示更新岗位状态请求参数为空。
	ErrPostUpdateStatusReqNil = errors.NewWithStatus(1224, http.StatusOK, "更新岗位状态参数不能为空")
	// ErrPostUpdateStatusFailed 表示更新岗位状态失败。
	ErrPostUpdateStatusFailed = errors.NewWithStatus(1225, http.StatusInternalServerError, "更新岗位状态失败")
	// ErrUserDeptDisabled 表示用户不能绑定禁用部门。
	ErrUserDeptDisabled = errors.NewWithStatus(1226, http.StatusOK, "用户不能绑定禁用部门")
	// ErrUserPostDisabled 表示用户不能绑定禁用岗位。
	ErrUserPostDisabled = errors.NewWithStatus(1227, http.StatusOK, "用户不能绑定禁用岗位")
	// ErrRoleDeptDisabled 表示角色不能绑定禁用部门。
	ErrRoleDeptDisabled = errors.NewWithStatus(1228, http.StatusOK, "角色不能绑定禁用部门")
	// ErrDictTypeCreateReqNil 表示创建字典类型请求为空。
	ErrDictTypeCreateReqNil = errors.NewWithStatus(1229, http.StatusOK, "创建字典类型参数不能为空")
	// ErrDictTypeNameRequired 表示字典类型名称为空。
	ErrDictTypeNameRequired = errors.NewWithStatus(1230, http.StatusOK, "字典类型名称不能为空")
	// ErrDictTypeCodeRequired 表示字典类型编码为空。
	ErrDictTypeCodeRequired = errors.NewWithStatus(1231, http.StatusOK, "字典类型编码不能为空")
	// ErrDictTypeStatusInvalid 表示字典类型状态非法。
	ErrDictTypeStatusInvalid = errors.NewWithStatus(1232, http.StatusOK, "字典类型状态只能是 0 或 1")
	// ErrDictTypeNameExists 表示字典类型名称已存在。
	ErrDictTypeNameExists = errors.NewWithStatus(1233, http.StatusOK, "字典类型名称已存在")
	// ErrDictTypeCodeExists 表示字典类型编码已存在。
	ErrDictTypeCodeExists = errors.NewWithStatus(1234, http.StatusOK, "字典类型编码已存在")
	// ErrDictTypeNameQueryFailed 表示查询字典类型名称失败。
	ErrDictTypeNameQueryFailed = errors.NewWithStatus(1235, http.StatusInternalServerError, "查询字典类型名称失败")
	// ErrDictTypeCodeQueryFailed 表示查询字典类型编码失败。
	ErrDictTypeCodeQueryFailed = errors.NewWithStatus(1236, http.StatusInternalServerError, "查询字典类型编码失败")
	// ErrDictTypeCreateFailed 表示创建字典类型失败。
	ErrDictTypeCreateFailed = errors.NewWithStatus(1237, http.StatusInternalServerError, "创建字典类型失败")
	// ErrDictTypeDeleteReqNil 表示删除字典类型请求为空。
	ErrDictTypeDeleteReqNil = errors.NewWithStatus(1238, http.StatusOK, "删除字典类型参数不能为空")
	// ErrDictTypeIDInvalid 表示字典类型 ID 非法。
	ErrDictTypeIDInvalid = errors.NewWithStatus(1239, http.StatusOK, "字典类型 ID 必须大于 0")
	// ErrDictTypeIDsRequired 表示字典类型 ID 集合为空。
	ErrDictTypeIDsRequired = errors.NewWithStatus(1240, http.StatusOK, "字典类型 ID 集合不能为空")
	// ErrDictTypeNotFound 表示字典类型不存在。
	ErrDictTypeNotFound = errors.NewWithStatus(1241, http.StatusOK, "字典类型不存在")
	// ErrDictTypeDisabled 表示字典类型已禁用。
	ErrDictTypeDisabled = errors.NewWithStatus(1242, http.StatusOK, "字典类型已禁用")
	// ErrDictTypeHasData 表示字典类型仍有数据。
	ErrDictTypeHasData = errors.NewWithStatus(1243, http.StatusOK, "字典类型下存在字典数据，不能删除")
	// ErrDictTypeQueryFailed 表示查询字典类型失败。
	ErrDictTypeQueryFailed = errors.NewWithStatus(1244, http.StatusInternalServerError, "查询字典类型失败")
	// ErrDictTypeDataQueryFailed 表示查询字典类型数据失败。
	ErrDictTypeDataQueryFailed = errors.NewWithStatus(1245, http.StatusInternalServerError, "查询字典类型数据失败")
	// ErrDictTypeDeleteFailed 表示删除字典类型失败。
	ErrDictTypeDeleteFailed = errors.NewWithStatus(1246, http.StatusInternalServerError, "删除字典类型失败")
	// ErrDictTypeUpdateReqNil 表示更新字典类型请求为空。
	ErrDictTypeUpdateReqNil = errors.NewWithStatus(1247, http.StatusOK, "更新字典类型参数不能为空")
	// ErrDictTypeUpdateFailed 表示更新字典类型失败。
	ErrDictTypeUpdateFailed = errors.NewWithStatus(1248, http.StatusInternalServerError, "更新字典类型失败")
	// ErrDictTypeListQueryFailed 表示查询字典类型列表失败。
	ErrDictTypeListQueryFailed = errors.NewWithStatus(1249, http.StatusInternalServerError, "查询字典类型列表失败")
	// ErrDictTypeDetailReqNil 表示查询字典类型详情请求为空。
	ErrDictTypeDetailReqNil = errors.NewWithStatus(1250, http.StatusOK, "查询字典类型详情参数不能为空")
	// ErrDictTypeUpdateStatusReqNil 表示更新字典类型状态请求为空。
	ErrDictTypeUpdateStatusReqNil = errors.NewWithStatus(1251, http.StatusOK, "更新字典类型状态参数不能为空")
	// ErrDictTypeUpdateStatusFailed 表示更新字典类型状态失败。
	ErrDictTypeUpdateStatusFailed = errors.NewWithStatus(1252, http.StatusInternalServerError, "更新字典类型状态失败")
	// ErrDictDataCreateReqNil 表示创建字典数据请求为空。
	ErrDictDataCreateReqNil = errors.NewWithStatus(1253, http.StatusOK, "创建字典数据参数不能为空")
	// ErrDictDataTypeCodeRequired 表示字典类型编码为空。
	ErrDictDataTypeCodeRequired = errors.NewWithStatus(1254, http.StatusOK, "字典类型编码不能为空")
	// ErrDictDataLabelRequired 表示字典显示文本为空。
	ErrDictDataLabelRequired = errors.NewWithStatus(1255, http.StatusOK, "字典显示文本不能为空")
	// ErrDictDataValueRequired 表示字典值为空。
	ErrDictDataValueRequired = errors.NewWithStatus(1256, http.StatusOK, "字典值不能为空")
	// ErrDictDataSortInvalid 表示字典数据排序非法。
	ErrDictDataSortInvalid = errors.NewWithStatus(1257, http.StatusOK, "字典数据排序不能小于 0")
	// ErrDictDataStatusInvalid 表示字典数据状态非法。
	ErrDictDataStatusInvalid = errors.NewWithStatus(1258, http.StatusOK, "字典数据状态只能是 0 或 1")
	// ErrDictDataTypeNotFound 表示字典数据所属类型不存在。
	ErrDictDataTypeNotFound = errors.NewWithStatus(1259, http.StatusOK, "字典数据所属类型不存在")
	// ErrDictDataValueExists 表示同类型字典值已存在。
	ErrDictDataValueExists = errors.NewWithStatus(1260, http.StatusOK, "同类型字典值已存在")
	// ErrDictDataDefaultDisabled 表示禁用数据不能设为默认项。
	ErrDictDataDefaultDisabled = errors.NewWithStatus(1261, http.StatusOK, "禁用字典数据不能设为默认项")
	// ErrDictDataTypeQueryFailed 表示查询字典数据所属类型失败。
	ErrDictDataTypeQueryFailed = errors.NewWithStatus(1262, http.StatusInternalServerError, "查询字典数据所属类型失败")
	// ErrDictDataValueQueryFailed 表示查询字典值失败。
	ErrDictDataValueQueryFailed = errors.NewWithStatus(1263, http.StatusInternalServerError, "查询字典值失败")
	// ErrDictDataCreateFailed 表示创建字典数据失败。
	ErrDictDataCreateFailed = errors.NewWithStatus(1264, http.StatusInternalServerError, "创建字典数据失败")
	// ErrDictDataDeleteReqNil 表示删除字典数据请求为空。
	ErrDictDataDeleteReqNil = errors.NewWithStatus(1265, http.StatusOK, "删除字典数据参数不能为空")
	// ErrDictDataIDInvalid 表示字典数据 ID 非法。
	ErrDictDataIDInvalid = errors.NewWithStatus(1266, http.StatusOK, "字典数据 ID 必须大于 0")
	// ErrDictDataIDsRequired 表示字典数据 ID 集合为空。
	ErrDictDataIDsRequired = errors.NewWithStatus(1267, http.StatusOK, "字典数据 ID 集合不能为空")
	// ErrDictDataNotFound 表示字典数据不存在。
	ErrDictDataNotFound = errors.NewWithStatus(1268, http.StatusOK, "字典数据不存在")
	// ErrDictDataQueryFailed 表示查询字典数据失败。
	ErrDictDataQueryFailed = errors.NewWithStatus(1269, http.StatusInternalServerError, "查询字典数据失败")
	// ErrDictDataDeleteFailed 表示删除字典数据失败。
	ErrDictDataDeleteFailed = errors.NewWithStatus(1270, http.StatusInternalServerError, "删除字典数据失败")
	// ErrDictDataUpdateReqNil 表示更新字典数据请求为空。
	ErrDictDataUpdateReqNil = errors.NewWithStatus(1271, http.StatusOK, "更新字典数据参数不能为空")
	// ErrDictDataUpdateFailed 表示更新字典数据失败。
	ErrDictDataUpdateFailed = errors.NewWithStatus(1272, http.StatusInternalServerError, "更新字典数据失败")
	// ErrDictDataListQueryFailed 表示查询字典数据列表失败。
	ErrDictDataListQueryFailed = errors.NewWithStatus(1273, http.StatusInternalServerError, "查询字典数据列表失败")
	// ErrDictDataDetailReqNil 表示查询字典数据详情请求为空。
	ErrDictDataDetailReqNil = errors.NewWithStatus(1274, http.StatusOK, "查询字典数据详情参数不能为空")
	// ErrDictDataUpdateStatusReqNil 表示更新字典数据状态请求为空。
	ErrDictDataUpdateStatusReqNil = errors.NewWithStatus(1275, http.StatusOK, "更新字典数据状态参数不能为空")
	// ErrDictDataUpdateStatusFailed 表示更新字典数据状态失败。
	ErrDictDataUpdateStatusFailed = errors.NewWithStatus(1276, http.StatusInternalServerError, "更新字典数据状态失败")
	// ErrDictValuesReqNil 表示查询字典值请求为空。
	ErrDictValuesReqNil = errors.NewWithStatus(1277, http.StatusOK, "查询字典值参数不能为空")
	// ErrDictValuesQueryFailed 表示查询字典值失败。
	ErrDictValuesQueryFailed = errors.NewWithStatus(1278, http.StatusInternalServerError, "查询字典值失败")
	// ErrUserAssignPostsReqNil 表示分配用户岗位请求参数为空。
	ErrUserAssignPostsReqNil = errors.NewWithStatus(1279, http.StatusOK, "分配用户岗位参数不能为空")
	// ErrUserAssignPostsFailed 表示分配用户岗位失败。
	ErrUserAssignPostsFailed = errors.NewWithStatus(1280, http.StatusInternalServerError, "分配用户岗位失败")
	// ErrConfigCreateReqNil 表示创建配置请求为空。
	ErrConfigCreateReqNil = errors.NewWithStatus(1281, http.StatusOK, "创建配置参数不能为空")
	// ErrConfigNameRequired 表示配置名称为空。
	ErrConfigNameRequired = errors.NewWithStatus(1282, http.StatusOK, "配置名称不能为空")
	// ErrConfigKeyRequired 表示配置键为空。
	ErrConfigKeyRequired = errors.NewWithStatus(1283, http.StatusOK, "配置键不能为空")
	// ErrConfigKeyInvalid 表示配置键格式非法。
	ErrConfigKeyInvalid = errors.NewWithStatus(1284, http.StatusOK, "配置键格式非法")
	// ErrConfigValueTypeInvalid 表示配置值类型非法。
	ErrConfigValueTypeInvalid = errors.NewWithStatus(1286, http.StatusOK, "配置值类型非法")
	// ErrConfigValueInvalid 表示配置值与声明类型不匹配。
	ErrConfigValueInvalid = errors.NewWithStatus(1287, http.StatusOK, "配置值与配置类型不匹配")
	// ErrConfigStatusInvalid 表示配置状态非法。
	ErrConfigStatusInvalid = errors.NewWithStatus(1288, http.StatusOK, "配置状态只能是 0 或 1")
	// ErrConfigKeyExists 表示配置键已存在。
	ErrConfigKeyExists = errors.NewWithStatus(1289, http.StatusOK, "配置键已存在")
	// ErrConfigKeyQueryFailed 表示查询配置键失败。
	ErrConfigKeyQueryFailed = errors.NewWithStatus(1290, http.StatusInternalServerError, "查询配置键失败")
	// ErrConfigCreateFailed 表示创建配置失败。
	ErrConfigCreateFailed = errors.NewWithStatus(1291, http.StatusInternalServerError, "创建配置失败")
	// ErrConfigDeleteReqNil 表示删除配置请求为空。
	ErrConfigDeleteReqNil = errors.NewWithStatus(1292, http.StatusOK, "删除配置参数不能为空")
	// ErrConfigIDInvalid 表示配置 ID 非法。
	ErrConfigIDInvalid = errors.NewWithStatus(1293, http.StatusOK, "配置 ID 必须大于 0")
	// ErrConfigIDsRequired 表示配置 ID 集合为空。
	ErrConfigIDsRequired = errors.NewWithStatus(1294, http.StatusOK, "配置 ID 集合不能为空")
	// ErrConfigNotFound 表示配置不存在或未启用。
	ErrConfigNotFound = errors.NewWithStatus(1295, http.StatusOK, "配置不存在")
	// ErrConfigBuiltinDelete 表示内置配置不能删除。
	ErrConfigBuiltinDelete = errors.NewWithStatus(1296, http.StatusOK, "系统内置配置不能删除")
	// ErrConfigQueryFailed 表示查询配置失败。
	ErrConfigQueryFailed = errors.NewWithStatus(1297, http.StatusInternalServerError, "查询配置失败")
	// ErrConfigDeleteFailed 表示删除配置失败。
	ErrConfigDeleteFailed = errors.NewWithStatus(1298, http.StatusInternalServerError, "删除配置失败")
	// ErrConfigUpdateReqNil 表示更新配置请求为空。
	ErrConfigUpdateReqNil = errors.NewWithStatus(1299, http.StatusOK, "更新配置参数不能为空")
	// ErrConfigUpdateFailed 表示更新配置失败。
	ErrConfigUpdateFailed = errors.NewWithStatus(1300, http.StatusInternalServerError, "更新配置失败")
	// ErrConfigListQueryFailed 表示查询配置列表失败。
	ErrConfigListQueryFailed = errors.NewWithStatus(1301, http.StatusInternalServerError, "查询配置列表失败")
	// ErrConfigDetailReqNil 表示查询配置详情请求为空。
	ErrConfigDetailReqNil = errors.NewWithStatus(1302, http.StatusOK, "查询配置详情参数不能为空")
	// ErrConfigUpdateStatusReqNil 表示更新配置状态请求为空。
	ErrConfigUpdateStatusReqNil = errors.NewWithStatus(1303, http.StatusOK, "更新配置状态参数不能为空")
	// ErrConfigUpdateStatusFailed 表示更新配置状态失败。
	ErrConfigUpdateStatusFailed = errors.NewWithStatus(1304, http.StatusInternalServerError, "更新配置状态失败")
	// ErrConfigValueTypeMismatch 表示类型化读取方法与配置类型不匹配。
	ErrConfigValueTypeMismatch = errors.NewWithStatus(1305, http.StatusInternalServerError, "配置读取类型不匹配")
	// ErrLoginLogRecordReqNil 表示登录日志写入参数为空。
	ErrLoginLogRecordReqNil = errors.NewWithStatus(1306, http.StatusInternalServerError, "登录日志参数不能为空")
	// ErrLoginLogRecordFailed 表示写入登录日志失败。
	ErrLoginLogRecordFailed = errors.NewWithStatus(1307, http.StatusInternalServerError, "写入登录日志失败")
	// ErrLoginLogStatusInvalid 表示登录日志状态非法。
	ErrLoginLogStatusInvalid = errors.NewWithStatus(1308, http.StatusOK, "登录状态只能是 0 或 1")
	// ErrLoginLogTimeInvalid 表示登录日志时间条件非法。
	ErrLoginLogTimeInvalid = errors.NewWithStatus(1309, http.StatusOK, "登录日志时间格式非法")
	// ErrLoginLogTimeRangeInvalid 表示登录日志时间范围非法。
	ErrLoginLogTimeRangeInvalid = errors.NewWithStatus(1310, http.StatusOK, "登录日志开始时间不能晚于结束时间")
	// ErrLoginLogListQueryFailed 表示查询登录日志列表失败。
	ErrLoginLogListQueryFailed = errors.NewWithStatus(1311, http.StatusInternalServerError, "查询登录日志列表失败")
	// ErrLoginLogDetailReqNil 表示查询登录日志详情参数为空。
	ErrLoginLogDetailReqNil = errors.NewWithStatus(1312, http.StatusOK, "查询登录日志详情参数不能为空")
	// ErrLoginLogIDInvalid 表示登录日志 ID 非法。
	ErrLoginLogIDInvalid = errors.NewWithStatus(1313, http.StatusOK, "登录日志 ID 必须大于 0")
	// ErrLoginLogNotFound 表示登录日志不存在。
	ErrLoginLogNotFound = errors.NewWithStatus(1314, http.StatusOK, "登录日志不存在")
	// ErrLoginLogQueryFailed 表示查询登录日志失败。
	ErrLoginLogQueryFailed = errors.NewWithStatus(1315, http.StatusInternalServerError, "查询登录日志失败")
	// ErrLoginLogDeleteReqNil 表示删除登录日志参数为空。
	ErrLoginLogDeleteReqNil = errors.NewWithStatus(1316, http.StatusOK, "删除登录日志参数不能为空")
	// ErrLoginLogIDsRequired 表示登录日志 ID 集合为空。
	ErrLoginLogIDsRequired = errors.NewWithStatus(1317, http.StatusOK, "登录日志 ID 集合不能为空")
	// ErrLoginLogDeleteFailed 表示删除登录日志失败。
	ErrLoginLogDeleteFailed = errors.NewWithStatus(1318, http.StatusInternalServerError, "删除登录日志失败")
	// ErrLoginLogCleanReqNil 表示清理登录日志参数为空。
	ErrLoginLogCleanReqNil = errors.NewWithStatus(1319, http.StatusOK, "清理登录日志参数不能为空")
	// ErrLoginLogCleanBeforeRequired 表示清理截止时间为空。
	ErrLoginLogCleanBeforeRequired = errors.NewWithStatus(1320, http.StatusOK, "清理截止时间不能为空")
	// ErrLoginLogCleanBeforeFuture 表示清理截止时间不早于当前时间。
	ErrLoginLogCleanBeforeFuture = errors.NewWithStatus(1321, http.StatusOK, "清理截止时间必须早于当前时间")
	// ErrLoginLogCleanFailed 表示清理登录日志失败。
	ErrLoginLogCleanFailed = errors.NewWithStatus(1322, http.StatusInternalServerError, "清理登录日志失败")
	// ErrOperLogRecordReqNil 表示操作日志写入参数为空。
	ErrOperLogRecordReqNil = errors.NewWithStatus(1323, http.StatusInternalServerError, "操作日志参数不能为空")
	// ErrOperLogModuleRequired 表示操作日志模块为空。
	ErrOperLogModuleRequired = errors.NewWithStatus(1324, http.StatusInternalServerError, "操作日志模块不能为空")
	// ErrOperLogActionRequired 表示操作日志动作为空。
	ErrOperLogActionRequired = errors.NewWithStatus(1325, http.StatusInternalServerError, "操作日志动作不能为空")
	// ErrOperLogMethodRequired 表示操作日志请求方法为空。
	ErrOperLogMethodRequired = errors.NewWithStatus(1326, http.StatusInternalServerError, "操作日志请求方法不能为空")
	// ErrOperLogPathRequired 表示操作日志请求路径为空。
	ErrOperLogPathRequired = errors.NewWithStatus(1327, http.StatusInternalServerError, "操作日志请求路径不能为空")
	// ErrOperLogRecordFailed 表示写入操作日志失败。
	ErrOperLogRecordFailed = errors.NewWithStatus(1328, http.StatusInternalServerError, "写入操作日志失败")
	// ErrOperLogStatusInvalid 表示操作日志状态非法。
	ErrOperLogStatusInvalid = errors.NewWithStatus(1329, http.StatusOK, "操作状态只能是 0 或 1")
	// ErrOperLogTimeInvalid 表示操作日志时间条件非法。
	ErrOperLogTimeInvalid = errors.NewWithStatus(1330, http.StatusOK, "操作日志时间格式非法")
	// ErrOperLogTimeRangeInvalid 表示操作日志时间范围非法。
	ErrOperLogTimeRangeInvalid = errors.NewWithStatus(1331, http.StatusOK, "操作日志开始时间不能晚于结束时间")
	// ErrOperLogListQueryFailed 表示查询操作日志列表失败。
	ErrOperLogListQueryFailed = errors.NewWithStatus(1332, http.StatusInternalServerError, "查询操作日志列表失败")
	// ErrOperLogDetailReqNil 表示查询操作日志详情参数为空。
	ErrOperLogDetailReqNil = errors.NewWithStatus(1333, http.StatusOK, "查询操作日志详情参数不能为空")
	// ErrOperLogIDInvalid 表示操作日志 ID 非法。
	ErrOperLogIDInvalid = errors.NewWithStatus(1334, http.StatusOK, "操作日志 ID 必须大于 0")
	// ErrOperLogNotFound 表示操作日志不存在。
	ErrOperLogNotFound = errors.NewWithStatus(1335, http.StatusOK, "操作日志不存在")
	// ErrOperLogQueryFailed 表示查询操作日志失败。
	ErrOperLogQueryFailed = errors.NewWithStatus(1336, http.StatusInternalServerError, "查询操作日志失败")
	// ErrOperLogDeleteReqNil 表示删除操作日志参数为空。
	ErrOperLogDeleteReqNil = errors.NewWithStatus(1337, http.StatusOK, "删除操作日志参数不能为空")
	// ErrOperLogIDsRequired 表示操作日志 ID 集合为空。
	ErrOperLogIDsRequired = errors.NewWithStatus(1338, http.StatusOK, "操作日志 ID 集合不能为空")
	// ErrOperLogDeleteFailed 表示删除操作日志失败。
	ErrOperLogDeleteFailed = errors.NewWithStatus(1339, http.StatusInternalServerError, "删除操作日志失败")
	// ErrOperLogCleanReqNil 表示清理操作日志参数为空。
	ErrOperLogCleanReqNil = errors.NewWithStatus(1340, http.StatusOK, "清理操作日志参数不能为空")
	// ErrOperLogCleanBeforeRequired 表示操作日志清理截止时间为空。
	ErrOperLogCleanBeforeRequired = errors.NewWithStatus(1341, http.StatusOK, "清理截止时间不能为空")
	// ErrOperLogCleanBeforeFuture 表示操作日志清理截止时间不早于当前时间。
	ErrOperLogCleanBeforeFuture = errors.NewWithStatus(1342, http.StatusOK, "清理截止时间必须早于当前时间")
	// ErrOperLogCleanFailed 表示清理操作日志失败。
	ErrOperLogCleanFailed = errors.NewWithStatus(1343, http.StatusInternalServerError, "清理操作日志失败")
	// ErrAuthForbidden 表示当前登录用户无权执行该操作。
	ErrAuthForbidden = errors.NewWithStatus(1344, http.StatusForbidden, "无权执行该操作")
)
