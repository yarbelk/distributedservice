package data

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/yarbelk/grpcstuff/proto"
)

// CustomerState is an aggregate root - need to move it to its own package
// TODO move to `customer` package.
// also - super basic example
type CustomerState struct {
	LastAction      string
	CurrentSequence int64
}

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	InvalidSequenceError Error = "Invalid sequence. Does not pass validation"
)

// Apply a log "cleanly", (actually, this is a completely unvalidated apply for proof of concept)
// actual consistency is left as an exercise for the reader.  the CustomerEventLog has a
// 'vector' clock (also fake vector, but probably good enough anyway), and
// between that and sequenceID you have a enough to implement a consistant
// event sourced Customer view
func (cs *CustomerState) Apply(l *proto.CustomerEventLog) {
	cs.LastAction = l.GetAction().GetAction() // fun semantics of protobuf
	cs.CurrentSequence++                      // should actually be validating the timestamp and prior sequences etc. but this is PoC
}

type Storer interface {
	GetCustomerState(id uint64) (CustomerState, error)
	WriteLog(id uint64, el proto.CustomerEventLog) error
}

// BadgerStore is a fast DB key value store that lets you very quickly iterate over keys in lexagraphical order
// its similar to BigTable and RocksDB and lots of others.  in this case, we only have a LogDB, you  could add a snapshot
// db, which has one value per customer ID, and stores the snapshot at a particular sequence.  then you skip replaying all
// the logs, and just replay from that sequenceID.
// this is left as an exercise for the reader.
//
// Another simplificaiton is in the keying/logging system.  we are completly skipping
// good design of the logging format: which should have a standardized way of looking up
// and versioning logs.  Typically i'd do something like (and this is proto)
type BadgerStore struct {
	LogDB *badger.DB
}

func (b *BadgerStore) Close() {
	b.LogDB.Close()
}

func New(path string) *BadgerStore {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		panic(err)
	}

	return &BadgerStore{LogDB: db}
}

// GetCustomerState to get a root for the customer.  totally could use snapshots etc
func (b *BadgerStore) GetCustomerState(id int64) (CustomerState, error) {
	cs := new(CustomerState)
	prefix := []byte(fmt.Sprintf("%d:", id))
	err := b.LogDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		customerLog := new(proto.CustomerEventLog)
		buf := make([]byte, 0, 1000) // make a capacity 1000 buffer, of length 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			buf = buf[:0] // we're reusing the buffer. reset length
			item := it.Item()
			key := item.Key()
			splits := strings.Split(string(key), ":")
			if len(splits) != 2 {
				panic("lazy panic stuffs bc/ PoC")
			}
			s, err := strconv.ParseUint(splits[1], 10, 64)
			buf, err := item.ValueCopy(buf)
			if err != nil {
				return err
			}
			protobuf.Unmarshal(buf, customerLog)
			// sanity checks
			if s != customerLog.SequenceId {
				log.Printf("heres a fun thing: the datamodel is borked for stored key %s: %+v\n", string(key), customerLog)
			}
			cs.Apply(customerLog)
		}
		defer it.Close()
		return nil
	})
	return *cs, err
}

// WriteLog is not optimized/batched up for speed.  It could be.
// also: copying nots on design from the protofile so they are not missed:

//     Another simplificaiton is in the keying/logging system.  we are completly skipping
//     good design of the logging format: which should have a standardized way of looking up
//     and versioning logs.  Typically i'd do something like
//
//     message LogMeta {
//       string EventType = 1;
//       int64 EventVersion = 2;
//       uint64 sequenceId = 3;
//       VectorTimestamp eventTimestamp = 4;
//       // a bunch of metadat
//       Any EventPayload = 10;  // or byte, or anything.
//     }
//     in this way; you can have many versions of the same EventName that cleanly apply
//     and its discoverable in a fast to deserialize way.  the meta data is moved
//     out of the 'XEventLog' message.
//     I'm not doing this because, while not hard to do, its too much effort for a
//     PoC; but I think its important to understand that as implemented: this is
//     _not_ a futureproof design.  Or a scalable design.
func (b *BadgerStore) WriteLog(id uint64, el *proto.CustomerEventLog) error {
	err := b.LogDB.Update(func(txn *badger.Txn) error {
		if !validSequenceID(id, el.SequenceId, txn) {
			return InvalidSequenceError
		}
		v, err := protobuf.Marshal(el)
		if err != nil {
			return err
		}
		txn.Set([]byte(fmt.Sprintf("%d:%d", id, el.SequenceId)), v)
		return nil
	})
	return err
}

func validSequenceID(cid, givenSID uint64, txn *badger.Txn) bool {
	var expectedSequenceId uint64

	prefix := []byte(fmt.Sprintf("%d:", cid))
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		key := it.Item().Key()
		splits := strings.Split(string(key), ":")
		if len(splits) != 2 {
			panic("lazy panic stuffs bc/ PoC")
		}
		s, err := strconv.ParseUint(splits[1], 10, 64)
		if err != nil {
			panic(err)
		}
		expectedSequenceId = s + 1
	}
	if givenSID != expectedSequenceId {
		return false
	}
	return true
}
