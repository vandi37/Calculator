package get

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vandi37/Calculator/internal/agent/module"
	"github.com/vandi37/Calculator/internal/agent/task"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func RunMultiple(ctx context.Context, num int, path string, port int, maxErrors int, periodicity int, logger *zap.Logger) {
	path = fmt.Sprintf("http://localhost:%d%s", port, path)
	for i := 0; i < num; i++ {
		go Run(ctx, path, maxErrors, periodicity, logger)
	}
	<-ctx.Done()
}

func Run(ctx context.Context, path string, maxErrors int, periodicity int, logger *zap.Logger) {
	client := &http.Client{}

	// -----------------------------------------------------------------------------------------------------------------------------------------------------------
	// send back
	sendBack := func(i int, f float64, r module.Request) {
		buf := bytes.NewBuffer([]byte{})
		err := json.NewEncoder(buf).Encode(module.Post{Id: i, Result: f})
		if err != nil {
			logger.Error("encoding failed", zap.Int("id", i), zap.Error(err))
			return
		}
		resp, err := client.Post(path, "application-json", buf)
		if err != nil {
			logger.Error("post request failed", zap.Int("id", i), zap.Error(err))
			return
		}

		switch resp.StatusCode {
		case http.StatusUnprocessableEntity:
			logger.Error("got status 422 Unprocessable Entity", zap.Int("id", i))
		case http.StatusBadRequest:
			logger.Error("got status 400 Bad Request", zap.Int("id", i))
		case http.StatusInternalServerError:
			logger.Error("got status 500 Internal Server Error", zap.Int("id", i))
		case http.StatusOK:
			logger.Debug("got status 200 OK", zap.Int("id", i), zap.String("task", r.String()), zap.Float64("result", f))
		default:
			logger.Warn(fmt.Sprintf("got unexpected status %d %s", resp.StatusCode, strings.TrimPrefix(resp.Status, strconv.Itoa(resp.StatusCode)+" ")), zap.Int("id", i))
		}
	}
	// -----------------------------------------------------------------------------------------------------------------------------------------------------------

	var n int
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		got, err := client.Get(path)
		if err != nil {
			logger.Error("request failed", zap.Error(err))
			n++
			if n >= maxErrors {
				return
			}
			continue
		}

		if got.StatusCode == http.StatusNotFound {
			time.Sleep(time.Millisecond * time.Duration(periodicity))
			continue
		}
		if got.StatusCode != http.StatusOK {
			logger.Error("status code isn't ok", zap.Int("status", got.StatusCode))
			n++
			if n >= maxErrors {
				return
			}
			continue
		}

		defer got.Body.Close()

		req := new(struct {
			Task module.Request `json:"task"`
		})
		err = json.NewDecoder(got.Body).Decode(req)
		if err != nil {
			logger.Error("decoding failed", zap.Error(err))
			n++
			if n >= maxErrors {
				return
			}
			continue
		}
		task.Task(req.Task, sendBack)
		time.Sleep(time.Millisecond * time.Duration(periodicity))
	}
}
