package enum

const (
	// ConfigValueTypeString 表示字符串配置值。
	ConfigValueTypeString = "string"
	// ConfigValueTypeInt 表示整数配置值。
	ConfigValueTypeInt = "int"
	// ConfigValueTypeBool 表示布尔配置值。
	ConfigValueTypeBool = "bool"
	// ConfigValueTypeJSON 表示 JSON 配置值。
	ConfigValueTypeJSON = "json"
	// DefaultConfigGroup 是默认配置分组。
	DefaultConfigGroup = "default"
)

// IsConfigValueTypeValid 判断配置值类型是否合法。
func IsConfigValueTypeValid(valueType string) bool {
	switch valueType {
	case ConfigValueTypeString, ConfigValueTypeInt, ConfigValueTypeBool, ConfigValueTypeJSON:
		return true
	default:
		return false
	}
}
