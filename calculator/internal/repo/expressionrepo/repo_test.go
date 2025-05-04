package expressionrepo_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/repo"
	"github.com/vandi37/Calculator/internal/repo/expressionrepo"
	"github.com/vandi37/Calculator/internal/status"
	"github.com/vandi37/Calculator/pkg/parsing/parser"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExpressionRepoTestSuite struct {
	suite.Suite
	mongoC         testcontainers.Container
	client         *mongo.Client
	expressionRepo *expressionrepo.Repo
	ctx            context.Context
	userId         primitive.ObjectID
	mockCallback   *MockCallback
}

func (suite *ExpressionRepoTestSuite) Clear() {
	_, err := suite.expressionRepo.GetCollection().DeleteMany(suite.ctx, bson.M{})
	require.NoError(suite.T(), err)
	_, err = suite.expressionRepo.GetNodeCollection().DeleteMany(suite.ctx, bson.M{})
	require.NoError(suite.T(), err)
}

func (suite *ExpressionRepoTestSuite) SetupSuite() {
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
	suite.expressionRepo = expressionrepo.New(db, 5*time.Minute)
	suite.userId = primitive.NewObjectID()
	suite.mockCallback = &MockCallback{}
	suite.expressionRepo.SetCallback(suite.ctx, suite.mockCallback)
}

func (suite *ExpressionRepoTestSuite) TearDownSuite() {
	_, err := suite.expressionRepo.GetCollection().DeleteMany(suite.ctx, bson.M{})
	require.NoError(suite.T(), err)

	_, err = suite.expressionRepo.GetNodeCollection().DeleteMany(suite.ctx, bson.M{})
	require.NoError(suite.T(), err)

	err = suite.client.Disconnect(suite.ctx)
	require.NoError(suite.T(), err)

	err = suite.mongoC.Terminate(suite.ctx)
	require.NoError(suite.T(), err)
}

func TestExpressionRepoTestSuite(t *testing.T) {
	suite.Run(t, new(ExpressionRepoTestSuite))
}

type MockCallback struct {
	lastTasks []pb.Task
	lastError error
}

func (m *MockCallback) SendResult(ctx context.Context, tasks []pb.Task) {
	m.lastTasks = tasks
	m.lastError = nil
}

func (m *MockCallback) SendError(ctx context.Context, err error) {
	m.lastTasks = nil
	m.lastError = err
}

func (suite *ExpressionRepoTestSuite) TestCreate() {
	suite.Clear()
	t := suite.T()
	ctx := context.Background()

	expressions := []struct {
		name, expression string
	}{
		{"Addition", "1 + 1"},
		{"Subtraction", "5 - 3"},
		{"Multiplication", "2 * 4"},
		{"Division", "10 / 2"},
		{"Addition and Subtraction", "1 + 2 - 3"},
		{"Multiplication and Division", "6 / 3 * 2"},
		{"Mixed Operations", "2 + 3 * 4"},
		{"Parentheses - Addition First", "(2 + 3) * 4"},
		{"Nested Parentheses", "((1 + 2) * 3) - 4"},
		{"Division and Subtraction with Parentheses", "10 - (20 / 5)"},
		{"Multiple Operations and Parentheses", "(1 + 2) * (3 - 1)"},
		{"More Complex Mixed Operations", "5 + 2 * (8 - 4) / 2"},
		{"Division and Multiplication with Parentheses", "(12 / 3) * (5 - 2)"},
		{"Chained Operations with Parentheses", "((1 + 1) * 2) / (4 - 2)"},
		{"Complex Nested Parentheses and Operations", "(10 / (2 + 3)) * (5 - (1 * 2))"},
		{"Harder Mixed Precedence and Parentheses", "10 + (5 * (4 - 2)) / (1 + 1)"},
		{"Even Harder with Multiple Nested Levels", "((1 + (2 * 3)) - (4 / 2)) * (5 - 1)"},
	}

	tests := []struct {
		name        string
		expression  models.Expression
		ast         tree.ExpressionType
		wantErr     bool
		expectedErr error
	}{
		{
			name: "invalid expression - nil AST",
			expression: models.Expression{
				UserID: suite.userId,
				Origin: "invalid",
				Status: status.Pending,
			},
			ast:         nil,
			wantErr:     true,
			expectedErr: repo.InvalidExpression,
		},
	}

	for _, data := range expressions {
		ast, err := parser.Build(data.expression)
		require.NoError(t, err)
		tests = append(tests, struct {
			name        string
			expression  models.Expression
			ast         tree.ExpressionType
			wantErr     bool
			expectedErr error
		}{
			name:       data.name,
			expression: models.Expression{UserID: suite.userId, Origin: data.expression, Status: status.Pending},
			ast:        ast.Expression,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			id, err := suite.expressionRepo.Create(ctx, tt.expression, tree.Ast{Expression: tt.ast})

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Equal(t, primitive.NilObjectID, id)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, primitive.NilObjectID, id)

				var expr models.Expression
				err := suite.expressionRepo.GetCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&expr)
				require.NoError(t, err)
				assert.Equal(t, tt.expression.Origin, expr.Origin)
				assert.Equal(t, tt.expression.UserID, expr.UserID)
				assert.Equal(t, tt.expression.Status, expr.Status)
				assert.NotEqual(t, primitive.NilObjectID, expr.NodeID)
				assert.False(t, expr.CreatedAt.IsZero())

				if tt.ast != nil {
					_, err := suite.expressionRepo.GetNode(ctx, expr.NodeID)
					require.NoError(t, err)
				}
			}
		})
	}
}

