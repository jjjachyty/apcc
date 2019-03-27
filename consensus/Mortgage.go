package consensus

import (
	"time"
)

type Mortgage struct {
	//来自交易的哈希
	TxHash []byte
	//在该交易VOuts里的下标
	Index int
	//未花费的交易输出
	// Output    *core.TXOutput
	Timestamp time.Time //抵押时间
}
