package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestSwaggerHandlerServesV2GeneratedDocument 验证内嵌 Swagger 文档
// 与 UI 资源可正常访问。
func TestSwaggerHandlerServesV2GeneratedDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/swagger/*any", newSwaggerHandler())

	documentResponse := httptest.NewRecorder()
	router.ServeHTTP(documentResponse, httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil))
	if documentResponse.Code != http.StatusOK {
		t.Fatalf("GET /swagger/doc.json status = %d", documentResponse.Code)
	}
	var document map[string]any
	if err := json.Unmarshal(documentResponse.Body.Bytes(), &document); err != nil {
		t.Fatalf("swagger document is not JSON: %v", err)
	}
	if document["swagger"] != "2.0" {
		t.Fatalf("swagger document version = %#v", document["swagger"])
	}

	uiResponse := httptest.NewRecorder()
	router.ServeHTTP(uiResponse, httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil))
	if uiResponse.Code != http.StatusOK {
		t.Fatalf("GET /swagger/index.html status = %d", uiResponse.Code)
	}
	initializerResponse := httptest.NewRecorder()
	router.ServeHTTP(
		initializerResponse,
		httptest.NewRequest(http.MethodGet, "/swagger/swagger-initializer.js", nil),
	)
	if initializerResponse.Code != http.StatusOK {
		t.Fatalf("GET /swagger/swagger-initializer.js status = %d", initializerResponse.Code)
	}
	if !strings.Contains(initializerResponse.Body.String(), "/swagger/doc.json") {
		t.Fatal("swagger UI does not reference the embedded v2 document")
	}
}
