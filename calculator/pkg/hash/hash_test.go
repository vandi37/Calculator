package hash_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vandi37/Calculator/pkg/hash"
	"golang.org/x/crypto/bcrypt"
)

func TestParsePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
		errType   error
	}{
		{
			name:     "Valid password",
			password: "validPassword123",
		},
		{
			name:     "Empty password",
			password: "",
		},
		{
			name:      "Password too long",
			password:  "thisPasswordIsWayTooLongAndExceedsTheMaximumAllowedLengthOf72CharactersWhichIsTheLimitForBcrypt",
			wantError: true,
			errType:   hash.InvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := hash.ParsePassword(tt.password)

			if tt.wantError {
				require.Error(t, err)
				assert.Equal(t, tt.errType, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.password, string(p.GetBytes()))
		})
	}
}

func TestPassword_GetBytes(t *testing.T) {
	t.Run("Returns copy of password", func(t *testing.T) {
		original := "password123"
		p, err := hash.ParsePassword(original)
		require.NoError(t, err)

		bytes := p.GetBytes()
		bytes[0] = 'P'

		assert.Equal(t, original, string(p.GetBytes()))
	})
}

func TestPasswordService_HashPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:     "Valid password",
			password: "securePassword!123",
		},
		{
			name:      "Too long password",
			password:  "thisPasswordIsWayTooLongAndExceedsTheMaximumAllowedLengthOf72CharactersWhichIsTheLimitForBcrypt",
			wantError: true,
		},
	}

	service := hash.NewPasswordService(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashStr, err := service.HashPassword(tt.password)

			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			byteHash, err := base64.StdEncoding.DecodeString(hashStr)
			require.NoError(t, err)

			assert.NotEmpty(t, byteHash)

			// Verify the hash can be decoded and is a valid bcrypt hash
			_, err = bcrypt.Cost(byteHash)
			assert.NoError(t, err)
		})
	}

	t.Run("Different costs produce different hashes", func(t *testing.T) {
		password := "samePassword"
		lowCostService := hash.NewPasswordService(intPtr(4))
		highCostService := hash.NewPasswordService(intPtr(10))

		hash1, err := lowCostService.HashPassword(password)
		require.NoError(t, err)

		hash2, err := highCostService.HashPassword(password)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestPasswordService_ComparePassword(t *testing.T) {
	service := hash.NewPasswordService(nil)
	password := "correctPassword"
	wrongPassword := "wrongPassword"

	t.Run("Correct password", func(t *testing.T) {
		hashStr, err := service.HashPassword(password)
		require.NoError(t, err)

		err = service.ComparePassword(password, hashStr)
		assert.NoError(t, err)
	})

	t.Run("Wrong password", func(t *testing.T) {
		hashStr, err := service.HashPassword(password)
		require.NoError(t, err)

		err = service.ComparePassword(wrongPassword, hashStr)
		assert.Error(t, err)
		assert.Equal(t, bcrypt.ErrMismatchedHashAndPassword, err)
	})

	t.Run("Invalid base64 hash", func(t *testing.T) {
		err := service.ComparePassword(password, "notaBase64string!!!")
		assert.Equal(t, hash.InvalidBase64, err)
	})

	t.Run("Empty password", func(t *testing.T) {
		hashStr, err := service.HashPassword("")
		require.NoError(t, err)

		err = service.ComparePassword("", hashStr)
		assert.NoError(t, err)

		err = service.ComparePassword("notEmpty", hashStr)
		assert.Error(t, err)
	})

	t.Run("Too long password", func(t *testing.T) {
		longPassword := "thisPasswordIsWayTooLongAndExceedsTheMaximumAllowedLengthOf72CharactersWhichIsTheLimitForBcrypt"
		err := service.ComparePassword(longPassword, "anyHash")
		assert.Equal(t, hash.InvalidPassword, err)
	})
}

func TestNewPasswordService(t *testing.T) {
	t.Run("Default cost", func(t *testing.T) {
		service := hash.NewPasswordService(nil)
		assert.Equal(t, bcrypt.DefaultCost, service.GetCost())
	})

	t.Run("Custom cost", func(t *testing.T) {
		customCost := 8
		service := hash.NewPasswordService(&customCost)
		assert.Equal(t, customCost, service.GetCost())
	})

	t.Run("Minimum cost", func(t *testing.T) {
		minCost := 4
		service := hash.NewPasswordService(&minCost)
		assert.Equal(t, minCost, service.GetCost())
	})
}

func intPtr(i int) *int {
	return &i
}
