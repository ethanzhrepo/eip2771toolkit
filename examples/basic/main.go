package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethanzhrepo/eip2771toolkit"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	fmt.Println("EIP-2771 Toolkit Basic Usage Example")
	fmt.Println("====================================")

	// Example 1: Generate private keys
	fmt.Println("\n1. Generating private keys...")

	userPrivKey, err := eip2771toolkit.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("Failed to generate user private key: %v", err)
	}
	userAddr := eip2771toolkit.AddressFromPrivateKey(userPrivKey)
	fmt.Printf("User address: %s\n", userAddr.Hex())

	relayerPrivKey, err := eip2771toolkit.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("Failed to generate relayer private key: %v", err)
	}
	relayerAddr := eip2771toolkit.AddressFromPrivateKey(relayerPrivKey)
	fmt.Printf("Relayer address: %s\n", relayerAddr.Hex())

	// Example 2: Create MetaTx
	fmt.Println("\n2. Creating MetaTx...")

	// Example addresses and values
	recipientAddr := common.HexToAddress("0x742b15cf35dF7bcdFAcE36CB4E8C4cF03B06cE85")
	tokenAddr := common.HexToAddress("0xA0b86a33E6411d01c8a96B60Dc2d7c8DB76A57E5") // Example ERC20 token
	amount := big.NewInt(1000000000000000000)                                      // 1 token (18 decimals)
	nonce := uint64(1)

	metaTx := eip2771toolkit.NewMetaTxWithDelay(
		userAddr,      // from
		recipientAddr, // to
		tokenAddr,     // token
		amount,        // amount
		100000,        // gas limit
		nonce,         // nonce
		3600,          // deadline in 1 hour
	)

	fmt.Printf("MetaTx created:\n")
	fmt.Printf("  From: %s\n", metaTx.From.Hex())
	fmt.Printf("  To: %s\n", metaTx.To.Hex())
	fmt.Printf("  Token: %s\n", metaTx.Token.Hex())
	fmt.Printf("  Amount: %s\n", metaTx.Amount.String())
	fmt.Printf("  Gas: %d\n", metaTx.Gas)
	fmt.Printf("  Nonce: %d\n", metaTx.Nonce)
	fmt.Printf("  Deadline: %d\n", metaTx.Deadline)

	// Example 3: Build domain separator
	fmt.Println("\n3. Building domain separator...")

	chainId := big.NewInt(1) // Ethereum mainnet
	forwarderAddr := common.HexToAddress("0x123456789012345678901234567890123456789")

	domainSeparator, err := eip2771toolkit.CreateDomainSeparatorForChain(chainId, forwarderAddr)
	if err != nil {
		log.Fatalf("Failed to build domain separator: %v", err)
	}
	fmt.Printf("Domain separator: %x\n", domainSeparator)

	// Example 4: Sign MetaTx
	fmt.Println("\n4. Signing MetaTx...")

	signature, err := eip2771toolkit.SignMetaTx(metaTx, userPrivKey, domainSeparator)
	if err != nil {
		log.Fatalf("Failed to sign MetaTx: %v", err)
	}
	fmt.Printf("Signature created:\n")
	fmt.Printf("  V: %d\n", signature.V)
	fmt.Printf("  R: %x\n", signature.R)
	fmt.Printf("  S: %x\n", signature.S)

	// Example 5: Verify signature
	fmt.Println("\n5. Verifying signature...")

	isValid, err := eip2771toolkit.VerifyMetaTxSignature(metaTx, signature, domainSeparator)
	if err != nil {
		log.Fatalf("Failed to verify signature: %v", err)
	}
	fmt.Printf("Signature is valid: %t\n", isValid)

	// Example 6: Demonstrate utility functions
	fmt.Println("\n6. Utility functions...")

	// Convert amounts
	etherAmount := big.NewFloat(1.5) // 1.5 ETH
	weiAmount := eip2771toolkit.ToWei(etherAmount)
	fmt.Printf("1.5 ETH = %s wei\n", weiAmount.String())

	backToEther := eip2771toolkit.FromWei(weiAmount)
	fmt.Printf("%s wei = %s ETH\n", weiAmount.String(), backToEther.String())

	// Generate random nonce
	randomNonce, err := eip2771toolkit.GenerateRandomNonce()
	if err != nil {
		log.Fatalf("Failed to generate random nonce: %v", err)
	}
	fmt.Printf("Random nonce: %d\n", randomNonce)

	// Current timestamp
	timestamp := eip2771toolkit.GetCurrentTimestamp()
	fmt.Printf("Current timestamp: %d\n", timestamp)

	fmt.Println("\n7. Example: How to relay transaction (requires live connection)")
	fmt.Println("// Connect to Ethereum node")
	fmt.Println("// client, err := ethclient.Dial(\"https://mainnet.infura.io/v3/YOUR-PROJECT-ID\")")
	fmt.Println("// if err != nil {")
	fmt.Println("//     log.Fatal(err)")
	fmt.Println("// }")
	fmt.Println("//")
	fmt.Println("// ctx := context.Background()")
	fmt.Println("// txHash, err := eip2771toolkit.RelayMetaTx(")
	fmt.Println("//     ctx,")
	fmt.Println("//     metaTx,")
	fmt.Println("//     signature,")
	fmt.Println("//     relayerPrivKey,")
	fmt.Println("//     erc2771ForwarderAddr,  // Updated: now uses ERC2771Forwarder")
	fmt.Println("//     client,")
	fmt.Println("// )")
	fmt.Println("// if err != nil {")
	fmt.Println("//     log.Fatal(err)")
	fmt.Println("// }")
	fmt.Println("// fmt.Printf(\"Transaction hash: %s\", txHash.Hex())")

	fmt.Println("\nExample completed successfully!")
}

