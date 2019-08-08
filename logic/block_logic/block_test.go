package block_logic

import (
	"encoding/hex"
	"fmt"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/transaction_base"
	"github.com/dappley/go-dappley/core/utxo"
	"github.com/dappley/go-dappley/logic/transaction_logic"
	"github.com/dappley/go-dappley/logic/utxo_logic"
	"sync"
	"testing"
	"time"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/util"
	"github.com/stretchr/testify/assert"
)

func TestHashTransactions(t *testing.T) {

	var parentBlk = block.NewBlockWithRawInfo(
		[]byte{'a'},
		[]byte{'e', 'c'},
		0,
		time.Now().Unix(),
		0,
		nil,
	)

	var expectHash = []uint8([]byte{0x5d, 0xf6, 0xe0, 0xe2, 0x76, 0x13, 0x59, 0xd3, 0xa, 0x82, 0x75, 0x5, 0x8e, 0x29, 0x9f, 0xcc, 0x3, 0x81, 0x53, 0x45, 0x45, 0xf5, 0x5c, 0xf4, 0x3e, 0x41, 0x98, 0x3f, 0x5d, 0x4c, 0x94, 0x56})

	blk := block.NewBlock([]*transaction.Transaction{{}}, parentBlk, "")
	hash := HashTransactions(blk)
	assert.Equal(t, expectHash, hash)
}

func TestBlock_VerifyHash(t *testing.T) {
	b1 := core.GenerateMockBlock()

	//The mocked block does not have correct hash Value
	assert.False(t, VerifyHash(b1))

	//calculate correct hash Value
	hash := CalculateHash(b1)
	b1.SetHash(hash)
	assert.True(t, VerifyHash(b1))

	//calculate a hash Value with a different nonce
	b1.SetNonce(b1.GetNonce() + 1)
	hash = CalculateHashWithNonce(b1)
	b1.SetHash(hash)
	assert.False(t, VerifyHash(b1))

	hash = CalculateHashWithoutNonce(b1)
	b1.SetHash(hash)
	assert.False(t, VerifyHash(b1))
}

func TestCalculateHashWithNonce(t *testing.T) {
	var parentBlk = block.NewBlockWithRawInfo(
		[]byte{'a'},
		[]byte{'e', 'c'},
		0,
		0,
		0,
		nil,
	)

	blk := block.NewBlock([]*transaction.Transaction{{}}, parentBlk, "")
	blk.SetTimestamp(0)
	expectHash1 := hash.Hash{0x3f, 0x2f, 0xec, 0xb4, 0x33, 0xf0, 0xd1, 0x1a, 0xa6, 0xf4, 0xf, 0xb8, 0x7f, 0x8f, 0x99, 0x11, 0xae, 0xe7, 0x42, 0xf4, 0x69, 0x7d, 0xf1, 0xaa, 0xc8, 0xd0, 0xfc, 0x40, 0xa2, 0xd8, 0xb1, 0xa5}
	blk.SetNonce(1)
	assert.Equal(t, hash.Hash(expectHash1), CalculateHashWithNonce(blk))
	expectHash2 := hash.Hash{0xe7, 0x57, 0x13, 0xc6, 0x8a, 0x98, 0x58, 0xb3, 0x5, 0x70, 0x6e, 0x33, 0xf0, 0x95, 0xd8, 0x1a, 0xbc, 0x76, 0xef, 0x30, 0x14, 0x59, 0x88, 0x11, 0x3c, 0x11, 0x59, 0x92, 0x65, 0xd5, 0xd3, 0x4c}
	blk.SetNonce(2)
	assert.Equal(t, hash.Hash(expectHash2), CalculateHashWithNonce(blk))
}

