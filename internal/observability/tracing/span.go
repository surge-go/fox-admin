package tracing

import (
	"net/http"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// FinishSpan 根据错误类型收尾业务 span。
func FinishSpan(span trace.Span, err error) {
	if err == nil {
		span.End()
		return
	}

	status := foxerrors.GetStatus(err)
	if status >= http.StatusInternalServerError {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return
	}

	// HTTP 200/4xx 的业务错误只记录事件，不把 span 标记为 Error，避免业务校验失败污染链路错误率。
	span.AddEvent("system.business_error", trace.WithAttributes(
		attribute.Int("error.code", foxerrors.GetCode(err)),
		attribute.Int("http.status_code", status),
		attribute.String("error.message", err.Error()),
	))
	span.End()
}