// ExampleWithRealConnection demonstrates usage with a real Ethereum connection
func ExampleWithRealConnection() {
	// This function shows how to use the toolkit with a real Ethereum connection
	// Uncomment and modify the following code when you have access to an Ethereum node

	/*
		// Connect to Ethereum node (replace with your node URL)
		client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
		if err != nil {
			log.Fatalf("Failed to connect to Ethereum node: %v", err)
		}

		// Setup
		userPrivKey, _ := eip2771toolkit.GeneratePrivateKey()
		relayerPrivKey, _ := eip2771toolkit.GeneratePrivateKey()
		userAddr := eip2771toolkit.AddressFromPrivateKey(userPrivKey)

		// Contract addresses (replace with real addresses)
		forwarderAddr := common.HexToAddress("0x...")
		tokenAddr := common.HexToAddress("0x...")
		recipientAddr := common.HexToAddress("0x...")

		// Get current nonce from the forwarder contract
		ctx := context.Background()
		nonce, err := eip2771toolkit.GetMetaTxNonce(ctx, forwarderAddr, userAddr, client)
		if err != nil {
			log.Fatalf("Failed to get nonce: %v", err)
		}

		// Create MetaTx
		amount := big.NewInt(1000000000000000000) // 1 token
		metaTx := eip2771toolkit.NewMetaTxWithDelay(userAddr, recipientAddr, tokenAddr, amount, 100000, nonce, 3600)

		// Get chain ID
		chainId, err := client.NetworkID(ctx)
		if err != nil {
			log.Fatalf("Failed to get chain ID: %v", err)
		}

		// Build domain separator
		domainSeparator, err := eip2771toolkit.CreateDomainSeparatorForChain(chainId, forwarderAddr)
		if err != nil {
			log.Fatalf("Failed to build domain separator: %v", err)
		}

		// Sign MetaTx
		signature, err := eip2771toolkit.SignMetaTx(metaTx, userPrivKey, domainSeparator)
		if err != nil {
			log.Fatalf("Failed to sign MetaTx: %v", err)
		}

		// Relay transaction
		txHash, err := eip2771toolkit.RelayMetaTx(ctx, metaTx, signature, relayerPrivKey, forwarderAddr, client)
		if err != nil {
			log.Fatalf("Failed to relay transaction: %v", err)
		}

		fmt.Printf("Transaction relayed successfully! Hash: %s\n", txHash.Hex())
	*/
}
