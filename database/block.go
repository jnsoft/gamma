package database

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/jnsoft/gamma/util/misc"
	"github.com/jnsoft/gamma/util/security"
)

const BlockReward = 100

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
	Parent Hash    `json:"parent"` // parent block reference
	Number uint64  `json:"number"`
	Nonce  uint32  `json:"nonce"`
	Time   uint64  `json:"time"`
	Miner  Address `json:"miner"`
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

func NewBlock(parent Hash, number uint64, miner Address, txs []SignedTx) Block {
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
