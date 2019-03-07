// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either pubKeyHash 3 of the License, or
// (at your option) any later pubKeyHash.
//
// the go-dappley library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//

package core

import (
	"encoding/hex"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/storage/mocks"
	"github.com/dappley/go-dappley/util"
	logger "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var bh1 = &BlockHeader{
	[]byte("hash"),
	nil,
	1,
	time.Now().Unix(),
	nil,
	0,
}

var bh2 = &BlockHeader{
	[]byte("hash1"),
	[]byte("hash"),
	1,
	time.Now().Unix(),
	nil,
	1,
}

// Padding Address to 32 Byte
var address1Bytes = []byte("address1000000000000000000000000")
var address2Bytes = []byte("address2000000000000000000000000")
var address1Hash, _ = NewUserPubKeyHash(address1Bytes)
var address2Hash, _ = NewUserPubKeyHash(address2Bytes)

func GenerateUtxoMockBlockWithoutInputs() *Block {

	t1 := MockUtxoTransactionWithoutInputs()
	return &Block{
		header:       bh1,
		transactions: []*Transaction{t1},
	}
}

func GenerateUtxoMockBlockWithInputs() *Block {

	t1 := MockUtxoTransactionWithInputs()
	return &Block{
		header:       bh2,
		transactions: []*Transaction{t1},
	}
}

func MockUtxoTransactionWithoutInputs() *Transaction {
	return &Transaction{
		ID:   []byte("tx1"),
		Vin:  []TXInput{},
		Vout: MockUtxoOutputsWithoutInputs(),
		Tip:  common.NewAmount(5),
	}
}

func MockUtxoTransactionWithInputs() *Transaction {
	return &Transaction{
		ID:   []byte("tx2"),
		Vin:  MockUtxoInputs(),
		Vout: MockUtxoOutputsWithInputs(),
		Tip:  common.NewAmount(5),
	}
}

func MockUtxoInputs() []TXInput {
	return []TXInput{
		{
			[]byte("tx1"),
			0,
			util.GenerateRandomAoB(2),
			address1Bytes},
		{
			[]byte("tx1"),
			1,
			util.GenerateRandomAoB(2),
			address1Bytes},
	}
}

func MockUtxoOutputsWithoutInputs() []TXOutput {
	return []TXOutput{
		{common.NewAmount(5), address1Hash, ""},
		{common.NewAmount(7), address1Hash, ""},
	}
}

func MockUtxoOutputsWithInputs() []TXOutput {
	return []TXOutput{
		{common.NewAmount(4), address1Hash, ""},
		{common.NewAmount(5), address2Hash, ""},
		{common.NewAmount(3), address2Hash, ""},
	}
}

func TestAddUTXO(t *testing.T) {
	db := storage.NewRamStorage()
	defer db.Close()

	txout := TXOutput{common.NewAmount(5), address1Hash, ""}
	utxoIndex := NewUTXOIndex(NewUTXOCache(storage.NewRamStorage()))

	utxoIndex.AddUTXO(txout, []byte{1}, 0)

	addr1UTXOs := utxoIndex.index[hex.EncodeToString(address1Hash)]
	assert.Equal(t, 1, addr1UTXOs.Size())
	assert.Equal(t, txout.Value, addr1UTXOs.GetAllUtxos()[0].Value)
	assert.Equal(t, []byte{1}, addr1UTXOs.GetAllUtxos()[0].Txid)
	assert.Equal(t, 0, addr1UTXOs.GetAllUtxos()[0].TxIndex)

	_, ok := utxoIndex.index["address2"]
	assert.Equal(t, false, ok)
}

