package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/yarbelk/grpcstuff/proto"
	"google.golang.org/grpc"
)

type ProtoStuffs struct {
	LastAction string
	Clock      time.Time
	log        []*proto.CustomerEventLog
	proto.UnimplementedProtoStuffServer
}

func (ps *ProtoStuffs) StreamEventLog(in *proto.Customer, s proto.ProtoStuff_StreamEventLogServer) error {
	for i := 0; true; i++ {
		<-time.After(1 * time.Second)
		s.Send(&proto.CustomerEventLog{
			SequenceId: int64(i),
			Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{time.Now().Unix()}},
			Action:     &proto.Action{Action: fmt.Sprintf("Action %d", i)},
		})
		i++
	}
	return nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		panic(err.Error())
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	proto.RegisterProtoStuffServer(grpcServer, &ProtoStuffs{log: make([]*proto.CustomerEventLog, 0)})

	go func() {
		log.Fatalln(grpcServer.Serve(lis))
	}()

	log.Println("Connecting to the grpcServer")
	conn, err := grpc.DialContext(
		context.Background(),
		"localhost:8081",
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()

	err = proto.RegisterProtoStuffHandler(context.Background(), gwmux, conn)
	if err != nil {
		panic(err)
	}

	gwServer := &http.Server{
		Addr:    "localhost:8080",
		Handler: gwmux,
	}
	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8080")
	log.Fatalln(gwServer.ListenAndServe())
}
