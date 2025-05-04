package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/Calculator/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SendError(ctx *gin.Context, err error) {
	code := service.GetCode(err)
	text := err.Error()
	if code == http.StatusInternalServerError {
		text = InternalError
	}
	ctx.AbortWithStatusJSON(code, models.ErrorResponse{Error: text})
	return
}

func (h *Handler) CalcHandler(ctx *gin.Context) {
	req := new(models.CalculationRequest)
	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil || req.Expression == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: InvalidBody})
		return
	}
	userId, ok := ctx.Get(UserIDKey)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: Unauthorized})
		return
	}

	id, err := h.Service.Add(ctx.Request.Context(), req.Expression, userId.(primitive.ObjectID))
	if err != nil {
		SendError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, models.CreatedResponse{Id: id})
}

func (h *Handler) ExpressionsHandler(ctx *gin.Context) {
	userId, ok := ctx.Get(UserIDKey)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: Unauthorized})
		return
	}
	expressions, err := h.Service.GetByUSer(ctx.Request.Context(), userId.(primitive.ObjectID))
	if err != nil {
		SendError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, models.ExpressionsResponse{Expressions: expressions})
}

func (h *Handler) GetByIdHandler(ctx *gin.Context) {
	id, err := primitive.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{Error: InvalidId})
		return
	}
	userId, ok := ctx.Get(UserIDKey)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: Unauthorized})
		return
	}
	expr, err := h.Service.Get(ctx.Request.Context(), id)
	if err != nil {
		SendError(ctx, err)
		return
	}

	if expr.UserID != userId.(primitive.ObjectID) {
		ctx.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{Error: Forbidden})
		return
	}
	ctx.JSON(http.StatusOK, models.ExpressionResponse{Expression: *expr})
}

func (h *Handler) RegisterHandler(ctx *gin.Context) {
	req := new(models.UserRequest)
	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil || req.Username == "" || req.Password == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: InvalidBody})
		return
	}
	id, err := h.Service.Register(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		SendError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, models.CreatedResponse{Id: id})
}

func (h *Handler) LoginHandler(ctx *gin.Context) {
	req := new(models.UserRequest)
	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: InvalidBody})
		return
	}
	token, err := h.Service.Login(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		SendError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, models.TokenResponse{Token: token})
}

func (h *Handler) ChangeUsernameHandler(ctx *gin.Context) {
	req := new(models.UsernameRequest)
	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil || req.Username == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: InvalidBody})
		return
	}
	userId, ok := ctx.Get(UserIDKey)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: Unauthorized})
		return
	}
	err = h.Service.UpdateUsername(ctx.Request.Context(), userId.(primitive.ObjectID), req.Username)
	if err != nil {
		SendError(ctx, err)
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}

func (h *Handler) ChangePasswordHandler(ctx *gin.Context) {
	req := new(models.PasswordRequest)
	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil || req.Password == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: InvalidBody})
		return
	}
	userId, ok := ctx.Get(UserIDKey)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: Unauthorized})
		return
	}
	err = h.Service.UpdatePassword(ctx.Request.Context(), userId.(primitive.ObjectID), req.Password)
	if err != nil {
		SendError(ctx, err)
		return
	}
	ctx.JSON(http.StatusNoContent, nil)

}

func (h *Handler) DeleteHandler(ctx *gin.Context) {
	userId, ok := ctx.Get(UserIDKey)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: Unauthorized})
		return
	}
	err := h.Service.Delete(ctx.Request.Context(), userId.(primitive.ObjectID))
	if err != nil {
		SendError(ctx, err)
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}

func (h *Handler) PingHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusNoContent, nil)
}
