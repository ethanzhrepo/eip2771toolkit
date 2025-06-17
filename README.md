# EIP-2771 Toolkit

A comprehensive Go library implementing EIP-2771 Meta Transactions for Ethereum. This toolkit provides everything needed to create, sign, and relay meta transactions, enabling gasless transactions for end users.

## Key Features

- **MetaTx Structure**: Core meta-transaction with `From`, `To`, `Token`, `Amount`, `Gas`, `Nonce`, and `Deadline` fields
- **EIP-712 Integration**: Complete domain separator construction, TypedData signing, and verification
- **ERC2771Forwarder Support**: Updated for OpenZeppelin v5.x ERC2771Forwarder contract
- **Batch Processing**: Efficient batch relay with both atomic and non-atomic execution modes
- **Gas Flexibility**: Configurable gas limits for different transaction complexities
- **Context Support**: Full context.Context integration for cancellation and timeout control
- **Utility Functions**: Comprehensive helper functions for key management, Wei conversion, and nonce handling

## Installation

```bash
go get github.com/ethanzhrepo/eip2771toolkit
```

## Core Components

### Types

```go
// MetaTx represents a meta transaction following EIP-2771 standard
type MetaTx struct {
    From     common.Address `json:"from"`
    To       common.Address `json:"to"`
    Token    common.Address `json:"token"`
    Amount   *big.Int       `json:"amount"`
    Gas      uint64         `json:"gas"`
    Nonce    uint64         `json:"nonce"`
    Deadline uint64         `json:"deadline"` // unix timestamp
}

// Signature represents an ECDSA signature
type Signature struct {
    V byte     `json:"v"`
    R [32]byte `json:"r"`
    S [32]byte `json:"s"`
}

// BatchMetaTxRequest represents a single request in a batch
type BatchMetaTxRequest struct {
    MetaTx    MetaTx    `json:"metaTx"`
    Signature Signature `json:"signature"`
}

// BatchMetaTxRequestList represents a list of batch requests
type BatchMetaTxRequestList []BatchMetaTxRequest
```

### Core Functions

#### User-side Functions

```go
// Generate EIP-712 typed data signature for MetaTx
func SignMetaTx(metaTx MetaTx, userPrivKey *ecdsa.PrivateKey, domainSeparator []byte) (Signature, error)

// Get MetaTx struct hash (digest)
func HashMetaTx(metaTx MetaTx, domainSeparator []byte) ([]byte, error)

// Verify MetaTx signature
func VerifyMetaTxSignature(metaTx MetaTx, sig Signature, domainSeparator []byte) (bool, error)
```

#### Relayer-side Functions

```go
// Single transaction relay
func RelayMetaTx(
    ctx context.Context,
    metaTx MetaTx,
    sig Signature,
    relayerPrivKey *ecdsa.PrivateKey,
    contractAddr common.Address,
    ethClient *ethclient.Client,
) (txHash common.Hash, err error)

// Batch transaction relay with refund receiver (non-atomic)
func RelayMetaTxBatch(
    ctx context.Context,
    batchRequests BatchMetaTxRequestList,
    refundReceiver common.Address,
    relayerPrivKey *ecdsa.PrivateKey,
    contractAddr common.Address,
    ethClient *ethclient.Client,
) (common.Hash, error)

// Atomic batch transaction relay (all-or-nothing)
func RelayMetaTxBatchAtomic(
    ctx context.Context,
    batchRequests BatchMetaTxRequestList,
    relayerPrivKey *ecdsa.PrivateKey,
    contractAddr common.Address,
    ethClient *ethclient.Client,
) (common.Hash, error)

// Get on-chain nonce for user
func GetMetaTxNonce(
    ctx context.Context,
    contractAddr common.Address,
    user common.Address,
    ethClient *ethclient.Client,
) (uint64, error)
```

#### Batch Utility Functions

```go
// Create batch from multiple MetaTx and corresponding private keys
func CreateBatchFromMetaTxs(metaTxs []MetaTx, userPrivKeys []*ecdsa.PrivateKey, domainSeparator []byte) (BatchMetaTxRequestList, error)

// Create batch from multiple MetaTx signed by single user
func CreateBatchFromSingleUser(metaTxs []MetaTx, userPrivKey *ecdsa.PrivateKey, domainSeparator []byte) (BatchMetaTxRequestList, error)

// Create multiple MetaTx with sequential nonces
func NewMetaTxBatch(from common.Address, recipients []common.Address, token common.Address, amounts []*big.Int, startingNonce uint64, deadline uint64) ([]MetaTx, error)

// Verify all signatures in a batch
func VerifyBatchRequests(batchRequests BatchMetaTxRequestList, domainSeparator []byte) ([]bool, error)

// Validate batch nonces are sequential
func ValidateBatchNonces(batch BatchMetaTxRequestList, expectedStartNonce uint64) error

// Validate all requests in batch are from same user
func ValidateBatchFromSameUser(batch BatchMetaTxRequestList) error
```

