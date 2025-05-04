package appservice

import (
	"context"

	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/ms"
	"github.com/vandi37/Calculator/internal/repo"
	"github.com/vandi37/Calculator/internal/service"
	"github.com/vandi37/Calculator/pkg/hash"
	"github.com/vandi37/Calculator/pkg/jwt"
	"github.com/vandi37/Calculator/pkg/parsing/parser"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

const (
	TASK_CAPACITY = 64
)

func New(
	logger *zap.Logger,
	msGetter *ms.MsGetter,
	userRepo repo.UserRepo,
	expressionRepo repo.ExpressionRepo,
	passwordService *hash.PasswordService,
	tokenService *jwt.TokenService,
) *Service {
	return &Service{msGetter, userRepo, expressionRepo, make(chan *pb.Task, TASK_CAPACITY), logger, passwordService, tokenService, false}
}

type Service struct {
	msGetter        *ms.MsGetter
	userRepo        repo.UserRepo
	expressionRepo  repo.ExpressionRepo
	tasks           chan *pb.Task
	logger          *zap.Logger
	passwordService *hash.PasswordService
	tokenService    *jwt.TokenService
	closed          bool
}

// Init implements service.Service.
func (s *Service) Init(ctx context.Context) {
	s.expressionRepo.SetCallback(ctx, s)
}

// Close implements service.Service.
func (s *Service) Close() error {
	if s.closed {
		return service.Closed
	}
	close(s.tasks)
	return nil
}

var _ service.Service = (*Service)(nil)

// Register implements service.Service.
func (s *Service) Register(ctx context.Context, username string, password string) (primitive.ObjectID, error) {
	hash, err := s.passwordService.HashPassword(password)
	if err != nil {
		s.logger.Debug("error while hashing password", zap.Error(err))
		return primitive.NilObjectID, err
	}
	id, err := s.userRepo.Register(ctx, models.User{
		Username: username,
		Password: hash,
	})
	if err != nil {
		s.logger.Debug("error while registering user", zap.Error(err))
		return primitive.NilObjectID, err
	}
	s.logger.Debug("user registered", zap.String("id", id.Hex()))
	return id, nil
}

// Login implements service.Service.
func (s *Service) Login(ctx context.Context, username string, password string) (string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Debug("error while getting user", zap.Error(err))
		return "", err
	}
	if err := s.passwordService.ComparePassword(password, user.Password); err != nil {
		s.logger.Debug("error while comparing passwords", zap.Error(err))
		return "", err
	}

	// Consider to add data to token if needed
	// e.g. user.Username, and other information
	token, err := s.tokenService.Generate(user.ID.Hex())
	if err != nil {
		s.logger.Debug("error while generating token", zap.Error(err))
		return "", err
	}
	s.logger.Debug("token generated", zap.String("id", user.ID.Hex()))
	return token, nil
}

// CheckToken implements service.Service.
func (s *Service) CheckToken(ctx context.Context, token string) (primitive.ObjectID, error) {
	parsedToken, err := s.tokenService.Parse(token)
	if err != nil {
		s.logger.Debug("error while parsing token", zap.Error(err))
		// Making error less specific
		return primitive.NilObjectID, service.InvalidToken
	}
	if !parsedToken.Valid {
		s.logger.Debug("invalid token")
		return primitive.NilObjectID, service.InvalidToken
	}
	sub, err := parsedToken.Claims.GetSubject()
	if err != nil {
		s.logger.Debug("error while getting subject", zap.Error(err))
		// Making error less specific
		return primitive.NilObjectID, service.InvalidToken
	}
	id, err := primitive.ObjectIDFromHex(sub)
	if err != nil {
		s.logger.Debug("error while parsing id", zap.Error(err))
		// Making error less specific
		return primitive.NilObjectID, service.InvalidToken
	}
	s.logger.Debug("token checked", zap.String("id", id.Hex()))
	return id, nil
}

