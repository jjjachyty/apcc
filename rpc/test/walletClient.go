package main

import (
	"context"
	"fmt"

	pb "apcchis.com/apcc/rpc/msg"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:3333", grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	// Creates a new CustomerClient
	client := pb.NewWalletClient(conn)
	blacnce, _ := client.GetBlance(context.Background(), &pb.Request{Passwd: "1234567890123456"})
	fmt.Println("blacnce", blacnce)
	fmt.Println(blacnce.Blances[0].Address, blacnce.Blances[0].UserableAmount, blacnce.Blances[0].FrzoneAmount)
}
