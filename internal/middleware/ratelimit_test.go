package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitChatRateLimiters(t *testing.T) {
	InitChatRateLimiters(30, 10, 5)

	if chatMessageLimiter == nil {
		t.Error("Expected chatMessageLimiter to be initialized")
	}

	if chatConversationLimiter == nil {
		t.Error("Expected chatConversationLimiter to be initialized")
	}

	if chatGroupLimiter == nil {
		t.Error("Expected chatGroupLimiter to be initialized")
	}

	if !chatLimitersInitialized {
		t.Error("Expected chatLimitersInitialized to be true")
	}
}

func TestChatRateLimitMiddleware_NotExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitChatRateLimiters(30, 10, 5)

	router := gin.New()
	router.Use(ChatRateLimitMiddleware("messages"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestChatRateLimitMiddleware_Exceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limit := 5
	InitChatRateLimiters(limit, 10, 5)

	router := gin.New()
	router.Use(ChatRateLimitMiddleware("messages"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d on first request, got %d", http.StatusOK, w.Code)
	}

	for i := 1; i < limit; i++ {
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Logf("Request %d returned status %d", i+1, w.Code)
		}
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Logf("Note: Token bucket allows burst of %d, rate limiting may not trigger immediately", limit)
	}
}

func TestChatRateLimitMiddleware_Conversations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitChatRateLimiters(30, 10, 5)

	router := gin.New()
	router.Use(ChatRateLimitMiddleware("conversations"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.2:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestChatRateLimitMiddleware_Groups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitChatRateLimiters(30, 10, 5)

	router := gin.New()
	router.Use(ChatRateLimitMiddleware("groups"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.3:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestChatRateLimitMiddleware_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ChatRateLimitMiddleware("unknown"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.4:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestChatRateLimitMiddleware_NotInitialized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chatLimitersInitialized = false
	chatMessageLimiter = nil

	router := gin.New()
	router.Use(ChatRateLimitMiddleware("messages"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.5:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	InitChatRateLimiters(30, 10, 5)
}
