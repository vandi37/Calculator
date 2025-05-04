//go:build integration

package integration

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vandi37/Calculator/internal/client"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/status"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test(ctx context.Context, t testing.TB, address string) {
	client := client.New(address)
	err := client.Ping(ctx)
	require.NoError(t, err)
	pavelName := "pavel_" + primitive.NewObjectID().Hex()
	pavelPassword := "12345"
	pavel, err := client.Register(ctx, pavelName, pavelPassword)
	require.NoError(t, err)
	defer func() {
		err := pavel.Delete(ctx)
		require.NoError(t, err)
	}()

	ritaName := "rita_" + primitive.NewObjectID().Hex()
	ritaPassword := "12345"
	rita, err := client.Register(ctx, ritaName, ritaPassword)
	require.NoError(t, err)
	defer func() {
		err := rita.Delete(ctx)
		require.NoError(t, err)
	}()

	time.Sleep(time.Second) // Waiting till the token is valid

	expressionId, err := pavel.Calculate(ctx, "2+2")
	require.NoError(t, err)
	time.Sleep(time.Second) // Expected to be enough to calculate
	expression, err := pavel.GetExpression(ctx, expressionId)
	require.NoError(t, err)
	require.NotNil(t, expression)

	assert.Equal(t, "2+2", expression.Origin)
	assert.Equal(t, status.Finished, expression.Status)
	require.NotNil(t, expression.Result)
	assert.Equal(t, 4.0, *expression.Result)

	nilExpression, err := rita.GetExpression(ctx, expressionId)
	assert.Nil(t, nilExpression)
	assert.ErrorContains(t, err, "forbidden")

	err = rita.ChangeUsername(ctx, "rita_"+primitive.NewObjectID().Hex())
	require.NoError(t, err)

	err = pavel.ChangePassword(ctx, "qwerty")
	require.NoError(t, err)

	otherExpressionId, err := pavel.Calculate(ctx, "(70/7)*10/((3+2)*(3+7))+10")
	require.NoError(t, err)
	time.Sleep(time.Second * 3) // Waiting a bit longer

	otherExpression, err := pavel.GetExpression(ctx, otherExpressionId)
	require.NoError(t, err)
	require.NotNil(t, otherExpression)

	assert.Equal(t, "(70/7)*10/((3+2)*(3+7))+10", otherExpression.Origin)
	assert.Equal(t, status.Finished, otherExpression.Status)
	require.NotNil(t, otherExpression.Result)
	assert.Equal(t, 12.0, *otherExpression.Result)

	expressions, err := pavel.GetExpressions(ctx)
	require.NoError(t, err)

	require.Len(t, expressions, 2)

	slices.SortFunc(expressions, func(a, b models.Expression) int { return a.CreatedAt.Compare(b.CreatedAt) })
	otherExpressions := []models.Expression{*otherExpression, *expression}
	slices.SortFunc(otherExpressions, func(a, b models.Expression) int { return a.CreatedAt.Compare(b.CreatedAt) })
	assert.Equal(t, otherExpressions, expressions)

	ritaExpressions, err := rita.GetExpressions(ctx)
	require.NoError(t, err)
	assert.Len(t, ritaExpressions, 0)
}
