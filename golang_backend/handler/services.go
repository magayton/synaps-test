package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type MintRequest struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	BaseURI     string `json:"base_uri"`
	SoulBoundTo string `json:"soul_bound_to"`
}

func MintAnimaToken(c *gin.Context) {
	// Bind the request body from the call to the endpoint to a MintRequest struct
	var request MintRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error ": err.Error()})
		return
	}

	// Load from .env
	privateKey := viper.GetString("PRIVATE_KEY")
	contract := viper.GetString("MINT_CONTRACT_ADDRESS")

	// Setup for Aptos client for devnet
	client, err := aptos.NewClient(aptos.DevnetConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error failed to create client ": err.Error()})
		return
	}

	// Create accoutn from private key
	key := crypto.Ed25519PrivateKey{}
	err = key.FromHex(privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error failed to decode private key ": err.Error()})
		return
	}
	sender, err := aptos.NewAccountFromSigner(&key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error failed to create sender ": err.Error()})
		return
	}

	// Create a module from the account with the address of the soulbound contract
	contractAddress := &aptos.AccountAddress{}
	contractAddress.ParseStringRelaxed(contract)
	// This module represents the soulbound contract
	soulboundModule := aptos.ModuleId{Address: *contractAddress, Name: "soulbound"}

	// Arguments for the mint_anima_token function must be Serialized
	receiverAddress := &aptos.AccountAddress{}
	receiverAddress.ParseStringRelaxed(request.SoulBoundTo)
	serializedReceiverAddress, err := bcs.Serialize(receiverAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error failed to serialize receiver address ": err.Error()})
		return
	}

	// First we need to convert the string to bytes because String serialization is not supported
	description, err := bcs.SerializeBytes([]byte(request.Description))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error failed to serialize description ": err.Error()})
		return
	}

	name, err := bcs.SerializeBytes([]byte(request.Name))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error failed to serialize name ": err.Error()})
		return
	}

	baseURI, err := bcs.SerializeBytes([]byte(request.BaseURI))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error failed to serialize baseURI ": err.Error()})
		return
	}

	var noTypeTags []aptos.TypeTag
	payload := aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module:   soulboundModule,
			Function: "mint_soulbound_token", // Entry function to call on the soulbound contract
			ArgTypes: noTypeTags,
			Args:     [][]byte{description, name, baseURI, serializedReceiverAddress},
		},
	}

	response, err := client.BuildSignAndSubmitTransaction(sender, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error failed to build and submit transaction ": err.Error()})
		return
	}

	_, err = client.WaitForTransaction(response.Hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error failed to wait for transaction ": err.Error()})
		return
	}

	// Return the transaction hash the submissions was successful
	c.JSON(http.StatusOK, gin.H{"txn_hash": response.Hash})
}

// -------------------------
//
//	Payment Monitoring
//
// -------------------------

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type TransactionData struct {
	TransactionVersion json.Number `json:"transaction_version"`
	Typename           string      `json:"__typename"`
}

type GraphQLResponse struct {
	Data struct {
		AccountTransactions []TransactionData `json:"account_transactions"`
	} `json:"data"`
}

func MonitorPayments(paymentContract string) {
	// lastTransactionVersion is set to 0 here, but in production it could be different (at least hte transaction version of the instantiation of the payment contract)
	lastTransactionVersion := uint64(0)
	graphqlEndpoint := "https://api.devnet.aptoslabs.com/v1/graphql"

	// This query retrieves all transactions for the payment contract address if they are greater than lastTransactionVersion
	query := `
        query GetAccountTransactionsData($address: String, $limit: Int, $gt: bigint) {
            account_transactions(
                where: { account_address: { _eq: $address }, transaction_version: {_gt: $gt} }
                order_by: {transaction_version: asc}
                limit: $limit
            ) {
                transaction_version
                __typename
            }
        }
    `

	for {
		variables := map[string]interface{}{
			"address": paymentContract,
			"limit":   100,
			"gt":      lastTransactionVersion,
		}
		requestBody := GraphQLRequest{
			Query:     query,
			Variables: variables,
		}

		// Serialize the request and POST it
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			log.Printf("Error marshaling request body: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		resp, err := http.Post(graphqlEndpoint, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			log.Printf("Error making GraphQL request: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// get raw response of the query
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Printf("Error reading response body: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Unserialize the response to handle the data
		var graphqlResponse GraphQLResponse
		if err := json.Unmarshal(body, &graphqlResponse); err != nil {
			log.Printf("Error unmarshaling response body: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// We iterate over all transactions we had in response
		newLastTransactionVersion := lastTransactionVersion
		for _, txn := range graphqlResponse.Data.AccountTransactions {
			txnVersion := txn.TransactionVersion.String()
			txnVersionUint, err := strconv.ParseUint(txnVersion, 10, 64)

			if err != nil {
				log.Printf("Error converting transaction version: %v", err)
				continue
			}

			// Logging every transactions that are greater than lastTransactionVersion
			if txnVersionUint > lastTransactionVersion {
				log.Printf("Interaction with contract detected: %+v", txn)
				txnDetails, err := GetTransactionDetails(txnVersion)
				if err != nil {
					log.Printf("Error getting transaction details: %v", err)
					continue
				}

				// Filter on receive_payment function
				contract_endpoint := fmt.Sprintf("%s::payment::receive_payment", paymentContract)

				if txnDetails.Payload.Function == contract_endpoint {
					log.Printf("Payment detected: Sender: %s, Amount: %s, Hash: %s, Timestamp: %s", txnDetails.Sender, txnDetails.Events[0].Data.Amount, txnDetails.Hash, txnDetails.Timestamp)
				} else {
					log.Printf("Interaction not for payment entry point")
				}

				if txnVersionUint > newLastTransactionVersion {
					newLastTransactionVersion = txnVersionUint
				}
			}
		}

		// Since we use a goroutine, updating lastTransactionVersion sets the new value for the next iteration
		lastTransactionVersion = newLastTransactionVersion

		// Limit is 500 calls for 5 min for the API, here we have 60 calls for 5 min
		time.Sleep(5 * time.Second)
	}
}

// Only a part of the transaction detail is used, full can be found here : https://aptos.dev/en/build/apis/fullnode-rest-api-reference?network=devnet#tag/transactions/get/transactions/by_version/{txn_version}
type TransactionDetail struct {
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
	Hash      string `json:"hash"`
	Payload   struct {
		Function string `json:"function"`
	} `json:"payload"`
	Events []Event `json:"events"`
}

type Event struct {
	Guid struct {
		CreationNumber string `json:"creation_number"`
		AccountAddress string `json:"account_address"`
	} `json:"guid"`
	SequenceNumber string `json:"sequence_number"`
	Type           string `json:"type"`
	Data           struct {
		Account  string `json:"account"`
		Amount   string `json:"amount"`
		CoinType string `json:"coin_type"`
	} `json:"data"`
}

// Calls the Aptos API to get full details of a transaction thanks to its transaction version
func GetTransactionDetails(version string) (*TransactionDetail, error) {
	url := fmt.Sprintf("https://api.devnet.aptoslabs.com/v1/transactions/by_version/%s", version)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var txnDetail TransactionDetail
	if err := json.Unmarshal(body, &txnDetail); err != nil {
		return nil, err
	}

	return &txnDetail, nil
}
