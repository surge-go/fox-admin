package enum

const (
	// StatusDisabled 表示禁用状态。
	StatusDisabled = 0
	// StatusEnabled 表示启用状态。
	StatusEnabled = 1
)

// IsStatusValid 判断状态值是否在系统支持范围内。
func IsStatusValid(status int) bool {
	return status == StatusDisabled || status == StatusEnabled
}
