package data_test

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/yarbelk/distributedservice/data"
	"github.com/yarbelk/distributedservice/proto"
)

func GetKeyList(b *data.BadgerStore, id uint64) ([]string, error) {
	prefix := []byte(fmt.Sprintf("%d:", id))

	keys := make([]string, 0)
	err := b.LogDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			keys = append(keys, string(item.Key()))
		}
		return nil
	})
	return keys, err
}

func TestBadgerImplementationBasics(t *testing.T) {
	t.Run("Write appends with correct sequences", func(t *testing.T) {
		dir := t.TempDir()
		ds := data.New(dir)
		defer ds.Close()
		el1 := proto.CustomerEventLog{
			SequenceId: 0,
			Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
			Action:     &proto.Action{Action: "Create First Record"},
		}
		el2 := proto.CustomerEventLog{
			SequenceId: 1,
			Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{2}},
			Action:     &proto.Action{Action: "Create Second Record"},
		}
		ds.WriteLog(1, &el1)
		ds.WriteLog(1, &el2)
		keys, err := GetKeyList(ds, 1)
		if err != nil {
			t.Fatal(err)
		}
		// ordering ftw - mat.MaxUint64.  this should be done programatically but...
		if !reflect.DeepEqual([]string{"1:000000000000000000000", "1:000000000000000000001"}, keys) {
			_, file, line, _ := runtime.Caller(0)

			t.Logf("%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\n\n", filepath.Base(file), line, []string{"1:000000000000000000000", "1:000000000000000000001"}, keys)
			t.FailNow()
		}
	})
	t.Run("Write returns error on incorrect sequenceId", func(t *testing.T) {
		dir := t.TempDir()
		ds := data.New(dir)
		defer ds.Close()
		table := []struct {
			name  string
			input []*proto.CustomerEventLog
		}{
			{
				name: "Duplicate sequenceId",
				input: []*proto.CustomerEventLog{
					&proto.CustomerEventLog{
						SequenceId: 0,
						Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
						Action:     &proto.Action{Action: "Create First Record"},
					},
					&proto.CustomerEventLog{
						SequenceId: 0,
						Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{2}},
						Action:     &proto.Action{Action: "This should fail with error"},
					},
				},
			}, {
				name: " sequenceId start on nonzero",
				input: []*proto.CustomerEventLog{
					&proto.CustomerEventLog{
						SequenceId: 1,
						Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
						Action:     &proto.Action{Action: "Create Bad First Record"},
					},
				},
			}, {
				name: " sequenceId skipping value",
				input: []*proto.CustomerEventLog{
					&proto.CustomerEventLog{
						SequenceId: 0,
						Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
						Action:     &proto.Action{Action: "Create First Record"},
					},
					&proto.CustomerEventLog{
						SequenceId: 2,
						Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
						Action:     &proto.Action{Action: "Create Bad First Record"},
					},
				},
			},
		}

		for _, tt := range table {
			var err error
			t.Run(tt.name, func(t *testing.T) {
				ds.LogDB.DropAll()
				var in *proto.CustomerEventLog
				for _, in = range tt.input {
					err = ds.WriteLog(1, in)
					if err != nil {
						break
					}
				}
				if err != data.InvalidSequenceError {
					t.Logf("expected to get %s, got %s error. CEL: %+v", data.InvalidSequenceError, err, in)
					t.FailNow()
				}
			})
		}

	})
	t.Run("Get Customer gets back expected state", func(t *testing.T) {
		t.Run("Sanity check for TDD", func(t *testing.T) {
			dir := t.TempDir()
			ds := data.New(dir)
			defer ds.Close()

			table := []struct {
				name          string
				input         []*proto.CustomerEventLog
				expectedState string
				expectedSID   uint64
			}{
				{
					name: "Single Entry sanity",
					input: []*proto.CustomerEventLog{
						&proto.CustomerEventLog{
							SequenceId: 0,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "Create First Record"},
						},
					},
					expectedState: "Create First Record",
					expectedSID:   0,
				}, {
					name: "Updates with two",
					input: []*proto.CustomerEventLog{
						&proto.CustomerEventLog{
							SequenceId: 0,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "Create First record"},
						},
						&proto.CustomerEventLog{
							SequenceId: 1,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "Updated Second"},
						},
					},
					expectedState: "Updated Second",
					expectedSID:   1,
				}, {
					name: "updates with 11 (lexagraphical sorting)",
					input: []*proto.CustomerEventLog{
						&proto.CustomerEventLog{
							SequenceId: 0,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "Create First record"},
						},
						&proto.CustomerEventLog{
							SequenceId: 1,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 2,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 3,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 4,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 5,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 6,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 7,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 8,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 9,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 10,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "logdata"},
						},
						&proto.CustomerEventLog{
							SequenceId: 11,
							Timestamp:  &proto.VectorTimestamp{Timestamps: []int64{1}},
							Action:     &proto.Action{Action: "Updated Last"},
						},
					},
					expectedState: "Updated Last",
					expectedSID:   11,
				},
			}
			for _, tt := range table {
				var err error
				t.Run(tt.name, func(t *testing.T) {
					ds.LogDB.DropAll()

					for _, in := range tt.input {
						err = ds.WriteLog(1, in)
						if err != nil {
							if err != nil {
								t.Fatalf("can't setup customer for %+v", err)
							}
						}
					}
					customer, err := ds.GetCustomerState(1)
					if err != nil {
						t.Fatalf("can't get customer %+v", err)
					}
					if customer.LastAction != tt.expectedState {

						t.Fatalf("expected state %s got %s", tt.expectedState, customer.LastAction)
					}
					if customer.CurrentSequence != tt.expectedSID {
						t.Fatalf("expected SID %d got %d", tt.expectedSID, customer.CurrentSequence)
					}
				})
			}
		})
	})
}

