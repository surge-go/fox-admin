package service

// isValidEnabledStatus 判断启禁用状态是否属于系统支持的数值范围。
func isValidEnabledStatus(status int) bool {
	return status == 0 || status == 1
}