func TestRemoveUTXO(t *testing.T) {
	db := storage.NewRamStorage()
	defer db.Close()

	utxoIndex := NewUTXOIndex(NewUTXOCache(storage.NewRamStorage()))

	addr1UtxoTx := NewUTXOTx()
	addr1UtxoTx = addr1UtxoTx.PutUtxo(&UTXO{TXOutput{common.NewAmount(5), address1Hash, ""}, []byte{1}, 0, UtxoNormal})
	addr1UtxoTx = addr1UtxoTx.PutUtxo(&UTXO{TXOutput{common.NewAmount(2), address1Hash, ""}, []byte{1}, 1, UtxoNormal})
	addr1UtxoTx = addr1UtxoTx.PutUtxo(&UTXO{TXOutput{common.NewAmount(2), address1Hash, ""}, []byte{2}, 0, UtxoNormal})

	addr2UtxoTx := NewUTXOTx()
	addr2UtxoTx = addr2UtxoTx.PutUtxo(&UTXO{TXOutput{common.NewAmount(4), address2Hash, ""}, []byte{1}, 2, UtxoNormal})

	utxoIndex.index[hex.EncodeToString(address1Hash)] = &addr1UtxoTx
	utxoIndex.index[hex.EncodeToString(address2Hash)] = &addr2UtxoTx

	err := utxoIndex.removeUTXO(address1Hash, []byte{1}, 0)

	assert.Nil(t, err)
	assert.Equal(t, 2, utxoIndex.index[hex.EncodeToString(address1Hash)].Size())
	assert.Equal(t, 1, utxoIndex.index[hex.EncodeToString(address2Hash)].Size())

	err = utxoIndex.removeUTXO(address2Hash, []byte{2}, 1) // Does not exists

	assert.NotNil(t, err)
	assert.Equal(t, 2, utxoIndex.index[hex.EncodeToString(address1Hash)].Size())
	assert.Equal(t, 1, utxoIndex.index[hex.EncodeToString(address2Hash)].Size())
}

