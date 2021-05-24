package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/yarbelk/grpcstuff/proto"
	"google.golang.org/grpc"
)

type ProtoStuffs struct {
	LastAction string
	Clock      time.Time
	log        []*proto.CustomerEventLog
	proto.UnimplementedProtoStuffServer
}

// StreamEventLog is 100% fake.  pretend its listening for new written logs using something
// like badger's DB.NewStream(); with prefix := []byte(fmt.Sprintf("%d:",in.Customer.ID))
// and applying/sending.  a bit more work than i want to do now; but this lets you send persisted
// events only in very timely manner to subscribers.  Streaming is fun.
func (ps *ProtoStuffs) StreamEventLog(in *proto.Customer, s proto.ProtoStuff_StreamEventLogServer) error {
	for i := 0; true; i++ {
		<-time.After(1 * time.Second)
		s.Send(&proto.CustomerEventLog{
			SequenceId: uint64(i),
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

	log.Fatalln(grpcServer.Serve(lis))
}
