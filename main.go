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
	"github.com/yarbelk/grpcstuff/protostuff"
	"google.golang.org/grpc"
)

type ProtoStuffs struct {
	LastAction string
	Clock      time.Time
	log        []*protostuff.PlayerEventLog
	protostuff.UnimplementedProtoStuffServer
}

func (ps *ProtoStuffs) StreamEventLog(in *protostuff.Player, s protostuff.ProtoStuff_StreamEventLogServer) error {
	for i := 0; true; i++ {
		<-time.After(1 * time.Second)
		s.Send(&protostuff.PlayerEventLog{
			SequenceId: int64(i),
			Timestamp:  &protostuff.VectorTimestamp{Timestamps: []int64{time.Now().Unix()}},
			Action:     &protostuff.Action{Action: fmt.Sprintf("Action %d", i)},
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

	protostuff.RegisterProtoStuffServer(grpcServer, &ProtoStuffs{log: make([]*protostuff.PlayerEventLog, 0)})

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

	err = protostuff.RegisterProtoStuffHandler(context.Background(), gwmux, conn)
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
