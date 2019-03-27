package main

import (
	"apcchis.com/apcc/server"
)

func main() {

	// common.WalletPassWd = "1234567890123456"
	// wallet.GetHDWallet()
	// key, _ := hex.DecodeString("0ee2573c8c7cc5374e2d198255c9af5d223b4fe226e868ca1cf23b0a16d053f9")

	// handler.Printchain()
	// core.ResetUTXOSet()
	// // cli.GetBlance()
	// // core.UpdateUTXO()
	// handler.GetBlance()

	// block := core.DeSerializeBlock(valu)
	// fmt.Println(block.Height)

	// for key, vl := range db.GetBucket(common.BlockTableName) {
	// 	fmt.Println(key, core.DeSerializeBlock(vl).Height)
	// }
	server.Start()

}