func TestUpdate(t *testing.T) {
	db := storage.NewRamStorage()
	defer db.Close()

	blk := GenerateUtxoMockBlockWithoutInputs()
	utxoIndex := NewUTXOIndex(NewUTXOCache(db))
	utxoIndex.UpdateUtxoState(blk.GetTransactions())
	utxoIndex.Save()
	utxoIndexInDB := NewUTXOIndex(NewUTXOCache(db))

	// test updating UTXO index with non-dependent transactions
	// Assert that both the original instance and the database copy are updated correctly
	for _, index := range []UTXOIndex{*utxoIndex, *utxoIndexInDB} {
		utxoTx := index.GetAllUTXOsByPubKeyHash(address1Hash)
		assert.Equal(t, 2, utxoTx.Size())
		utxo0 := utxoTx.GetUtxo(blk.transactions[0].ID, 0)
		utx1 := utxoTx.GetUtxo(blk.transactions[0].ID, 1)
		assert.Equal(t, blk.transactions[0].ID, utxo0.Txid)
		assert.Equal(t, 0, utxo0.TxIndex)
		assert.Equal(t, blk.transactions[0].Vout[0].Value, utxo0.Value)
		assert.Equal(t, blk.transactions[0].ID, utx1.Txid)
		assert.Equal(t, 1, utx1.TxIndex)
		assert.Equal(t, blk.transactions[0].Vout[1].Value, utx1.Value)
	}

	// test updating UTXO index with dependent transactions
	var prikey1 = "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa71"
	var pubkey1 = GetKeyPairByString(prikey1).PublicKey
	var pkHash1, _ = NewUserPubKeyHash(pubkey1)
	var prikey2 = "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa72"
	var pubkey2 = GetKeyPairByString(prikey2).PublicKey
	var pkHash2, _ = NewUserPubKeyHash(pubkey2)
	var prikey3 = "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa73"
	var pubkey3 = GetKeyPairByString(prikey3).PublicKey
	var pkHash3, _ = NewUserPubKeyHash(pubkey3)
	var prikey4 = "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa74"
	var pubkey4 = GetKeyPairByString(prikey4).PublicKey
	var pkHash4, _ = NewUserPubKeyHash(pubkey4)
	var prikey5 = "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa75"
	var pubkey5 = GetKeyPairByString(prikey5).PublicKey
	var pkHash5, _ = NewUserPubKeyHash(pubkey5)

	var dependentTx1 = Transaction{
		ID: nil,
		Vin: []TXInput{
			{tx1.ID, 1, nil, pubkey1},
		},
		Vout: []TXOutput{
			{common.NewAmount(5), pkHash1, ""},
			{common.NewAmount(10), pkHash2, ""},
		},
		Tip: common.NewAmount(3),
	}
	dependentTx1.ID = dependentTx1.Hash()

	var dependentTx2 = Transaction{
		ID: nil,
		Vin: []TXInput{
			{dependentTx1.ID, 1, nil, pubkey2},
		},
		Vout: []TXOutput{
			{common.NewAmount(5), pkHash3, ""},
			{common.NewAmount(3), pkHash4, ""},
		},
		Tip: common.NewAmount(2),
	}
	dependentTx2.ID = dependentTx2.Hash()

	var dependentTx3 = Transaction{
		ID: nil,
		Vin: []TXInput{
			{dependentTx2.ID, 0, nil, pubkey3},
		},
		Vout: []TXOutput{
			{common.NewAmount(1), pkHash4, ""},
		},
		Tip: common.NewAmount(4),
	}
	dependentTx3.ID = dependentTx3.Hash()

	var dependentTx4 = Transaction{
		ID: nil,
		Vin: []TXInput{
			{dependentTx2.ID, 1, nil, pubkey4},
			{dependentTx3.ID, 0, nil, pubkey4},
		},
		Vout: []TXOutput{
			{common.NewAmount(3), pkHash1, ""},
		},
		Tip: common.NewAmount(1),
	}
	dependentTx4.ID = dependentTx4.Hash()

	var dependentTx5 = Transaction{
		ID: nil,
		Vin: []TXInput{
			{dependentTx1.ID, 0, nil, pubkey1},
			{dependentTx4.ID, 0, nil, pubkey1},
		},
		Vout: []TXOutput{
			{common.NewAmount(4), pkHash5, ""},
		},
		Tip: common.NewAmount(4),
	}
	dependentTx5.ID = dependentTx5.Hash()

	utxoPk2 := &UTXO{dependentTx1.Vout[1], dependentTx1.ID, 1, UtxoNormal}
	utxoPk1 := &UTXO{dependentTx1.Vout[0], dependentTx1.ID, 0, UtxoNormal}

	utxoTxPk2 := NewUTXOTx()
	utxoTxPk2 = utxoTxPk2.PutUtxo(utxoPk2)

	utxoTxPk1 := NewUTXOTx()
	utxoTxPk1 = utxoTxPk1.PutUtxo(utxoPk1)

	utxoIndex2 := NewUTXOIndex(NewUTXOCache(storage.NewRamStorage()))

	utxoIndex2.index[hex.EncodeToString(pkHash2)] = &utxoTxPk2
	utxoIndex2.index[hex.EncodeToString(pkHash1)] = &utxoTxPk1

	tx2Utxo1 := UTXO{dependentTx2.Vout[0], dependentTx2.ID, 0, UtxoNormal}
	tx2Utxo2 := UTXO{dependentTx2.Vout[1], dependentTx2.ID, 1, UtxoNormal}
	tx2Utxo3 := UTXO{dependentTx3.Vout[0], dependentTx3.ID, 0, UtxoNormal}
	tx2Utxo4 := UTXO{dependentTx1.Vout[0], dependentTx1.ID, 0, UtxoNormal}
	tx2Utxo5 := UTXO{dependentTx4.Vout[0], dependentTx4.ID, 0, UtxoNormal}
	dependentTx2.Sign(GetKeyPairByString(prikey2).PrivateKey, utxoIndex2.index[hex.EncodeToString(pkHash2)].GetAllUtxos())
	dependentTx3.Sign(GetKeyPairByString(prikey3).PrivateKey, []*UTXO{&tx2Utxo1})
	dependentTx4.Sign(GetKeyPairByString(prikey4).PrivateKey, []*UTXO{&tx2Utxo2, &tx2Utxo3})
	dependentTx5.Sign(GetKeyPairByString(prikey1).PrivateKey, []*UTXO{&tx2Utxo4, &tx2Utxo5})

	txsForUpdate := []*Transaction{&dependentTx2, &dependentTx3}
	utxoIndex2.UpdateUtxoState(txsForUpdate)
	assert.Equal(t, 1, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash1).Size())
	assert.Equal(t, 0, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash2).Size())
	assert.Equal(t, 0, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash3).Size())
	assert.Equal(t, 2, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash4).Size())
	txsForUpdate = []*Transaction{&dependentTx2, &dependentTx3, &dependentTx4}
	utxoIndex2.UpdateUtxoState(txsForUpdate)
	assert.Equal(t, 2, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash1).Size())
	assert.Equal(t, 0, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash2).Size())
	assert.Equal(t, 0, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash3).Size())
	txsForUpdate = []*Transaction{&dependentTx2, &dependentTx3, &dependentTx4, &dependentTx5}
	utxoIndex2.UpdateUtxoState(txsForUpdate)
	assert.Equal(t, 0, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash1).Size())
	assert.Equal(t, 0, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash2).Size())
	assert.Equal(t, 0, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash3).Size())
	assert.Equal(t, 0, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash4).Size())
	assert.Equal(t, 1, utxoIndex2.GetAllUTXOsByPubKeyHash(pkHash5).Size())
}