func (suite *ExpressionRepoTestSuite) TestGet() {
	suite.Clear()
	t := suite.T()
	ctx := context.Background()
	finExpr := models.Expression{UserID: suite.userId, Origin: "1", Status: status.Pending}
	finishedId, err := suite.expressionRepo.Create(ctx, finExpr, tree.Ast{Expression: tree.Num(1)})
	require.NoError(t, err)
	doingExpr := models.Expression{UserID: suite.userId, Origin: "2+2*2", Status: status.Pending}
	id, err := suite.expressionRepo.Create(ctx, doingExpr, tree.Ast{Expression: tree.Expression{Left: tree.Num(2), Operation: tree.Operation(pb.Operation_ADD), Right: tree.Expression{Left: tree.Num(2), Operation: tree.Operation(pb.Operation_MULTIPLY), Right: tree.Num(2)}}})
	require.NoError(t, err)
	tests := []struct {
		name        string
		id          primitive.ObjectID
		wantErr     bool
		expectedErr error
		checkResult func(*models.Expression)
	}{
		{
			name:    "existing expression with nodes",
			id:      id,
			wantErr: false,
			checkResult: func(expr *models.Expression) {
				assert.Equal(t, doingExpr.Origin, expr.Origin)
				assert.Equal(t, doingExpr.UserID, expr.UserID)
				assert.Equal(t, status.Pending, expr.Status)
				assert.NotEqual(t, primitive.NilObjectID, expr.NodeID)
			},
		},
		{
			name:        "non-existent expression",
			id:          primitive.NewObjectID(),
			wantErr:     true,
			expectedErr: repo.ExpressionNotFound,
		},
		{
			name:    "finished expression with result",
			id:      finishedId,
			wantErr: false,
			checkResult: func(expr *models.Expression) {

				result, err := suite.expressionRepo.Get(ctx, expr.ID)
				require.NoError(t, err)
				assert.Equal(t, 1.0, *result.Result)
				assert.Equal(t, status.Finished, result.Status)
				assert.Equal(t, primitive.NilObjectID, result.NodeID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := suite.expressionRepo.Get(ctx, tt.id)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, expr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, expr)
				if tt.checkResult != nil {
					tt.checkResult(expr)
				}
			}
		})
	}
}

func (suite *ExpressionRepoTestSuite) TestGetByUser() {
	suite.Clear()
	t := suite.T()
	ctx := context.Background()
	for range 10 {
		createHugeRandomTree(ctx, suite.userId, suite.expressionRepo, 4)
	}
	tests := []struct {
		name        string
		userID      primitive.ObjectID
		wantErr     bool
		expectedLen int
	}{
		{
			name:        "user with expressions",
			userID:      suite.userId,
			wantErr:     false,
			expectedLen: 10,
		},
		{
			name:        "user without expressions",
			userID:      primitive.NewObjectID(),
			wantErr:     false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expressions, err := suite.expressionRepo.GetByUser(ctx, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, expressions)
			} else {
				require.NoError(t, err)
				assert.Len(t, expressions, tt.expectedLen)
				for _, expr := range expressions {
					assert.Equal(t, tt.userID, expr.UserID)
				}
			}
		})
	}
}

