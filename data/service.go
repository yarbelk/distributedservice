package data

import (
	"fmt"

	"github.com/yarbelk/grpcstuff/protostuff"
)

type CustomerState struct {
	LastAction      string
	CurrentSequence int64
}

type Storer interface {
	GetCustomerState(id int64) CustomerState
	WriteLog(id int64, el protostuff.CustomerEventLog)
}

type BadgerStore struct {
	DB *badger.DB
}

func New(path, fn string) *BadgerStore {
	db, err := Open(DefaultOptions(dir))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.View(func(txn *Txn) error {
		_, err := txn.Get([]byte("key"))
		// We expect ErrKeyNotFound
		fmt.Println(err)
		return nil
	})

	if err != nil {
		panic(err)
	}

	txn := db.NewTransaction(true) // Read-write txn
	err = txn.SetEntry(NewEntry([]byte("key"), []byte("value")))
	if err != nil {
		panic(err)
	}
	err = txn.Commit()
	if err != nil {
		panic(err)
	}

	err = db.View(func(txn *Txn) error {
		item, err := txn.Get([]byte("key"))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(val))
		return nil
	})

	if err != nil {
		panic(err)
	}
}
