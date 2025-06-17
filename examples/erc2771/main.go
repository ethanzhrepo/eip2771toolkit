package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethanzhrepo/eip2771toolkit"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	fmt.Println("EIP-2771 Toolkit - ERC2771Forwarder Integration Example")
	fmt.Println("======================================================")

	// This example demonstrates the updated integration with OpenZeppelin's
	// ERC2771Forwarder contract (v5.x)

	// 1. Generate keys for demonstration
	fmt.Println("\n1. Setting up accounts...")

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

	// 2. Contract addresses (example addresses - replace with real ones)
	fmt.Println("\n2. Contract addresses...")

	// Example ERC2771Forwarder contract address
	forwarderAddr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	fmt.Printf("ERC2771Forwarder address: %s\n", forwarderAddr.Hex())

	// Example ERC20 token contract address
	tokenAddr := common.HexToAddress("0xA0b86a33E6411d01c8a96B60Dc2d7c8DB76A57E5")
	fmt.Printf("Token contract address: %s\n", tokenAddr.Hex())

	// Recipient address
	recipientAddr := common.HexToAddress("0x742b15cf35dF7bcdFAcE36CB4E8C4cF03B06cE85")
	fmt.Printf("Recipient address: %s\n", recipientAddr.Hex())

	// 3. Create MetaTx
	fmt.Println("\n3. Creating MetaTx for ERC2771Forwarder...")

	amount := big.NewInt(1000000000000000000) // 1 token (18 decimals)
	nonce := uint64(1)

	metaTx := eip2771toolkit.NewMetaTxWithDelay(
		userAddr,      // from
		recipientAddr, // to
		tokenAddr,     // token contract
		amount,        // amount
		100000,        // gas limit
		nonce,         // nonce
		3600,          // deadline in 1 hour
	)

	fmt.Printf("MetaTx for ERC2771Forwarder:\n")
	fmt.Printf("  From: %s\n", metaTx.From.Hex())
	fmt.Printf("  To (recipient): %s\n", metaTx.To.Hex())
	fmt.Printf("  Token: %s\n", metaTx.Token.Hex())
	fmt.Printf("  Amount: %s\n", metaTx.Amount.String())
	fmt.Printf("  Nonce: %d\n", metaTx.Nonce)
	fmt.Printf("  Deadline: %d\n", metaTx.Deadline)

	// 4. Build domain separator for ERC2771Forwarder
	fmt.Println("\n4. Building ERC2771Forwarder domain separator...")

	chainId := big.NewInt(1) // Ethereum mainnet
	domainSeparator, err := eip2771toolkit.CreateDomainSeparatorForChain(chainId, forwarderAddr)
	if err != nil {
		log.Fatalf("Failed to build domain separator: %v", err)
	}
	fmt.Printf("Domain separator: %x\n", domainSeparator)
	fmt.Printf("Domain uses: name='ERC2771Forwarder', version='1'\n")

	// 5. Sign MetaTx with ERC2771Forwarder structure
	fmt.Println("\n5. Signing MetaTx for ERC2771Forwarder...")

	signature, err := eip2771toolkit.SignMetaTx(metaTx, userPrivKey, domainSeparator)
	if err != nil {
		log.Fatalf("Failed to sign MetaTx: %v", err)
	}
	fmt.Printf("Signature created:\n")
	fmt.Printf("  V: %d\n", signature.V)
	fmt.Printf("  R: %x\n", signature.R)
	fmt.Printf("  S: %x\n", signature.S)
	fmt.Printf("  Signature bytes: %x\n", signature.ToBytes())

	// 6. Verify signature
	fmt.Println("\n6. Verifying signature...")

	isValid, err := eip2771toolkit.VerifyMetaTxSignature(metaTx, signature, domainSeparator)
	if err != nil {
		log.Fatalf("Failed to verify signature: %v", err)
	}
	fmt.Printf("Signature is valid: %t\n", isValid)

	// 7. Show ForwardRequest structure that would be sent to ERC2771Forwarder
	fmt.Println("\n7. ERC2771Forwarder ForwardRequestData structure:")
	fmt.Printf("ForwardRequestData {\n")
	fmt.Printf("  from: %s\n", metaTx.From.Hex())
	fmt.Printf("  to: %s  // Target contract (token)\n", metaTx.Token.Hex())
	fmt.Printf("  value: 0  // No ETH for ERC20 transfer\n")
	fmt.Printf("  gas: 100000  // Gas limit for inner call\n")
	fmt.Printf("  deadline: %d  // uint48 deadline\n", metaTx.Deadline)
	fmt.Printf("  data: transfer(%s, %s)  // ERC20 transfer call\n", metaTx.To.Hex(), metaTx.Amount.String())
	fmt.Printf("  signature: %x\n", signature.ToBytes())
	fmt.Printf("}\n")

	// 8. Example with real connection (commented out)
	fmt.Println("\n8. Example with real ERC2771Forwarder connection:")
	fmt.Println("// To use with a real ERC2771Forwarder contract:")
	fmt.Println("//")
	fmt.Println("// client, err := ethclient.Dial(\"https://mainnet.infura.io/v3/YOUR-PROJECT-ID\")")
	fmt.Println("// if err != nil {")
	fmt.Println("//     log.Fatal(err)")
	fmt.Println("// }")
	fmt.Println("//")
	fmt.Println("// ctx := context.Background()")
	fmt.Println("//")
	fmt.Println("// // Get current nonce from ERC2771Forwarder")
	fmt.Println("// currentNonce, err := eip2771toolkit.GetMetaTxNonce(ctx, forwarderAddr, userAddr, client)")
	fmt.Println("// if err != nil {")
	fmt.Println("//     log.Fatal(err)")
	fmt.Println("// }")
	fmt.Println("//")
	fmt.Println("// // Create MetaTx with current nonce")
	fmt.Println("// metaTx := eip2771toolkit.NewMetaTxWithDelay(userAddr, recipientAddr, tokenAddr, amount, currentNonce, 3600)")
	fmt.Println("//")
	fmt.Println("// // Get chain ID and build domain separator")
	fmt.Println("// chainId, _ := client.NetworkID(ctx)")
	fmt.Println("// domainSeparator, _ := eip2771toolkit.CreateDomainSeparatorForChain(chainId, forwarderAddr)")
	fmt.Println("//")
	fmt.Println("// // Sign and relay")
	fmt.Println("// signature, _ := eip2771toolkit.SignMetaTx(metaTx, userPrivKey, domainSeparator)")
	fmt.Println("// txHash, err := eip2771toolkit.RelayMetaTx(ctx, metaTx, signature, relayerPrivKey, forwarderAddr, client)")
	fmt.Println("// if err != nil {")
	fmt.Println("//     log.Fatal(err)")
	fmt.Println("// }")
	fmt.Println("// fmt.Printf(\"Transaction hash: %s\", txHash.Hex())")

	fmt.Println("\n9. Key differences from MinimalForwarder:")
	fmt.Println("✓ Updated to ERC2771Forwarder contract structure")
	fmt.Printf("✓ Domain name: 'ERC2771Forwarder' (was 'MinimalForwarder')\n")
	fmt.Printf("✓ Domain version: '1' (was '0.0.1')\n")
	fmt.Printf("✓ Method: nonces() instead of getNonce()\n")
	fmt.Printf("✓ Deadline field: uint48 instead of uint256\n")
	fmt.Printf("✓ Signature included in ForwardRequestData struct\n")
	fmt.Printf("✓ Updated TypeHash for ForwardRequest\n")

	fmt.Println("\nERC2771Forwarder integration example completed successfully!")
}

// DemoERC2771ForwarderStructure shows the exact structure expected by the contract
func DemoERC2771ForwarderStructure() {
	fmt.Println("\nERC2771Forwarder expects this exact structure:")
	fmt.Println("struct ForwardRequestData {")
	fmt.Println("    address from;      // Signer address")
	fmt.Println("    address to;        // Target contract")
	fmt.Println("    uint256 value;     // ETH value")
	fmt.Println("    uint256 gas;       // Gas limit")
	fmt.Println("    uint48 deadline;   // Expiration timestamp")
	fmt.Println("    bytes data;        // Call data")
	fmt.Println("    bytes signature;   // EIP-712 signature")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("TypeHash:")
	fmt.Println("'ForwardRequest(address from,address to,uint256 value,uint256 gas,uint256 nonce,uint48 deadline,bytes data)'")
}
