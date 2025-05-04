package models

import (
	"time"

	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/status"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type NodeType bool

const (
	Number    NodeType = true
	Operation NodeType = false
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"-"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Node struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type     NodeType           `bson:"type" json:"type"`
	Tree     *TreeNode          `bson:"tree,omitempty" json:"tree"`
	Number   *float64           `bson:"number,omitempty" json:"number"`
	SendedAt *time.Time         `bson:"sended_at,omitempty" json:"sended_at,omitempty"`
}

type TreeNode struct {
	Operator pb.Operation       `bson:"operator" json:"operator"`
	Left     primitive.ObjectID `bson:"left" json:"left"`
	Right    primitive.ObjectID `bson:"right" json:"right"`
}

type Expression struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Origin    string             `bson:"origin" json:"origin"`
	Error     string            `bson:"error,omitempty" json:"error,omitempty"`
	Result    *float64           `bson:"result,omitempty" json:"result,omitempty"`
	NodeID    primitive.ObjectID `bson:"node_id,omitempty" json:"-"`
	Status    status.Status      `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

func (e *Expression) ZapField() zap.Field {
	fields := []zap.Field{
		zap.String("id", e.ID.Hex()),
		zap.String("user_id", e.UserID.Hex()),
		zap.String("origin", e.Origin),
		zap.String("status", e.Status.String()),
	}
	if e.Error != "" {
		fields = append(fields, zap.String("error", e.Error))
	} else if e.Result != nil {
		fields = append(fields, zap.Float64("result", *e.Result))
	}
	fields = append(fields, zap.Time("created_at", e.CreatedAt))
	return zap.Dict("expression", fields...)
}

type AggregatedNode struct {
	ID   primitive.ObjectID `bson:"_id"`
	Tree struct {
		Operator pb.Operation `bson:"operator"`
	} `bson:"tree"`
	LeftNode struct {
		Number float64 `bson:"number"`
	} `bson:"leftNode"`
	RightNode struct {
		Number float64 `bson:"number"`
	} `bson:"rightNode"`
}
