package handler

import (
	"github.com/sirupsen/logrus"

	"apcchis.com/apcc/core"
	"apcchis.com/apcc/wallet"
)

type Transfer struct {
	From       string
	ChangeType int
	To         string
	Amount     int64
	Free       int64
	Frozen     bool
}

//Send 钱包转账
func Send(transfer []Transfer) {

	if !wallet.IsWalletExists() {
		logrus.Fatalln("未找到钱包apcc.wallet,请先创建钱包")
	}

	// 遍历每一笔转账构造交易
	txs := []*core.Transaction{}
	for _, trs := range transfer {

		// if 1 == trs.ChangeType {
		// 	changeAddr = wt.GetNewAddress(trs.ChangeType).Value
		// 	fmt.Println("新增零钱地址:", changeAddr)
		// }
		if tx, err := core.NewTransaction(trs.From, trs.ChangeType, trs.To, trs.Amount, trs.Free, trs.Frozen, txs); err == nil {
			logrus.Debugf("新建交易成功 form=%s to=%s amount=%d", trs.From, trs.To, trs.Amount)

			txs = append(txs, tx)
			core.MemTxPool = append(core.MemTxPool, *tx)
		}

	}

}

//余额查询
func GetBlance() map[string][2]int64 {
	wt := wallet.GetHDWallet()
	blance := make(map[string][2]int64)
	for i, external := range wt.ExternalAddress {
		var amount, frozenAmount int64

		UTXOS := core.FindUTXOsForAddress(external.Value)

		for _, utxo := range UTXOS {
			amountTmp, frozenAmountTmp := utxo.GetBlacnce()
			amount += amountTmp
			frozenAmount += frozenAmountTmp
		}
		logrus.Debugf("地址 %d :%s 可用余额 %0.8f APCC 待释放金额%f  APCC \n", i+1, external.Value, float64(amount)/100000000, float64(frozenAmount)/100000000)
		blance[external.Value] = [2]int64{amount, frozenAmount}
	}

	for i, internal := range wt.InternalAddress {
		var amount, frozenAmount int64
		UTXOS := core.FindUTXOsForAddress(internal.Value)

		for _, utxo := range UTXOS {
			amountTmp, frozenAmountTmp := utxo.GetBlacnce()
			amount += amountTmp
			frozenAmount += frozenAmountTmp
		}
		logrus.Debugf("零钱 %d :%s 可用余额 %0.8f APCC 待释放金额%0.8f  APCC \n", i+1, internal.Value, float64(amount)/100000000, float64(frozenAmount)/100000000)
		blance[internal.Value] = [2]int64{amount, frozenAmount}
	}

	return blance
}
