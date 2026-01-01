package block

import (
	"encoding/json"

	"github.com/jnsoft/gamma/src/pkg/common"
	"github.com/jnsoft/gamma/src/pkg/crypto"
	"github.com/jnsoft/gamma/src/pkg/domain/signedtx"
)

type Block struct {
	Header BlockHeader
	TXs    []signedtx.SignedTx // new transactions only (payload)
}

type BlockHeader struct {
	ParentHash       [32]byte // parent block reference (hash)
	StateRoot        [32]byte
	TransactionsRoot [32]byte
	ReceiptsRoot     []byte
	Number           uint64 `json:"number"`
	Nonce            uint32 `json:"nonce"`
	ExtraData        []byte
	Time             uint64
}

func NewBlock(parentHash, stateRoot [32]byte, blockNumber uint64, extraData []byte, txs []signedtx.SignedTx) (Block, error) {
	troot, err := computeTransactionsRoot(txs)
	if err != nil {
		return Block{}, err
	}
	return Block{
		Header: BlockHeader{
			ParentHash:       parentHash,
			StateRoot:        stateRoot,
			TransactionsRoot: troot,
			ReceiptsRoot:     nil,
			Number:           blockNumber,
			Nonce:            0,
			ExtraData:        extraData,
			Time:             uint64(common.NowUnixUTC()),
		},
		TXs: txs,
	}, nil
}

func (b Block) Hash() ([32]byte, error) {
	bJson, err := b.Encode()
	if err != nil {
		return [32]byte{}, err
	}

	var out [32]byte
	hash := crypto.Sha3_256(bJson)
	copy(out[:], hash)
	return out, nil
}

func (b Block) Encode() ([]byte, error) {
	return json.Marshal(b)
	/*
		hasher := sha256.New()

		hasher.Write(h.ParentHash[:])
		hasher.Write(h.StateRoot[:])
		hasher.Write(h.TransactionsRoot[:])
		hasher.Write(h.ReceiptsRoot[:])

		buf := make([]byte, 8)

		binary.BigEndian.PutUint64(buf, h.BlockNumber)
		hasher.Write(buf)

		binary.BigEndian.PutUint64(buf, h.Timestamp)
		hasher.Write(buf)

		binary.BigEndian.PutUint64(buf, h.GasLimit)
		hasher.Write(buf)

		binary.BigEndian.PutUint64(buf, h.GasUsed)
		hasher.Write(buf)

		hasher.Write(h.ExtraData)
		hasher.Write(h.Nonce[:])

		var out [32]byte
		copy(out[:], hasher.Sum(nil))
		return out
	*/
}

func computeTransactionsRoot(txs []signedtx.SignedTx) ([32]byte, error) {
	if len(txs) == 0 {
		var root [32]byte
		hash := crypto.Sha3_256(nil)
		copy(root[:], hash)
		return root, nil
	}

	var txHashes [][]byte
	for _, tx := range txs {
		h, err := tx.Hash()
		if err != nil {
			return [32]byte{}, err
		}
		txHashes = append(txHashes, h[:])
	}
	return computeMerkleRoot(txHashes)
}

func computeMerkleRoot(hashes [][]byte) ([32]byte, error) {
	if len(hashes) == 0 {
		return [32]byte{}, nil
	}
	for len(hashes) > 1 {
		var next [][]byte
		for i := 0; i < len(hashes); i += 2 {
			if i+1 == len(hashes) {
				next = append(next, hashes[i])
			} else {
				h := crypto.Sha3_256(append(hashes[i], hashes[i+1]...))
				next = append(next, h[:])
			}
		}
		hashes = next
	}
	var root [32]byte
	copy(root[:], hashes[0])
	return root, nil
}
