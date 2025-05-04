package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/repo"
	"github.com/vandi37/Calculator/internal/service/mock_service"
	"github.com/vandi37/Calculator/internal/status"
	"github.com/vandi37/Calculator/internal/transport/handler"
	"github.com/vandi37/Calculator/pkg/hash"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestCalcHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mock_service.MockService, primitive.ObjectID)
		userId         interface{}
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			requestBody: models.CalculationRequest{
				Expression: "2+2",
			},
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().Add(gomock.Any(), "2+2", userId).Return(primitive.NewObjectID(), nil)
			},
			userId:         primitive.NewObjectID(),
			expectedStatus: http.StatusCreated,
			expectedBody:   models.CreatedResponse{Id: primitive.NewObjectID()},
		},
		{
			name:           "Empty expression",
			requestBody:    models.CalculationRequest{Expression: ""},
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         primitive.NewObjectID(),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name:           "Invalid body",
			requestBody:    "invalid",
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         primitive.NewObjectID(),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name: "Service error",
			requestBody: models.CalculationRequest{
				Expression: "2+2",
			},
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().Add(gomock.Any(), "2+2", userId).Return(primitive.ObjectID{}, errors.New("some error"))
			},
			userId:         primitive.NewObjectID(),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.ErrorResponse{Error: handler.InternalError},
		},
		{
			name: "Unauthorized",
			requestBody: models.CalculationRequest{
				Expression: "2+2",
			},
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   models.ErrorResponse{Error: handler.Unauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			h := handler.New(mockService, zap.NewNop())

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/calculate", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			if tt.userId != nil {
				ctx.Set(handler.UserIDKey, tt.userId)
			}

			if tt.setupMock != nil && tt.userId != nil {
				tt.setupMock(mockService, tt.userId.(primitive.ObjectID))
			}

			h.CalcHandler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response interface{}
				if tt.expectedStatus == http.StatusCreated {
					response = &models.CreatedResponse{}
				} else {
					response = &models.ErrorResponse{}
				}
				_ = json.Unmarshal(w.Body.Bytes(), response)
				if createdResp, ok := tt.expectedBody.(models.CreatedResponse); ok {
					assert.NotEmpty(t, createdResp.Id)
				} else {
					assert.Equal(t, tt.expectedBody, *response.(*models.ErrorResponse))
				}
			}
		})
	}
}

func TestExpressionsHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mock_service.MockService, primitive.ObjectID)
		userId         interface{}
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().GetByUSer(gomock.Any(), userId).Return([]models.Expression{
					{ID: primitive.NewObjectID(), Origin: "2+2", Status: status.Finished},
				}, nil)
			},
			userId:         primitive.NewObjectID(),
			expectedStatus: http.StatusOK,
			expectedBody: models.ExpressionsResponse{
				Expressions: []models.Expression{
					{ID: primitive.NewObjectID(), Origin: "2+2", Status: status.Finished},
				},
			},
		},
		{
			name:           "Unauthorized",
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   models.ErrorResponse{Error: handler.Unauthorized},
		},
		{
			name: "Service error",
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().GetByUSer(gomock.Any(), userId).Return(nil, errors.New("some error"))
			},
			userId:         primitive.NewObjectID(),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.ErrorResponse{Error: handler.InternalError},
		},
		{
			name: "Empty expressions",
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().GetByUSer(gomock.Any(), userId).Return([]models.Expression{}, nil)
			},
			userId:         primitive.NewObjectID(),
			expectedStatus: http.StatusOK,
			expectedBody: models.ExpressionsResponse{
				Expressions: []models.Expression{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			h := handler.New(mockService, zap.NewNop())

			req, _ := http.NewRequest(http.MethodGet, "/expressions", nil)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			if tt.userId != nil {
				ctx.Set(handler.UserIDKey, tt.userId)
			}

			if tt.setupMock != nil && tt.userId != nil {
				tt.setupMock(mockService, tt.userId.(primitive.ObjectID))
			}

			h.ExpressionsHandler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response interface{}
				if tt.expectedStatus == http.StatusOK {
					response = &models.ExpressionsResponse{}
				} else {
					response = &models.ErrorResponse{}
				}
				_ = json.Unmarshal(w.Body.Bytes(), response)
				if exprResp, ok := tt.expectedBody.(models.ExpressionsResponse); ok {
					if len(exprResp.Expressions) > 0 {
						assert.NotEmpty(t, response.(*models.ExpressionsResponse).Expressions[0].ID)
					} else {
						assert.Empty(t, response.(*models.ExpressionsResponse).Expressions)
					}
				} else {
					assert.Equal(t, tt.expectedBody, *response.(*models.ErrorResponse))
				}
			}
		})
	}
}