func TestBlock_VerifyTransactions(t *testing.T) {
	// Prepare test data
	// Padding Address to 32 Byte
	var address1Bytes = []byte("address1000000000000000000000000")
	var address1Hash, _ = account.NewUserPubKeyHash(address1Bytes)

	normalCoinbaseTX := transaction_logic.NewCoinbaseTX(address1Hash.GenerateAddress(), "", 1, common.NewAmount(0))
	rewardTX := transaction.NewRewardTx(1, map[string]string{address1Hash.GenerateAddress().String(): "10"})
	userPubKey := account.NewKeyPair().GetPublicKey()
	userPubKeyHash, _ := account.NewUserPubKeyHash(userPubKey)
	userAddr := userPubKeyHash.GenerateAddress()
	contractPubKeyHash := account.NewContractPubKeyHash()
	contractAddr := contractPubKeyHash.GenerateAddress()

	txIdStr := "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa71"
	generatedTxId, err := hex.DecodeString(txIdStr)
	assert.Nil(t, err)
	fmt.Println(hex.EncodeToString(generatedTxId))
	generatedTX := &transaction.Transaction{
		generatedTxId,
		[]transaction_base.TXInput{
			{[]byte("prevtxid"), 0, []byte("txid"), []byte(contractPubKeyHash)},
			{[]byte("prevtxid"), 1, []byte("txid"), []byte(contractPubKeyHash)},
		},
		[]transaction_base.TXOutput{
			*transaction_base.NewTxOut(common.NewAmount(23), userAddr, ""),
			*transaction_base.NewTxOut(common.NewAmount(10), contractAddr, ""),
		},
		common.NewAmount(7),
		common.NewAmount(0),
		common.NewAmount(0),
	}

	var prikey1 = "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa71"
	var pubkey1 = account.GenerateKeyPairByPrivateKey(prikey1).GetPublicKey()
	var pkHash1, _ = account.NewUserPubKeyHash(pubkey1)
	var prikey2 = "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa72"
	var pubkey2 = account.GenerateKeyPairByPrivateKey(prikey2).GetPublicKey()
	var pkHash2, _ = account.NewUserPubKeyHash(pubkey2)

	dependentTx1 := transaction.NewTransactionByVin(util.GenerateRandomAoB(1), 1, pubkey1, 10, pkHash2, 3)
	dependentTx2 := transaction.NewTransactionByVin(dependentTx1.ID, 0, pubkey2, 5, pkHash1, 5)
	dependentTx3 := transaction.NewTransactionByVin(dependentTx2.ID, 0, pubkey1, 1, pkHash2, 4)

	tx2Utxo1 := utxo.UTXO{dependentTx2.Vout[0], dependentTx2.ID, 0, utxo.UtxoNormal}

	tx1Utxos := map[string][]*utxo.UTXO{
		pkHash2.String(): {&utxo.UTXO{dependentTx1.Vout[0], dependentTx1.ID, 0, utxo.UtxoNormal}},
	}
	transaction_logic.Sign(account.GenerateKeyPairByPrivateKey(prikey2).GetPrivateKey(), tx1Utxos[pkHash2.String()], &dependentTx2)
	transaction_logic.Sign(account.GenerateKeyPairByPrivateKey(prikey1).GetPrivateKey(), []*utxo.UTXO{&tx2Utxo1}, &dependentTx3)

	tests := []struct {
		name  string
		txs   []*transaction.Transaction
		utxos map[string][]*utxo.UTXO
		ok    bool
	}{
		{
			"normal txs",
			[]*transaction.Transaction{&normalCoinbaseTX},
			map[string][]*utxo.UTXO{},
			true,
		},
		{
			"no txs",
			[]*transaction.Transaction{},
			make(map[string][]*utxo.UTXO),
			true,
		},
		{
			"invalid normal txs",
			[]*transaction.Transaction{{
				ID: []byte("txid"),
				Vin: []transaction_base.TXInput{{
					[]byte("tx1"),
					0,
					util.GenerateRandomAoB(2),
					address1Bytes,
				}},
				Vout: core.MockUtxoOutputsWithInputs(),
				Tip:  common.NewAmount(5),
			}},
			map[string][]*utxo.UTXO{},
			false,
		},
		{
			"normal dependent txs",
			[]*transaction.Transaction{&dependentTx2, &dependentTx3},
			tx1Utxos,
			true,
		},
		{
			"invalid dependent txs",
			[]*transaction.Transaction{&dependentTx3, &dependentTx2},
			tx1Utxos,
			false,
		},
		{
			"reward tx",
			[]*transaction.Transaction{&rewardTX},
			map[string][]*utxo.UTXO{
				contractPubKeyHash.String(): {
					{*transaction_base.NewTXOutput(common.NewAmount(0), contractAddr), []byte("prevtxid"), 0, utxo.UtxoNormal},
				},
				userPubKeyHash.String(): {
					{*transaction_base.NewTXOutput(common.NewAmount(1), userAddr), []byte("txinid"), 0, utxo.UtxoNormal},
				},
			},
			false,
		},
		{
			"generated tx",
			[]*transaction.Transaction{generatedTX},
			map[string][]*utxo.UTXO{
				contractPubKeyHash.String(): {
					{*transaction_base.NewTXOutput(common.NewAmount(20), contractAddr), []byte("prevtxid"), 0, utxo.UtxoNormal},
					{*transaction_base.NewTXOutput(common.NewAmount(20), contractAddr), []byte("prevtxid"), 1, utxo.UtxoNormal},
				},
				userPubKeyHash.String(): {
					{*transaction_base.NewTXOutput(common.NewAmount(1), userAddr), []byte("txinid"), 0, utxo.UtxoNormal},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := storage.NewRamStorage()
			index := make(map[string]*utxo.UTXOTx)

			for key, addrUtxos := range tt.utxos {
				utxoTx := utxo.NewUTXOTx()
				for _, addrUtxo := range addrUtxos {
					utxoTx.PutUtxo(addrUtxo)
				}
				index[key] = &utxoTx
			}

			utxoIndex := utxo_logic.UTXOIndex{index, utxo.NewUTXOCache(db), &sync.RWMutex{}}
			scState := core.NewScState()
			var parentBlk = block.NewBlockWithRawInfo(
				[]byte{'a'},
				[]byte{'e', 'c'},
				0,
				time.Now().Unix(),
				0,
				nil,
			)
			blk := block.NewBlock(tt.txs, parentBlk, "")
			assert.Equal(t, tt.ok, VerifyTransactions(blk, &utxoIndex, scState, nil, parentBlk))
		})
	}
}
