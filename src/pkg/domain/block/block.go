package block

import (
	"encoding/json"

	"github.com/jnsoft/gamma/src/pkg/crypto"
	"github.com/jnsoft/gamma/src/pkg/domain/signedtx"
)

type Block struct {
	Header BlockHeader
	TXs    []signedtx.SignedTx // new transactions only (payload)
}

type BlockHeader struct {
	Parent [32]byte // parent block reference (hash)
	Number uint64   `json:"number"`
	Nonce  uint32   `json:"nonce"`
	Time   uint64
}

func NewBlock(parentHash []byte, time uint64, txs []signedtx.SignedTx) Block {
	return Block{
		Header: BlockHeader{
			Parent: parentHash,
			Time:   time,
		},
		TXs: txs,
	}
}

func (b Block) Hash() ([]byte, error) {
	bJson, err := b.Encode()
	if err != nil {
		return nil, err
	}

	return crypto.Sha3_256(bJson), nil
}

func (b Block) Encode() ([]byte, error) {
	return json.Marshal(b)
}