// UpdatePassword implements service.Service.
func (s *Service) UpdatePassword(ctx context.Context, id primitive.ObjectID, password string) error {
	hash, err := s.passwordService.HashPassword(password)
	if err != nil {
		s.logger.Debug("error while hashing password", zap.Error(err))
		return err
	}
	err = s.userRepo.UpdatePassword(ctx, id, hash)
	if err != nil {
		s.logger.Debug("error while updating password", zap.Error(err))
		return err
	}
	s.logger.Debug("password updated", zap.String("id", id.Hex()))
	return nil
}

// UpdateUsername implements service.Service.
func (s *Service) UpdateUsername(ctx context.Context, id primitive.ObjectID, username string) error {
	err := s.userRepo.UpdateUsername(ctx, id, username)
	if err != nil {
		s.logger.Debug("error while updating username", zap.Error(err))
		return err
	}
	s.logger.Debug("username updated", zap.String("id", id.Hex()))
	return nil
}

// Delete implements service.Service.
func (s *Service) Delete(ctx context.Context, id primitive.ObjectID) error {
	err := s.userRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Debug("error while deleting user", zap.Error(err))
		return err
	}
	s.logger.Debug("user deleted", zap.String("id", id.Hex()))
	return nil
}

// Add implements service.Service.
func (s *Service) Add(ctx context.Context, expression string, userId primitive.ObjectID) (primitive.ObjectID, error) {
	ast, err := parser.Build(expression)
	if err != nil {
		s.logger.Debug("error while parsing expression", zap.Error(err))
		return primitive.NilObjectID, err
	}
	expr := models.Expression{
		UserID: userId,
		Origin: expression,
	}
	id, err := s.expressionRepo.Create(ctx, expr, ast)
	if err != nil {
		s.logger.Debug("error while creating expression", zap.Error(err))
		return primitive.NilObjectID, err
	}
	s.logger.Debug("expression created", zap.String("id", id.Hex()), zap.String("expression", expression))
	return id, nil
}

// DoTask implements service.Service.
func (s *Service) DoTask(ctx context.Context, result *pb.Result) error {
	realId, err := primitive.ObjectIDFromHex(result.Id)
	if err != nil {
		s.logger.Debug("error while converting id", zap.Error(err))
		return err
	}
	err = s.expressionRepo.SetToNum(ctx, realId, result.Result)
	if err != nil {
		s.logger.Debug("error while setting result", zap.Error(err))
		return err
	}
	s.logger.Debug("result set", zap.String("id", result.Id), zap.Float64("result", result.Result))
	return nil
}

// DoError implements service.Service.
func (s *Service) DoError(ctx context.Context, res *pb.Error) error {
	realId, err := primitive.ObjectIDFromHex(res.Id)
	if err != nil {
		s.logger.Debug("error while converting id", zap.Error(err))
		return err
	}
	err = s.expressionRepo.SetToError(ctx, realId, res.Error)
	if err != nil {
		s.logger.Debug("error while setting error", zap.Error(err))
		return err
	}
	s.logger.Debug("error set", zap.String("id", res.Id), zap.String("error", res.Error))
	return nil
}

// Get implements service.Service.
func (s *Service) Get(ctx context.Context, id primitive.ObjectID) (*models.Expression, error) {
	expr, err := s.expressionRepo.Get(ctx, id)
	if err != nil {
		s.logger.Debug("error while getting expression", zap.Error(err))
		return nil, err
	}
	s.logger.Debug("expression got", expr.ZapField())
	return expr, nil
}

// GetAll implements service.Service.
func (s *Service) GetByUSer(ctx context.Context, userId primitive.ObjectID) ([]models.Expression, error) {
	expressions, err := s.expressionRepo.GetByUser(ctx, userId)
	if err != nil {
		s.logger.Debug("error while getting expressions", zap.Error(err))
		return nil, err
	}
	s.logger.Debug("expressions got", zap.String("user_id", userId.Hex()), zap.Int("count", len(expressions)))
	return expressions, nil
}

// Tasks implements service.Service.
func (s *Service) Tasks() <-chan *pb.Task {
	return s.tasks
}
