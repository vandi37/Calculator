package expressionrepo

import (
	"context"
	"errors"
	"sync"
	"time"

	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/repo"
	"github.com/vandi37/Calculator/internal/status"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
	"github.com/vandi37/ferror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	collectionName     = "expressions"
	nodeCollectionName = "nodes"
)

type Repo struct {
	collection     *mongo.Collection
	nodeCollection *mongo.Collection
	d              time.Duration
	callback       repo.Callback
	mu             sync.Mutex
}

// GetCollection implements repo.ExpressionRepo.
func (r *Repo) GetCollection() *mongo.Collection {
	return r.collection
}

// GetNodeCollection implements repo.ExpressionRepo.
func (r *Repo) GetNodeCollection() *mongo.Collection {
	return r.nodeCollection
}

// SetCallback implements repo.ExpressionRepo.
func (r *Repo) SetCallback(ctx context.Context, callback repo.Callback) {
	r.callback = callback
	go r.DoCallback()
}

// SetToError implements repo.ExpressionRepo.
func (r *Repo) SetToError(ctx context.Context, id primitive.ObjectID, errVal string) error {
	var save = ferror.Save("expressionrepo.Repo.SetToError")
	for {
		var node models.Node
		if err := r.nodeCollection.FindOne(ctx, bson.M{"tree.left": id}).Decode(&node); err == mongo.ErrNoDocuments {
		} else if err != nil {
			return save.New(err)
		} else {
			id = node.ID
			continue
		}

		if err := r.nodeCollection.FindOne(ctx, bson.M{"tree.right": id}).Decode(&node); err == mongo.ErrNoDocuments {
			return r.setToError(ctx, save, id, errVal)
		} else if err != nil {
			return save.New(err)
		} else {
			id = node.ID
		}
	}
}

func (r *Repo) setToError(ctx context.Context, save ferror.Save, id primitive.ObjectID, errVal string) error {
	update := bson.M{
		"$set":   bson.M{"status": status.Error, "error": errVal},
		"$unset": bson.M{"result": 1, "node_id": 1},
	}
	if res, err := r.collection.UpdateMany(ctx, bson.M{"node_id": id}, update); err != nil {
		return save.New(err)
	} else if res.MatchedCount == 0 || res.ModifiedCount == 0 {
		return repo.ExpressionNotFound
	}
	if err := r.deleteNodes(ctx, id); err != nil {
		return save.New(err)
	}
	return nil
}

// SetToNum implements repo.ExpressionRepo.
func (r *Repo) SetToNum(ctx context.Context, nodeId primitive.ObjectID, result float64) error {
	var save = ferror.Save("expressionrepo.Repo.SetToNum")
	if res, err := r.collection.UpdateMany(ctx, bson.M{"node_id": nodeId}, bson.M{"$set": bson.M{
		"status": status.Finished,
		"result": result,
	}, "$unset": bson.M{"node_id": 1}}); err != nil {
		return save.New(err)
	} else if res.MatchedCount != 0 && res.ModifiedCount != 0 {
		if err := r.deleteNodes(ctx, nodeId); err != nil {
			return save.New(err)
		}
		return nil
	}

	update := bson.M{
		"$set":   bson.M{"type": models.Number, "number": result},
		"$unset": bson.M{"tree": 1},
	}

	if res, err := r.nodeCollection.UpdateOne(ctx, bson.M{"_id": nodeId}, update); err != nil {
		return save.New(err)
	} else if res.MatchedCount == 0 || res.ModifiedCount == 0 {
		return repo.NodeNotFound
	}
	go r.DoCallback()
	return nil
}

