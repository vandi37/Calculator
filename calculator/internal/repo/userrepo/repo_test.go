package userrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/repo"
	"github.com/vandi37/Calculator/internal/repo/userrepo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepoTestSuite struct {
	suite.Suite
	mongoC      testcontainers.Container
	client      *mongo.Client
	userRepo    *userrepo.Repo
	testUserID  primitive.ObjectID
	testUserID2 primitive.ObjectID
	ctx         context.Context
}

func (suite *UserRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(20 * time.Second),
	}

	mongoC, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.mongoC = mongoC

	endpoint, err := mongoC.Endpoint(suite.ctx, "")
	require.NoError(suite.T(), err)

	client, err := mongo.Connect(suite.ctx, options.Client().ApplyURI("mongodb://"+endpoint))
	require.NoError(suite.T(), err)
	suite.client = client

	db := client.Database("test_db")
	suite.userRepo = userrepo.New(db)

	suite.testUserID = primitive.NewObjectID()
	suite.testUserID2 = primitive.NewObjectID()

	testUsers := []models.User{
		{
			ID:        suite.testUserID,
			Username:  "testuser1",
			Password:  "password1",
			CreatedAt: time.Now(),
		},
		{
			ID:        suite.testUserID2,
			Username:  "testuser2",
			Password:  "password2",
			CreatedAt: time.Now(),
		},
	}

	var docs []interface{}
	for _, u := range testUsers {
		docs = append(docs, u)
	}

	_, err = suite.userRepo.GetCollection().InsertMany(suite.ctx, docs)
	require.NoError(suite.T(), err)
}

func (suite *UserRepoTestSuite) TearDownSuite() {
	_, err := suite.userRepo.GetCollection().DeleteMany(suite.ctx, bson.M{})
	require.NoError(suite.T(), err)

	err = suite.client.Disconnect(suite.ctx)
	require.NoError(suite.T(), err)

	err = suite.mongoC.Terminate(suite.ctx)
	require.NoError(suite.T(), err)
}

func TestUserRepoTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepoTestSuite))
}

func (suite *UserRepoTestSuite) TestRegister() {
	t := suite.T()
	ctx := context.Background()

	tests := []struct {
		name        string
		user        models.User
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful registration",
			user: models.User{
				Username: "newuser",
				Password: "newpassword",
			},
			wantErr: false,
		},
		{
			name: "duplicate username",
			user: models.User{
				Username: "testuser1",
				Password: "password",
			},
			wantErr:     true,
			expectedErr: repo.UsernameTaken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := suite.userRepo.Register(ctx, tt.user)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Equal(t, primitive.NilObjectID, id)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, primitive.NilObjectID, id)

				var user models.User
				err := suite.userRepo.GetCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&user)
				require.NoError(t, err)
				assert.Equal(t, tt.user.Username, user.Username)
				assert.Equal(t, tt.user.Password, user.Password)
				assert.False(t, user.CreatedAt.IsZero())
			}
		})
	}
}

func (suite *UserRepoTestSuite) TestGet() {
	t := suite.T()
	ctx := context.Background()

	tests := []struct {
		name        string
		id          primitive.ObjectID
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "existing user",
			id:      suite.testUserID,
			wantErr: false,
		},
		{
			name:        "non-existent user",
			id:          primitive.NewObjectID(),
			wantErr:     true,
			expectedErr: repo.UserNotFound,
		},
		{
			name:        "nil object id",
			id:          primitive.NilObjectID,
			wantErr:     true,
			expectedErr: repo.UserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := suite.userRepo.Get(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.id, user.ID)
				assert.Equal(t, "testuser1", user.Username)
				assert.Equal(t, "password1", user.Password)
				assert.False(t, user.CreatedAt.IsZero())
			}
		})
	}
}

func (suite *UserRepoTestSuite) TestGetByUsername() {
	t := suite.T()
	ctx := context.Background()

	tests := []struct {
		name        string
		username    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "existing user",
			username: "testuser1",
			wantErr:  false,
		},
		{
			name:        "non-existent user",
			username:    "nonexistent",
			wantErr:     true,
			expectedErr: repo.UserNotFound,
		},
		{
			name:        "empty username",
			username:    "",
			wantErr:     true,
			expectedErr: repo.UserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := suite.userRepo.GetByUsername(ctx, tt.username)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
				if tt.username == "testuser1" {
					assert.Equal(t, suite.testUserID, user.ID)
					assert.Equal(t, "password1", user.Password)
				}
				assert.False(t, user.CreatedAt.IsZero())
			}
		})
	}
}

