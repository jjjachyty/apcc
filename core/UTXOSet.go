package core

import (
	"apcc_demo/wallet"
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"

	"apcchis.com/apcc/common"

	"apcchis.com/apcc/db"
)

// 1.重置数据库表
func ResetUTXOSet() {
	logrus.Debugln("重置UXTO中....")

	var err error
	defer db.Close()
	//fmt.Println("resetUTXO:\n")
	//utxoSet.Blockchain.Printchain()
	if err = db.DeleteBucket(common.UTXOTableName); err == nil {
		if err = db.CreateBucket(common.UTXOTableName); err == nil {
			//找到链上所有UTXO并存入数据库
			txOutputsMap := FindUTXOMap()
			fmt.Println("txOutputsMap=", txOutputsMap)
			ds := make([]db.KV, 0)
			for keyHash, outs := range txOutputsMap {

				ds = append(ds, db.KV{Key: keyHash, Value: outs.Serialize()})
			}
			if err = db.UpdateBatch(common.UTXOTableName, ds); err == nil {
				logrus.Debugln("重置UXTO成功....", err)
				return
			}
		}
	}
	logrus.Errorln("重置UXTO失败")
}

// 2.查询某个地址的UTXO
func FindUTXOsForAddress(address string) (utxos map[string]*UTXO) {
	utxos = make(map[string]*UTXO)
	db.GetBucket(common.UTXOTableName, func(key, value []byte) {
		TXOutputs := DeserializeTXOutputs(value)

		for _, UTXO := range TXOutputs.UTXOS {

			if UTXO.Output.UnLockScriptPubKeyWithAddress(address) {
				fmt.Printf("地址%s TXHASH=%x  COIn= %d\n", address, UTXO.TxHash, UTXO.Output.Coin)
				utxos[hex.EncodeToString(UTXO.TxHash)] = UTXO
			}
		}
	})

	return utxos
}

// 查找未花费的UTXO[string]*TXOutputs 返回字典  键为所属交易的哈希，值为TXOutput数组
func FindUTXOMap() map[string]*TXOutputs {

	if NewestBlock != nil {
		key := NewestBlock.Hash
		// 存储已花费的UTXO的信息
		spentableUTXOsMap := make(map[string][]*TXInput)

		utxoMaps := make(map[string]*TXOutputs)

		for {
			logrus.Debugf("FindUTXOMap key= %x \n", key)
			if blockByts := db.Get(common.BlockTableName, key); len(blockByts) > 0 {
				block := DeSerializeBlock(blockByts)
				key = block.PrevBlockHash

				for _, tx := range block.Txs {

					//所有花费的input
					if !tx.IsCoinbaseTransaction() { //非创世交易
						for _, TXInput := range tx.Vins {

							TxHash := hex.EncodeToString(TXInput.TxHash)
							spentableUTXOsMap[TxHash] = append(spentableUTXOsMap[TxHash], TXInput)
						}
					}
				}
				for _, tx := range block.Txs {
					TXOutputs := &TXOutputs{[]*UTXO{}}
					TxHash := hex.EncodeToString(tx.TxHash)

				WorkOutLoop:
					//所有未花费的output
					for index, out := range tx.Vouts {

						txInputs := spentableUTXOsMap[TxHash]

						if len(txInputs) > 0 {
							isUnSpent := true

							for _, in := range txInputs {

								outPublicKey := out.Ripemd160Hash
								inPublicKey := in.PublicKey

								if bytes.Compare(outPublicKey, common.Ripemd160Hash(inPublicKey)) == 0 {
									if index == in.Vout {
										isUnSpent = false
										continue WorkOutLoop
									}
								}

							}

							if isUnSpent {

								UTXO := &UTXO{tx.TxHash, index, out, block.Timestamp}
								TXOutputs.UTXOS = append(TXOutputs.UTXOS, UTXO)
							}

						} else {
							UTXO := &UTXO{tx.TxHash, index, out, block.Timestamp}
							TXOutputs.UTXOS = append(TXOutputs.UTXOS, UTXO)
						}

					}
					if len(TXOutputs.UTXOS) > 0 {
						utxoMaps[TxHash] = TXOutputs
					}

				}

				//找到创世区块推出循环
				if block.IsGenesisBlock() {
					goto END
				}
			}

		}
	END:
		// for hash, vals := range utxoMaps {
		// 	logrus.Debugf("UTXO:hash【%s】txhash【%v】coin【%v】\n", hash, vals.UTXOS[0].TxHash, vals.UTXOS[0].Output.Coin)

		// }
		return utxoMaps
	}
	return nil
}

