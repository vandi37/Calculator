package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vandi37/Calculator/internal/agent/module"
	"github.com/vandi37/Calculator/internal/transport/input"
	"github.com/vandi37/Calculator/internal/transport/resp"
	"github.com/vandi37/Calculator/internal/wait"
	"github.com/vandi37/Calculator/pkg/calc"
	"github.com/vandi37/vanerrors"
)

func (h *Handler) CalcHandler(ctx *gin.Context) {
	req := new(input.CalcRequest)
	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil || req.Expression == "" {
		ctx.JSON(http.StatusBadRequest, resp.ResponseError{Error: InvalidBody})
		return
	}

	id := h.Waiter.Add(req.Expression)
	ast, err := calc.Pre(req.Expression)
	if err != nil {
		h.Waiter.Error(id, err)
		ctx.JSON(calc.GetCode(err), resp.ResponseError{Error: err.Error()})
		return
	}

	// Using a goroutine to make
	// the long calculating process
	// is not affecting on the handler working
	go func() {
		res, err := calc.Calc(ast, h.Waiter.WaitingFunc(id))
		if err != nil {
			h.Waiter.Error(id, err)
			return
		}
		h.Waiter.Finish(id, res)
	}()

	ctx.JSON(http.StatusCreated, resp.Created{Id: id})
}

func (h *Handler) ExpressionsHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, resp.All{Expressions: h.Waiter.GetAll()})
}

func (h *Handler) ById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || !h.Waiter.Exist(id) {
		ctx.JSON(http.StatusNotFound, resp.ResponseError{Error: InvalidId})
		return
	}
	ctx.JSON(http.StatusOK, resp.Expression{Expression: h.Waiter.Get(id)})
}

func (h *Handler) Task(ctx *gin.Context) {
	job, ok := h.Waiter.GetJob()
	if !ok {
		ctx.JSON(http.StatusNotFound, resp.ResponseError{Error: NoJobsFound})
		return
	}
	ctx.JSON(http.StatusOK, resp.Job{Task: job})
}

func (h *Handler) FinishTask(ctx *gin.Context) {
	req := new(module.Post)
	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, resp.ResponseError{Error: InvalidBody})
		return
	}

	if !h.Waiter.Exist(req.Id) {
		ctx.JSON(http.StatusNotFound, resp.ResponseError{Error: NoJobsFound})
		return
	}

	if err := h.Waiter.FinishJob(req.Id, req.Result); errors.Is(err, vanerrors.Simple(wait.StatusIsNotProcessing)) {
		ctx.JSON(http.StatusUnprocessableEntity, resp.ResponseError{Error: err.Error()})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.ResponseError{Error: NoJobsFound})
		return
	}

	ctx.Writer.WriteHeader(http.StatusOK)
}