func TestUpdate_Failed(t *testing.T) {
	db := new(mocks.Storage)

	simulatedFailure := errors.New("simulated storage failure")
	db.On("Put", mock.Anything, mock.Anything).Return(simulatedFailure)
	db.On("Get", mock.Anything, mock.Anything).Return(nil, nil)

	blk := GenerateUtxoMockBlockWithoutInputs()
	utxoIndex := NewUTXOIndex(NewUTXOCache(db))
	utxoIndex.UpdateUtxoState(blk.GetTransactions())
	err := utxoIndex.Save()
	assert.Equal(t, simulatedFailure, err)
	assert.Equal(t, 2, utxoIndex.GetAllUTXOsByPubKeyHash(address1Hash).Size())
}

func TestGetUTXOIndexAtBlockHash(t *testing.T) {
	genesisAddr := NewAddress("##@@")
	genesisBlock := NewGenesisBlock(genesisAddr)

	// prepareBlockchainWithBlocks returns a blockchain that contains the given blocks with correct utxoIndex in RAM
	prepareBlockchainWithBlocks := func(blks []*Block) *Blockchain {
		bc := CreateBlockchain(genesisAddr, storage.NewRamStorage(), nil, 128, nil, 100000)
		for _, blk := range blks {
			err := bc.AddBlockContextToTail(PrepareBlockContext(bc, blk))
			if err != nil {
				logger.Fatal("TestGetUTXOIndexAtBlockHash: cannot add the blocks to blockchain.")
			}
		}
		return bc
	}

	// utxoIndexFromTXs creates a utxoIndex containing all vout of transactions in txs
	utxoIndexFromTXs := func(txs []*Transaction, cache *UTXOCache) *UTXOIndex {
		utxoIndex := NewUTXOIndex(cache)
		utxosMap := make(map[string]*UTXOTx)
		for _, tx := range txs {
			for i, vout := range tx.Vout {
				utxos, ok := utxosMap[hex.EncodeToString(vout.PubKeyHash)]
				if !ok {
					newUtxos := NewUTXOTx()
					utxos = &newUtxos
				}
				*utxos = utxos.PutUtxo(newUTXO(vout, tx.ID, i, UtxoNormal))
				utxosMap[hex.EncodeToString(vout.PubKeyHash)] = utxos
			}
		}
		utxoIndex.index = utxosMap
		return utxoIndex
	}

	keypair := NewKeyPair()
	pbkh, _ := NewUserPubKeyHash(keypair.PublicKey)
	addr := pbkh.GenerateAddress()

	normalTX := NewCoinbaseTX(addr, "", 1, common.NewAmount(5))
	normalTX2 := Transaction{
		Hash("normal2"),
		[]TXInput{{normalTX.ID, 0, nil, keypair.PublicKey}},
		[]TXOutput{{common.NewAmount(5), pbkh, ""}},
		common.NewAmount(0),
	}
	abnormalTX := Transaction{
		Hash("abnormal"),
		[]TXInput{{normalTX.ID, 1, nil, nil}},
		[]TXOutput{{common.NewAmount(5), PubKeyHash([]byte("pkh")), ""}},
		common.NewAmount(0),
	}
	prevBlock := NewBlock([]*Transaction{}, genesisBlock)
	prevBlock.SetHash(prevBlock.CalculateHash())
	emptyBlock := NewBlock([]*Transaction{}, prevBlock)
	emptyBlock.SetHash(emptyBlock.CalculateHash())
	normalBlock := NewBlock([]*Transaction{&normalTX}, genesisBlock)
	normalBlock.SetHash(normalBlock.CalculateHash())
	normalBlock2 := NewBlock([]*Transaction{&normalTX2}, normalBlock)
	normalBlock2.SetHash(normalBlock2.CalculateHash())
	abnormalBlock := NewBlock([]*Transaction{&abnormalTX}, normalBlock)
	abnormalBlock.SetHash(abnormalBlock.CalculateHash())
	corruptedUTXOBlockchain := prepareBlockchainWithBlocks([]*Block{normalBlock, normalBlock2})
	err := utxoIndexFromTXs([]*Transaction{&normalTX}, corruptedUTXOBlockchain.GetUtxoCache()).Save()
	if err != nil {
		logger.Fatal("TestGetUTXOIndexAtBlockHash: cannot corrupt the utxoIndex in database.")
	}

	bcs := []*Blockchain{
		prepareBlockchainWithBlocks([]*Block{normalBlock}),
		prepareBlockchainWithBlocks([]*Block{normalBlock, normalBlock2}),
		CreateBlockchain(NewAddress(""), storage.NewRamStorage(), nil, 128, nil, 100000),
		prepareBlockchainWithBlocks([]*Block{prevBlock, emptyBlock}),
		prepareBlockchainWithBlocks([]*Block{normalBlock, normalBlock2}),
		prepareBlockchainWithBlocks([]*Block{normalBlock, abnormalBlock}),
		corruptedUTXOBlockchain,
	}
	tests := []struct {
		name     string
		bc       *Blockchain
		hash     Hash
		expected *UTXOIndex
		err      error
	}{
		{
			name:     "current block",
			bc:       bcs[0],
			hash:     normalBlock.GetHash(),
			expected: utxoIndexFromTXs([]*Transaction{&normalTX}, bcs[0].GetUtxoCache()),
			err:      nil,
		},
		{
			name:     "previous block",
			bc:       bcs[1],
			hash:     normalBlock.GetHash(),
			expected: utxoIndexFromTXs([]*Transaction{&normalTX}, bcs[1].GetUtxoCache()), // should not have utxo from normalTX2
			err:      nil,
		},
		{
			name:     "block not found",
			bc:       bcs[2],
			hash:     Hash("not there"),
			expected: NewUTXOIndex(bcs[2].GetUtxoCache()),
			err:      ErrBlockDoesNotExist,
		},
		{
			name:     "no txs in blocks",
			bc:       bcs[3],
			hash:     emptyBlock.GetHash(),
			expected: utxoIndexFromTXs(genesisBlock.transactions, bcs[3].GetUtxoCache()),
			err:      nil,
		},
		{
			name:     "genesis block",
			bc:       bcs[4],
			hash:     genesisBlock.GetHash(),
			expected: utxoIndexFromTXs(genesisBlock.transactions, bcs[4].GetUtxoCache()),
			err:      nil,
		},
		{
			name:     "utxo not found",
			bc:       bcs[5],
			hash:     normalBlock.GetHash(),
			expected: NewUTXOIndex(bcs[5].GetUtxoCache()),
			err:      ErrUTXONotFound,
		},
		{
			name:     "corrupted utxoIndex",
			bc:       bcs[6],
			hash:     normalBlock.GetHash(),
			expected: NewUTXOIndex(bcs[6].GetUtxoCache()),
			err:      ErrUTXONotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetUTXOIndexAtBlockHash(tt.bc.GetDb(), tt.bc, tt.hash)
			if !assert.Equal(t, tt.err, err) {
				return
			}
			//if len(tt.expected.index) != len(utxoIndex.index) {
			//	// The utxoIndex maps may contain empty lists for certain pubkeyhashes, which are equivalent to
			//	// nils on the other utxoIndex map we are comparing with.
			//	for pkh, utxos := range tt.expected.index {
			//		if len(utxos) == 0 && utxoIndex.index[pkh] == nil {
			//			continue
			//		}
			//		assert.Equal(t, tt.expected.index[pkh], utxoIndex.index[pkh])
			//	}
			//	for pkh, utxos := range utxoIndex.index {
			//		if len(utxos) == 0 && tt.expected.index[pkh] == nil {
			//			continue
			//		}
			//		if pkh == "" {
			//			continue
			//		}
			//		assert.Equal(t, tt.expected.index[pkh], utxoIndex.index[pkh])
			//	}
			//	return
			//}
			//assert.Equal(t, tt.expected, utxoIndex)
		})
	}
}

