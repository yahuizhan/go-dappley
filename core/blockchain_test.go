// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
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
	"testing"
	"github.com/dappley/go-dappley/storage"
	"github.com/stretchr/testify/assert"
	"os"
	logger "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logger.SetLevel(logger.WarnLevel)
	retCode := m.Run()
	os.Exit(retCode)
}

func TestCreateBlockchain(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	addr := NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	bc:= CreateBlockchain(addr, s,nil)

	//find next block. This block should be the genesis block and its prev hash should be empty
	blk,err := bc.Next()
	assert.Nil(t, err)
	assert.Empty(t, blk.GetPrevHash())
}

func TestBlockchain_HigherThanBlockchainTestHigher(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	addr := NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	bc:= CreateBlockchain(addr, s,nil)
	blk := GenerateMockBlock()
	blk.height = 1
	assert.True(t,bc.HigherThanBlockchain(blk))
}

func TestBlockchain_HigherThanBlockchainTestLower(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	addr := NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	bc:= CreateBlockchain(addr, s,nil)

	blk := GenerateMockBlock()
	blk.height = 1
	bc.UpdateNewBlock(blk)

	assert.False(t,bc.HigherThanBlockchain(blk))
}

func TestBlockchain_IsInBlockchain(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	addr := NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	bc:= CreateBlockchain(addr, s,nil)

	blk := GenerateMockBlock()
	blk.SetHash([]byte("hash1"))
	blk.height = 1
	bc.UpdateNewBlock(blk)

	isFound := bc.IsInBlockchain([]byte("hash1"))
	assert.True(t,isFound)

	isFound = bc.IsInBlockchain([]byte("hash2"))
	assert.False(t,isFound)
}

func TestBlockchain_RollbackToABlock(t *testing.T) {
	//create a mock blockchain with max height of 5
	bc := GenerateMockBlockchain(5)
	defer bc.db.Close()

	blk,err := bc.GetTailBlock()
	assert.Nil(t,err)

	//find the hash at height 3 (5-2)
	for i:=0; i<2; i++{
		blk,err = bc.GetBlockByHash(blk.GetPrevHash())
		assert.Nil(t,err)
	}

	//rollback to height 3
	bc.RollbackToABlockHeight(blk.GetHash())

	//the height 3 block should be the new tail block
	newTailBlk,err := bc.GetTailBlock()
	assert.Nil(t,err)
	assert.Equal(t,blk.GetHash(),newTailBlk.GetHash())

}

func TestBlockchain_ConcatenateForkToBlockchain(t *testing.T) {

	//mock a blockchain and a fork whose parent is the tail of the blockchain
	bc := GenerateMockBlockchain(5)
	defer bc.db.Close()
	tailBlk,err:= bc.GetTailBlock()
	assert.Nil(t, err)
	bc.GetBlockPool().forkPool = GenerateMockFork(5,tailBlk)
	forkTailBlockHash := bc.GetBlockPool().forkPool[0].GetHash()

	//add the fork to the end of the blockchain
	bc.concatenateForkToBlockchain()
	//the highest block should have the height of 10
	assert.Equal(t, uint64(10), bc.GetMaxHeight())
	tailBlkHash,err := bc.GetTailHash()
	assert.Nil(t, err)
	assert.ElementsMatch(t,forkTailBlockHash,tailBlkHash)

}

func TestBlockchain_MergeForkCoinbaseTxOnly(t *testing.T) {
	//mock a blockchain and a fork whose parent is the tail of the blockchain
	bc := GenerateMockBlockchainWithCoinbaseTxOnly(5)
	defer bc.db.Close()
	blk,err:= bc.GetTailBlock()
	assert.Nil(t, err)

	//find the hash at height 3 (5-2)
	for i:=0; i<2; i++{
		blk,err = bc.GetBlockByHash(blk.GetPrevHash())
		assert.Nil(t,err)
	}

	//generate a fork that is forked from height 3
	bc.GetBlockPool().forkPool = GenerateMockFork(5,blk)

	//get the last fork hash
	forkTailBlockHash := bc.GetBlockPool().forkPool[0].GetHash()

	bc.MergeFork()

	//the highest block should have the height of 8 -> 3+5
	assert.Equal(t, uint64(8), bc.GetMaxHeight())
	tailBlkHash,err := bc.GetTailHash()
	assert.Nil(t, err)
	assert.ElementsMatch(t,forkTailBlockHash,tailBlkHash)

}

func TestBlockchain_MergeForkInvalidTransaction(t *testing.T) {
	//mock a blockchain and a fork whose parent is the tail of the blockchain
	bc := GenerateMockBlockchainWithCoinbaseTxOnly(5)
	defer bc.db.Close()
	blk,err:= bc.GetTailBlock()
	assert.Nil(t, err)

	//find the hash at height 3 (5-2)
	for i:=0; i<2; i++{
		blk,err = bc.GetBlockByHash(blk.GetPrevHash())
		assert.Nil(t,err)
	}

	tailBlkHash,err := bc.GetTailHash()
	assert.Nil(t, err)

	//generate a fork that is forked from height 3
	bc.GetBlockPool().forkPool = GenerateMockForkWithInvalidTx(5,blk)

	//the merge should fail since the transactions are invalid
	bc.MergeFork()

	//the highest block should have the height of 5
	assert.Equal(t, uint64(5), bc.GetMaxHeight())
	newTailBlkHash,err := bc.GetTailHash()
	assert.Nil(t, err)
	assert.ElementsMatch(t,tailBlkHash,newTailBlkHash)
}