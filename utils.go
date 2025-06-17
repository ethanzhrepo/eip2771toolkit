package eip2771toolkit

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// GeneratePrivateKey generates a new ECDSA private key
func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateKey()
}

// PrivateKeyFromHex creates a private key from hex string
func PrivateKeyFromHex(hexKey string) (*ecdsa.PrivateKey, error) {
	return crypto.HexToECDSA(hexKey)
}

// AddressFromPrivateKey derives the Ethereum address from a private key
func AddressFromPrivateKey(privKey *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(privKey.PublicKey)
}

// NewMetaTx creates a new MetaTx with the given parameters
func NewMetaTx(from, to, token common.Address, amount *big.Int, gas, nonce uint64, deadline uint64) MetaTx {
	return MetaTx{
		From:     from,
		To:       to,
		Token:    token,
		Amount:   amount,
		Gas:      gas,
		Nonce:    nonce,
		Deadline: deadline,
	}
}

// NewMetaTxWithDelay creates a new MetaTx with deadline set to current time + delay
func NewMetaTxWithDelay(from, to, token common.Address, amount *big.Int, gas, nonce uint64, delaySeconds uint64) MetaTx {
	deadline := uint64(time.Now().Unix()) + delaySeconds
	return NewMetaTx(from, to, token, amount, gas, nonce, deadline)
}

// NewMetaTxWithDefaultGas creates a new MetaTx with default gas limit of 100000
func NewMetaTxWithDefaultGas(from, to, token common.Address, amount *big.Int, nonce uint64, deadline uint64) MetaTx {
	return NewMetaTx(from, to, token, amount, 100000, nonce, deadline)
}

// IsValidAddress checks if the given address is valid (not zero address)
func IsValidAddress(addr common.Address) bool {
	return addr != (common.Address{})
}

// ToWei converts ether amount to wei
func ToWei(ether *big.Float) *big.Int {
	wei := new(big.Float)
	wei.Mul(ether, big.NewFloat(1e18))

	result := new(big.Int)
	wei.Int(result)
	return result
}

// FromWei converts wei to ether
func FromWei(wei *big.Int) *big.Float {
	ether := new(big.Float)
	ether.SetInt(wei)
	ether.Quo(ether, big.NewFloat(1e18))
	return ether
}

// GenerateRandomNonce generates a cryptographically secure random nonce
func GenerateRandomNonce() (uint64, error) {
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(64), nil).Sub(max, big.NewInt(1))

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, fmt.Errorf("failed to generate random nonce: %w", err)
	}

	return n.Uint64(), nil
}

// ValidateDeadline checks if the deadline is valid (not expired)
func ValidateDeadline(deadline uint64) error {
	if uint64(time.Now().Unix()) > deadline {
		return ErrExpiredDeadline
	}
	return nil
}

// GetCurrentTimestamp returns the current Unix timestamp
func GetCurrentTimestamp() uint64 {
	return uint64(time.Now().Unix())
}

// CreateDomainSeparatorForChain creates a domain separator for a specific chain using ERC2771Forwarder
func CreateDomainSeparatorForChain(chainId *big.Int, contractAddr common.Address) ([]byte, error) {
	return BuildDomainSeparator("ERC2771Forwarder", "1", chainId, contractAddr)
}

// CreateBatchFromMetaTxs creates a BatchMetaTxRequestList from MetaTx slice and user private keys
func CreateBatchFromMetaTxs(ctx context.Context, metaTxs []MetaTx, userPrivKeys []*ecdsa.PrivateKey, domainSeparator []byte) (BatchMetaTxRequestList, error) {
	if len(metaTxs) != len(userPrivKeys) {
		return nil, fmt.Errorf("metaTxs and userPrivKeys length mismatch: %d vs %d", len(metaTxs), len(userPrivKeys))
	}

	batch := make(BatchMetaTxRequestList, len(metaTxs))

	for i, metaTx := range metaTxs {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		batchReq, err := CreateBatchRequest(metaTx, userPrivKeys[i], domainSeparator)
		if err != nil {
			return nil, fmt.Errorf("failed to create batch request at index %d: %w", i, err)
		}
		batch[i] = batchReq
	}

	return batch, nil
}

// CreateBatchFromSingleUser creates a BatchMetaTxRequestList where all MetaTxs are signed by the same user
func CreateBatchFromSingleUser(ctx context.Context, metaTxs []MetaTx, userPrivKey *ecdsa.PrivateKey, domainSeparator []byte) (BatchMetaTxRequestList, error) {
	batch := make(BatchMetaTxRequestList, len(metaTxs))

	for i, metaTx := range metaTxs {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		batchReq, err := CreateBatchRequest(metaTx, userPrivKey, domainSeparator)
		if err != nil {
			return nil, fmt.Errorf("failed to create batch request at index %d: %w", i, err)
		}
		batch[i] = batchReq
	}

	return batch, nil
}

// NewMetaTxBatch creates multiple MetaTx with sequential nonces
func NewMetaTxBatch(
	from common.Address,
	recipients []common.Address,
	token common.Address,
	amounts []*big.Int,
	gas uint64,
	startingNonce uint64,
	deadline uint64,
) ([]MetaTx, error) {
	if len(recipients) != len(amounts) {
		return nil, fmt.Errorf("recipients and amounts length mismatch: %d vs %d", len(recipients), len(amounts))
	}

	metaTxs := make([]MetaTx, len(recipients))

	for i := range recipients {
		metaTxs[i] = NewMetaTx(
			from,
			recipients[i],
			token,
			amounts[i],
			gas,
			startingNonce+uint64(i),
			deadline,
		)
	}

	return metaTxs, nil
}

// NewMetaTxBatchWithDefaultGas creates multiple MetaTx with sequential nonces and default gas limit
func NewMetaTxBatchWithDefaultGas(
	from common.Address,
	recipients []common.Address,
	token common.Address,
	amounts []*big.Int,
	startingNonce uint64,
	deadline uint64,
) ([]MetaTx, error) {
	return NewMetaTxBatch(from, recipients, token, amounts, 100000, startingNonce, deadline)
}

// ValidateBatchNonces checks if all nonces in the batch are sequential and starting from expected nonce
func ValidateBatchNonces(batch BatchMetaTxRequestList, expectedStartNonce uint64) error {
	for i, req := range batch {
		expectedNonce := expectedStartNonce + uint64(i)
		if req.MetaTx.Nonce != expectedNonce {
			return fmt.Errorf("invalid nonce at index %d: expected %d, got %d", i, expectedNonce, req.MetaTx.Nonce)
		}
	}
	return nil
}

// ValidateBatchFromSameUser checks if all requests in the batch are from the same user
func ValidateBatchFromSameUser(batch BatchMetaTxRequestList) error {
	if len(batch) == 0 {
		return nil
	}

	expectedFrom := batch[0].MetaTx.From
	for i, req := range batch {
		if req.MetaTx.From != expectedFrom {
			return fmt.Errorf("request at index %d has different from address: expected %s, got %s",
				i, expectedFrom.Hex(), req.MetaTx.From.Hex())
		}
	}
	return nil
}
