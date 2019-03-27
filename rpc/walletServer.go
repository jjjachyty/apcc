package rpc

import (
	"context"
	"fmt"

	"apcchis.com/apcc/common"

	"apcchis.com/apcc/handler"

	pb "apcchis.com/apcc/rpc/msg"
)

type Wallet struct {
}

func (*Wallet) GetBlance(ctx context.Context, in *pb.Request) (*pb.BlanceResponse, error) {
	var resp = make([]*pb.AddressBlance, 0)
	common.WalletPassWd = in.Passwd
	myBlance := handler.GetBlance()

	fmt.Println("server", in.Passwd)
	if in.Address != "" {

		resp = append(resp, &pb.AddressBlance{Address: in.Address, UserableAmount: myBlance[in.Address][0], FrzoneAmount: myBlance[in.Address][1]})
	} else {
		for k := range myBlance {
			resp = append(resp, &pb.AddressBlance{Address: k, UserableAmount: myBlance[k][0], FrzoneAmount: myBlance[k][1]})

		}
	}
	return &pb.BlanceResponse{Reply: "1", Blances: resp}, nil
}
