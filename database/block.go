package database

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/jnsoft/gamma/common"
	"github.com/jnsoft/gamma/util/misc"
	"github.com/jnsoft/gamma/util/security"
)

const BlockReward = 100

type SimpleBlock struct {
	Header BlockHeader `json:"header"`
	TXs    []SimpleTx  `json:"payload"` // new transactions only (payload)
}

type Block struct {
	Header BlockHeader `json:"header"`  // metadata (parent block hash + time)
	TXs    []SignedTx  `json:"payload"` // new transactions only (payload)
}

type BlockHeader struct {
	Parent common.Hash    `json:"parent"` // parent block reference
	Number uint64         `json:"number"`
	Nonce  uint32         `json:"nonce"`
	Time   uint64         `json:"time"`
	Miner  common.Address `json:"miner"`
}

type BlockFS struct {
	Key   common.Hash `json:"hash"`
	Value Block       `json:"block"`
}

type SimpleBlockFS struct {
	Key   common.Hash `json:"hash"`
	Value SimpleBlock `json:"block"`
}

func NewSimpleBlock(parent common.Hash, number uint64, miner common.Address, txs []SimpleTx) SimpleBlock {
	return SimpleBlock{BlockHeader{parent, number, security.GenerateNonce(), misc.GetTime(), miner}, txs}
}

func NewBlock(parent common.Hash, number uint64, miner common.Address, txs []SignedTx) Block {
	return Block{BlockHeader{parent, number, security.GenerateNonce(), misc.GetTime(), miner}, txs}
}

func (b Block) Hash() (common.Hash, error) {
	blockJson, err := json.Marshal(b)
	if err != nil {
		return common.Hash{}, err
	}

	return sha256.Sum256(blockJson), nil
}

func (b SimpleBlock) Hash() (common.Hash, error) {
	blockJson, err := json.Marshal(b)
	if err != nil {
		return common.Hash{}, err
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

func (b SimpleBlock) GasReward() uint {
	reward := uint(0)

	for _, tx := range b.TXs {
		reward += tx.GasCost()
	}

	return reward
}

func IsBlockHashValid(hash common.Hash, miningDifficulty uint) bool {
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
