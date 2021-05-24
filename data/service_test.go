package data_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/yarbelk/grpcstuff/data"
	"github.com/yarbelk/grpcstuff/proto"
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
		dir, err := os.MkdirTemp("/tmp", "badger_store_test_*.db")
		if err != nil {
			t.Fatalf("cant create test db file %+v", err)
		}
		defer os.RemoveAll(dir)
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
		if !reflect.DeepEqual([]string{"1:0", "1:1"}, keys) {
			_, file, line, _ := runtime.Caller(0)
			fmt.Printf("%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\n\n", filepath.Base(file), line, []string{"1:0", "1:1"}, keys)
			t.FailNow()
		}
	})
	t.Run("Write returns error on incorrect sequenceId", func(t *testing.T) {
		dir, err := os.MkdirTemp("/tmp", "badger_store_test_*.db")
		if err != nil {
			t.Fatalf("cant create test db file %+v", err)
		}
		defer os.RemoveAll(dir)
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
			dir, err := os.MkdirTemp("/tmp", "badger_store_test_*.db")
			if err != nil {
				t.Fatalf("cant create test db file %+v", err)
			}
			defer os.RemoveAll(dir)
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
