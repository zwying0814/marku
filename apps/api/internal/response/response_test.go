package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAsAppError(t *testing.T) {
	known := New(http.StatusBadRequest, CodeBadRequest, "bad request")
	if got := AsAppError(known); got != known {
		t.Fatalf("AsAppError(known) = %p, want %p", got, known)
	}

	wrapped := fmt.Errorf("outer: %w", known)
	if got := AsAppError(wrapped); got.Code != CodeBadRequest {
		t.Fatalf("wrapped code = %d, want %d", got.Code, CodeBadRequest)
	}

	cause := errors.New("boom")
	unknown := AsAppError(cause)
	if unknown.HTTPStatus != http.StatusInternalServerError || unknown.Code != CodeInternal {
		t.Fatalf("unknown error = %+v", unknown)
	}
	if !errors.Is(unknown, cause) {
		t.Fatal("unknown error should preserve its cause")
	}
}

func TestOKAndAbort(t *testing.T) {
	gin.SetMode(gin.TestMode)

	okRecorder := httptest.NewRecorder()
	okContext, _ := gin.CreateTestContext(okRecorder)
	OK(okContext, gin.H{"status": "ok"})

	var okEnvelope Envelope
	if err := json.Unmarshal(okRecorder.Body.Bytes(), &okEnvelope); err != nil {
		t.Fatalf("decode ok response: %v", err)
	}
	if okRecorder.Code != http.StatusOK || okEnvelope.Code != CodeOK || okEnvelope.Message != "ok" {
		t.Fatalf("unexpected ok response: status=%d body=%+v", okRecorder.Code, okEnvelope)
	}

	errRecorder := httptest.NewRecorder()
	errContext, _ := gin.CreateTestContext(errRecorder)
	Abort(errContext, New(http.StatusNotFound, CodeNotFound, "not found"))

	var errEnvelope Envelope
	if err := json.Unmarshal(errRecorder.Body.Bytes(), &errEnvelope); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errRecorder.Code != http.StatusNotFound || errEnvelope.Code != CodeNotFound || errEnvelope.Message != "not found" {
		t.Fatalf("unexpected error response: status=%d body=%+v", errRecorder.Code, errEnvelope)
	}
}
