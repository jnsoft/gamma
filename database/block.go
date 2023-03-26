package database

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jnsoft/gamma/util/misc"
	"github.com/jnsoft/gamma/util/security"
)

// Lengths of hashes and addresses in bytes.
const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the address
	AddressLength = 20

	BlockReward = 100
)

type Hash [HashLength]byte
type Address [AddressLength]byte

func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(h.Hex()), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

func (h Hash) IsEmpty() bool {
	emptyHash := Hash{}

	return bytes.Equal(emptyHash[:], h[:])
}

type SimpleBlock struct {
	Parent Hash       `json:"parent"` // parent block reference
	Time   uint64     `json:"time"`
	TXs    []SimpleTx `json:"payload"` // new transactions only (payload)
}

type Block struct {
	Header BlockHeader `json:"header"`  // metadata (parent block hash + time)
	TXs    []SignedTx  `json:"payload"` // new transactions only (payload)
}

type BlockHeader struct {
	Parent Hash           `json:"parent"` // parent block reference
	Number uint64         `json:"number"`
	Nonce  uint32         `json:"nonce"`
	Time   uint64         `json:"time"`
	Miner  common.Address `json:"miner"`
}

type BlockFS struct {
	Key   Hash  `json:"hash"`
	Value Block `json:"block"`
}

type SimpleBlockFS struct {
	Key   Hash        `json:"hash"`
	Value SimpleBlock `json:"block"`
}

func NewSimpleBlock(parent Hash, txs []SimpleTx) SimpleBlock {
	return SimpleBlock{parent, misc.GetTime(), txs}
}

func NewBlock(parent Hash, number uint64, miner common.Address, txs []SignedTx) Block {
	return Block{BlockHeader{parent, number, security.GenerateNonce(), misc.GetTime(), miner}, txs}
}

func (b Block) Hash() (Hash, error) {
	blockJson, err := json.Marshal(b)
	if err != nil {
		return Hash{}, err
	}

	return sha256.Sum256(blockJson), nil
}

func (b SimpleBlock) Hash() (Hash, error) {
	blockJson, err := json.Marshal(b)
	if err != nil {
		return Hash{}, err
	}

	return sha256.Sum256(blockJson), nil
}

func (b Block) GasReward() uint {
	reward := uint(0)

	for _, tx := range b.TXs {
		reward += tx.GasCost()
	}

	return reward
}

func IsBlockHashValid(hash Hash, miningDifficulty uint) bool {
	zeroesCount := uint(0)

	for i := uint(0); i < miningDifficulty; i++ {
		if fmt.Sprintf("%x", hash[i]) == "0" {
			zeroesCount++
		}
	}

	if fmt.Sprintf("%x", hash[miningDifficulty]) == "0" {
		return false
	}

	return zeroesCount == miningDifficulty
}