#### Other Utility Functions

```go
// Build EIP-712 domain separator
func BuildDomainSeparator(name, version string, chainId *big.Int, verifyingContract common.Address) ([]byte, error)

// Create domain separator for specific chain
func CreateDomainSeparatorForChain(chainId *big.Int, contractAddr common.Address) ([]byte, error)

// Generate cryptographic keys
func GeneratePrivateKey() (*ecdsa.PrivateKey, error)
func PrivateKeyFromHex(hexKey string) (*ecdsa.PrivateKey, error)
func AddressFromPrivateKey(privKey *ecdsa.PrivateKey) common.Address

// Amount conversion utilities
func ToWei(ether *big.Float) *big.Int
func FromWei(wei *big.Int) *big.Float

// Helper functions
func NewMetaTx(from, to, token common.Address, amount *big.Int, gas uint64, nonce uint64, deadline uint64) MetaTx
func NewMetaTxWithDefaultGas(from, to, token common.Address, amount *big.Int, nonce uint64, deadline uint64) MetaTx
func GenerateRandomNonce() (uint64, error)
func GetCurrentTimestamp() uint64
func ValidateDeadline(deadline uint64) error
func IsValidAddress(addr common.Address) bool
```

## Quick Start

```go
// Create MetaTx with custom gas limit
metaTx := eip2771toolkit.NewMetaTx(
    userAddr,      // from
    recipientAddr, // to  
    tokenAddr,     // token contract
    amount,        // amount to transfer
    150000,        // gas limit (customizable)
    nonce,         // nonce
    deadline,      // deadline timestamp
)

// Or use default gas limit (100,000)
metaTx := eip2771toolkit.NewMetaTxWithDefaultGas(
    userAddr, recipientAddr, tokenAddr, amount, nonce, deadline,
)
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/eip2771toolkit"
)

func main() {
    // 1. Generate keys
    userPrivKey, _ := eip2771toolkit.GeneratePrivateKey()
    relayerPrivKey, _ := eip2771toolkit.GeneratePrivateKey()
    userAddr := eip2771toolkit.AddressFromPrivateKey(userPrivKey)

    // 2. Create MetaTx
    recipientAddr := common.HexToAddress("0x742b15cf35dF7bcdFAcE36CB4E8C4cF03B06cE85")
    tokenAddr := common.HexToAddress("0xA0b86a33E6411d01c8a96B60Dc2d7c8DB76A57E5")
    amount := big.NewInt(1000000000000000000) // 1 token
    nonce := uint64(1)
    
    metaTx := eip2771toolkit.NewMetaTxWithDefaultGas(
        userAddr, recipientAddr, tokenAddr, amount, nonce, eip2771toolkit.GetCurrentTimestamp()+3600,
    )

    // 3. Build domain separator
    chainId := big.NewInt(1) // Ethereum mainnet
    forwarderAddr := common.HexToAddress("0x123456789012345678901234567890123456789")
    domainSeparator, _ := eip2771toolkit.CreateDomainSeparatorForChain(chainId, forwarderAddr)

    // 4. Sign MetaTx
    signature, _ := eip2771toolkit.SignMetaTx(metaTx, userPrivKey, domainSeparator)

    // 5. Verify signature
    isValid, _ := eip2771toolkit.VerifyMetaTxSignature(metaTx, signature, domainSeparator)
    fmt.Printf("Signature is valid: %t\n", isValid)

    // 6. Relay transaction (requires live connection)
    // client, _ := ethclient.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
    // ctx := context.Background()
    // txHash, _ := eip2771toolkit.RelayMetaTx(ctx, metaTx, signature, relayerPrivKey, forwarderAddr, client)
    // fmt.Printf("Transaction hash: %s\n", txHash.Hex())
}
```

### Batch Relay Usage

