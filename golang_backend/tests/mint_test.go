package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"aptos/services/handler"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// Init the .env file
func init() {
	viper.SetConfigFile("../.env")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func TestMintAnimaToken_Success(t *testing.T) {
	// Mock JSON request body
	jsonBody := `{
        "description": "MINTED FROM GO",
        "name": "TECHNICAL TEST TOKEN",
        "base_uri": "URI/FOR/TEST",
        "soul_bound_to": "0xCAFE"
    }`
	body := bytes.NewBufferString(jsonBody)

	// Create a new Gin router
	r := gin.Default()
	r.POST("/mint", handler.MintAnimaToken)

	// Create a new request
	req, _ := http.NewRequest("POST", "/mint", body)
	req.Header.Set("Content-Type", "application/json")

	// Get request response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "txn_hash")
}

func TestMintAnimaToken_BadRequest(t *testing.T) {
	// Mock JSON request body with missing fields
	// If fields are missing, the reques will not fait, "" will be used as params
	jsonBody := `{
        "soul_bound_to": "0xCAFE"
    }`
	body := bytes.NewBufferString(jsonBody)

	// Create a new Gin router
	r := gin.Default()
	r.POST("/mint", handler.MintAnimaToken)

	// Create a new request
	req, _ := http.NewRequest("POST", "/mint", body)
	req.Header.Set("Content-Type", "application/json")

	// Get request response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "txn_hash")
}

func TestMintAnimaToken_InternalServerError(t *testing.T) {
	// Mock JSON request body with invalid private key
	viper.Set("PRIVATE_KEY", "invalid_private_key")

	jsonBody := `{
        "description": "MINTED FROM G0",
        "name": "TECHNICAL TEST TOKEN",
        "base_uri": "URI/FOR/TEST",
        "soul_bound_to": "0xCAFE"
    }`
	body := bytes.NewBufferString(jsonBody)

	// Create a new Gin router
	r := gin.Default()
	r.POST("/mint", handler.MintAnimaToken)

	// Create a new request
	req, _ := http.NewRequest("POST", "/mint", body)
	req.Header.Set("Content-Type", "application/json")

	/// Get request response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