func (suite *ExpressionRepoTestSuite) TestDelete() {
	suite.Clear()
	t := suite.T()
	ctx := context.Background()

	_, _, expr, _, err := createHugeRandomTree(ctx, suite.userId, suite.expressionRepo, 64)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          primitive.ObjectID
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "existing expression",
			id:      expr.ID,
			wantErr: false,
		},
		{
			name:        "non-existent expression",
			id:          primitive.NewObjectID(),
			wantErr:     true,
			expectedErr: repo.ExpressionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := suite.expressionRepo.Delete(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (suite *ExpressionRepoTestSuite) TestDeleteByUser() {
	suite.Clear()
	t := suite.T()
	ctx := context.Background()

	for range 10 {
		createHugeRandomTree(ctx, suite.userId, suite.expressionRepo, 4)
	}

	tests := []struct {
		name        string
		userID      primitive.ObjectID
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "user with expressions",
			userID:  suite.userId,
			wantErr: false,
		},
		{
			name:    "user without expressions",
			userID:  primitive.NewObjectID(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := suite.expressionRepo.DeleteByUser(ctx, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)

				count, err := suite.expressionRepo.GetCollection().CountDocuments(ctx, bson.M{"user_id": tt.userID})
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			}
		})
	}
}

func (suite *ExpressionRepoTestSuite) TestSetToNum() {
	suite.Clear()
	t := suite.T()
	ctx := context.Background()

	tempNode := models.Node{
		ID:   primitive.NewObjectID(),
		Type: models.Operation,
		Tree: &models.TreeNode{
			Operator: pb.Operation_ADD,
			Left:     primitive.NewObjectID(),
			Right:    primitive.NewObjectID(),
		},
	}
	_, err := suite.expressionRepo.GetNodeCollection().InsertOne(ctx, tempNode)
	require.NoError(t, err)

	tests := []struct {
		name        string
		nodeID      primitive.ObjectID
		result      float64
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "successful update",
			nodeID:  tempNode.ID,
			result:  42.0,
			wantErr: false,
		},
		{
			name:        "non-existent node",
			nodeID:      primitive.NewObjectID(),
			result:      42.0,
			wantErr:     true,
			expectedErr: repo.NodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := suite.expressionRepo.SetToNum(ctx, tt.nodeID, tt.result)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)

				var node models.Node
				err := suite.expressionRepo.GetNodeCollection().FindOne(ctx, bson.M{"_id": tt.nodeID}).Decode(&node)
				require.NoError(t, err)
				assert.Equal(t, models.Number, node.Type)
				assert.Equal(t, tt.result, *node.Number)
				assert.Nil(t, node.Tree)

				assert.NotNil(t, suite.mockCallback.lastTasks)
			}
		})
	}
}

