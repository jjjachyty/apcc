package rpc

import (
	"net"

	msg "apcchis.com/apcc/rpc/msg"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const port = ":3333"

func Run() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.Fatal("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	msg.RegisterWalletServer(s, &Wallet{})
	s.Serve(lis)
}
