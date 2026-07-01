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
	// ErrMenuPermissionRequired 表示菜单权限标识为空。
	ErrMenuPermissionRequired = errors.NewWithStatus(1009, http.StatusOK, "菜单权限标识不能为空")
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
	// ErrRoleAssignMenusReqNil 表示分配角色菜单请求参数为空。
	ErrRoleAssignMenusReqNil = errors.NewWithStatus(1061, http.StatusOK, "分配角色菜单参数不能为空")
	// ErrRoleAssignMenusFailed 表示分配角色菜单失败。
	ErrRoleAssignMenusFailed = errors.NewWithStatus(1062, http.StatusInternalServerError, "分配角色菜单失败")
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
	// ErrAuthUserNotFound 表示登录用户不存在。
	ErrAuthUserNotFound = errors.NewWithStatus(1105, http.StatusOK, "账号或密码错误")
	// ErrAuthPasswordInvalid 表示登录密码错误。
	ErrAuthPasswordInvalid = errors.NewWithStatus(1106, http.StatusOK, "账号或密码错误")
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
)
