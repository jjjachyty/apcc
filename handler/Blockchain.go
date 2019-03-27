package handler

import (
	"apcc_demo/utils"
	"fmt"
	"log"
	"time"

	"apcchis.com/apcc/core"
	"github.com/boltdb/bolt"

	"apcchis.com/apcc/common"
	"apcchis.com/apcc/db"
)

//NewestHash var 最新区块的Hash

func Printchain() {

	if core.NewestBlock != nil {

		defer db.Close()
		fmt.Println("common.NewestBlockHashKey", core.NewestBlock.Hash)
		var hashKey = core.NewestBlock.Hash
		for {
			currentBlockBytes := db.Get(common.BlockTableName, hashKey)
			if currentBlockBytes != nil {

				// 获取到当前迭代器里面的currentHash所对应的区块
				block := core.DeSerializeBlock(currentBlockBytes)

				fmt.Println("------------------------------")
				fmt.Printf("Height：%d\n", block.Height)
				fmt.Printf("PrevBlockHash：%x\n", block.PrevBlockHash)
				fmt.Printf("Timestamp：%s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
				fmt.Printf("Hash：%x\n", block.Hash)
				fmt.Printf("Nonce:%x\n", block.Nonce)
				fmt.Println("Txs:")
				for i, tx := range block.Txs {
					fmt.Println("------------------------------")
					fmt.Println("交易", i)
					fmt.Printf("%x\n", tx.TxHash)
					fmt.Println("Vins:")
					for _, in := range tx.Vins {
						fmt.Printf("TxHash:%x\n", in.TxHash)
						fmt.Printf("Vout:%d\n", in.Vout)
						fmt.Printf("Signature:%x\n\n", in.Signature)
						fmt.Printf("PublicKey:%x\n\n", in.PublicKey)
					}

					fmt.Println("Vouts:")
					for _, out := range tx.Vouts {
						fmt.Printf("Value:%d\n", out.Coin)
						fmt.Printf("Frozen:%t\n", out.Frozen)
						fmt.Printf("Ripemd160Hash:%x\n\n", out.Ripemd160Hash)
					}

				}

				fmt.Println("------------------------------")
				// 更新迭代器里面CurrentHash
				hashKey = block.PrevBlockHash
				if block.IsGenesisBlock() {
					break
				}
			} else {
				break
			}

		}
	}
}

// 获取区块所有哈希
func GetBlockHashes() [][]byte {

	key := core.NewestBlock.Hash

	var blockHashs [][]byte

	// var key1 []byte
	db.BucketIterator(common.BlockTableName, func(b *bolt.Bucket) {

		for {
			if val := b.Get(key); len(val) > 0 {

				block := core.DeSerializeBlock(val)
				blockHashs = append(blockHashs, block.Hash)
				key = block.PrevBlockHash

				if block.IsGenesisBlock() {
					break
				}
				continue
			}
			break
		}
	})
	return blockHashs
}

// 获取对应哈希的区块
func GetBlock(bHash []byte) []byte {

	return db.Get(common.BlockTableName, bHash)
}

// 将同步请求的主链区块添加到区块链

func AddBlock(block *core.Block) error {

	var err error
	db.UpdateFunc(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(common.BlockTableName))
		if b != nil {

			blockExist := b.Get(block.Hash)
			if blockExist != nil {
				// 如果存在，不需要做任何过多的处理
				return nil
			}

			err := b.Put(block.Hash, block.Serialize())
			if err != nil {

				log.Panic(err)
			}

			// 最新的区块链的Hash
			blockHash := b.Get([]byte(common.NewestBlockKey))
			blockInDB := core.DeSerializeBlock(b.Get(blockHash))

			if blockInDB.Height < block.Height {
				b.Put([]byte(utils.NewestBlockKey), block.Hash)
				core.NewestBlock = block
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return err
}