func createHugeRandomTree(ctx context.Context, userId primitive.ObjectID, expressionRepo *expressionrepo.Repo, max int) (countFit int, nodeNum int, rootExpr *models.Expression, furthestNodeID primitive.ObjectID, expr error) {
	expressionCollection := expressionRepo.GetCollection()
	nodeCollection := expressionRepo.GetNodeCollection()

	rootNodeID := primitive.NewObjectID()
	expressionID := primitive.NewObjectID()

	rootExpr = &models.Expression{
		ID:     expressionID,
		UserID: userId,
		Origin: "randomly generated tree",
		NodeID: rootNodeID,
		Status: status.Pending,
	}

	_, err := expressionCollection.InsertOne(ctx, rootExpr)
	if err != nil {
		return 0, 0, nil, primitive.NilObjectID, err
	}

	var buildTree func(nodeID primitive.ObjectID, remainingNodes int) error

	buildTree = func(nodeID primitive.ObjectID, remainingNodes int) error {
		nodeNum++
		if remainingNodes <= 0 {
			numberNode := models.Node{
				ID:     nodeID,
				Type:   models.Number,
				Number: func() *float64 { v := rand.Float64() * 100; return &v }(),
			}
			_, err := nodeCollection.InsertOne(ctx, numberNode)
			if err != nil {
				return err
			}
			furthestNodeID = nodeID
			return nil
		}

		leftNodeID := primitive.NewObjectID()
		rightNodeID := primitive.NewObjectID()
		operators := []pb.Operation{pb.Operation_ADD, pb.Operation_SUBTRACT, pb.Operation_MULTIPLY, pb.Operation_DIVIDE}
		operator := operators[rand.Intn(len(operators))]
		opNode := models.Node{
			ID:   nodeID,
			Type: models.Operation,
			Tree: &models.TreeNode{
				Operator: operator,
				Left:     leftNodeID,
				Right:    rightNodeID,
			},
		}
		_, err := nodeCollection.InsertOne(ctx, opNode)
		if err != nil {
			return err
		}

		if remainingNodes%2 != 0 {
			if err := buildTree(leftNodeID, 0); err != nil {
				return err
			}
			if err := buildTree(rightNodeID, remainingNodes-1); err != nil {
				return err
			}
			if remainingNodes == 1 {
				countFit++
			}

			return nil
		}

		subtreeSize := remainingNodes / 2
		if err := buildTree(leftNodeID, subtreeSize); err != nil {
			return err
		}
		if err := buildTree(rightNodeID, subtreeSize); err != nil {
			return err
		}

		return nil
	}
	if err := buildTree(rootNodeID, rand.Intn(max/2)); err != nil {
		return 0, 0, nil, primitive.NilObjectID, err
	}
	return
}

func (suite *ExpressionRepoTestSuite) TestSetToError() {
	suite.Clear()
	t := suite.T()
	ctx := context.Background()

	tests := []struct {
		name        string
		id          primitive.ObjectID
		expr        *models.Expression
		nodeNum     int
		errVal      string
		wantErr     bool
		expectedErr error
	}{

		{
			name:        "non-existent node",
			id:          primitive.NewObjectID(),
			errVal:      "error",
			wantErr:     true,
			expectedErr: repo.ExpressionNotFound,
		},
	}

	for i := range 10 {
		_, nodes, expr, nodeId, err := createHugeRandomTree(ctx, suite.userId, suite.expressionRepo, 64)
		require.NoError(t, err)
		tests = append(tests, struct {
			name        string
			id          primitive.ObjectID
			expr        *models.Expression
			nodeNum     int
			errVal      string
			wantErr     bool
			expectedErr error
		}{
			name:   fmt.Sprintf("successful error %d(%d)", i, nodes),
			id:     nodeId,
			expr:   expr,
			errVal: fmt.Sprintf("some error %d", i),
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := suite.expressionRepo.SetToError(ctx, tt.id, tt.errVal)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				var expr models.Expression
				err = suite.expressionRepo.GetCollection().FindOne(ctx, bson.M{"_id": tt.expr.ID}).Decode(&expr)
				require.NoError(t, err)
				assert.Equal(t, status.Error, expr.Status)
				assert.Equal(t, tt.errVal, expr.Error)
				assert.Nil(t, expr.Result)
				assert.Equal(t, tt.expr.Origin, expr.Origin)
				assert.Equal(t, primitive.NilObjectID, expr.NodeID)
			}
		})
	}
}

func (suite *ExpressionRepoTestSuite) TestGetFitNodes() {
	suite.Clear()
	t := suite.T()
	ctx := context.Background()

	var fit int
	for range 10 {
		fitLoc, _, _, _, err := createHugeRandomTree(ctx, suite.userId, suite.expressionRepo, 64)
		require.NoError(t, err)
		fit += fitLoc
	}

	tests := []struct {
		name         string
		expectedLen  int
		expectedTask *pb.Task
	}{
		{
			name:        "get fit nodes without duration check",
			expectedLen: fit,
		},
		// More tests...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := suite.expressionRepo.GetFitNodes(ctx)
			require.NoError(t, err)
			require.Len(t, tasks, tt.expectedLen)
		})
	}
}

func (suite *ExpressionRepoTestSuite) TestDoCallback() {
	suite.Clear()
	t := suite.T()

	suite.mockCallback.lastTasks = nil
	suite.mockCallback.lastError = nil

	suite.expressionRepo.DoCallback()
	assert.NotNil(t, suite.mockCallback.lastTasks)
	assert.Nil(t, suite.mockCallback.lastError)
}
