package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strconv"
    "strings"
    "sync"
)

// EthParser contains the state and synchronization primitives.
type EthParser struct {
    currentBlock  int
    subscriptions map[string]bool
    mutex         sync.Mutex
}

// NewEthParser creates a new Ethereum Parser.
func NewEthParser() *EthParser {
    return &EthParser{
        subscriptions: make(map[string]bool),
    }
}

// GetCurrentBlock retrieves the latest block number from the Ethereum blockchain.
func (p *EthParser) GetCurrentBlock() int {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    response, err := http.Post(ethereumNodeURL, "application/json", bytes.NewBuffer([]byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)))
    if err != nil {
        fmt.Println("Error fetching current block:", err)
        return -1
    }
    defer response.Body.Close()
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        return -1
    }

    fmt.Println("Response Body:", string(body))

    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        fmt.Println("Error parsing JSON:", err)
        return -1
    }

    blockNumberHex, ok := result["result"].(string)
    if !ok {
        fmt.Println("Error: result is not a string.")
        return -1
    }

    blockNumber, err := strconv.ParseInt(blockNumberHex[2:], 16, 64) // Parse the hex string to an integer
    if err != nil {
        fmt.Println("Error converting hex to integer:", err)
        return -1
    }

    p.currentBlock = int(blockNumber)
    return p.currentBlock
}

// Subscribe adds an address to watch for transactions.
func (p *EthParser) Subscribe(address string) bool {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    normalizedAddress := strings.ToLower(address)
    if _, ok := p.subscriptions[normalizedAddress]; !ok {
        p.subscriptions[normalizedAddress] = true
        return true
    }
    return false
}

// GetTransactions retrieves transactions for a specific address from the current block.
func (p *EthParser) GetTransactions(address string) ([]Transaction, error) {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    var transactions []Transaction

    response, err := http.Post(ethereumNodeURL, "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":1}`, p.currentBlock))))
    if err != nil {
        return nil, fmt.Errorf("error fetching block transactions: %w", err)
    }
    defer response.Body.Close()
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %w", err)
    }

    var result map[string]interface{}
    json.Unmarshal(body, &result)
    blockData, ok := result["result"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("error parsing block data")
    }

    normalizedAddress := strings.ToLower(address)
    if txs, found := blockData["transactions"].([]interface{}); found {
        fmt.Printf("Number of transactions in block %d: %d\n", p.currentBlock, len(txs))
        for _, tx := range txs {
            txMap := tx.(map[string]interface{})
            from, fromOk := txMap["from"].(string)
            to, toOk := txMap["to"].(string)
            if fromOk && toOk {
                fmt.Printf("Transaction from %s to %s\n", from, to)
                if strings.ToLower(from) == normalizedAddress || strings.ToLower(to) == normalizedAddress {
                    transaction := Transaction{
                        From:  from,
                        To:    to,
                        Value: fmt.Sprintf("%v", txMap["value"]), // Convert to string safely
                    }
                    transactions = append(transactions, transaction)
                    fmt.Printf("Added transaction: %+v\n", transaction)
                }
            }
        }
    }
    if len(transactions) == 0 {
        fmt.Println("No transactions found for the address in the queried block.")
    }
    return transactions, nil
}

