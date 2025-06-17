package eip2771toolkit

import "errors"

var (
	// ErrInvalidSignatureLength is returned when signature length is not 65 bytes
	ErrInvalidSignatureLength = errors.New("invalid signature length, expected 65 bytes")

	// ErrInvalidSignature is returned when signature verification fails
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrExpiredDeadline is returned when the deadline has passed
	ErrExpiredDeadline = errors.New("deadline has expired")

	// ErrInvalidNonce is returned when nonce is invalid
	ErrInvalidNonce = errors.New("invalid nonce")

	// ErrZeroAddress is returned when address is zero
	ErrZeroAddress = errors.New("address cannot be zero")

	// ErrInvalidAmount is returned when amount is invalid
	ErrInvalidAmount = errors.New("invalid amount")

	// ErrContractCallFailed is returned when contract call fails
	ErrContractCallFailed = errors.New("contract call failed")
)
