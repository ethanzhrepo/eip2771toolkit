package eip2771toolkit

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC2771Forwarder ABI for meta transaction execution
const ERC2771ForwarderABI = `[
	{
		"inputs": [
			{
				"components": [
					{"internalType": "address", "name": "from", "type": "address"},
					{"internalType": "address", "name": "to", "type": "address"},
					{"internalType": "uint256", "name": "value", "type": "uint256"},
					{"internalType": "uint256", "name": "gas", "type": "uint256"},
					{"internalType": "uint48", "name": "deadline", "type": "uint48"},
					{"internalType": "bytes", "name": "data", "type": "bytes"},
					{"internalType": "bytes", "name": "signature", "type": "bytes"}
				],
				"internalType": "struct ERC2771Forwarder.ForwardRequestData",
				"name": "request",
				"type": "tuple"
			}
		],
		"name": "execute",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{"internalType": "address", "name": "from", "type": "address"},
					{"internalType": "address", "name": "to", "type": "address"},
					{"internalType": "uint256", "name": "value", "type": "uint256"},
					{"internalType": "uint256", "name": "gas", "type": "uint256"},
					{"internalType": "uint48", "name": "deadline", "type": "uint48"},
					{"internalType": "bytes", "name": "data", "type": "bytes"},
					{"internalType": "bytes", "name": "signature", "type": "bytes"}
				],
				"internalType": "struct ERC2771Forwarder.ForwardRequestData[]",
				"name": "requests",
				"type": "tuple[]"
			},
			{
				"internalType": "address payable",
				"name": "refundReceiver",
				"type": "address"
			}
		],
		"name": "executeBatch",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{"internalType": "address", "name": "from", "type": "address"},
					{"internalType": "address", "name": "to", "type": "address"},
					{"internalType": "uint256", "name": "value", "type": "uint256"},
					{"internalType": "uint256", "name": "gas", "type": "uint256"},
					{"internalType": "uint48", "name": "deadline", "type": "uint48"},
					{"internalType": "bytes", "name": "data", "type": "bytes"},
					{"internalType": "bytes", "name": "signature", "type": "bytes"}
				],
				"internalType": "struct ERC2771Forwarder.ForwardRequestData",
				"name": "request",
				"type": "tuple"
			}
		],
		"name": "verify",
		"outputs": [
			{"internalType": "bool", "name": "", "type": "bool"}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "address", "name": "owner", "type": "address"}
		],
		"name": "nonces",
		"outputs": [
			{"internalType": "uint256", "name": "", "type": "uint256"}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

// ERC20Transfer ABI for token transfer
const ERC20TransferABI = `[
	{
		"inputs": [
			{"internalType": "address", "name": "to", "type": "address"},
			{"internalType": "uint256", "name": "amount", "type": "uint256"}
		],
		"name": "transfer",
		"outputs": [
			{"internalType": "bool", "name": "", "type": "bool"}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

// RelayMetaTx submits a meta transaction to the blockchain through a relayer
func RelayMetaTx(
	ctx context.Context,
	metaTx MetaTx,
	sig Signature,
	relayerPrivKey *ecdsa.PrivateKey,
	contractAddr common.Address,
	ethClient *ethclient.Client,
) (common.Hash, error) {
	// Validate inputs
	if err := validateMetaTx(metaTx); err != nil {
		return common.Hash{}, fmt.Errorf("invalid MetaTx: %w", err)
	}

	// Check deadline
	if uint64(time.Now().Unix()) > metaTx.Deadline {
		return common.Hash{}, ErrExpiredDeadline
	}

	// Get relayer address
	relayerAddr := crypto.PubkeyToAddress(relayerPrivKey.PublicKey)

	// Parse ERC2771Forwarder contract ABI
	parsedABI, err := abi.JSON(strings.NewReader(ERC2771ForwarderABI))
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Prepare ERC20 transfer data
	transferData, err := metaTx.TransferData()
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to prepare transfer data: %w", err)
	}

	// Create ForwardRequestData struct for new ERC2771Forwarder
	forwardRequestData := struct {
		From      common.Address
		To        common.Address
		Value     *big.Int
		Gas       *big.Int
		Deadline  *big.Int // uint48 in contract but use uint256 for ABI encoding
		Data      []byte
		Signature []byte
	}{
		From:      metaTx.From,
		To:        metaTx.Token,                       // Target is the token contract
		Value:     big.NewInt(0),                      // No ETH value for ERC20 transfer
		Gas:       new(big.Int).SetUint64(metaTx.Gas), // Use MetaTx.Gas field
		Deadline:  new(big.Int).SetUint64(metaTx.Deadline),
		Data:      transferData,
		Signature: sig.ToBytes(),
	}

	// Pack the execute method call
	data, err := parsedABI.Pack("execute", forwardRequestData)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack execute call: %w", err)
	}

	// Get current gas price
	gasPrice, err := ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Get nonce for relayer
	nonce, err := ethClient.PendingNonceAt(ctx, relayerAddr)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get relayer nonce: %w", err)
	}

	// Estimate gas
	msg := ethereum.CallMsg{
		From:     relayerAddr,
		To:       &contractAddr,
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
		Data:     data,
	}
	gasLimit, err := ethClient.EstimateGas(ctx, msg)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(nonce, contractAddr, big.NewInt(0), gasLimit, gasPrice, data)

	// Get chain ID
	chainID, err := ethClient.NetworkID(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), relayerPrivKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash(), nil
}