```go
func BatchRelayExample() {
    // Setup (same as basic example)
    userPrivKey, _ := eip2771toolkit.GeneratePrivateKey()
    relayerPrivKey, _ := eip2771toolkit.GeneratePrivateKey()
    userAddr := eip2771toolkit.AddressFromPrivateKey(userPrivKey)
    
    // Contract addresses
    forwarderAddr := common.HexToAddress("0x...")
    tokenAddr := common.HexToAddress("0x...")
    
    // Create multiple recipients and amounts
    recipients := []common.Address{
        common.HexToAddress("0x742b15cf35dF7bcdFAcE36CB4E8C4cF03B06cE85"),
        common.HexToAddress("0x123456789abcdef123456789abcdef123456789a"),
    }
    amounts := []*big.Int{
        big.NewInt(1000000000000000000), // 1 token
        big.NewInt(2000000000000000000), // 2 tokens
    }
    
    // Create batch MetaTx
    metaTxs, _ := eip2771toolkit.NewMetaTxBatch(
        userAddr, recipients, tokenAddr, amounts, 10, // starting nonce: 10
        eip2771toolkit.GetCurrentTimestamp()+3600, // deadline: 1 hour
    )
    
    // Build domain separator and create batch
    chainId := big.NewInt(1)
    domainSeparator, _ := eip2771toolkit.CreateDomainSeparatorForChain(chainId, forwarderAddr)
    batchRequests, _ := eip2771toolkit.CreateBatchFromSingleUser(metaTxs, userPrivKey, domainSeparator)
    
    // Verify batch
    verificationResults, _ := eip2771toolkit.VerifyBatchRequests(batchRequests, domainSeparator)
    fmt.Printf("All signatures valid: %v\n", verificationResults)
    
    // Relay batch (requires live connection)
    // client, _ := ethclient.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
    // ctx := context.Background()
    
    // Non-atomic batch (with refund receiver)
    // refundReceiver := common.HexToAddress("0x...")
    // txHash, _ := eip2771toolkit.RelayMetaTxBatch(ctx, batchRequests, refundReceiver, relayerPrivKey, forwarderAddr, client)
    
    // Atomic batch (all-or-nothing)
    // txHash, _ := eip2771toolkit.RelayMetaTxBatchAtomic(ctx, batchRequests, relayerPrivKey, forwarderAddr, client)
}
```

## Batch Execution Strategies

### 1. Non-Atomic Batch (with refund receiver)
- Failed requests are skipped
- Unused ETH value is refunded to `refundReceiver`
- More robust against individual transaction failures
- Use when some failures are acceptable

### 2. Atomic Batch (no refund receiver)
- All requests must succeed or entire batch reverts
- No partial execution
- More gas efficient on success
- Use when all-or-nothing behavior is required

### 3. Gas Optimization
- **Optimal batch size**: 5-20 requests for best gas efficiency
- **Sequential nonces**: Reduces storage writes and gas costs
- **Same target contract**: Enables better gas estimation
- **Atomic vs non-atomic**: Atomic saves gas on success, non-atomic more robust

## Examples

The toolkit includes comprehensive examples:

- **Basic Usage**: Run `cd examples/basic && go run main.go`
- **Batch Processing**: Run `cd examples/batch && go run main.go`  
- **ERC2771Forwarder**: Run `cd examples/erc2771 && go run main.go`

Each example demonstrates different aspects of the toolkit with detailed output and explanations.

## References

- [EIP-2771: Secure Protocol for Native Meta Transactions](https://eips.ethereum.org/EIPS/eip-2771)
- [EIP-712: Typed structured data hashing and signing](https://eips.ethereum.org/EIPS/eip-712)
- [OpenZeppelin ERC2771Forwarder](https://docs.openzeppelin.com/contracts/5.x/api/metatx#ERC2771Forwarder)

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please ensure:

1. No mock or simulation code
2. Real implementations only
3. Comprehensive testing
4. Clear documentation
5. Follow Go best practices

## Support

For questions or issues, please open an issue on the GitHub repository.

### Batch Operations

```go
// Create batch MetaTx with custom gas
metaTxs, err := eip2771toolkit.NewMetaTxBatch(
    userAddr, recipients, tokenAddr, amounts, 
    100000,        // gas limit per transaction
    startingNonce, deadline,
)

// Create batch requests with context support
ctx := context.Background()
batchRequests, err := eip2771toolkit.CreateBatchFromSingleUser(
    ctx, metaTxs, userPrivKey, domainSeparator,
)

// Verify batch signatures with context
verificationResults, err := eip2771toolkit.VerifyBatchRequests(
    ctx, batchRequests, domainSeparator,
)
```