func BenchmarkLookupSpeed(b *testing.B) {
	// or: fun explorations in typecasting int types to get random data sets.

	users := 10000
	// Test performance of datastore
	dir := b.TempDir()
	ds := data.New(dir)
	defer ds.Close()
	var i uint64
	for i = 0; i < uint64(users); i++ {
		// this is safe because Int63n will panic on negative inputs, and is always positive
		var j int64
		for j = 0; int64(j) <= rand.Int63n(200); j++ {
			b.Log(i, j)
			el := &proto.CustomerEventLog{
				SequenceId: uint64(j),
				Timestamp: &proto.VectorTimestamp{
					Timestamps: []int64{j},
				},
				Action: &proto.Action{Action: "test it"},
			}
			err := ds.WriteLog(i, el)
			if err != nil {
				b.Fatal(el, err)
			}
		}
	}
	b.Run("Test 10 random lookups", func(b *testing.B) {
		// change lookups to mess with caching - worst case
		lookups := make([]uint64, 0, 10)
		for i = 0; i < 10; i++ {
			lookups = append(lookups, uint64(rand.Int63n(int64(users))))
		}
		b.ResetTimer()
		for _, id := range lookups {
			cs, err := ds.GetCustomerState(id)
			if err != nil {
				b.Log(err)
				b.Fail()
			}
			if cs.LastAction != "test it" {
				b.Logf("%+v\n", cs)
				b.Fail()
			}
		}
	})
	b.Run("Test 1000 random lookups", func(b *testing.B) {
		// change lookups to mess with caching - worst case
		lookups := make([]uint64, 0, 1000)
		for i = 0; i < 1000; i++ {
			lookups = append(lookups, uint64(rand.Int63n(int64(users))))
		}
		b.ResetTimer()
		for _, id := range lookups {
			cs, err := ds.GetCustomerState(id)
			if err != nil {
				b.Log(err)
				b.Fail()
			}
			if cs.LastAction != "test it" {
				b.Logf("%+v\n", cs)
				b.Fail()
			}
		}
	})
	b.Run("Test 5000 random lookups", func(b *testing.B) {
		// change lookups to mess with caching - worst case
		lookups := make([]uint64, 0, 5000)
		for i = 0; i < 5000; i++ {
			lookups = append(lookups, uint64(rand.Int63n(int64(users))))
		}
		b.ResetTimer()
		for _, id := range lookups {
			cs, err := ds.GetCustomerState(id)
			if err != nil {
				b.Log(err)
				b.Fail()
			}
			if cs.LastAction != "test it" {
				b.Logf("%+v\n", cs)
				b.Fail()
			}
		}
	})
}
