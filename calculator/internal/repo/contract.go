package repo

import (
	"context"

	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IntoCollection interface {
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
}

type Callback interface {
	SendError(context.Context, error)
	SendResult(context.Context, []pb.Task)
}

type UserRepo interface {
	Register(ctx context.Context, user models.User) (primitive.ObjectID, error)
	Get(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUsername(ctx context.Context, id primitive.ObjectID, username string) error
	UpdatePassword(ctx context.Context, id primitive.ObjectID, password string) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetCollection() *mongo.Collection
}

type ExpressionRepo interface {
	SetCallback(ctx context.Context, callback Callback)
	Create(ctx context.Context, expression models.Expression, ast tree.Ast) (primitive.ObjectID, error)
	Get(ctx context.Context, id primitive.ObjectID) (*models.Expression, error)
	GetByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Expression, error)
	GetNode(ctx context.Context, id primitive.ObjectID) (*models.Node, error)
	GetFitNodes(ctx context.Context) ([]pb.Task, error)
	SetToError(ctx context.Context, id primitive.ObjectID, err string) error
	SetToNum(ctx context.Context, nodeId primitive.ObjectID, result float64) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByUser(ctx context.Context, userID primitive.ObjectID) error
	GetCollection() *mongo.Collection
	GetNodeCollection() *mongo.Collection
}