func TestGetByIdHandler(t *testing.T) {
	validId := primitive.NewObjectID()
	invalidId := "invalid"
	validUserId := primitive.NewObjectID()

	tests := []struct {
		name           string
		idParam        string
		userId         interface{}
		setupMock      func(*mock_service.MockService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:    "Success",
			idParam: validId.Hex(),
			userId:  validUserId,
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Get(gomock.Any(), validId).Return(&models.Expression{
					ID:     validId,
					UserID: validUserId,
					Origin: "2+2",
					Status: status.Finished,
					Result: func() *float64 { f := 4.0; return &f }(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: models.ExpressionResponse{
				Expression: models.Expression{
					ID:     validId,
					Origin: "2+2",
					Status: status.Finished,
					Result: func() *float64 { f := 4.0; return &f }(),
				},
			},
		},
		{
			name:           "Invalid ID",
			idParam:        invalidId,
			userId:         primitive.NewObjectID(),
			setupMock:      func(m *mock_service.MockService) {},
			expectedStatus: http.StatusNotFound,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidId},
		},
		{
			name:    "Not found",
			idParam: validId.Hex(),
			userId:  primitive.NewObjectID(),
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Get(gomock.Any(), validId).Return(nil, repo.ExpressionNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   models.ErrorResponse{Error: repo.ExpressionNotFound.Error()},
		},
		{
			name:    "Service error",
			idParam: validId.Hex(),
			userId:  primitive.NewObjectID(),
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Get(gomock.Any(), validId).Return(nil, errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.ErrorResponse{Error: handler.InternalError},
		},
		{
			name:    "Unauthorized",
			idParam: validId.Hex(),
			userId:  nil,
			setupMock: func(m *mock_service.MockService) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   models.ErrorResponse{Error: handler.Unauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			h := handler.New(mockService, zap.NewNop())

			req, _ := http.NewRequest(http.MethodGet, "/expressions/"+tt.idParam, nil)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			if tt.userId != nil {
				ctx.Set(handler.UserIDKey, tt.userId)
			}

			if tt.setupMock != nil && tt.idParam != invalidId {
				tt.setupMock(mockService)
			}

			h.GetByIdHandler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response interface{}
				if tt.expectedStatus == http.StatusOK {
					response = &models.ExpressionResponse{}
				} else {
					response = &models.ErrorResponse{}
				}
				_ = json.Unmarshal(w.Body.Bytes(), response)
				if exprResp, ok := tt.expectedBody.(models.ExpressionResponse); ok {
					assert.Equal(t, exprResp.Expression.Origin, response.(*models.ExpressionResponse).Expression.Origin)
					assert.Equal(t, exprResp.Expression.Status, response.(*models.ExpressionResponse).Expression.Status)
					assert.Equal(t, exprResp.Expression.Result, response.(*models.ExpressionResponse).Expression.Result)
				} else {
					assert.Equal(t, tt.expectedBody, *response.(*models.ErrorResponse))
				}
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mock_service.MockService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			requestBody: models.UserRequest{
				Username: "user",
				Password: "pass",
			},
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Register(gomock.Any(), "user", "pass").Return(primitive.NewObjectID(), nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   models.CreatedResponse{Id: primitive.NewObjectID()},
		},
		{
			name:           "Empty username",
			requestBody:    models.UserRequest{Password: "pass"},
			setupMock:      func(m *mock_service.MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name:           "Empty password",
			requestBody:    models.UserRequest{Username: "user"},
			setupMock:      func(m *mock_service.MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name:           "Invalid body",
			requestBody:    "invalid",
			setupMock:      func(m *mock_service.MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name: "Username exists",
			requestBody: models.UserRequest{
				Username: "user",
				Password: "pass",
			},
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Register(gomock.Any(), "user", "pass").Return(primitive.ObjectID{}, repo.UsernameTaken)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   models.ErrorResponse{Error: repo.UsernameTaken.Error()},
		},
		{
			name: "Service error",
			requestBody: models.UserRequest{
				Username: "user",
				Password: "pass",
			},
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Register(gomock.Any(), "user", "pass").Return(primitive.ObjectID{}, errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.ErrorResponse{Error: handler.InternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			h := handler.New(mockService, zap.NewNop())

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			h.RegisterHandler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response interface{}
				if tt.expectedStatus == http.StatusCreated {
					response = &models.CreatedResponse{}
				} else {
					response = &models.ErrorResponse{}
				}
				_ = json.Unmarshal(w.Body.Bytes(), response)
				if createdResp, ok := tt.expectedBody.(models.CreatedResponse); ok {
					assert.NotEmpty(t, createdResp.Id)
				} else {
					assert.Equal(t, tt.expectedBody, *response.(*models.ErrorResponse))
				}
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mock_service.MockService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			requestBody: models.UserRequest{
				Username: "user",
				Password: "pass",
			},
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Login(gomock.Any(), "user", "pass").Return("token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &models.TokenResponse{Token: "token"},
		},
		{
			name:           "Invalid body",
			requestBody:    "invalid",
			setupMock:      func(m *mock_service.MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   &models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name: "Invalid credentials",
			requestBody: models.UserRequest{
				Username: "user",
				Password: "wrong",
			},
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Login(gomock.Any(), "user", "wrong").Return("", hash.InvalidPassword)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   &models.ErrorResponse{Error: hash.InvalidPassword.Error()},
		},
		{
			name: "User not found",
			requestBody: models.UserRequest{
				Username: "nonexistent",
				Password: "pass",
			},
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Login(gomock.Any(), "nonexistent", "pass").Return("", repo.UserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   &models.ErrorResponse{Error: repo.UserNotFound.Error()},
		},
		{
			name: "Service error",
			requestBody: models.UserRequest{
				Username: "user",
				Password: "pass",
			},
			setupMock: func(m *mock_service.MockService) {
				m.EXPECT().Login(gomock.Any(), "user", "pass").Return("", errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   &models.ErrorResponse{Error: handler.InternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			h := handler.New(mockService, zap.NewNop())

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			h.LoginHandler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response interface{}
				if tt.expectedStatus == http.StatusOK {
					response = &models.TokenResponse{}
				} else {
					response = &models.ErrorResponse{}
				}
				_ = json.Unmarshal(w.Body.Bytes(), response)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestChangeUsernameHandler(t *testing.T) {
	userId := primitive.NewObjectID()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mock_service.MockService, primitive.ObjectID)
		userId         interface{}
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			requestBody: models.UsernameRequest{
				Username: "newuser",
			},
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().UpdateUsername(gomock.Any(), userId, "newuser").Return(nil)
			},
			userId:         userId,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Unauthorized",
			requestBody:    models.UsernameRequest{Username: "newuser"},
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   models.ErrorResponse{Error: handler.Unauthorized},
		},
		{
			name:           "Empty username",
			requestBody:    models.UsernameRequest{Username: ""},
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         userId,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name:           "Invalid body",
			requestBody:    "invalid",
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         userId,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name: "Username exists",
			requestBody: models.UsernameRequest{
				Username: "existing",
			},
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().UpdateUsername(gomock.Any(), userId, "existing").Return(repo.UsernameTaken)
			},
			userId:         userId,
			expectedStatus: http.StatusConflict,
			expectedBody:   models.ErrorResponse{Error: repo.UsernameTaken.Error()},
		},
		{
			name: "Service error",
			requestBody: models.UsernameRequest{
				Username: "newuser",
			},
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().UpdateUsername(gomock.Any(), userId, "newuser").Return(errors.New("some error"))
			},
			userId:         userId,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.ErrorResponse{Error: handler.InternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			h := handler.New(mockService, zap.NewNop())

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/username", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			if tt.userId != nil {
				ctx.Set(handler.UserIDKey, tt.userId)
			}

			if tt.setupMock != nil && tt.userId != nil {
				tt.setupMock(mockService, tt.userId.(primitive.ObjectID))
			}

			h.ChangeUsernameHandler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response models.ErrorResponse
				_ = json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestChangePasswordHandler(t *testing.T) {
	userId := primitive.NewObjectID()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mock_service.MockService, primitive.ObjectID)
		userId         interface{}
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			requestBody: models.PasswordRequest{
				Password: "newpass",
			},
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().UpdatePassword(gomock.Any(), userId, "newpass").Return(nil)
			},
			userId:         userId,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Unauthorized",
			requestBody:    models.PasswordRequest{Password: "newpass"},
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   models.ErrorResponse{Error: handler.Unauthorized},
		},
		{
			name:           "Empty password",
			requestBody:    models.PasswordRequest{Password: ""},
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         userId,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name:           "Invalid body",
			requestBody:    "invalid",
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         userId,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.ErrorResponse{Error: handler.InvalidBody},
		},
		{
			name: "Service error",
			requestBody: models.PasswordRequest{
				Password: "newpass",
			},
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().UpdatePassword(gomock.Any(), userId, "newpass").Return(errors.New("some error"))
			},
			userId:         userId,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.ErrorResponse{Error: handler.InternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			h := handler.New(mockService, zap.NewNop())

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/password", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			if tt.userId != nil {
				ctx.Set(handler.UserIDKey, tt.userId)
			}

			if tt.setupMock != nil && tt.userId != nil {
				tt.setupMock(mockService, tt.userId.(primitive.ObjectID))
			}

			h.ChangePasswordHandler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response models.ErrorResponse
				_ = json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestDeleteHandler(t *testing.T) {
	userId := primitive.NewObjectID()

	tests := []struct {
		name           string
		setupMock      func(*mock_service.MockService, primitive.ObjectID)
		userId         interface{}
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().Delete(gomock.Any(), userId).Return(nil)
			},
			userId:         userId,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Unauthorized",
			setupMock:      func(m *mock_service.MockService, userId primitive.ObjectID) {},
			userId:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   models.ErrorResponse{Error: handler.Unauthorized},
		},
		{
			name: "Service error",
			setupMock: func(m *mock_service.MockService, userId primitive.ObjectID) {
				m.EXPECT().Delete(gomock.Any(), userId).Return(errors.New("some error"))
			},
			userId:         userId,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.ErrorResponse{Error: handler.InternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			h := handler.New(mockService, zap.NewNop())

			req, _ := http.NewRequest(http.MethodDelete, "/account", nil)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			if tt.userId != nil {
				ctx.Set(handler.UserIDKey, tt.userId)
			}

			if tt.setupMock != nil && tt.userId != nil {
				tt.setupMock(mockService, tt.userId.(primitive.ObjectID))
			}

			h.DeleteHandler(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response models.ErrorResponse
				_ = json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}
