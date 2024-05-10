package main

import (
    "fmt"
)

const (
    ethereumNodeURL = "https://cloudflare-eth.com"
)

func main() {
    parser := NewEthParser()
    currentBlock := parser.GetCurrentBlock()
    fmt.Println("Current Ethereum Block:", currentBlock)

    address := "0x00000000219ab540356cBB839Cbe05303d7705Fa"
    if parser.Subscribe(address) {
        fmt.Println("Subscription Successful: true")
        transactions, err := parser.GetTransactions(address)
        if err != nil {
            fmt.Println("Error fetching transactions:", err)
        } else {
            fmt.Println("Transactions involving the address:", transactions)
        }
    } else {
        fmt.Println("Subscription Failed")
    }
}

