package parser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEthereumParser_Subscribe(t *testing.T) {
	parser := NewEthereumParser("", http.Client{})

	address := "0x1234"

	// First subscription should succeed
	if !parser.Subscribe(address) {
		t.Fatalf("expected subscription to succeed for address: %s", address)
	}

	// Subscribing again should return false (duplicate subscription)
	if parser.Subscribe(address) {
		t.Fatalf("expected subscription to fail for duplicate address: %s", address)
	}
}

func TestEthereumParser_GetTransactions(t *testing.T) {
	parser := NewEthereumParser("", http.Client{})

	address := "0x1234"
	parser.Subscribe(address)

	// No transactions should exist initially
	transactions := parser.GetTransactions(address)
	if len(transactions) != 0 {
		t.Fatalf("expected no transactions, got: %d", len(transactions))
	}

	// Add a transaction manually for the subscribed address
	parser.transactions[address] = []Transaction{
		{Hash: "0xhash", From: address, To: "0x5678", Value: "100"},
	}

	// Fetch transactions and verify
	transactions = parser.GetTransactions(address)
	if len(transactions) != 1 {
		t.Fatalf("expected 1 transaction, got: %d", len(transactions))
	}

	if transactions[0].Hash != "0xhash" {
		t.Fatalf("expected transaction hash 0xhash, got: %s", transactions[0].Hash)
	}
}

func TestEthereumParser_GetCurrentBlock(t *testing.T) {
	parser := NewEthereumParser("", http.Client{})

	// Initial block should be 0
	if parser.GetCurrentBlock() != 0 {
		t.Fatalf("expected initial block to be 0")
	}

	// Update block number
	parser.currentBlock = 100

	// Current block should be 100
	if parser.GetCurrentBlock() != 100 {
		t.Fatalf("expected current block to be 100")
	}
}

// Mock Ethereum RPC server response for a block with transactions
func mockEthereumRPCResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		if req["method"] == "eth_getBlockByNumber" {
			block := Block{
				Number: "0x1", // Block number in hexadecimal (1 in decimal)
				Transactions: []Transaction{
					{
						Hash:        "0xabc123",
						From:        "0x1234",
						To:          "0x5678",
						Value:       "1000000000000000000",
						BlockNumber: "1",
					},
				},
			}

			// JSON-RPC response format
			resp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req["id"],
				"result":  block,
			}

			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	// Default: return error response for unexpected requests
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"error": "Invalid request"}`))
}

func TestParseBlock(t *testing.T) {
	// Create a mock Ethereum RPC server
	mockServer := httptest.NewServer(http.HandlerFunc(mockEthereumRPCResponse))
	defer mockServer.Close()

	// Create a new parser instance with the mock server URL
	parser := NewEthereumParser(mockServer.URL, http.Client{})

	// Subscribe to the "from" and "to" addresses that we expect in the mock response
	parser.Subscribe("0x1234") // From address
	parser.Subscribe("0x5678") // To address

	// Call the ParseBlock function with block number 1
	err := parser.ParseBlock(1)
	if err != nil {
		t.Fatalf("ParseBlock failed: %v", err)
	}

	// Check if transactions were correctly parsed and stored for both addresses
	fromTransactions := parser.GetTransactions("0x1234")
	toTransactions := parser.GetTransactions("0x5678")

	if len(fromTransactions) == 0 {
		t.Errorf("Expected transactions for address 0x1234, but got none")
	}

	if len(toTransactions) == 0 {
		t.Errorf("Expected transactions for address 0x5678, but got none")
	}

	// Check specific transaction details
	expectedHash := "0xabc123"
	if fromTransactions[0].Hash != expectedHash || toTransactions[0].Hash != expectedHash {
		t.Errorf("Expected transaction hash %s, but got %s", expectedHash, fromTransactions[0].Hash)
	}
}
