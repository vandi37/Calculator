//go:build integration

package integration_test

import (
	"os"
	"testing"

	"github.com/vandi37/Calculator/integration"
)

func TestApp(t *testing.T) {
	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "http://localhost:8080/api/v1"
	}

	integration.Test(t.Context(), t, address)
}