func (suite *UserRepoTestSuite) TestDelete() {
	t := suite.T()
	ctx := context.Background()

	tempUser := models.User{
		ID:        primitive.NewObjectID(),
		Username:  "tempuser",
		Password:  "temppass",
		CreatedAt: time.Now(),
	}
	_, err := suite.userRepo.GetCollection().InsertOne(ctx, tempUser)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          primitive.ObjectID
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "existing user",
			id:      tempUser.ID,
			wantErr: false,
		},
		{
			name:        "non-existent user",
			id:          primitive.NewObjectID(),
			wantErr:     true,
			expectedErr: repo.UserNotFound,
		},
		{
			name:        "nil object id",
			id:          primitive.NilObjectID,
			wantErr:     true,
			expectedErr: repo.UserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := suite.userRepo.Delete(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)

				count, err := suite.userRepo.GetCollection().CountDocuments(ctx, bson.M{"_id": tt.id})
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			}
		})
	}
}

func (suite *UserRepoTestSuite) TestUpdateUsername() {
	t := suite.T()
	ctx := context.Background()

	tempUser := models.User{
		ID:        primitive.NewObjectID(),
		Username:  "updateuser",
		Password:  "updatepass",
		CreatedAt: time.Now(),
	}
	_, err := suite.userRepo.GetCollection().InsertOne(ctx, tempUser)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          primitive.ObjectID
		newUsername string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "successful update",
			id:          tempUser.ID,
			newUsername: "updatedusername",
			wantErr:     false,
		},
		{
			name:        "duplicate username",
			id:          tempUser.ID,
			newUsername: "testuser1",
			wantErr:     true,
			expectedErr: repo.UsernameTaken,
		},
		{
			name:        "non-existent user",
			id:          primitive.NewObjectID(),
			newUsername: "newusername",
			wantErr:     true,
			expectedErr: repo.UserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := suite.userRepo.UpdateUsername(ctx, tt.id, tt.newUsername)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)

				var user models.User
				err := suite.userRepo.GetCollection().FindOne(ctx, bson.M{"_id": tt.id}).Decode(&user)
				require.NoError(t, err)
				assert.Equal(t, tt.newUsername, user.Username)
			}
		})
	}
}

func (suite *UserRepoTestSuite) TestUpdatePassword() {
	t := suite.T()
	ctx := context.Background()

	tempUser := models.User{
		ID:        primitive.NewObjectID(),
		Username:  "passuser",
		Password:  "oldpassword",
		CreatedAt: time.Now(),
	}
	_, err := suite.userRepo.GetCollection().InsertOne(ctx, tempUser)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          primitive.ObjectID
		newPassword string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "successful update",
			id:          tempUser.ID,
			newPassword: "newpassword",
			wantErr:     false,
		},
		{
			name:        "non-existent user",
			id:          primitive.NewObjectID(),
			newPassword: "newpassword",
			wantErr:     true,
			expectedErr: repo.UserNotFound,
		},
		{
			name:        "nil object id",
			id:          primitive.NilObjectID,
			newPassword: "newpassword",
			wantErr:     true,
			expectedErr: repo.UserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := suite.userRepo.UpdatePassword(ctx, tt.id, tt.newPassword)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)

				var user models.User
				err := suite.userRepo.GetCollection().FindOne(ctx, bson.M{"_id": tt.id}).Decode(&user)
				require.NoError(t, err)
				assert.Equal(t, tt.newPassword, user.Password)
			}
		})
	}
}

func (suite *UserRepoTestSuite) TestUsernameExists() {
	t := suite.T()
	ctx := context.Background()

	tests := []struct {
		name     string
		username string
		expected bool
		wantErr  bool
	}{
		{
			name:     "existing username",
			username: "testuser1",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "non-existing username",
			username: "nonexistent",
			expected: false,
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			expected: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := suite.userRepo.GetCollection().CountDocuments(ctx, bson.M{"username": tt.username})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, count > 0)
			}
		})
	}
}