// GetMetaTxNonce retrieves the current nonce for a user from the ERC2771Forwarder contract
func GetMetaTxNonce(
	ctx context.Context,
	contractAddr common.Address,
	user common.Address,
	ethClient *ethclient.Client,
) (uint64, error) {
	// Parse ERC2771Forwarder contract ABI
	parsedABI, err := abi.JSON(strings.NewReader(ERC2771ForwarderABI))
	if err != nil {
		return 0, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the nonces method call (changed from getNonce to nonces)
	data, err := parsedABI.Pack("nonces", user)
	if err != nil {
		return 0, fmt.Errorf("failed to pack nonces call: %w", err)
	}

	// Call contract
	msg := ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}
	result, err := ethClient.CallContract(ctx, msg, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call contract: %w", err)
	}

	// Unpack result
	var nonce *big.Int
	err = parsedABI.UnpackIntoInterface(&nonce, "nonces", result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack result: %w", err)
	}

	return nonce.Uint64(), nil
}

// validateMetaTx validates the MetaTx struct
func validateMetaTx(metaTx MetaTx) error {
	if metaTx.From == (common.Address{}) {
		return ErrZeroAddress
	}
	if metaTx.To == (common.Address{}) {
		return ErrZeroAddress
	}
	if metaTx.Token == (common.Address{}) {
		return ErrZeroAddress
	}
	if metaTx.Amount == nil || metaTx.Amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	if metaTx.Deadline == 0 {
		return ErrExpiredDeadline
	}
	return nil
}

// RelayMetaTxBatch submits multiple meta transactions to the blockchain through a relayer using executeBatch
func RelayMetaTxBatch(
	ctx context.Context,
	batchRequests BatchMetaTxRequestList,
	refundReceiver common.Address,
	relayerPrivKey *ecdsa.PrivateKey,
	contractAddr common.Address,
	ethClient *ethclient.Client,
) (common.Hash, error) {
	if len(batchRequests) == 0 {
		return common.Hash{}, fmt.Errorf("batch cannot be empty")
	}

	// Validate all requests in the batch
	for i, req := range batchRequests {
		if err := validateMetaTx(req.MetaTx); err != nil {
			return common.Hash{}, fmt.Errorf("invalid MetaTx at index %d: %w", i, err)
		}

		// Check deadline for each request
		if uint64(time.Now().Unix()) > req.MetaTx.Deadline {
			return common.Hash{}, fmt.Errorf("request at index %d has expired deadline", i)
		}
	}

	// Get relayer address
	relayerAddr := crypto.PubkeyToAddress(relayerPrivKey.PublicKey)

	// Parse ERC2771Forwarder contract ABI
	parsedABI, err := abi.JSON(strings.NewReader(ERC2771ForwarderABI))
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Prepare batch requests
	forwardRequestDataList, totalValue, err := prepareBatchRequests(batchRequests)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to prepare batch requests: %w", err)
	}

	// Pack the executeBatch method call
	data, err := parsedABI.Pack("executeBatch", forwardRequestDataList, refundReceiver)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack executeBatch call: %w", err)
	}

	// Get current gas price
	gasPrice, err := ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Get nonce for relayer
	nonce, err := ethClient.PendingNonceAt(ctx, relayerAddr)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get relayer nonce: %w", err)
	}

	// Estimate gas
	msg := ethereum.CallMsg{
		From:     relayerAddr,
		To:       &contractAddr,
		GasPrice: gasPrice,
		Value:    totalValue,
		Data:     data,
	}
	gasLimit, err := ethClient.EstimateGas(ctx, msg)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(nonce, contractAddr, totalValue, gasLimit, gasPrice, data)

	// Get chain ID
	chainID, err := ethClient.NetworkID(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), relayerPrivKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash(), nil
}

