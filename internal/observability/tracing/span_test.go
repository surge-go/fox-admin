package tracing

import (
	"context"
	"net/http"
	"testing"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestFinishSpanRecordsBusinessErrorAsEvent(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	_, span := provider.Tracer("test").Start(context.Background(), "business")

	FinishSpan(span, foxerrors.NewWithStatus(1001, http.StatusOK, "业务错误"))

	ended := recorder.Ended()
	if len(ended) != 1 {
		t.Fatalf("ended span count = %d, want 1", len(ended))
	}
	got := tracetest.SpanStubFromReadOnlySpan(ended[0])
	if got.Status.Code == codes.Error {
		t.Fatalf("business error span status = %v, want not error", got.Status.Code)
	}
	if len(got.Events) != 1 || got.Events[0].Name != "system.business_error" {
		t.Fatalf("business error events = %#v, want system.business_error", got.Events)
	}
}

func TestFinishSpanRecordsServerErrorAsSpanError(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	_, span := provider.Tracer("test").Start(context.Background(), "server")

	FinishSpan(span, foxerrors.NewWithStatus(1002, http.StatusInternalServerError, "系统错误"))

	ended := recorder.Ended()
	if len(ended) != 1 {
		t.Fatalf("ended span count = %d, want 1", len(ended))
	}
	got := tracetest.SpanStubFromReadOnlySpan(ended[0])
	if got.Status.Code != codes.Error {
		t.Fatalf("server error span status = %v, want error", got.Status.Code)
	}
}
