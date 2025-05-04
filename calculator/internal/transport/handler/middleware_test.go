package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/service"
	"github.com/vandi37/Calculator/internal/service/mock_service"
	"github.com/vandi37/Calculator/internal/transport/handler"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestContentTypeMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		expectedHeader string
	}{
		{
			name:           "Sets JSON content type",
			expectedHeader: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			middleware := handler.ContentType()
			middleware(ctx)

			assert.Equal(t, tt.expectedHeader, w.Header().Get("Content-Type"))
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		path           string
		method         string
		expectErrorLog bool
	}{
		{
			name:           "Success request logs info",
			statusCode:     http.StatusOK,
			path:           "/test",
			method:         http.MethodGet,
			expectErrorLog: false,
		},
		{
			name:           "Error request logs error",
			statusCode:     http.StatusInternalServerError,
			path:           "/error",
			method:         http.MethodPost,
			expectErrorLog: true,
		},
		{
			name:           "Empty path uses URL",
			statusCode:     http.StatusOK,
			path:           "",
			method:         http.MethodGet,
			expectErrorLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			defer logger.Sync()

			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			// Set up expected logging
			if tt.path == "" {
				tt.path = req.URL.String()
			}

			middleware := handler.Logging(logger)
			middleware(ctx)

			ctx.Writer.WriteHeader(tt.statusCode)
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		expectedHeaders  map[string]string
		expectedStatus   int
		isOptionsRequest bool
	}{
		{
			name:   "Regular request sets CORS headers",
			method: http.MethodGet,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST, OPTIONS, GET",
			},
			expectedStatus:   http.StatusOK,
			isOptionsRequest: false,
		},
		{
			name:   "OPTIONS request returns NoContent",
			method: http.MethodOptions,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST, OPTIONS, GET",
			},
			expectedStatus:   http.StatusNoContent,
			isOptionsRequest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			middleware := handler.CORSMiddleware()
			middleware(ctx)

			if !tt.isOptionsRequest {
				ctx.Next()
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			for key, value := range tt.expectedHeaders {
				assert.Equal(t, value, w.Header().Get(key))
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	validToken := "valid_token"
	invalidToken := "invalid_token"
	userId := primitive.NewObjectID()

	tests := []struct {
		name           string
		token          string
		setupMock      func(*mock_service.MockService)
		expectedStatus int
		expectedBody   interface{}
		expectUserId   bool
	}{
		{
			name:  "Valid token",
			token: validToken,
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().CheckToken(gomock.Any(), validToken).Return(userId, nil)
			},
			expectedStatus: http.StatusOK,
			expectUserId:   true,
		},
		{
			name:           "Empty token",
			token:          "",
			setupMock:      func(m *mock_service.MockService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   models.ErrorResponse{Error: handler.Unauthorized},
			expectUserId:   false,
		},
		{
			name:  "Invalid token",
			token: invalidToken,
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().CheckToken(gomock.Any(), invalidToken).Return(primitive.ObjectID{}, service.InvalidToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   models.ErrorResponse{Error: handler.Unauthorized},
			expectUserId:   false,
		},
		{
			name:  "Service error",
			token: validToken,
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().CheckToken(gomock.Any(), validToken).Return(primitive.ObjectID{}, errors.New("some error"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   models.ErrorResponse{Error: handler.Unauthorized},
			expectUserId:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			h := handler.New(mockService, zap.NewNop())

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			middleware := h.AuthMiddleware()
			middleware(ctx)

			if tt.expectedStatus != http.StatusOK {
				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.expectedBody != nil {
					var response models.ErrorResponse
					_ = json.Unmarshal(w.Body.Bytes(), &response)
					assert.Equal(t, tt.expectedBody, response)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.expectUserId {
					assert.Equal(t, userId, ctx.MustGet(handler.UserIDKey))
				}
			}
		})
	}
}

func TestMiddlewareChain(t *testing.T) {
	t.Run("Multiple middlewares execute in order", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mock_service.NewMockService(ctrl)
		h := handler.New(mockService, zap.NewNop())
		logger := zaptest.NewLogger(t)
		defer logger.Sync()

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		router := gin.New()
		router.Use(
			handler.ContentType(),
			handler.Logging(logger),
			handler.CORSMiddleware(),
			h.AuthMiddleware(),
		)

		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		router.ServeHTTP(w, req)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
