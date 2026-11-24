package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"marku/api/internal/response"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	Health(context)

	var envelope response.Envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if recorder.Code != http.StatusOK || envelope.Code != response.CodeOK {
		t.Fatalf("unexpected health response: status=%d body=%+v", recorder.Code, envelope)
	}
}

func TestReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	readyHandler := Ready(db)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	readyHandler(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ready status = %d, want 200", recorder.Code)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	unreadyRecorder := httptest.NewRecorder()
	unreadyContext, _ := gin.CreateTestContext(unreadyRecorder)
	readyHandler(unreadyContext)
	if unreadyRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("unready status = %d, want 503", unreadyRecorder.Code)
	}
}
