package eip2771toolkit

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// EIP712_DOMAIN_TYPEHASH is the EIP-712 domain separator typehash
	EIP712_DOMAIN_TYPEHASH = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"

	// FORWARD_REQUEST_TYPEHASH is the ForwardRequest struct typehash for ERC2771Forwarder
	FORWARD_REQUEST_TYPEHASH = "ForwardRequest(address from,address to,uint256 value,uint256 gas,uint256 nonce,uint48 deadline,bytes data)"
)

// BuildDomainSeparator creates EIP-712 domain separator
func BuildDomainSeparator(name, version string, chainId *big.Int, verifyingContract common.Address) ([]byte, error) {
	// Calculate domain typehash
	domainTypeHash := crypto.Keccak256([]byte(EIP712_DOMAIN_TYPEHASH))

	// Calculate name hash
	nameHash := crypto.Keccak256([]byte(name))

	// Calculate version hash
	versionHash := crypto.Keccak256([]byte(version))

	// Convert chainId to 32 bytes
	chainIdBytes := make([]byte, 32)
	chainId.FillBytes(chainIdBytes)

	// Concatenate all parts
	data := make([]byte, 0, 32*5)
	data = append(data, domainTypeHash...)
	data = append(data, nameHash...)
	data = append(data, versionHash...)
	data = append(data, chainIdBytes...)
	data = append(data, verifyingContract.Bytes()...)

	// Hash the concatenated data
	domainSeparator := crypto.Keccak256(data)
	return domainSeparator, nil
}

// HashMetaTx generates the EIP-712 digest for a MetaTx (compatible with ERC2771Forwarder)
func HashMetaTx(metaTx MetaTx, domainSeparator []byte) ([]byte, error) {
	// Calculate struct typehash
	structTypeHash := crypto.Keccak256([]byte(FORWARD_REQUEST_TYPEHASH))

	// Prepare ERC20 transfer data
	transferData, err := metaTx.TransferData()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare transfer data: %w", err)
	}

	// Encode ForwardRequest struct according to new ERC2771Forwarder format
	// ForwardRequest(address from,address to,uint256 value,uint256 gas,uint256 nonce,uint48 deadline,bytes data)
	structData := make([]byte, 0, 32*7)
	structData = append(structData, structTypeHash...)
	structData = append(structData, metaTx.From.Bytes()...)
	structData = append(structData, metaTx.Token.Bytes()...) // 'to' field points to token contract

	// Value is 0 for ERC20 transfers
	valueBytes := make([]byte, 32)
	structData = append(structData, valueBytes...)

	// Gas limit from MetaTx.Gas field
	gasBytes := make([]byte, 32)
	new(big.Int).SetUint64(metaTx.Gas).FillBytes(gasBytes)
	structData = append(structData, gasBytes...)

	// Convert nonce to 32 bytes
	nonceBytes := make([]byte, 32)
	new(big.Int).SetUint64(metaTx.Nonce).FillBytes(nonceBytes)
	structData = append(structData, nonceBytes...)

	// Convert deadline to 32 bytes (uint48 but encoded as uint256 in hash)
	deadlineBytes := make([]byte, 32)
	new(big.Int).SetUint64(metaTx.Deadline).FillBytes(deadlineBytes)
	structData = append(structData, deadlineBytes...)

	// Hash of the data field
	dataHash := crypto.Keccak256(transferData)
	structData = append(structData, dataHash...)

	// Hash the struct data
	structHash := crypto.Keccak256(structData)

	// Create EIP-712 digest: "\x19\x01" || domainSeparator || structHash
	digest := make([]byte, 0, 2+32+32)
	digest = append(digest, 0x19, 0x01)
	digest = append(digest, domainSeparator...)
	digest = append(digest, structHash...)

	// Final hash
	finalHash := crypto.Keccak256(digest)
	return finalHash, nil
}

// SignMetaTx signs a MetaTx using EIP-712
func SignMetaTx(metaTx MetaTx, userPrivKey *ecdsa.PrivateKey, domainSeparator []byte) (Signature, error) {
	var sig Signature

	// Get the hash to sign
	hash, err := HashMetaTx(metaTx, domainSeparator)
	if err != nil {
		return sig, fmt.Errorf("failed to hash MetaTx: %w", err)
	}

	// Sign the hash
	sigBytes, err := crypto.Sign(hash, userPrivKey)
	if err != nil {
		return sig, fmt.Errorf("failed to sign hash: %w", err)
	}

	// Convert to our Signature format
	err = sig.FromBytes(sigBytes)
	if err != nil {
		return sig, fmt.Errorf("failed to parse signature: %w", err)
	}

	return sig, nil
}

// VerifyMetaTxSignature verifies a MetaTx signature
func VerifyMetaTxSignature(metaTx MetaTx, sig Signature, domainSeparator []byte) (bool, error) {
	// Get the hash that was signed
	hash, err := HashMetaTx(metaTx, domainSeparator)
	if err != nil {
		return false, fmt.Errorf("failed to hash MetaTx: %w", err)
	}

	// Convert signature to bytes
	sigBytes := sig.ToBytes()

	// Recover public key from signature
	recoveredPubKey, err := crypto.SigToPub(hash, sigBytes)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get the address from recovered public key
	recoveredAddr := crypto.PubkeyToAddress(*recoveredPubKey)

	// Check if recovered address matches the from address
	return recoveredAddr == metaTx.From, nil
}
