package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/buraksezer/consistent"
	"github.com/hashicorp/memberlist"
	"github.com/yarbelk/distributedservice/data"
	"github.com/yarbelk/distributedservice/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WrappedNode struct {
	*memberlist.Node
}

func (wn WrappedNode) String() string {
	return wn.Name
}

type Customer struct {
	Storage data.Storer

	MemberList *memberlist.Memberlist

	HashList *consistent.Consistent

	proto.UnimplementedProtoStuffServer
}

// StreamEventLog is 100% fake.  pretend its listening for new written logs using something
// like badger's DB.NewStream(); with prefix := []byte(fmt.Sprintf("%d:",in.Customer.ID))
// and applying/sending.  a bit more work than i want to do now; but this lets you send persisted
// events only in very timely manner to subscribers.  Streaming is fun.
func (c *Customer) StreamEventLog(in *proto.Customer, s proto.ProtoStuff_StreamEventLogServer) error {
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

// CustomerState is a straight lookup.  Internally badger uses ristretto now (I believe; it is in
// its dependency graph)
// If this is a cache and you want to fall back, you can add a fall back lookup to a slower
// but canonical store
func (c *Customer) CustomerState(ctx context.Context, in *proto.Customer) (*proto.CustomerState, error) {
	// we are assuming its asking the right node.
	cs, err := c.Storage.GetCustomerState(in.Id)

	if err != nil {
		// if you have slower canonical backing store:
		//   if cs, err = c.SlowStorage.CustomerState(ctx, in); err == nil {
		//       return cs, nil
		//   }
		return nil, status.Errorf(codes.NotFound, "Cant Find it, originally: %s", err)
	}
	out := &proto.CustomerState{
		Id:              in.Id,
		LastAction:      cs.LastAction,
		CurrentSequence: cs.CurrentSequence,
	}
	return out, err
}

// WriteLog could be of two forms: like this, or streamed.
// streaming would be much faster; and let batching work much better; but you need to
// have a service streaming to it.
// If you're using this as a caching layer; then WriteLog is only here for hot loading data based
// on predicted usage.
func (c *Customer) WriteLog(ctx context.Context, el *proto.NewCustomerLog) (*proto.ErrorDetails, error) {
	// handle filtering on member list and consistent hash
	// (not really tested for replicationFactors)
	if c.HashList.LocateKey([]byte(strconv.FormatUint(el.GetCustomerID(), 10))).String() != c.MemberList.LocalNode().Name {
		return &proto.ErrorDetails{
				Failed:    true,
				ErrorCode: 1,
				ErrorMsg:  "Wrong Node",
			},
			status.Errorf(codes.FailedPrecondition, "Wrong member")
	}

	err := c.Storage.WriteLog(el.GetCustomerID(), el.GetLog())
	if err != nil {
		return &proto.ErrorDetails{
			Failed:    true,
			ErrorCode: 1,
			ErrorMsg:  err.Error(),
		}, status.Errorf(codes.Unknown, err.Error())
	}

	return new(proto.ErrorDetails), nil
}
