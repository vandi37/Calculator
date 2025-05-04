package service

import (
	"context"

	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Function to get a task
type PostFunc func(float64, float64, tree.Operation) (float64, error)

type Service interface {
	// Adds a new expression
	Add(ctx context.Context, expression string, userId primitive.ObjectID) (primitive.ObjectID, error)
	// Getting the expression
	Get(ctx context.Context, id primitive.ObjectID) (*models.Expression, error)
	// Getting expressions by user id
	GetByUSer(ctx context.Context, userId primitive.ObjectID) ([]models.Expression, error)
	// Sending task result
	DoTask(ctx context.Context, result *pb.Result) error
	// Sending task error
	DoError(ctx context.Context, error *pb.Error) error
	// Getting task chan
	Tasks() <-chan *pb.Task
	// Create a new user
	Register(ctx context.Context, username, password string) (primitive.ObjectID, error)
	// Get jwt token
	Login(ctx context.Context, username, password string) (string, error)
	// Checks the token
	CheckToken(ctx context.Context, token string) (primitive.ObjectID, error)
	// Update username
	UpdateUsername(ctx context.Context, id primitive.ObjectID, username string) error
	// Update password
	UpdatePassword(ctx context.Context, id primitive.ObjectID, password string) error
	// Delete
	Delete(ctx context.Context, id primitive.ObjectID) error
	// Close tasks
	Close() error
	// Init service
	Init(ctx context.Context)
}
