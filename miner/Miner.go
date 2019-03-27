package miner

import (
	"bytes"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"

	"apcchis.com/apcc/db"
	"apcchis.com/apcc/handler"

	"apcchis.com/apcc/common"

	"apcchis.com/apcc/core"
)

// 挖矿需要满足的最小交易数
const minMinerTxCount = 1

// 矿工地址
// var miningAddress string

//开始挖矿
func Mining() {

	// miningAddress = wt.GetNewAddress(1).Value

	for {

		var txs []*core.Transaction

		for indx, trx := range core.MemTxPool {

			if !trx.VerifyTransaction(txs) {
				logrus.Errorf("交易验证未通过%x \n", trx.TxHash)

				continue
			}
			logrus.Infof("交易验证通过,%d \n", trx.Vouts[0].Coin)
			for _, in := range trx.Vins {
				logrus.Infof("交易IN HASH,%x \n", in.TxHash)

			}
			txs = append(txs, &core.MemTxPool[indx])

		}

		if len(txs) > 0 {

			addr, flag := CheckMiner()
			if !flag {
				logrus.Fatalf("挖矿抵押币不足,至少需要%d APCC \n", common.MiningCost)
				return
			}

			logrus.Debugf("发现有可打包的交易信息,开始打包出块%x", txs[0].Vins[0].PublicKey)
			txs = GetTXFree(txs, addr[0])

			block := core.NewBlock(txs, core.NewestBlock.Height+1, core.NewestBlock.Hash)
			logrus.Debugf("交易费手续费 HASH=%x\n\n\n\n\n", block.Txs[len(block.Txs)-1].TxHash)
			if db.UpdateFunc(func(tx *bolt.Tx) (err error) {
				var b *bolt.Bucket
				if b, err = tx.CreateBucketIfNotExists([]byte(common.BlockTableName)); err == nil {

					if err = b.Put(block.Hash, block.Serialize()); err == nil {
						if err = b.Put([]byte(common.NewestBlockKey), block.Hash); err == nil {
							return
						}
						return
					}
					logrus.Errorln("新增区块错误", err)
				}
				return
			}) == nil {
				// 	//更新成功
				core.NewestBlock = block
				core.UpdateUTXO()
				// for _, tx := range core.NewestBlock.Txs {
				// 	for _, in := range tx.Vins {
				// 		fmt.Printf("txHASH %x , inHASH=%x\n", tx.TxHash, in.TxHash)
				// 	}
				// }
				// 去除内存池中打包到区块的交易
				for _, tx := range txs {
					for i, menTx := range core.MemTxPool {
						if bytes.Compare(tx.TxHash, menTx.TxHash) == 0 {
							if i == len(core.MemTxPool) {
								core.MemTxPool = append(core.MemTxPool[:i], core.MemTxPool[i:]...)
							} else {
								core.MemTxPool = append(core.MemTxPool[:i], core.MemTxPool[i+1:]...)
							}
						}
					}

				}

				txs = []*core.Transaction{}
				logrus.Debugf("新增区块高度%d,开始全网广播", core.NewestBlock.Height)

				//广播区块

			}

		}

		time.Sleep(time.Second * 1)

		fmt.Printf("\r挖矿中.....(交易数量%d)", len(core.MemTxPool))

	}
}

func CheckMiner() (adds []string, flag bool) {
	var blacnce int64
	for k, v := range handler.GetBlance() {
		if v[0] > 0 {

			blacnce += v[0]
			adds = append(adds, k)
			if blacnce >= common.MiningCost {
				break
			}
		}
	}

	return adds, blacnce >= common.MiningCost
}

//获取交易手续费
func GetTXFree(txs []*core.Transaction, address string) []*core.Transaction {
	var inputValue, outputValue int64

	for _, tx := range txs {
		for _, vin := range tx.Vins {
			outPuts := core.FindUXTOByTxHash(vin.TxHash)
			for _, utxo := range outPuts.UTXOS {
				userable, _ := utxo.GetBlacnce()
				inputValue += userable
			}
		}
		for _, outPut := range tx.Vouts {

			outputValue += outPut.Coin

		}
	}
	tx := core.NewCoinbaseTransaction(address, inputValue-outputValue)
	fmt.Println("\n\n TX HASH = \n\n", tx.TxHash)
	return append(txs, tx)
}
