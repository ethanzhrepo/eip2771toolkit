package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethanzhrepo/eip2771toolkit"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	fmt.Println("EIP-2771 Toolkit - Batch Relay Example")
	fmt.Println("======================================")

	// This example demonstrates batch relay functionality using executeBatch

	// 1. Setup accounts
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

	// 2. Contract addresses
	fmt.Println("\n2. Contract addresses...")

	forwarderAddr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	tokenAddr := common.HexToAddress("0xA0b86a33E6411d01c8a96B60Dc2d7c8DB76A57E5")
	fmt.Printf("ERC2771Forwarder: %s\n", forwarderAddr.Hex())
	fmt.Printf("Token contract: %s\n", tokenAddr.Hex())

	// 3. Create multiple recipients and amounts for batch transfer
	fmt.Println("\n3. Creating batch transfer data...")

	recipients := []common.Address{
		common.HexToAddress("0x742b15cf35dF7bcdFAcE36CB4E8C4cF03B06cE85"),
		common.HexToAddress("0x123456789abcdef123456789abcdef123456789a"),
		common.HexToAddress("0x987654321fedcba987654321fedcba987654321f"),
	}

	amounts := []*big.Int{
		big.NewInt(1000000000000000000), // 1 token
		big.NewInt(2000000000000000000), // 2 tokens
		big.NewInt(500000000000000000),  // 0.5 token
	}

	startingNonce := uint64(10)
	deadline := eip2771toolkit.GetCurrentTimestamp() + 3600 // 1 hour from now

	fmt.Printf("Recipients: %d\n", len(recipients))
	fmt.Printf("Starting nonce: %d\n", startingNonce)
	fmt.Printf("Deadline: %d\n", deadline)

	// 4. Create batch MetaTx
	fmt.Println("\n4. Creating batch MetaTx...")

	metaTxs, err := eip2771toolkit.NewMetaTxBatchWithDefaultGas(
		userAddr,
		recipients,
		tokenAddr,
		amounts,
		startingNonce,
		deadline,
	)
	if err != nil {
		log.Fatalf("Failed to create MetaTx batch: %v", err)
	}

	fmt.Printf("Created %d MetaTx:\n", len(metaTxs))
	for i, metaTx := range metaTxs {
		fmt.Printf("  [%d] To: %s, Amount: %s, Gas: %d, Nonce: %d\n",
			i, metaTx.To.Hex(), metaTx.Amount.String(), metaTx.Gas, metaTx.Nonce)
	}

	// 5. Build domain separator
	fmt.Println("\n5. Building domain separator...")

	chainId := big.NewInt(1) // Ethereum mainnet
	domainSeparator, err := eip2771toolkit.CreateDomainSeparatorForChain(chainId, forwarderAddr)
	if err != nil {
		log.Fatalf("Failed to build domain separator: %v", err)
	}
	fmt.Printf("Domain separator: %x\n", domainSeparator)

	// 6. Create batch requests (sign all MetaTx)
	fmt.Println("\n6. Creating and signing batch requests...")

	ctx := context.Background()
	batchRequests, err := eip2771toolkit.CreateBatchFromSingleUser(ctx, metaTxs, userPrivKey, domainSeparator)
	if err != nil {
		log.Fatalf("Failed to create batch requests: %v", err)
	}

	fmt.Printf("Created batch with %d requests\n", batchRequests.Count())
	fmt.Printf("Total ETH value needed: %s wei\n", batchRequests.TotalValue().String())

	// 7. Verify all signatures in the batch
	fmt.Println("\n7. Verifying batch signatures...")

	verificationResults, err := eip2771toolkit.VerifyBatchRequests(ctx, batchRequests, domainSeparator)
	if err != nil {
		log.Fatalf("Failed to verify batch requests: %v", err)
	}

	allValid := true
	for i, isValid := range verificationResults {
		fmt.Printf("  [%d] Signature valid: %t\n", i, isValid)
		if !isValid {
			allValid = false
		}
	}
	fmt.Printf("All signatures valid: %t\n", allValid)

	// 8. Validate batch properties
	fmt.Println("\n8. Validating batch properties...")

	// Check nonces are sequential
	err = eip2771toolkit.ValidateBatchNonces(batchRequests, startingNonce)
	if err != nil {
		log.Fatalf("Batch nonce validation failed: %v", err)
	}
	fmt.Printf("✓ Nonces are sequential starting from %d\n", startingNonce)

	// Check all from same user
	err = eip2771toolkit.ValidateBatchFromSameUser(batchRequests)
	if err != nil {
		log.Fatalf("Batch user validation failed: %v", err)
	}
	fmt.Printf("✓ All requests from same user: %s\n", userAddr.Hex())

	// 9. Show batch structure for executeBatch
	fmt.Println("\n9. Batch structure for executeBatch:")
	fmt.Printf("ForwardRequestData[] requests = [\n")
	for i, req := range batchRequests {
		fmt.Printf("  [%d] {\n", i)
		fmt.Printf("    from: %s\n", req.MetaTx.From.Hex())
		fmt.Printf("    to: %s  // Token contract\n", req.MetaTx.Token.Hex())
		fmt.Printf("    value: 0  // No ETH for ERC20\n")
		fmt.Printf("    gas: 100000\n")
		fmt.Printf("    deadline: %d\n", req.MetaTx.Deadline)
		fmt.Printf("    data: transfer(%s, %s)\n", req.MetaTx.To.Hex(), req.MetaTx.Amount.String())
		fmt.Printf("    signature: %x\n", req.Signature.ToBytes())
		fmt.Printf("  }\n")
	}
	fmt.Printf("]\n")

	// 10. Example batch relay calls
	fmt.Println("\n10. Example batch relay calls:")

	// Example refund receiver
	refundReceiver := common.HexToAddress("0xRefundReceiver123456789abcdef123456789abc")

	fmt.Println("// Batch with refund receiver (non-atomic):")
	fmt.Println("// If some requests fail, failed requests will be skipped and value refunded")
	fmt.Println("// client, _ := ethclient.Dial(\"https://mainnet.infura.io/v3/YOUR-PROJECT-ID\")")
	fmt.Println("// ctx := context.Background()")
	fmt.Printf("// refundReceiver := common.HexToAddress(\"%s\")\n", refundReceiver.Hex())
	fmt.Println("// txHash, err := eip2771toolkit.RelayMetaTxBatch(")
	fmt.Println("//     ctx,")
	fmt.Println("//     batchRequests,")
	fmt.Println("//     refundReceiver,")
	fmt.Println("//     relayerPrivKey,")
	fmt.Println("//     forwarderAddr,")
	fmt.Println("//     client,")
	fmt.Println("// )")
	fmt.Println("//")
	fmt.Println("// Atomic batch (all-or-nothing):")
	fmt.Println("// If any request fails, entire batch reverts")
	fmt.Println("// txHash, err := eip2771toolkit.RelayMetaTxBatchAtomic(")
	fmt.Println("//     ctx,")
	fmt.Println("//     batchRequests,")
	fmt.Println("//     relayerPrivKey,")
	fmt.Println("//     forwarderAddr,")
	fmt.Println("//     client,")
	fmt.Println("// )")

	// 11. Demonstrate different batch creation methods
	fmt.Println("\n11. Different batch creation methods:")

	// Method 1: Single user, multiple recipients
	fmt.Println("✓ Single user, multiple recipients (demonstrated above)")

	// Method 2: Multiple users (different private keys)
	fmt.Println("✓ Multiple users with different private keys:")

	user2PrivKey, _ := eip2771toolkit.GeneratePrivateKey()
	user3PrivKey, _ := eip2771toolkit.GeneratePrivateKey()

	multiUserKeys := []*ecdsa.PrivateKey{userPrivKey, user2PrivKey, user3PrivKey}
	multiUserTxs := metaTxs[:3] // Use first 3 transactions

	// Update from addresses to match the different users
	multiUserTxs[1].From = eip2771toolkit.AddressFromPrivateKey(user2PrivKey)
	multiUserTxs[2].From = eip2771toolkit.AddressFromPrivateKey(user3PrivKey)

	multiUserBatch, err := eip2771toolkit.CreateBatchFromMetaTxs(ctx, multiUserTxs, multiUserKeys, domainSeparator)
	if err != nil {
		log.Fatalf("Failed to create multi-user batch: %v", err)
	}

	fmt.Printf("   Multi-user batch created with %d requests\n", multiUserBatch.Count())
	for i, req := range multiUserBatch {
		fmt.Printf("   [%d] From: %s\n", i, req.MetaTx.From.Hex())
	}

	// 12. Gas optimization considerations
	fmt.Println("\n12. Gas optimization considerations:")
	fmt.Println("✓ Batch size: 3 requests (recommended: 5-20 for optimal gas efficiency)")
	fmt.Println("✓ Sequential nonces: reduces gas cost per transaction")
	fmt.Println("✓ Same token contract: enables better gas estimation")
	fmt.Println("✓ Atomic vs non-atomic: atomic saves gas on success, non-atomic more robust")

	fmt.Println("\nBatch relay example completed successfully!")
}

// DemoBatchExecutionStrategies shows different execution strategies
func DemoBatchExecutionStrategies() {
	fmt.Println("\nBatch Execution Strategies:")
	fmt.Println("==========================")

	fmt.Println("1. Non-Atomic (with refund receiver):")
	fmt.Println("   - Failed requests are skipped")
	fmt.Println("   - Unused ETH value is refunded to refundReceiver")
	fmt.Println("   - More robust against individual transaction failures")
	fmt.Println("   - Use when some failures are acceptable")

	fmt.Println("\n2. Atomic (no refund receiver):")
	fmt.Println("   - All requests must succeed or entire batch reverts")
	fmt.Println("   - No partial execution")
	fmt.Println("   - More gas efficient on success")
	fmt.Println("   - Use when all-or-nothing behavior is required")

	fmt.Println("\n3. Gas Considerations:")
	fmt.Println("   - Batch overhead: ~21000 gas + per-request overhead")
	fmt.Println("   - Optimal batch size: 5-20 requests")
	fmt.Println("   - Sequential nonces: reduces storage writes")
	fmt.Println("   - Same target contract: enables better gas estimation")
}
