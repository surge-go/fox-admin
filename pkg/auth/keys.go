package auth

import (
	"fmt"
	"strconv"
)

type keyBuilder struct {
	prefix string
}

func (k keyBuilder) session(sessionID string) string {
	return k.prefix + ":session:" + sessionID
}

func (k keyBuilder) sessionMeta(sessionID string) string {
	return k.prefix + ":session_meta:" + sessionID
}

func (k keyBuilder) subjectSessions(subjectType SubjectType, subjectID int64) string {
	return fmt.Sprintf("%s:subject_sessions:%s:%s", k.prefix, subjectType, strconv.FormatInt(subjectID, 10))
}

func (k keyBuilder) platformSessions(subjectType SubjectType, subjectID int64, platform Platform) string {
	return fmt.Sprintf("%s:platform_sessions:%s:%s:%s", k.prefix, subjectType, strconv.FormatInt(subjectID, 10), platform)
}

func (k keyBuilder) deviceSessions(subjectType SubjectType, subjectID int64, deviceID string) string {
	return fmt.Sprintf("%s:device_sessions:%s:%s:%s", k.prefix, subjectType, strconv.FormatInt(subjectID, 10), deviceID)
}

func (k keyBuilder) refresh(hash string) string {
	return k.prefix + ":refresh:" + hash
}

func (k keyBuilder) sessionRefresh(sessionID string) string {
	return k.prefix + ":session_refresh:" + sessionID
}

func (k keyBuilder) refreshReuse(hash string) string {
	return k.prefix + ":refresh_reuse:" + hash
}
