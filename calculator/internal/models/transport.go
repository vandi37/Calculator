package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type CalculationRequest struct {
	Expression string `json:"expression"`
}
type CreatedResponse struct {
	Id primitive.ObjectID `json:"id"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type ExpressionsResponse struct {
	Expressions []Expression `json:"expressions"`
}

type ExpressionResponse struct {
	Expression Expression `json:"expression"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UsernameRequest struct {
	Username string `json:"username"`
}

type PasswordRequest struct {
	Password string `json:"password"`
}
