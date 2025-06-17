package eip2771toolkit

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// MetaTx represents a meta transaction following EIP-2771 standard
type MetaTx struct {
	From     common.Address `json:"from"`
	To       common.Address `json:"to"`
	Token    common.Address `json:"token"`
	Amount   *big.Int       `json:"amount"`
	Gas      uint64         `json:"gas"` // Gas limit for the inner transaction
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

// ToBytes converts signature to bytes representation
func (s *Signature) ToBytes() []byte {
	result := make([]byte, 65)
	copy(result[0:32], s.R[:])
	copy(result[32:64], s.S[:])
	result[64] = s.V
	return result
}

// FromBytes sets signature from bytes representation
func (s *Signature) FromBytes(data []byte) error {
	if len(data) != 65 {
		return ErrInvalidSignatureLength
	}
	copy(s.R[:], data[0:32])
	copy(s.S[:], data[32:64])
	s.V = data[64]
	return nil
}

// TotalValue calculates the total ETH value needed for the batch
func (batch BatchMetaTxRequestList) TotalValue() *big.Int {
	total := big.NewInt(0)
	// For ERC20 transfers, we don't send ETH value, so this returns 0
	// But this method is available for future extensibility
	return total
}

// Count returns the number of requests in the batch
func (batch BatchMetaTxRequestList) Count() int {
	return len(batch)
}

// TransferData creates the calldata for ERC20 transfer
func (m *MetaTx) TransferData() ([]byte, error) {
	// ERC20 transfer function signature: transfer(address,uint256)
	transferSignature := crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

	data := make([]byte, 0, 4+32+32)
	data = append(data, transferSignature...)

	// to address (32 bytes, padded)
	toBytes := make([]byte, 32)
	copy(toBytes[12:], m.To.Bytes())
	data = append(data, toBytes...)

	// amount (32 bytes)
	amountBytes := make([]byte, 32)
	m.Amount.FillBytes(amountBytes)
	data = append(data, amountBytes...)

	return data, nil
}