// RelayMetaTxBatchAtomic submits multiple meta transactions atomically (no refund receiver)
// If any request fails, the entire batch will revert
func RelayMetaTxBatchAtomic(
	ctx context.Context,
	batchRequests BatchMetaTxRequestList,
	relayerPrivKey *ecdsa.PrivateKey,
	contractAddr common.Address,
	ethClient *ethclient.Client,
) (common.Hash, error) {
	// Use zero address as refund receiver for atomic execution
	zeroAddress := common.Address{}
	return RelayMetaTxBatch(ctx, batchRequests, zeroAddress, relayerPrivKey, contractAddr, ethClient)
}

// prepareBatchRequests converts BatchMetaTxRequestList to the format expected by executeBatch
func prepareBatchRequests(batchRequests BatchMetaTxRequestList) ([]interface{}, *big.Int, error) {
	forwardRequestDataList := make([]interface{}, len(batchRequests))
	totalValue := big.NewInt(0)

	for i, req := range batchRequests {
		// Prepare ERC20 transfer data for this request
		transferData, err := req.MetaTx.TransferData()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to prepare transfer data for request %d: %w", i, err)
		}

		// Create ForwardRequestData struct
		forwardRequestData := struct {
			From      common.Address
			To        common.Address
			Value     *big.Int
			Gas       *big.Int
			Deadline  *big.Int
			Data      []byte
			Signature []byte
		}{
			From:      req.MetaTx.From,
			To:        req.MetaTx.Token,
			Value:     big.NewInt(0), // No ETH value for ERC20 transfer
			Gas:       new(big.Int).SetUint64(req.MetaTx.Gas),
			Deadline:  new(big.Int).SetUint64(req.MetaTx.Deadline),
			Data:      transferData,
			Signature: req.Signature.ToBytes(),
		}

		forwardRequestDataList[i] = forwardRequestData
		// Add to total value (for ERC20 transfers, this is always 0)
		totalValue.Add(totalValue, forwardRequestData.Value)
	}

	return forwardRequestDataList, totalValue, nil
}

// VerifyBatchRequests verifies all signatures in a batch
func VerifyBatchRequests(ctx context.Context, batchRequests BatchMetaTxRequestList, domainSeparator []byte) ([]bool, error) {
	results := make([]bool, len(batchRequests))

	for i, req := range batchRequests {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		isValid, err := VerifyMetaTxSignature(req.MetaTx, req.Signature, domainSeparator)
		if err != nil {
			return nil, fmt.Errorf("failed to verify signature for request %d: %w", i, err)
		}
		results[i] = isValid
	}

	return results, nil
}

// CreateBatchRequest creates a BatchMetaTxRequest from MetaTx and private key
func CreateBatchRequest(metaTx MetaTx, userPrivKey *ecdsa.PrivateKey, domainSeparator []byte) (BatchMetaTxRequest, error) {
	signature, err := SignMetaTx(metaTx, userPrivKey, domainSeparator)
	if err != nil {
		return BatchMetaTxRequest{}, fmt.Errorf("failed to sign MetaTx: %w", err)
	}

	return BatchMetaTxRequest{
		MetaTx:    metaTx,
		Signature: signature,
	}, nil
}
