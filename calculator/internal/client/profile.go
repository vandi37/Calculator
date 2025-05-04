package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vandi37/Calculator/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Profile struct {
	token string
	*Client
}

func (p *Profile) Token() string {
	return p.token
}

func (p *Profile) AddAuth(req *http.Request) {
	req.Header.Add("Authorization", p.token)
}

func (p *Profile) Calculate(ctx context.Context, expression string) (primitive.ObjectID, error) {
	b, err := json.Marshal(models.CalculationRequest{
		Expression: expression,
	})
	if err != nil {
		return primitive.NilObjectID, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Address+"/calculate", strings.NewReader(string(b)))
	if err != nil {
		return primitive.NilObjectID, err
	}
	p.AddAuth(req)

	resp, err := p.Request(req)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if resp.StatusCode != http.StatusCreated {
		return primitive.NilObjectID, p.ReadError(resp)
	}
	id := new(models.CreatedResponse)
	if err := p.Read(resp, id); err != nil {
		return primitive.NilObjectID, err
	}
	return id.Id, nil
}

func (p *Profile) GetExpressions(ctx context.Context) ([]models.Expression, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.Address+"/expressions", nil)
	if err != nil {
		return nil, err
	}
	p.AddAuth(req)

	resp, err := p.Request(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, p.ReadError(resp)
	}
	expressions := new(models.ExpressionsResponse)
	if err := p.Read(resp, expressions); err != nil {
		return nil, err
	}
	return expressions.Expressions, nil
}

func (p *Profile) GetExpression(ctx context.Context, id primitive.ObjectID) (*models.Expression, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.Address+"/expressions/"+id.Hex(), nil)
	if err != nil {
		return nil, err
	}
	p.AddAuth(req)

	resp, err := p.Request(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, p.ReadError(resp)
	}
	expression := new(models.ExpressionResponse)
	if err := p.Read(resp, expression); err != nil {
		return nil, err
	}
	return &expression.Expression, nil
}

func (p *Profile) ChangeUsername(ctx context.Context, username string) error {
	b, err := json.Marshal(models.UsernameRequest{
		Username: username,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, p.Address+"/username", strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	p.AddAuth(req)

	resp, err := p.Request(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return p.ReadError(resp)
	}
	return nil
}

func (p *Profile) ChangePassword(ctx context.Context, password string) error {
	b, err := json.Marshal(models.PasswordRequest{
		Password: password,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, p.Address+"/password", strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	p.AddAuth(req)

	resp, err := p.Request(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return p.ReadError(resp)
	}
	return nil
}
func (p *Profile) Delete(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, p.Address+"/delete", nil)
	if err != nil {
		return err
	}
	p.AddAuth(req)

	resp, err := p.Request(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return p.ReadError(resp)
	}
	return nil
}
