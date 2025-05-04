package appservice_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/config"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/ms"
	"github.com/vandi37/Calculator/internal/repo/mock_repo"
	"github.com/vandi37/Calculator/internal/service/appservice"
	"github.com/vandi37/Calculator/internal/status"
	"github.com/vandi37/Calculator/pkg/hash"
	"github.com/vandi37/Calculator/pkg/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func TestService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repo.NewMockUserRepo(ctrl)
	mockExprRepo := mock_repo.NewMockExpressionRepo(ctrl)

	passwordService := hash.NewPasswordService(nil)
	tokenService := jwt.New("secret", time.Hour, 0)
	msGetter := ms.From(config.Time{
		AdditionMs:       100,
		SubtractionMs:    200,
		MultiplicationMs: 300,
		DivisionMs:       400,
	})

	svc := appservice.New(zap.NewNop(), msGetter, mockUserRepo, mockExprRepo, passwordService, tokenService)

	tests := []struct {
		name        string
		username    string
		password    string
		mockSetup   func()
		expectError bool
	}{
		{
			name:     "Successful registration",
			username: "testuser",
			password: "testpass",
			mockSetup: func() {
				mockUserRepo.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(primitive.NewObjectID(), nil)
			},
		},
		{
			name:     "Password too long",
			username: "testuser",
			password: "thispasswordiswaytoolongandexceedsthemaximumallowedlengthof72characterswhichisthelimitforbcrypt",
			mockSetup: func() {
				// No repo calls expected
			},
			expectError: true,
		},
		{
			name:     "Repository error",
			username: "testuser",
			password: "testpass",
			mockSetup: func() {
				mockUserRepo.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(primitive.NilObjectID, errors.New("repo error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			id, err := svc.Register(context.Background(), tt.username, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, primitive.NilObjectID, id)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, primitive.NilObjectID, id)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repo.NewMockUserRepo(ctrl)
	mockExprRepo := mock_repo.NewMockExpressionRepo(ctrl)

	passwordService := hash.NewPasswordService(nil)
	tokenService := jwt.New("secret", time.Hour, 0)
	msGetter := ms.From(config.Time{
		AdditionMs:       100,
		SubtractionMs:    200,
		MultiplicationMs: 300,
		DivisionMs:       400,
	})

	svc := appservice.New(zap.NewNop(), msGetter, mockUserRepo, mockExprRepo, passwordService, tokenService)

	userID := primitive.NewObjectID()
	hashedPass, _ := passwordService.HashPassword("correctpass")

	tests := []struct {
		name        string
		username    string
		password    string
		mockSetup   func()
		expectError bool
	}{
		{
			name:     "Successful login",
			username: "testuser",
			password: "correctpass",
			mockSetup: func() {
				mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").
					Return(&models.User{
						ID:       userID,
						Username: "testuser",
						Password: hashedPass,
					}, nil)
			},
		},
		{
			name:     "User not found",
			username: "nonexistent",
			password: "anypass",
			mockSetup: func() {
				mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "nonexistent").
					Return(nil, errors.New("not found"))
			},
			expectError: true,
		},
		{
			name:     "Wrong password",
			username: "testuser",
			password: "wrongpass",
			mockSetup: func() {
				mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").
					Return(&models.User{
						ID:       userID,
						Username: "testuser",
						Password: hashedPass,
					}, nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			token, err := svc.Login(context.Background(), tt.username, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestService_CheckToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repo.NewMockUserRepo(ctrl)
	mockExprRepo := mock_repo.NewMockExpressionRepo(ctrl)

	passwordService := hash.NewPasswordService(nil)
	tokenService := jwt.New("secret", time.Hour, 0)
	msGetter := ms.From(config.Time{
		AdditionMs:       100,
		SubtractionMs:    200,
		MultiplicationMs: 300,
		DivisionMs:       400,
	})

	svc := appservice.New(zap.NewNop(), msGetter, mockUserRepo, mockExprRepo, passwordService, tokenService)

	userID := primitive.NewObjectID()
	validToken, _ := tokenService.Generate(userID.Hex())

	tests := []struct {
		name        string
		token       string
		expectError bool
		expectedID  primitive.ObjectID
	}{
		{
			name:       "Valid token",
			token:      validToken,
			expectedID: userID,
		},
		{
			name:        "Invalid token",
			token:       "invalid.token.here",
			expectError: true,
		},
		{
			name:        "Empty token",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := svc.CheckToken(context.Background(), tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, primitive.NilObjectID, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestService_AddExpression(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repo.NewMockUserRepo(ctrl)
	mockExprRepo := mock_repo.NewMockExpressionRepo(ctrl)

	passwordService := hash.NewPasswordService(nil)
	tokenService := jwt.New("secret", time.Hour, 0)
	msGetter := ms.From(config.Time{
		AdditionMs:       100,
		SubtractionMs:    200,
		MultiplicationMs: 300,
		DivisionMs:       400,
	})

	svc := appservice.New(zap.NewNop(), msGetter, mockUserRepo, mockExprRepo, passwordService, tokenService)

	userID := primitive.NewObjectID()

	tests := []struct {
		name        string
		expression  string
		mockSetup   func()
		expectError bool
	}{
		{
			name:       "Valid expression",
			expression: "2+2",
			mockSetup: func() {
				mockExprRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(primitive.NewObjectID(), nil)
			},
		},
		{
			name:        "Invalid expression",
			expression:  "2+",
			expectError: true,
		},
		{
			name:       "Repository error",
			expression: "2+2",
			mockSetup: func() {
				mockExprRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(primitive.NilObjectID, errors.New("repo error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			id, err := svc.Add(context.Background(), tt.expression, userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, primitive.NilObjectID, id)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, primitive.NilObjectID, id)
			}
		})
	}
}

func TestService_DoTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repo.NewMockUserRepo(ctrl)
	mockExprRepo := mock_repo.NewMockExpressionRepo(ctrl)

	passwordService := hash.NewPasswordService(nil)
	tokenService := jwt.New("secret", time.Hour, 0)
	msGetter := ms.From(config.Time{
		AdditionMs:       100,
		SubtractionMs:    200,
		MultiplicationMs: 300,
		DivisionMs:       400,
	})

	svc := appservice.New(zap.NewNop(), msGetter, mockUserRepo, mockExprRepo, passwordService, tokenService)

	tests := []struct {
		name        string
		result      *pb.Result
		mockSetup   func()
		expectError bool
	}{
		{
			name: "Successful result processing",
			result: &pb.Result{
				Id:     primitive.NewObjectID().Hex(),
				Result: 42.0,
			},
			mockSetup: func() {
				mockExprRepo.EXPECT().SetToNum(gomock.Any(), gomock.Any(), 42.0).
					Return(nil)
			},
		},
		{
			name: "Invalid ID format",
			result: &pb.Result{
				Id:     "invalid-id",
				Result: 42.0,
			},
			expectError: true,
		},
		{
			name: "Repository error",
			result: &pb.Result{
				Id:     primitive.NewObjectID().Hex(),
				Result: 42.0,
			},
			mockSetup: func() {
				mockExprRepo.EXPECT().SetToNum(gomock.Any(), gomock.Any(), 42.0).
					Return(errors.New("repo error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			err := svc.DoTask(context.Background(), tt.result)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_DoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repo.NewMockUserRepo(ctrl)
	mockExprRepo := mock_repo.NewMockExpressionRepo(ctrl)

	passwordService := hash.NewPasswordService(nil)
	tokenService := jwt.New("secret", time.Hour, 0)
	msGetter := ms.From(config.Time{
		AdditionMs:       100,
		SubtractionMs:    200,
		MultiplicationMs: 300,
		DivisionMs:       400,
	})

	svc := appservice.New(zap.NewNop(), msGetter, mockUserRepo, mockExprRepo, passwordService, tokenService)

	tests := []struct {
		name        string
		errorRes    *pb.Error
		mockSetup   func()
		expectError bool
	}{
		{
			name: "Successful error processing",
			errorRes: &pb.Error{
				Id:    primitive.NewObjectID().Hex(),
				Error: "division by zero",
			},
			mockSetup: func() {
				mockExprRepo.EXPECT().SetToError(gomock.Any(), gomock.Any(), "division by zero").
					Return(nil)
			},
		},
		{
			name: "Invalid ID format",
			errorRes: &pb.Error{
				Id:    "invalid-id",
				Error: "division by zero",
			},
			expectError: true,
		},
		{
			name: "Repository error",
			errorRes: &pb.Error{
				Id:    primitive.NewObjectID().Hex(),
				Error: "division by zero",
			},
			mockSetup: func() {
				mockExprRepo.EXPECT().SetToError(gomock.Any(), gomock.Any(), "division by zero").
					Return(errors.New("repo error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			err := svc.DoError(context.Background(), tt.errorRes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetExpressions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repo.NewMockUserRepo(ctrl)
	mockExprRepo := mock_repo.NewMockExpressionRepo(ctrl)

	passwordService := hash.NewPasswordService(nil)
	tokenService := jwt.New("secret", time.Hour, 0)
	msGetter := ms.From(config.Time{
		AdditionMs:       100,
		SubtractionMs:    200,
		MultiplicationMs: 300,
		DivisionMs:       400,
	})

	svc := appservice.New(zap.NewNop(), msGetter, mockUserRepo, mockExprRepo, passwordService, tokenService)

	userID := primitive.NewObjectID()
	result := 4.0
	expr := models.Expression{
		ID:     primitive.NewObjectID(),
		UserID: userID,
		Origin: "2+2",
		Status: status.Finished,
		Result: &result,
	}

	tests := []struct {
		name        string
		mockSetup   func()
		expectError bool
	}{
		{
			name: "Successful get expressions",
			mockSetup: func() {
				mockExprRepo.EXPECT().GetByUser(gomock.Any(), userID).
					Return([]models.Expression{expr}, nil)
			},
		},
		{
			name: "Repository error",
			mockSetup: func() {
				mockExprRepo.EXPECT().GetByUser(gomock.Any(), userID).
					Return(nil, errors.New("repo error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			expressions, err := svc.GetByUSer(context.Background(), userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, expressions)
			} else {
				assert.NoError(t, err)
				assert.Len(t, expressions, 1)
				assert.Equal(t, expr.Origin, expressions[0].Origin)
				assert.Equal(t, expr.Result, expressions[0].Result)
			}
		})
	}
}

func TestService_SendResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repo.NewMockUserRepo(ctrl)
	mockExprRepo := mock_repo.NewMockExpressionRepo(ctrl)

	passwordService := hash.NewPasswordService(nil)
	tokenService := jwt.New("secret", time.Hour, 0)
	msGetter := ms.From(config.Time{
		AdditionMs:       100,
		SubtractionMs:    200,
		MultiplicationMs: 300,
		DivisionMs:       400,
	})

	svc := appservice.New(zap.NewNop(), msGetter, mockUserRepo, mockExprRepo, passwordService, tokenService)

	tasks := []pb.Task{
		{
			Id:        primitive.NewObjectID().Hex(),
			Operation: pb.Operation_ADD,
			Arg1:      1,
			Arg2:      2,
		},
		{
			Id:        primitive.NewObjectID().Hex(),
			Operation: pb.Operation_DIVIDE,
			Arg1:      10,
			Arg2:      2,
		},
	}

	// Start a goroutine to consume tasks from the channel
	done := make(chan struct{})
	go func() {
		defer close(done)
		for task := range svc.Tasks() {
			switch task.Operation {
			case pb.Operation_ADD:
				assert.Equal(t, int32(100), task.OperationTime)
			case pb.Operation_DIVIDE:
				assert.Equal(t, int32(400), task.OperationTime)
			}
		}
	}()

	// Call SendResult
	svc.SendResult(context.Background(), tasks)

	// Close tasks channel and wait for goroutine to finish
	svc.Close()
	<-done
}