func (r *Repo) createNodes(ctx context.Context, expr tree.ExpressionType, isFirst bool) (*float64, primitive.ObjectID, error) {
	if expr == nil {
		return nil, primitive.NilObjectID, repo.InvalidExpression
	}

	switch v := expr.(type) {
	case tree.Num:
		var num = float64(v)
		if isFirst {
			return &num, primitive.NilObjectID, nil
		}
		node := models.Node{
			Type:   models.Number,
			Number: &num,
		}
		if id, err := r.nodeCollection.InsertOne(ctx, node); err != nil {
			return nil, primitive.NilObjectID, err
		} else {
			return nil, id.InsertedID.(primitive.ObjectID), nil
		}
	case tree.Expression:
		_, leftId, err := r.createNodes(ctx, v.Left, false)
		if err != nil {
			return nil, primitive.NilObjectID, err
		}
		_, rightId, err := r.createNodes(ctx, v.Right, false)
		if err != nil {
			return nil, primitive.NilObjectID, err
		}
		node := models.Node{
			Type: models.Operation,
			Tree: &models.TreeNode{
				Operator: pb.Operation(v.Operation),
				Left:     leftId,
				Right:    rightId,
			},
		}
		if id, err := r.nodeCollection.InsertOne(ctx, node); err != nil {
			return nil, primitive.NilObjectID, err
		} else {
			return nil, id.InsertedID.(primitive.ObjectID), nil
		}
	default:
		return nil, primitive.NilObjectID, repo.InvalidExpression
	}
}

// Create implements repo.ExpressionRepo.
func (r *Repo) Create(ctx context.Context, expression models.Expression, ast tree.Ast) (primitive.ObjectID, error) {
	var save = ferror.Save("expressionrepo.Repo.Create")
	num, id, err := r.createNodes(ctx, ast.Expression, true)
	if err != nil {
		return primitive.NilObjectID, save.New(err)
	}
	if num != nil {
		expression.NodeID = primitive.NilObjectID
		expression.Result = num
		expression.Error = ""
		expression.Status = status.Finished
	} else {
		expression.NodeID = id
		expression.Error = ""
		expression.Result = nil
		expression.Status = status.Pending
	}
	expression.CreatedAt = time.Now()
	if res, err := r.collection.InsertOne(ctx, expression); err != nil {
		return primitive.NilObjectID, save.New(err)
	} else {
		go r.DoCallback()
		return res.InsertedID.(primitive.ObjectID), nil
	}
}

func (r *Repo) deleteNodes(ctx context.Context, nodeID primitive.ObjectID) error {
	var save = ferror.Save("expressionrepo.Repo.deleteNodes")
	var node models.Node
	if err := r.nodeCollection.FindOneAndDelete(ctx, bson.M{"_id": nodeID}).Decode(&node); err == mongo.ErrNoDocuments {
		return repo.NodeNotFound
	} else if err != nil {
		return save.New(err)
	}
	if node.Tree != nil {
		if err := r.deleteNodes(ctx, node.Tree.Left); err != nil {
			return err
		}
		if err := r.deleteNodes(ctx, node.Tree.Right); err != nil {
			return err
		}
	}
	return nil
}

// Delete implements repo.ExpressionRepo.
func (r *Repo) Delete(ctx context.Context, id primitive.ObjectID) error {
	var save = ferror.Save("expressionrepo.Repo.Delete")
	var expr models.Expression
	if err := r.collection.FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(&expr); err == mongo.ErrNoDocuments {
		return repo.ExpressionNotFound
	} else if err != nil {
		return save.New(err)
	}
	if expr.NodeID != primitive.NilObjectID {
		if err := r.deleteNodes(ctx, expr.NodeID); err != nil {
			return err
		}
	}
	return nil
}

// DeleteByUser implements repo.ExpressionRepo.
func (r *Repo) DeleteByUser(ctx context.Context, userID primitive.ObjectID) error {
	var save = ferror.Save("expressionrepo.Repo.DeleteByUser")
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return save.New(err)
	}
	defer cursor.Close(ctx)

	var expressions []models.Expression
	if err := cursor.All(ctx, &expressions); err != nil {
		return save.New(err)
	}
	multiErrors := []error{}

	for _, expr := range expressions {
		if expr.NodeID != primitive.NilObjectID {
			if err := r.deleteNodes(ctx, expr.NodeID); err != nil {
				multiErrors = append(multiErrors, err)
			}
		}
		if _, err := r.collection.DeleteOne(ctx, bson.M{"_id": expr.ID}); err != nil {
			multiErrors = append(multiErrors, err)
		}
	}
	if len(multiErrors) > 0 {
		return save.New(errors.Join(multiErrors...))
	}
	return nil
}

