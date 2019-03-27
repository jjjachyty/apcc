package wallet

import (
	"crypto/ecdsa"

	"apcchis.com/apcc/common"
)

func GetHDWallet() *HDWallet {
	wt := &HDWallet{Password: common.WalletPassWd}
	wt.OpenWallet()
	return wt
}

func (wt *HDWallet) GetNewAddress(changeType int) (addr Address) {
	if wt != nil {
		switch changeType {
		case 0:
			//ExternalAddress
			extermalNum := len(wt.ExternalAddress)
			lastAddress := wt.ExternalAddress[extermalNum-1]
			if extermalNum >= 20 {
				addr = wt.CreateAddress(lastAddress.Account+1, 0, 0)
			} else {
				addr = wt.CreateAddress(lastAddress.Account, 0, lastAddress.AddrIndex+1)
			}
		case 1:

			internalNum := len(wt.InternalAddress)
			if internalNum == 0 {
				addr = wt.CreateAddress(0, 1, 0)
				return
			}
			lastAddress := wt.InternalAddress[internalNum-1]
			if internalNum >= 20 {
				addr = wt.CreateAddress(lastAddress.Account+1, 1, 0)
			} else {
				addr = wt.CreateAddress(lastAddress.Account, 1, lastAddress.AddrIndex+1)
			}
		}

	}
	return
}

func (wt *HDWallet) GetAddress(addrstr string, changeType int) (addr Address) {
	if changeType == 0 {
		for _, addr := range wt.ExternalAddress {
			if addr.Value == addrstr {
				return addr
			}
		}
	}
	for _, addr := range wt.InternalAddress {
		if addr.Value == addrstr {
			return addr
		}
	}

	return
}

//GetPrivKey根据地址获取私钥
func (wt *HDWallet) GetPrivKey(address string) *ecdsa.PrivateKey {
	for _, addr := range wt.ExternalAddress {
		if addr.Value == address {
			privKey, _ := common.ToECDSA(addr.PrivKey, true)

			return privKey
		}
	}

	for _, addr := range wt.InternalAddress {
		if addr.Value == address {
			privKey, _ := common.ToECDSA(addr.PrivKey, true)
			return privKey

		}
	}
	return nil
}
