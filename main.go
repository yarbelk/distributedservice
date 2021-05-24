package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash"
	"github.com/hashicorp/memberlist"
	"github.com/yarbelk/grpcstuff/data"
	"github.com/yarbelk/grpcstuff/proto"
	"github.com/yarbelk/grpcstuff/service"
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

var (
	name              = flag.String("name", "", "name for node. must be unique")
	address           = flag.String("address", "0.0.0.0:8080", "address to bind server too")
	clusterAddr       = flag.String("cluster-addr", "", "address to bind server too")
	clusterPort       = flag.Int("cluster-port", 9999, "port to bind server too")
	config            = flag.String("cfg", "local", "default config type from memberlist")
	partitions        = flag.Int("partitions", 1051, "chose a big enough prime for balancing")
	replicationFactor = flag.Int("rep-factor", 3, "how many replications")

	dataStorageDir = flag.String("data", "customer_data/", "which directory to store the event data in")
)

type hasher struct{}

// Sum64 on type to conform to expectations of consistent library
func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func main() {
	flag.Parse()
	var cfg *memberlist.Config
	switch *config {
	case "local":
		cfg = memberlist.DefaultLocalConfig()
	case "lan":
		cfg = memberlist.DefaultLANConfig()
	}
	if *name != "" {
		cfg.Name = *name
	}
	members, err := memberlist.Create(cfg)
	if err != nil {
		panic(err.Error())
	}

	joinList := flag.Args()

	_, err = members.Join(joinList)

	if err != nil {
		panic(err)
	}

	ch := consistent.New(nil, consistent.Config{
		Hasher:            hasher{},
		ReplicationFactor: *replicationFactor,
		Load:              1.25,
		PartitionCount:    *partitions,
	})

	for _, node := range members.Members() {
		ch.Add(service.WrappedNode{node})
	}

	// makeing some huge assumptions here about readyness of the memberlist.
	// i'm also not at all acounting for rebalancing the nodes; but this library
	// supports that stuff

	cs := service.Customer{
		Storage:    data.New(*dataStorageDir),
		MemberList: members,
		HashList:   ch,
	}

	lis, err := net.Listen("tcp", *address)
	if err != nil {
		panic(err.Error())
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	proto.RegisterProtoStuffServer(grpcServer, &cs)

	log.Fatalln(grpcServer.Serve(lis))
}