// 多比交易需要计算当前地址是否有多的h'h'h
func FindSpendableUTXOs(address string, amount int64, txs []*Transaction) (int64, int64, map[string][]int) {

	spentTXOutputs, unPackageUTXOS := FindUnPackageSpendableUTXOS(address, txs)

	spentableUTXO := make(map[string][]int)

	var value int64
	var frozen int64
	for _, UTXO := range unPackageUTXOS {

		value += UTXO.Output.Coin
		TxHash := hex.EncodeToString(UTXO.TxHash)
		spentableUTXO[TxHash] = append(spentableUTXO[TxHash], UTXO.Index)

		if value >= amount {

			return 0, value, spentableUTXO
		}
	}

	// 钱还不够

	logrus.Debugf("当前交易中余额【%d】不足【%d】,从UXTO集中查找余额\n", value, amount)

	var userableAmount, frozenAmount int64
	db.GetBucket(common.UTXOTableName, func(key, v []byte) {
		txOutputs := DeserializeTXOutputs(v)
		for _, utxo := range txOutputs.UTXOS {
			if len(spentTXOutputs[hex.EncodeToString(utxo.TxHash)]) == 0 && utxo.Output.UnLockScriptPubKeyWithAddress(address) {
				logrus.Debugf("UXTO 找到未花费的交易 TXHASH %x COIN %d \n\n", utxo.TxHash, utxo.Output.Coin)

				userableAmount, frozenAmount = utxo.GetBlacnce()

				value += userableAmount
				frozen += frozenAmount
				TxHash := hex.EncodeToString(utxo.TxHash)
				spentableUTXO[TxHash] = append(spentableUTXO[TxHash], utxo.Index)

				if value >= amount {
					break
				}
			}
		}
	})

	if value < amount {
		logrus.Errorf("可用余额[%d]不足转出[%d],待释放金额[%d]", value, amount, frozenAmount)
	}
	return value, frozen, spentableUTXO
}