// Get implements repo.ExpressionRepo.
func (r *Repo) Get(ctx context.Context, id primitive.ObjectID) (*models.Expression, error) {
	var save = ferror.Save("expressionrepo.Repo.Get")
	var expr models.Expression
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&expr); err == mongo.ErrNoDocuments {
		return nil, repo.ExpressionNotFound
	} else if err != nil {
		return nil, save.New(err)
	}
	return &expr, nil
}

// GetByUser implements repo.ExpressionRepo.
func (r *Repo) GetByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Expression, error) {
	var save = ferror.Save("expressionrepo.Repo.GetByUser")
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, save.New(err)
	}
	defer cursor.Close(ctx)
	var expressions []models.Expression
	if err := cursor.All(ctx, &expressions); err != nil {
		return nil, save.New(err)
	}
	return expressions, nil
}

// GetNode implements repo.ExpressionRepo.
func (r *Repo) GetNode(ctx context.Context, id primitive.ObjectID) (*models.Node, error) {
	var save = ferror.Save("expressionrepo.Repo.GetNode")
	var node models.Node
	if err := r.nodeCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&node); err == mongo.ErrNoDocuments {
		return nil, repo.NodeNotFound
	} else if err != nil {
		return nil, save.New(err)
	}
	return &node, nil
}

func New(db repo.IntoCollection, d time.Duration) *Repo {
	return &Repo{
		collection:     db.Collection(collectionName),
		nodeCollection: db.Collection(nodeCollectionName),
		d:              d,
	}
}

func (r *Repo) GetFitNodes(ctx context.Context) ([]pb.Task, error) {
	var save = ferror.Save("expressionrepo.Repo.GetFitNodes")

	filter := bson.M{
		"type": models.Operation,
		"tree": bson.M{"$exists": true},
		"$or": []bson.M{
			{"sended_at": bson.M{"$exists": false}},
			{"sended_at": bson.M{"$lt": time.Now().Add(-r.d)}},
		},
	}

	pipeline := []bson.M{
		{"$match": filter},
		{"$lookup": bson.M{
			"from":         nodeCollectionName,
			"localField":   "tree.left",
			"foreignField": "_id",
			"as":           "leftNode",
		}},
		{"$unwind": bson.M{"path": "$leftNode", "preserveNullAndEmptyArrays": false}},
		{"$lookup": bson.M{
			"from":         nodeCollectionName,
			"localField":   "tree.right",
			"foreignField": "_id",
			"as":           "rightNode",
		}},
		{"$unwind": bson.M{"path": "$rightNode", "preserveNullAndEmptyArrays": false}},
		{"$match": bson.M{
			"leftNode.type":    models.Number,
			"rightNode.type":   models.Number,
			"leftNode.number":  bson.M{"$exists": true},
			"rightNode.number": bson.M{"$exists": true},
		}},
		{"$project": bson.M{
			"_id":              1,
			"tree.operator":    1,
			"leftNode.number":  1,
			"rightNode.number": 1,
		}},
	}
	cursor, err := r.nodeCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, save.New(err)
	}
	defer cursor.Close(ctx)
	var results []models.AggregatedNode
	if err := cursor.All(ctx, &results); err != nil {
		return nil, save.New(err)
	}

	tasks := make([]pb.Task, len(results))
	ids := make(bson.A, len(results))
	for i, result := range results {
		ids[i] = result.ID
		tasks[i] = pb.Task{
			Id:        result.ID.Hex(),
			Arg1:      result.LeftNode.Number,
			Arg2:      result.RightNode.Number,
			Operation: pb.Operation(result.Tree.Operator),
		}
	}
	if _, err := r.nodeCollection.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": ids}}, bson.M{"$set": bson.M{"sended_at": time.Now()}}); err != nil {
		return nil, save.New(err)
	}

	return tasks, nil
}

func (r *Repo) DoCallback() {
	r.mu.Lock()
	defer r.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()
	tasks, err := r.GetFitNodes(ctx)
	if err != nil {
		r.callback.SendError(ctx, err)
		return
	}
	r.callback.SendResult(ctx, tasks)
}

var _ repo.ExpressionRepo = (*Repo)(nil)
