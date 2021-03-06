package blockchain

import (
	"github.com/sirupsen/logrus"
	. "yu/common"
	"yu/storage/kv"
)

// the Key Name of last finalized blockID
var LastFinalizedKey = []byte("Last-Finalized-BlockID")

type BlockChain struct {
	kvdb kv.KV
}

func NewBlockChain(kvCfg *kv.KVconf) *BlockChain {
	kvdb, err := kv.NewKV(kvCfg)
	if err != nil {
		logrus.Panicln("cannot load kvdb")
	}
	return &BlockChain{
		kvdb: kvdb,
	}
}

func (bc *BlockChain) AppendBlock(ib IBlock) error {
	var b *Block = ib.(*Block)
	blockId := b.BlockId().Bytes()
	blockByt, err := b.Encode()
	if err != nil {
		return err
	}
	return bc.kvdb.Set(blockId, blockByt)
}

func (bc *BlockChain) GetBlock(id BlockId) (IBlock, error) {
	blockByt, err := bc.kvdb.Get(id.Bytes())
	if err != nil {
		return nil, err
	}
	return DecodeBlock(blockByt)
}

func (bc *BlockChain) Children(prevId BlockId) ([]IBlock, error) {
	prevBlockNum, prevHash := prevId.Separate()
	blockNum := prevBlockNum + 1
	iter, err := bc.kvdb.Iter(blockNum.Bytes())
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var blocks []IBlock
	for iter.Valid() {
		_, blockByt, err := iter.Entry()
		if err != nil {
			return nil, err
		}
		block, err := DecodeBlock(blockByt)
		if err != nil {
			return nil, err
		}
		if block.PrevHash() == prevHash {
			blocks = append(blocks, block)
		}

		err = iter.Next()
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

func (bc *BlockChain) Finalize(id BlockId) error {
	return bc.kvdb.Set(LastFinalizedKey, id.Bytes())
}

func (bc *BlockChain) LastFinalized() (IBlock, error) {
	lfBlockIdByt, err := bc.kvdb.Get(LastFinalizedKey)
	if err != nil {
		return nil, err
	}
	blockByt, err := bc.kvdb.Get(lfBlockIdByt)
	if err != nil {
		return nil, err
	}
	return DecodeBlock(blockByt)
}

func (bc *BlockChain) Leaves() ([]IBlock, error) {
	iter, err := bc.kvdb.Iter(BlockNum(0).Bytes())
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var blocks []IBlock
	for iter.Valid() {
		_, blockByt, err := iter.Entry()
		if err != nil {
			return nil, err
		}
		block, err := DecodeBlock(blockByt)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}
