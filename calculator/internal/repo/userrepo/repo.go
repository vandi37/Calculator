package userrepo

import (
	"context"
	"time"

	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/repo"
	"github.com/vandi37/ferror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	collectionName = "users"
)

type Repo struct {
	collection *mongo.Collection
}

// GetCollection implements repo.UserRepo.
func (r *Repo) GetCollection() *mongo.Collection {
	return r.collection
}

func (r *Repo) usernameExists(ctx context.Context, username string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// UpdatePassword implements repo.UserRepo.
func (r *Repo) UpdatePassword(ctx context.Context, id primitive.ObjectID, password string) error {
	save := ferror.Save("userrepo.Repo.UpdatePassword")
	if res, err := r.collection.UpdateByID(
		ctx,
		id,
		bson.M{"$set": bson.M{"password": password}},
	); err != nil {
		return save.New(err)
	} else if  res.MatchedCount == 0 || res.ModifiedCount == 0 {
		return repo.UserNotFound
	}
	return nil
}

// UpdateUsername implements repo.UserRepo.
func (r *Repo) UpdateUsername(ctx context.Context, id primitive.ObjectID, username string) error {
	save := ferror.Save("userrepo.Repo.UpdateUsername")
	if exists, err := r.usernameExists(ctx, username); err != nil {
		return save.New(err)
	} else if exists {
		return repo.UsernameTaken
	}
	if res, err := r.collection.UpdateByID(
		ctx,
		id,
		bson.M{"$set": bson.M{"username": username}},
	); err != nil {
		return save.New(err)
	} else if res.MatchedCount == 0 || res.ModifiedCount == 0 {
		return repo.UserNotFound
	}
	return nil
}

// Delete implements repo.UserRepo.
func (r *Repo) Delete(ctx context.Context, id primitive.ObjectID) error {
	var save = ferror.Save("userrepo.Repo.Delete")
	if res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return save.New(err)
	} else if res.DeletedCount == 0 {
		return repo.UserNotFound
	}
	return nil
}

// Get implements repo.UserRepo.
func (r *Repo) Get(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var save = ferror.Save("userrepo.Repo.Get")
	var user models.User
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user); err == mongo.ErrNoDocuments {
		return nil, repo.UserNotFound
	} else if err != nil {
		return nil, save.New(err)
	}
	return &user, nil
}

// GetByUsername implements repo.UserRepo.
func (r *Repo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var save = ferror.Save("userrepo.Repo.GetByUsername")
	var user models.User
	if err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user); err == mongo.ErrNoDocuments {
		return nil, repo.UserNotFound
	} else if err != nil {
		return nil, save.New(err)
	}
	return &user, nil
}

// Register implements repo.UserRepo.
func (r *Repo) Register(ctx context.Context, user models.User) (primitive.ObjectID, error) {
	var save = ferror.Save("userrepo.Repo.Register")
	if exists, err := r.usernameExists(ctx, user.Username); err != nil {
		return primitive.NilObjectID, save.New(err)
	} else if exists {
		return primitive.NilObjectID, repo.UsernameTaken
	}
	user.CreatedAt = time.Now()
	if res, err := r.collection.InsertOne(ctx, user); err != nil {
		return primitive.NilObjectID, save.New(err)
	} else {
		return res.InsertedID.(primitive.ObjectID), nil
	}
}

func New(db repo.IntoCollection) *Repo {
	return &Repo{
		collection: db.Collection(collectionName),
	}
}

var _ repo.UserRepo = (*Repo)(nil)
