package tests

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"

	"aptos/services/handler"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetTransactionDetails(t *testing.T) {
	// This transaction number will never move, so we can use it to test the function
	txnDetail, err := handler.GetTransactionDetails("100000")
	assert.NoError(t, err)
	assert.NotNil(t, txnDetail)
	assert.Equal(t, "0x974da4cd12516fdc0ec7a46e8b99d0f9df9264325177de237d3f9ff46f32b0ad", txnDetail.Sender)
	assert.Equal(t, "1719522534293117", txnDetail.Timestamp)
	assert.Equal(t, "0x70ea96bc03c064651e9636d506fd7a7e572aba087d6ab87a0b316cea35b5a93d", txnDetail.Hash)
	assert.Equal(t, "0xa108583a34fccc38cdc41e4218fb9bd44bf7212efc0db8541aa4e76d760f0de6::resource_groups_example::set_and_read_p", txnDetail.Payload.Function)

	// An other way to test the function through httpmock if you do not want to use a real request

	/*httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.devnet.aptoslabs.com/v1/transactions/by_version/1001",
		httpmock.NewStringResponder(200, `{
			"sender": "0x1",
			"timestamp": "1627593982",
			"hash": "somehash",
			"payload": {
				"function": "0x1::aptos_coin::receive_payment"
			}
		}`))

	txnDetail, err := handler.GetTransactionDetails("1001")
	assert.NoError(t, err)
	assert.NotNil(t, txnDetail)
	assert.Equal(t, "0x1", txnDetail.Sender)
	assert.Equal(t, "1627593982", txnDetail.Timestamp)
	assert.Equal(t, "somehash", txnDetail.Hash)
	assert.Equal(t, "0x1::aptos_coin::receive_payment", txnDetail.Payload.Function)*/
}

func TestMonitorPayments(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://api.devnet.aptoslabs.com/v1/graphql",
		httpmock.NewStringResponder(200, `{
			"data": {
				"account_transactions": [
					{
						"transaction_version": "1001",
						"__typename": "account_transactions"
					}
				]
			}
		}`))

	httpmock.RegisterResponder("GET", "https://api.devnet.aptoslabs.com/v1/transactions/by_version/1001",
		httpmock.NewStringResponder(200, `{
			"sender": "0x1",
			"timestamp": "1627593982",
			"hash": "somehash",
			"payload": {
				"function": "0xCAFE::payment::receive_payment"
			},
			"events": [
				{
					"guid": {
						"creation_number": "0",
						"account_address": "0x0"
					},
					"sequence_number": "0",
					"type": "0x1::coin::CoinWithdraw",
					"data": {
						"account": "0xa8a1f3b50a3b156273e3df4ed26c267c853808dd9cced2b8308985a58e9c42a7",
						"amount": "20",
						"coin_type": "0x1::aptos_coin::AptosCoin"
					}
				}
			]
		}`))

	viper.Set("PAYMENT_CONTRACT_ADDRESS", "0xCAFE")

	output := getLogOutput(func() {
		go handler.MonitorPayments("0xCAFE")
		time.Sleep(2 * time.Second)
	})

	assert.Contains(t, output, "Interaction with contract detected: {TransactionVersion:1001 Typename:account_transactions}")
	assert.Contains(t, output, "Payment detected: Sender: 0x1")
}

// Helper function to get what is logged by the MonitorPayments function
func getLogOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stdout)

	f()
	return buf.String()
}
