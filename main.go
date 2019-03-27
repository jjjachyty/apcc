package main

import (
	"time"

	"apcchis.com/apcc/handler"
	"apcchis.com/apcc/miner"

	"apcchis.com/apcc/core"

	"apcchis.com/apcc/common"

	"apcchis.com/apcc/wallet"
)

func main() {
	common.WalletPassWd = "1234567890123456"

	// wt := wallet.Create("1234567890123456")
	wallet.GetHDWallet()
	core.CreateBlockchainWithGensisBlock()

	// wt.GetNewAddress(0)
	// wt.GetNewAddress(0)
	// wt := wallet.GetHDWallet("1234567890123456")
	// pk := wt.GetPrivKey("1Dt3UmL9z65vchGVUyXtsDaPKgp9SS7yv9")
	// r, s, err := ecdsa.Sign(rand.Reader, pk, []byte("1472089438dfa546b3d2642e536494a8ac49f9317ca58be7a465c1ff3652cb72567aed9c21b7a83a9e4ba2e0d469e8ec152606fc14f87a44c32e609b5ddf3a82"))
	// if err != nil {
	// 	log.Panic(err)
	// }
	// //一个ECDSA签名就是一对数字，我们对这对数字连接起来就是signature
	// signature := append(r.Bytes(), s.Bytes()...)

	// r1 := big.Int{}
	// s1 := big.Int{}
	// sigLen := len(signature)
	// r1.SetBytes(signature[:(sigLen / 2)])
	// s1.SetBytes(signature[(sigLen / 2):])

	// fmt.Println(ecdsa.Verify(&pk.PublicKey, []byte("1472089438dfa546b3d2642e536494a8ac49f9317ca58be7a465c1ff3652cb72567aed9c21b7a83a9e4ba2e0d469e8ec152606fc14f87a44c32e609b5ddf3a82"), &r1, &s1))

	go miner.Mining()
	// core.CreateBlockchainWithGensisBlock()
	handler.Printchain()
	// core.ResetUTXOSet()
	// // // //wallet.Create("1234567890123456")
	// // // wt := wallet.GetHDWallet("1234567890123456")

	// cli.GetBlance()
	// cli.Printchain()
	time.Sleep(2 * time.Second)
	send := []handler.Transfer{
		handler.Transfer{From: "16UebB4eAnyAbDH9oi5ohAup5iNBPS86h5", ChangeType: 1, Frozen: true, To: "1KHsMwK9t6XYYG7CRyTCGgCtVVxS7xEfvL",
			Amount: 100000000},
		// cli.Transfer{From: "1KHsMwK9t6XYYG7CRyTCGgCtVVxS7xEfvL", ChangeType: 0, To: "1Hq7QE4iQGiaScko51sS2mHe4SzESfbenV",
		// 	Amount: 100000000},
		// cli.Transfer{From: "1EULVuR2AXjemKXT4zYvz6CDXqyRHc4TFg", ChangeType: 1, Frozen: true, To: "1Kcbj8SgJzM5T1qwRssweYsrfmG3MPPcHy",
		// 	Amount: 200},
		// 	// wallet.Transfer{From: "1P1xnPW2pESUkdeUEJnPN9aNyi4zZWRSNz", ChangeType: 1, To: "1KUnhEJzBkMaJZx2sVkkgXrEBQPit2SjvM",
		// 	// 	Amount: 200000000},
	}
	handler.Send(send)

	// wallet.Send("1234567890123456", send)
	// core.Printchain()
	// cli.GetBlance("1234567890123456")

	// p1, _ := new(big.Int).SetString("112132100000000231111278979797979123123132132119999999991111111313218789797797877896654654564549999999956", 0)
	// fmt.Println(p1.Int64())
	// p2 := new(big.Int).SetBytes([]byte("9853393445385562019"))
	// fmt.Println(p2.String())
	// test, _ := new(big.Int).SetString("1644649884951448510757860025008616070635679144429003516725940070333041275844581436357181190635160199511120249279753197239322510161963657079928275058084249929669100343336674746635197995034150147581416555241935848223332304133533731327457442941290036190164843131910476891455893925894672332911", 0)
	// ripemd160 := ripemd160.New()
	// ripemd160.Write(test.Bytes())
	// fmt.Println(string(common.Base58Encode(ripemd160.Sum(nil))))

	// version_publicKey_checksumBytes := common.Base58Decode([]byte("14XXMBpAAhQLFYCYz4xmjmTH3oVmX"))
	// decode := new(big.Int).SetBytes(version_publicKey_checksumBytes)
	time.Sleep(time.Second * 50)
}