func TestCopyAndRevertUtxos(t *testing.T) {
	db := storage.NewRamStorage()
	defer db.Close()

	coinbaseAddr := Address{"testaddress"}
	bc := CreateBlockchain(coinbaseAddr, db, nil, 128, nil, 100000)

	blk1 := GenerateUtxoMockBlockWithoutInputs() // contains 2 UTXOs for address1
	blk2 := GenerateUtxoMockBlockWithInputs()    // contains tx that transfers address1's UTXOs to address2 with a change

	bc.AddBlockContextToTail(PrepareBlockContext(bc, blk1))
	bc.AddBlockContextToTail(PrepareBlockContext(bc, blk2))

	utxoIndex := NewUTXOIndex(bc.GetUtxoCache())
	addr1UTXOs := utxoIndex.GetAllUTXOsByPubKeyHash([]byte(address1Hash))
	addr2UTXOs := utxoIndex.GetAllUTXOsByPubKeyHash([]byte(address2Hash))
	// Expect address1 to have 1 utxo of $4
	assert.Equal(t, 1, addr1UTXOs.Size())
	_, utxo1, _ := addr1UTXOs.Iterator()
	assert.Equal(t, common.NewAmount(4), utxo1.Value)

	// Expect address2 to have 2 utxos totaling $8
	assert.Equal(t, 2, addr2UTXOs.Size())

	// Rollback to blk1, address1 has a $5 utxo and a $7 utxo, total $12, and address2 has nothing
	indexSnapshot, err := GetUTXOIndexAtBlockHash(db, bc, blk1.GetHash())
	if err != nil {
		panic(err)
	}

	addr1UtxoTx := indexSnapshot.GetAllUTXOsByPubKeyHash(address1Hash)
	assert.Equal(t, 2, addr1UtxoTx.Size())

	tx1 := MockUtxoTransactionWithoutInputs()

	assert.Equal(t, common.NewAmount(5), addr1UtxoTx.GetUtxo(tx1.ID, 0).Value)
	assert.Equal(t, common.NewAmount(7), addr1UtxoTx.GetUtxo(tx1.ID, 1).Value)
	assert.Equal(t, 0, indexSnapshot.GetAllUTXOsByPubKeyHash(address2Hash).Size())
}