// 返回要凑多少钱
func FindUnPackageSpendableUTXOS(address string, txs []*Transaction) (map[string][]int, []*UTXO) {

	var unUTXOs []*UTXO
	spentTXOutputs := make(map[string][]int)
	for _, tx := range txs {

		if tx.IsCoinbaseTransaction() == false {

			for _, in := range tx.Vins {
				logrus.Debugf("校验输入【%x】是否是【%t】当前地址【%s】", in.TxHash, in.UnlockWithAddress(address), address)
				//是否能够解锁
				if in.UnlockWithAddress(address) {

					key := hex.EncodeToString(in.TxHash)
					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
	}

	for _, tx := range txs {

	Work:
		for index, out := range tx.Vouts {

			if out.UnLockScriptPubKeyWithAddress(address) {

				if len(spentTXOutputs) != 0 {

					for hash, indexArray := range spentTXOutputs {

						TxHashStr := hex.EncodeToString(tx.TxHash)

						if hash == TxHashStr {

							var isUnSpent = true
							for _, outIndex := range indexArray {

								if index == outIndex {

									isUnSpent = false
									continue Work
								}

								if isUnSpent {

									utxo := &UTXO{tx.TxHash, index, out, 0}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {

							utxo := &UTXO{tx.TxHash, index, out, 0}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				} else {

					utxo := &UTXO{tx.TxHash, index, out, 0}
					unUTXOs = append(unUTXOs, utxo)
				}
			}
		}
	}

	return spentTXOutputs, unUTXOs
}

//更新UTXO
func UpdateUTXO() {

	// 未花费的UTXO  键为对应交易哈希，值为TXOutput数组
	outsMap := make(map[string]*TXOutputs)
	// 新区快的交易输入,这些交易输入引用的TXOutput被消耗，应该从UTXOSet删除
	ins := []*TXInput{}

	for _, tx := range NewestBlock.Txs {
		for _, in := range tx.Vins {
			fmt.Printf("UpdateUTXO   txHASH %x , inHASH=%x\n", tx.TxHash, in.TxHash)
		}
	}

	// 2.遍历区块交易找出交易输入
	for _, tx := range NewestBlock.Txs {

		//遍历交易输入，
		for i, in := range tx.Vins {
			fmt.Printf("交易输入 i HASH=%x \n inHASH=%x \n", tx.Vins[i].TxHash, in.TxHash)
			ins = append(ins, tx.Vins[i])
		}
	}

	// 2.遍历交易输出
	for _, tx := range NewestBlock.Txs {

		utxos := []*UTXO{}
		fmt.Println("遍历输出", hex.EncodeToString(tx.TxHash), tx.Vouts[0].Coin)
		for index, out := range tx.Vouts {

			//未花费标志
			isUnSpent := true
			for _, in := range ins {
				fmt.Printf("in.Vout=%d tx.TxHash=%x, in.TxHash=%x,COIN =%d \n", in.Vout, tx.TxHash, in.TxHash, out.Coin)
				if in.Vout == index && bytes.Compare(tx.TxHash, in.TxHash) == 0 &&
					bytes.Compare(out.Ripemd160Hash, wallet.Ripemd160Hash(in.PublicKey)) == 0 {
					fmt.Println("----------------------已花费")
					isUnSpent = false
					continue
				}
			}

			if isUnSpent {
				fmt.Printf("该笔未花费 TXHASH=%x COIN=%d\n", tx.TxHash, out.Coin)
				utxo := &UTXO{tx.TxHash, index, out, NewestBlock.Timestamp}
				utxos = append(utxos, utxo)
			}
		}

		if len(utxos) > 0 {

			TxHash := hex.EncodeToString(tx.TxHash)
			outsMap[TxHash] = &TXOutputs{utxos}
		}
	}
	logrus.Debugln("最新区块信息准备完成")
	defer db.Close()
	//3. 删除已消耗的TXOutput
	err := db.GetDB().Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(common.UTXOTableName))
		if b != nil {

			for _, in := range ins {

				if txOutputsBytes := b.Get(in.TxHash); len(txOutputsBytes) == 0 {

					//如果该交易输入无引用的交易哈希

					continue
				} else {
					txOutputs := DeserializeTXOutputs(txOutputsBytes)
					fmt.Printf("原来UXTO里面的交易HASH=%x 被花费", txOutputs.UTXOS[0].TxHash)
					// 判断是否需要
					isNeedDelete := false

					//缓存来自该交易还未花费的UTXO
					utxos := []*UTXO{}

					for _, utxo := range txOutputs.UTXOS {

						if in.Vout == utxo.Index && bytes.Compare(utxo.Output.Ripemd160Hash, wallet.Ripemd160Hash(in.PublicKey)) == 0 {

							fmt.Printf("需要删除这笔交易 交易HASH=%x COIN=%d被花费\n", utxo.TxHash, utxo.Output.Coin)
							isNeedDelete = true
						} else {
							fmt.Printf("需要新增这笔交易 交易HASH=%x COIN=%d被花费\n", utxo.TxHash, utxo.Output.Coin)

							//txOutputs中剩余未花费的txOutput
							utxos = append(utxos, utxo)
						}
					}
					if isNeedDelete {
						fmt.Printf("需要删除UXTO里面的交易HASH=%x\n", in.TxHash)
						b.Delete(in.TxHash)
						fmt.Println("---------------------------------------------")
						for keyHash, outPuts := range outsMap {
							fmt.Println("新增交易到UTXO")
							fmt.Printf("keyHash= %s  coin=%d \n", keyHash, outPuts.UTXOS[0].Output.Coin)

						}
						fmt.Println("---------------------------------------------")

						if len(utxos) > 0 {

							preTXOutputs := outsMap[hex.EncodeToString(in.TxHash)]

							if preTXOutputs == nil { //之前的交易有OUTPUT 未花费
								preTXOutputs = &TXOutputs{[]*UTXO{}}
							}
							preTXOutputs.UTXOS = append(preTXOutputs.UTXOS, utxos...)
							outsMap[hex.EncodeToString(in.TxHash)] = preTXOutputs
						}
					}
				}

			}
			// 4.新增交易输出到UTXOSet
			for keyHash, outPuts := range outsMap {
				fmt.Println("新增交易到UTXO")
				fmt.Printf("keyHash= %s  coin=%d \n", keyHash, outPuts.UTXOS[0].Output.Coin)
				keyHashBytes, _ := hex.DecodeString(keyHash)
				b.Put(keyHashBytes, outPuts.Serialize())
			}
		}

		return nil
	})
	logrus.Debugln("更新UXTO成功")
	if err != nil {

		log.Panic(err)
	}
}

func FindUXTOByTxHash(txHash []byte) *TXOutputs {
	if utxoBytes := db.Get(common.UTXOTableName, txHash); len(utxoBytes) > 0 {
		return DeserializeTXOutputs(utxoBytes)

	}
	return nil
}
