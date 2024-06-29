package main

import (
	"aptos/services/config"
	"aptos/services/handler"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration to acces .env variables
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Start goroutine to monitor payments on payment contract
	contract := viper.GetString("PAYMENT_CONTRACT_ADDRESS")
	go handler.MonitorPayments(contract)

	// Expose API endpoint /mint to mint a token on soulbound contract
	r := gin.Default()
	r.POST("/mint", handler.MintAnimaToken)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