func TestFindUTXO(t *testing.T) {
	Txin := MockTxInputs()
	Txin = append(Txin, MockTxInputs()...)
	utxo1 := &UTXO{TXOutput{common.NewAmount(10), PubKeyHash([]byte("addr1")), ""}, Txin[0].Txid, Txin[0].Vout, UtxoNormal}
	utxo2 := &UTXO{TXOutput{common.NewAmount(9), PubKeyHash([]byte("addr1")), ""}, Txin[1].Txid, Txin[1].Vout, UtxoNormal}
	utxoTx1 := NewUTXOTxWithData(utxo1)
	utxoTx2 := NewUTXOTxWithData(utxo2)

	assert.Equal(t, utxo1, utxoTx1.GetUtxo(Txin[0].Txid, Txin[0].Vout))
	assert.Equal(t, utxo2, utxoTx2.GetUtxo(Txin[1].Txid, Txin[1].Vout))
	assert.Nil(t, utxoTx1.GetUtxo(Txin[2].Txid, Txin[2].Vout))
	assert.Nil(t, utxoTx2.GetUtxo(Txin[3].Txid, Txin[3].Vout))
}

func TestConcurrentUTXOindexReadWrite(t *testing.T) {
	index := NewUTXOIndex(NewUTXOCache(storage.NewRamStorage()))

	var mu sync.Mutex
	var readOps uint64
	var addOps uint64
	var deleteOps uint64
	const concurrentUsers = 10
	exists := false

	// start 10 simultaneous goroutines to execute repeated
	// reads and writes, once per millisecond in
	// each goroutine.
	for r := 0; r < concurrentUsers; r++ {
		go func() {
			for {
				//perform a read
				index.GetAllUTXOsByPubKeyHash([]byte("asd"))
				atomic.AddUint64(&readOps, 1)
				//perform a write

				mu.Lock()
				tmpExists := exists
				mu.Unlock()
				if !tmpExists {
					index.AddUTXO(TXOutput{}, []byte("asd"), 65)
					atomic.AddUint64(&addOps, 1)
					mu.Lock()
					exists = true
					mu.Unlock()

				} else {
					index.removeUTXO([]byte("asd"), []byte("asd"), 65)
					atomic.AddUint64(&deleteOps, 1)
					mu.Lock()
					exists = false
					mu.Unlock()
				}

				time.Sleep(time.Millisecond * 1)
			}
		}()
	}

	time.Sleep(time.Second * 1)

	//if reports concurrent map writes, then test is broken, if passes, then test is correct
	assert.True(t, true)
}

