package blockchain

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"log"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

const (
	dbPath = "./tmp/blocks"
)

func InitBlockChain() *BlockChain {
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalln(err)
	}
	err2 := db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing block")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			if err != nil {
				log.Fatalln(err)
			}
			var lastHash []byte

			err = item.Value(func(val []byte) error {

				// Copying or parsing val is valid.
				lastHash = append([]byte{}, val...)

				return nil
			})

			return err
		}
	})

	if err2 != nil {
		log.Fatalln(err2)
	}
	blockChain := BlockChain{lastHash, db}
	return &blockChain
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Fatalln(err)
		}
		var lastHash []byte

		err = item.Value(func(val []byte) error {

			// Copying or parsing val is valid.
			lastHash = append([]byte{}, val...)

			return nil
		})
		return err

	})
	if err != nil {
		log.Fatalln(err)
	}
	newBlock := CreateBlock(data, lastHash)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Fatalln(err)
		}
		chain.LastHash = newBlock.Hash
		return err
	})
	if err != nil {
		log.Fatalln(err)
	}
}
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}
	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)

		var encodedBlock []byte

		err = item.Value(func(val []byte) error {

			// Copying or parsing val is valid.
			encodedBlock = append([]byte{}, val...)

			return nil
		})
		if err != nil {
			log.Fatalln(err)
		}
		block = Deserialize(encodedBlock)
		return err
	})

	if err != nil {
		log.Fatalln(err)
	}
	iter.CurrentHash = block.PrevHash
	return block
}
