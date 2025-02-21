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
)

func RunMultiple(num int, path string, port int, maxErrors int, periodicity int, logger *zap.Logger) {
	path = fmt.Sprintf("http://localhost:%d%s", port, path)
	for i := 0; i < num; i++ {
		go Run(path, maxErrors, periodicity, logger)
	}
}

func Run(path string, maxErrors int, periodicity int, logger *zap.Logger) {
	sendBack := SendBack(path, logger)
	var n int
	for {
		got, err := http.Get(path)
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
		task.Task(req.Task, sendBack)()
		time.Sleep(time.Millisecond * time.Duration(periodicity))
	}
}

func SendBack(path string, logger *zap.Logger) func(int, float64) {
	return func(i int, f float64) {
		buf := bytes.NewBuffer([]byte{})
		err := json.NewEncoder(buf).Encode(module.Post{Id: i, Result: f})
		if err != nil {
			logger.Error("encoding failed", zap.Error(err))
			return
		}
		resp, err := http.Post(path, "application-json", buf)
		if err != nil {
			logger.Error("post request failed", zap.Error(err))
			return
		}

		switch resp.StatusCode {
		case http.StatusUnprocessableEntity:
			logger.Error("got status 422 [unprocessable entity]")
		case http.StatusBadRequest:
			logger.Error("got status 400 [bad request]")
		case http.StatusInternalServerError:
			logger.Error("got status 500 [internal server error]")
		case http.StatusOK:
			logger.Debug("got status 200 [ok]")
		default:
			logger.Warn(fmt.Sprintf("got unexpected status %d [%s]", resp.StatusCode, strings.TrimPrefix(resp.Status, strconv.Itoa(resp.StatusCode)+" ")))
		}
	}
}