func TestUTXOIndex_GetUTXOsByAmount(t *testing.T) {

	contractPkh := NewContractPubKeyHash()
	//preapre 3 utxos in the utxo index
	txoutputs := []TXOutput{
		{common.NewAmount(3), address1Hash, ""},
		{common.NewAmount(4), address2Hash, ""},
		{common.NewAmount(5), address2Hash, ""},
		{common.NewAmount(2), contractPkh, "helloworld!"},
		{common.NewAmount(4), contractPkh, ""},
	}

	index := NewUTXOIndex(NewUTXOCache(storage.NewRamStorage()))
	for i, txoutput := range txoutputs {
		index.AddUTXO(txoutput, []byte("01"), i)
	}

	//start the test
	tests := []struct {
		name   string
		amount *common.Amount
		pubKey []byte
		err    error
	}{
		{"enoughUtxo",
			common.NewAmount(3),
			[]byte(address2Hash),
			nil},

		{"notEnoughUtxo",
			common.NewAmount(4),
			[]byte(address1Hash),
			ErrInsufficientFund},

		{"justEnoughUtxo",
			common.NewAmount(9),
			[]byte(address2Hash),
			nil},
		{"notEnoughUtxo2",
			common.NewAmount(10),
			[]byte(address2Hash),
			ErrInsufficientFund},
		{"smartContractUtxo",
			common.NewAmount(3),
			[]byte(contractPkh),
			nil},
		{"smartContractUtxoInsufficient",
			common.NewAmount(5),
			[]byte(contractPkh),
			ErrInsufficientFund},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utxos, err := index.GetUTXOsByAmount(tt.pubKey, tt.amount)
			assert.Equal(t, tt.err, err)
			if err != nil {
				return
			}
			sum := common.NewAmount(0)
			for _, utxo := range utxos {
				sum = sum.Add(utxo.Value)
			}
			assert.True(t, sum.Cmp(tt.amount) >= 0)
		})
	}

}

func TestUTXOIndex_DeepCopy(t *testing.T) {
	utxoIndex := NewUTXOIndex(NewUTXOCache(storage.NewRamStorage()))
	utxoCopy := utxoIndex.DeepCopy()
	assert.Equal(t, 0, len(utxoIndex.index))
	assert.Equal(t, 0, len(utxoCopy.index))

	addr1UtxoTx := NewUTXOTx()
	utxoIndex.index[string(address1Hash)] = &addr1UtxoTx
	assert.Equal(t, 1, len(utxoIndex.index))
	assert.Equal(t, 0, len(utxoCopy.index))

	copyUtxoTx := NewUTXOTxWithData(&UTXO{MockUtxoOutputsWithoutInputs()[0], []byte{}, 0, UtxoNormal})
	utxoCopy.index[string(address1Hash)] = &copyUtxoTx
	assert.Equal(t, 1, len(utxoIndex.index))
	assert.Equal(t, 1, len(utxoCopy.index))
	assert.Equal(t, 0, utxoIndex.index[string(address1Hash)].Size())
	assert.Equal(t, 1, utxoCopy.index[string(address1Hash)].Size())

	copyUtxoTx1 := NewUTXOTx()
	copyUtxoTx1 = copyUtxoTx1.PutUtxo(&UTXO{MockUtxoOutputsWithoutInputs()[0], []byte{}, 0, UtxoNormal})
	copyUtxoTx1 = copyUtxoTx1.PutUtxo(&UTXO{MockUtxoOutputsWithoutInputs()[1], []byte{}, 1, UtxoNormal})
	utxoCopy.index["1"] = &copyUtxoTx1

	utxoCopy2 := utxoCopy.DeepCopy()
	copy2UtxoTx1 := NewUTXOTx()
	copy2UtxoTx1 = copy2UtxoTx1.PutUtxo(&UTXO{MockUtxoOutputsWithoutInputs()[0], []byte{}, 0, UtxoNormal})
	utxoCopy2.index["1"] = &copy2UtxoTx1
	assert.Equal(t, 2, len(utxoCopy.index))
	assert.Equal(t, 2, len(utxoCopy2.index))
	assert.Equal(t, 2, utxoCopy.index["1"].Size())
	assert.Equal(t, 1, utxoCopy2.index["1"].Size())
	assert.Equal(t, 1, len(utxoIndex.index))

	assert.EqualValues(t, utxoCopy.index[hex.EncodeToString(address1Hash)], utxoCopy2.index[hex.EncodeToString(address1Hash)])
}
