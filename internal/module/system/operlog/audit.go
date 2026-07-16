package operlog

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"fox-admin/internal/middleware"
	"fox-admin/internal/module/system/enum"
	authcore "fox-admin/pkg/auth"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

const (
	maxAuditRequestBytes  = 16 * 1024
	maxAuditResponseBytes = 16 * 1024
	maxAuditSummaryBytes  = 4 * 1024
)

var auditSkipPaths = map[string]struct{}{
	"/api/v1/system/auth/login":   {},
	"/api/v1/system/auth/refresh": {},
}

// Audit 创建系统操作审计中间件。
func Audit(recorder *Recorder, logger *zap.Logger) fox.HandlerFunc {
	if recorder == nil {
		panic("operation log audit recorder is nil")
	}
	if logger == nil {
		panic("operation log audit logger is nil")
	}
	return func(c *fox.Context) {
		req := c.RawRequest()
		if !shouldAudit(req.Method, req.URL.Path) {
			c.Next()
			return
		}

		policy, exists := auditPolicies[req.Method+" "+req.URL.Path]
		if !exists {
			policy = deriveAuditPolicy(req.URL.Path)
		}
		requestData := captureRequestSummary(req, policy.Fields)
		start := time.Now()
		originalWriter, canCapture := c.RawWriter().(auditGinResponseWriter)
		var capture *auditResponseWriter
		if canCapture {
			capture = &auditResponseWriter{auditGinResponseWriter: originalWriter, limit: maxAuditResponseBytes}
			c.SetRawWriter(capture)
		}

		defer func() {
			recovered := recover()
			costMillis := time.Since(start).Milliseconds()
			if canCapture {
				c.SetRawWriter(originalWriter)
			}
			statusCode := c.Status()
			if statusCode <= 0 {
				statusCode = http.StatusOK
			}
			if recovered != nil {
				statusCode = http.StatusInternalServerError
			}
			businessCode, message := auditResult(statusCode, capture, recovered)
			status := enum.StatusEnabled
			if businessCode != 200 {
				status = enum.StatusDisabled
			}
			userID := auditUserID(c)
			path := c.FullPath()
			if path == "" {
				path = req.URL.Path
			}
			input := &RecordInput{
				RequestID:    c.RequestID(),
				TraceID:      c.TraceID(),
				UserID:       userID,
				Module:       policy.Module,
				Action:       policy.Action,
				Method:       req.Method,
				Path:         path,
				IP:           c.ClientIP(),
				UserAgent:    req.UserAgent(),
				RequestData:  requestData,
				Status:       status,
				StatusCode:   statusCode,
				BusinessCode: businessCode,
				CostMillis:   costMillis,
			}
			if status == enum.StatusDisabled {
				input.ErrorMessage = message
			}
			recorder.Enqueue(input)
			if recovered != nil {
				panic(recovered)
			}
		}()

		c.Next()
	}
}

type auditResponseWriter struct {
	auditGinResponseWriter
	body  bytes.Buffer
	limit int
}

type auditGinResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier
	Status() int
	Size() int
	WriteString(string) (int, error)
	Written() bool
	WriteHeaderNow()
	Pusher() http.Pusher
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	n, err := w.auditGinResponseWriter.Write(data)
	w.capture(data[:min(n, len(data))])
	return n, err
}

func (w *auditResponseWriter) WriteString(value string) (int, error) {
	n, err := w.auditGinResponseWriter.WriteString(value)
	if n > len(value) {
		n = len(value)
	}
	w.capture([]byte(value[:n]))
	return n, err
}

func (w *auditResponseWriter) capture(data []byte) {
	remaining := w.limit - w.body.Len()
	if remaining <= 0 {
		return
	}
	w.body.Write(data[:min(len(data), remaining)])
}

type auditResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func auditResult(statusCode int, capture *auditResponseWriter, recovered any) (int, string) {
	if recovered != nil {
		return 500, "服务器内部错误"
	}
	if capture != nil && capture.body.Len() > 0 {
		var response auditResponse
		if err := json.Unmarshal(capture.body.Bytes(), &response); err == nil && response.Code != 0 {
			return response.Code, strings.TrimSpace(response.Message)
		}
	}
	if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
		return 200, ""
	}
	return statusCode, http.StatusText(statusCode)
}

func shouldAudit(method, path string) bool {
	if _, skip := auditSkipPaths[path]; skip {
		return false
	}
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func auditUserID(c *fox.Context) *int64 {
	value, exists := c.Get(middleware.AuthClaimsKey)
	claims, ok := value.(*authcore.Claims)
	if !exists || !ok || claims == nil || claims.SubjectID <= 0 {
		return nil
	}
	userID := claims.SubjectID
	return &userID
}

func deriveAuditPolicy(path string) AuditPolicy {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	start := -1
	for i := range parts {
		if parts[i] == "system" {
			start = i
			break
		}
	}
	if start < 0 || start+1 >= len(parts) {
		return AuditPolicy{Module: "system.unknown", Action: "unknown"}
	}
	module := strings.ReplaceAll(parts[start+1], "-", "_")
	actionParts := parts[start+2:]
	for i := range actionParts {
		actionParts[i] = strings.ReplaceAll(actionParts[i], "-", "_")
	}
	action := strings.Join(actionParts, "_")
	if action == "" {
		action = "unknown"
	}
	return AuditPolicy{Module: "system." + module, Action: action}
}

type replayReadCloser struct {
	io.Reader
	io.Closer
}

func captureRequestSummary(req *http.Request, fields []string) string {
	if req == nil || req.Body == nil || len(fields) == 0 || !strings.HasPrefix(strings.ToLower(req.Header.Get("Content-Type")), "application/json") {
		return ""
	}
	original := req.Body
	prefix, err := io.ReadAll(io.LimitReader(original, maxAuditRequestBytes+1))
	req.Body = &replayReadCloser{Reader: io.MultiReader(bytes.NewReader(prefix), original), Closer: original}
	if err != nil || len(prefix) > maxAuditRequestBytes {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(prefix, &payload); err != nil {
		return ""
	}
	selected := make(map[string]any, len(fields))
	for _, field := range fields {
		if isSensitiveAuditField(field) {
			continue
		}
		if value, exists := payload[field]; exists {
			selected[field] = redactAuditValue(value)
		}
	}
	if len(selected) == 0 {
		return ""
	}
	encoded, err := json.Marshal(selected)
	if err != nil || len(encoded) > maxAuditSummaryBytes {
		return ""
	}
	return string(encoded)
}

func redactAuditValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for key, child := range typed {
			if !isSensitiveAuditField(key) {
				redacted[key] = redactAuditValue(child)
			}
		}
		return redacted
	case []any:
		redacted := make([]any, len(typed))
		for i := range typed {
			redacted[i] = redactAuditValue(typed[i])
		}
		return redacted
	default:
		return value
	}
}

func isSensitiveAuditField(field string) bool {
	field = strings.ToLower(strings.TrimSpace(field))
	switch field {
	case "password", "old_password", "new_password", "token", "access_token", "refresh_token",
		"authorization", "cookie", "secret", "token_secret", "dsn", "config_value":
		return true
	default:
		return false
	}
}
