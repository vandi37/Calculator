// This tests cover
//
// 1. `../internal/status/` (using status)
//
// 2. `../internal/status/` (manipulation with processes)
// //
// In this case i think it isn't necessary to create separate tests for this modules
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/vandi37/Calculator/internal/agent/module"
	"github.com/vandi37/Calculator/internal/ms"
	"github.com/vandi37/Calculator/internal/status"
	"github.com/vandi37/Calculator/internal/wait"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
	"github.com/vandi37/vanerrors"
	"go.uber.org/zap"
)

func NewWait() *wait.Waiter {
	return wait.New(&ms.MsGetter{}, zap.NewNop())
}

func MakeJson(id int, expression string, status status.StatusLevel, value string) string {
	return fmt.Sprintf(`{"id":%d,"expression":"%s","status":"%s","value":%s}`, id, expression, status.String(), value)
}

// *(wait.Waiter).WaitingFunc(id int) isn't tested because it's functionality depends on *(wait.Waiter).StartWaiting(id int, arg1, arg2 float64, operation tree.ExprSep)
func TestWait(t *testing.T) {
	w := NewWait()

	// _________________________________________________________________________________________________________________________________________________________________
	// Adding

	expression := "my_expression"
	id := w.Add(expression)

	if !w.Exist(id) {
		t.Fatal("id isn't valid")
	}

	// _________________________________________________________________________________________________________________________________________________________________
	// Get

	expectedJson := MakeJson(id, expression, status.Nothing, "null")
	s, _ := json.Marshal(w.Get(id))

	if got := strings.TrimSpace(string(s)); got != expectedJson {
		t.Fatalf("json values don't match. expected: '%s', got: '%s'", expectedJson, got)
	}

	// _________________________________________________________________________________________________________________________________________________________________
	// Other add

	other := "other"
	id2 := w.Add(other)

	// _________________________________________________________________________________________________________________________________________________________________
	// Get all

	s, _ = json.Marshal(w.GetAll())

	expectedJson = "[" +
		MakeJson(id, expression, status.Nothing, "null") +
		"," +
		MakeJson(id2, other, status.Nothing, "null") +
		"]"

	if got := strings.TrimSpace(string(s)); got != expectedJson {
		t.Fatalf("json values don't match. expected: '%s', got: '%s'", expectedJson, got)
	}

	// _________________________________________________________________________________________________________________________________________________________________
	// Start waiting

	req := module.Request{Arg1: 10, Arg2: 20, Operation: tree.Addition, OperationTimeMs: 0}
	ch, err := w.StartWaiting(id, req.Arg1, req.Arg2, req.Operation)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// _________________________________________________________________________________________________________________________________________________________________
	// Get waiting
	jTask, _ := json.Marshal(req)
	expectedJson = MakeJson(id, expression, status.Waiting, string(jTask))
	s, _ = json.Marshal(w.Get(id))

	if got := strings.TrimSpace(string(s)); got != expectedJson {
		t.Fatalf("json values don't match. expected: '%s', got: '%s'", expectedJson, got)
	}

	// _________________________________________________________________________________________________________________________________________________________________
	// Get task
	task, ok := w.GetJob()
	if !ok {
		t.Fatal("expected an existing job")
	}

	if task != any(req) {
		t.Fatalf("got task and expected task don't match. expected: %v, got %v", req, task)
	}

	// _________________________________________________________________________________________________________________________________________________________________
	// Finish job
	res := 30.0
	go t.Run("finish job", func(t *testing.T) {
		err := w.FinishJob(id, res)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// _________________________________________________________________________________________________________________________________________________________________
	// Get job result
	jobRes := <-ch

	if jobRes != res {
		t.Fatalf("results don't match expected %3.f, got %3.f", res, jobRes)
	}

	// _________________________________________________________________________________________________________________________________________________________________
	// Finish
	result := 30.0
	err = w.Finish(id, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedJson = MakeJson(id, expression, status.Finished, fmt.Sprint(result))
	s, _ = json.Marshal(w.Get(id))

	if got := strings.TrimSpace(string(s)); got != expectedJson {
		t.Fatalf("json values don't match. expected: '%s', got: '%s'", expectedJson, got)
	}

	// _________________________________________________________________________________________________________________________________________________________________
	// Error
	sendError := vanerrors.Simple("error")
	err = w.Error(id, sendError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedJson = MakeJson(id, expression, status.Error, "\""+sendError.Error()+"\"")
	s, _ = json.Marshal(w.Get(id))

	if got := strings.TrimSpace(string(s)); got != expectedJson {
		t.Fatalf("json values don't match. expected: '%s', got: '%s'", expectedJson, got)
	}
}
