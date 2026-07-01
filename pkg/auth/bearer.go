package auth

import "strings"

// ParseBearer 从 Authorization 头中解析 Bearer token。
func ParseBearer(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", ErrTokenRequired
	}
	fields := strings.Fields(header)
	if len(fields) == 1 {
		if strings.EqualFold(fields[0], tokenTypeBearer) {
			return "", ErrTokenRequired
		}
		return fields[0], nil
	}
	if len(fields) != 2 || !strings.EqualFold(fields[0], tokenTypeBearer) {
		return "", ErrTokenMalformed
	}
	if strings.TrimSpace(fields[1]) == "" {
		return "", ErrTokenRequired
	}
	return fields[1], nil
}
