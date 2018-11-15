package sc

import (
	"testing"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestScEngine_Execute(t *testing.T) {
	script := `'use strict';

var AddrChecker = function(){
	
};

AddrChecker.prototype = {
		check:function(addr,dummy){
    	return Blockchain.verifyAddress(addr)+dummy;
    }
};
var addrChecker = new AddrChecker;
`

	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	assert.Equal(t, "35", sc.Execute("check", "\"dastXXWLe5pxbRYFhcyUq8T3wb5srWkHKa\",34"))
}

func TestScEngine_BlockchainTransfer(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	script := `'use strict';
var TransferTest = function(){};
TransferTest.prototype = {
    transfer: function(to, amount, tip){
        return Blockchain.transfer(to, amount, tip);
    }
};
var transferTest = new TransferTest;`

	contractPubKeyHash := core.NewContractPubKeyHash()
	contractAddr := contractPubKeyHash.GenerateAddress()
	contractUTXOs := []*core.UTXO{
		{
			Txid:     []byte("1"),
			TxIndex:  1,
			TXOutput: *core.NewTxOut(common.NewAmount(15), contractAddr.String(), ""),
		},
		{
			Txid:     []byte("2"),
			TxIndex:  0,
			TXOutput: *core.NewTxOut(common.NewAmount(3), contractAddr.String(), ""),
		},
	}

	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportContractAddr(contractAddr)
	sc.ImportSourceTXID([]byte("thatTX"))
	sc.ImportUTXOs(contractUTXOs)
	result := sc.Execute("transfer", "'16PencPNnF8CiSx2EBGEd1axhf7vuHCouj','10','2'")

	assert.Equal(t, "0", result)
	if assert.Equal(t, 1, len(sc.generatedTXs)) {
		if assert.Equal(t, 1, len(sc.generatedTXs[0].Vin)) {
			assert.Equal(t, []byte("1"), sc.generatedTXs[0].Vin[0].Txid)
			assert.Equal(t, 1, sc.generatedTXs[0].Vin[0].Vout)
			assert.Equal(t, []byte("thatTX"), sc.generatedTXs[0].Vin[0].Signature)
			assert.Equal(t, contractPubKeyHash.GetPubKeyHash(), sc.generatedTXs[0].Vin[0].PubKey)
		}
		if assert.Equal(t, 2, len(sc.generatedTXs[0].Vout)) {
			// payout
			assert.Equal(t, common.NewAmount(10), sc.generatedTXs[0].Vout[0].Value)
			// change
			assert.Equal(t, common.NewAmount(15-10-2), sc.generatedTXs[0].Vout[1].Value)

			assert.Equal(t, core.HashAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj"), sc.generatedTXs[0].Vout[0].PubKeyHash.GetPubKeyHash())
			assert.Equal(t, contractPubKeyHash.GetPubKeyHash(), sc.generatedTXs[0].Vout[1].PubKeyHash.GetPubKeyHash())
		}
	}
}

func TestScEngine_StorageGet(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	script := `'use strict';

var StorageTest = function(){
	
};

StorageTest.prototype = {
	set:function(key,value){
    	return LocalStorage.set(key,value);
    },
	get:function(key){
    	return LocalStorage.get(key);
    }
};
var storageTest = new StorageTest;
`
	ss := make(map[string]string)
	ss["key"] = "7"
	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportLocalStorage(ss)
	assert.Equal(t, "7", sc.Execute("get", "\"key\""))
}

func TestScEngine_StorageSet(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	script := `'use strict';

var StorageTest = function(){
	
};

StorageTest.prototype = {
	set:function(key,value){
    	return LocalStorage.set(key,value);
    },
	get:function(key){
    	return LocalStorage.get(key);
    }
};
var storageTest = new StorageTest;
`
	ss := make(map[string]string)
	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportLocalStorage(ss)

	assert.Equal(t, "0", sc.Execute("set", "\"key\",6"))
	assert.Equal(t, "6", sc.Execute("get", "\"key\""))
}

func TestScEngine_StorageDel(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	script := `'use strict';

var StorageTest = function(){
	
};

StorageTest.prototype = {
	set:function(key,value){
    	return LocalStorage.set(key,value);
    },
	get:function(key){
    	return LocalStorage.get(key);
    },
	del:function(key){
    	return LocalStorage.del(key);
    }
};
var storageTest = new StorageTest;
`
	ss := make(map[string]string)
	sc := NewV8Engine()
	sc.ImportSourceCode(script)
	sc.ImportLocalStorage(ss)
	assert.Equal(t, "0", sc.Execute("set", "\"key\",6"))
	assert.Equal(t, "0", sc.Execute("del", "\"key\""))
	assert.Equal(t, "null", sc.Execute("get", "\"key\""))
}
