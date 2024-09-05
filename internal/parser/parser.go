package parser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// Transaction represents an Ethereum transaction
type Transaction struct {
	Hash        string
	From        string
	To          string
	Value       string
	BlockNumber string
}

// Block represents an Ethereum block
type Block struct {
	Number       string        `json:"number"`
	Transactions []Transaction `json:"transactions"`
}

// Parser is the interface that defines the required operations
type Parser interface {
	// GetCurrentBlock returns the last parsed block number
	GetCurrentBlock() int

	// Subscribe adds an address to the observer list
	Subscribe(address string) bool

	// GetTransactions returns a list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

type EthereumParser struct {
	currentBlock  int
	subscriptions map[string]bool
	transactions  map[string][]Transaction
	mu            sync.RWMutex
	rpcURL        string
	client        http.Client
}

// NewEthereumParser creates a new instance of EthereumParser
func NewEthereumParser(rpcURL string, client http.Client) *EthereumParser {
	return &EthereumParser{
		currentBlock:  0,
		subscriptions: make(map[string]bool),
		transactions:  make(map[string][]Transaction),
		rpcURL:        rpcURL,
		client:        client,
	}
}

// GetCurrentBlock returns the last parsed block number
func (p *EthereumParser) GetCurrentBlock() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentBlock
}

// Subscribe adds an address to the observer list
func (p *EthereumParser) Subscribe(address string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.subscriptions[address]; exists {
		return false
	}
	p.subscriptions[address] = true
	return true
}

// GetTransactions returns a list of inbound or outbound transactions for an address
func (p *EthereumParser) GetTransactions(address string) []Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.transactions[address]
}

// ParseBlock fetches and parses a block for transactions involving subscribed addresses
func (p *EthereumParser) ParseBlock(blockNumber uint64) error {
	block, err := p.getBlockByNumber(blockNumber)
	if err != nil {
		return fmt.Errorf("failed to get block: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, tx := range block.Transactions {
		if p.subscriptions[tx.From] {
			p.transactions[tx.From] = append(p.transactions[tx.From], tx)
		}
		if p.subscriptions[tx.To] {
			p.transactions[tx.To] = append(p.transactions[tx.To], tx)
		}
	}

	p.currentBlock = int(blockNumber)
	return nil
}

// getBlockByNumber fetches a block by its number using Ethereum JSON-RPC
func (p *EthereumParser) getBlockByNumber(blockNumber uint64) (*Block, error) {
	var block Block
	params := []interface{}{fmt.Sprintf("0x%x", blockNumber), true}
	if err := p.callRPC("eth_getBlockByNumber", params, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// callRPC is a helper function to make JSON-RPC calls
func (p *EthereumParser) callRPC(method string, params []interface{}, result interface{}) error {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON Marshal Error: %w", err)
	}

	resp, err := p.client.Post(p.rpcURL, "application/json", strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("RPC Error: %w", err)
	}
	defer resp.Body.Close()

	var rpcResponse map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return fmt.Errorf("JSON Decode Error: %w", err)
	}

	if errorMsg, exists := rpcResponse["error"]; exists {
		var rpcErr RPCError
		if err := json.Unmarshal(errorMsg, &rpcErr); err != nil {
			return fmt.Errorf("RPC Error: %s", errorMsg)
		}
		return fmt.Errorf("RPC Error %d: %s", rpcErr.Code, rpcErr.Message)
	}

	return json.Unmarshal(rpcResponse["result"], result)
}
