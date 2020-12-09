package util_ron

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/transactionbase"
	"github.com/dappley/go-dappley/core/utxo"
	utxopb "github.com/dappley/go-dappley/core/utxo/pb"
	"github.com/dappley/go-dappley/logic/ltransaction"
	rpcpb "github.com/dappley/go-dappley/rpc/pb"
	account_ron "github.com/dappley/go-dappley/tool/dependable_testing/sdk"
	logger "github.com/sirupsen/logrus"
)

func CreateTransaction(respon *rpcpb.GetUTXOResponse, acc *account_ron.AccountInfo, amount uint64, fromAccount, toAccount string) (transaction.Transaction, error) {
	//从服务器返回的utxo集合里找到满足转账所需金额的utxo
	tx_utxos, err := getUTXOsWithAmount(
		respon.GetUtxos(),
		common.NewAmount(amount),
		common.NewAmount(0),
		common.NewAmount(0),
		common.NewAmount(0))
	if err != nil {
		logger.Error("Error:", err.Error())
		return transaction.Transaction{}, err
	}
	//组装交易参数
	sendTxParam := transaction.NewSendTxParam(
		account.NewAddress(fromAccount),
		acc.GetAccount(fromAccount).GetKeyPair(),
		account.NewAddress(toAccount),
		common.NewAmount(amount),
		common.NewAmount(0),
		common.NewAmount(0),
		common.NewAmount(0),
		"")

	return ltransaction.NewUTXOTransaction(tx_utxos, sendTxParam)
}

func CreateTransactionByUTXOs(utxoTx *utxo.UTXOTx, acc *account_ron.AccountInfo, amount uint64, fromAccount, toAccount string,
		tip uint64, gasLimit uint64, gasPrice uint64, contract string) (transaction.Transaction, error) {
	//从服务器返回的utxo集合里找到满足转账所需金额的utxo
	tx_utxos, err := getUTXOsWithAmountByUTXOs(
		utxoTx,
		common.NewAmount(amount),
		common.NewAmount(tip),
		common.NewAmount(gasLimit),
		common.NewAmount(gasPrice))
	if err != nil {
		acc.Lock()
		acc.Balances[fromAccount] = 0
		acc.Unlock()
		logger.Error("Error:", err.Error())
		return transaction.Transaction{}, err
	}
	//组装交易参数
	sendTxParam := transaction.NewSendTxParam(
		account.NewAddress(fromAccount),
		acc.GetAccount(fromAccount).GetKeyPair(),
		account.NewAddress(toAccount),
		common.NewAmount(amount),
		common.NewAmount(tip),
		common.NewAmount(gasLimit),
		common.NewAmount(gasPrice),
		contract)

	return ltransaction.NewUTXOTransaction(tx_utxos, sendTxParam) //这里会去计算找零
}

//从utxo集合里拿到足够交易的utxo集合
func getUTXOsWithAmountByUTXOs(utxoTx *utxo.UTXOTx, amount *common.Amount, tip *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount) ([]*utxo.UTXO, error) {
	if tip != nil {
		amount = amount.Add(tip)
	}
	if gasLimit != nil {
		limitedFee := gasLimit.Mul(gasPrice)
		amount = amount.Add(limitedFee)
	}

	var retUtxos []*utxo.UTXO
	sum := common.NewAmount(0)
	for _, u := range utxoTx.Indices {
		sum = sum.Add(u.Value)
		retUtxos = append(retUtxos, u)
		if sum.Cmp(amount) >= 0 {
			break
		}
	}

	if sum.Cmp(amount) < 0 {
		fmt.Println("sum:",sum)
		fmt.Println("amount:",amount)
		return nil, errors.New("util: the balance is insufficient")
	}

	return retUtxos, nil
}

//从服务器返回的utxo集合里找到满足转账所需金额的utxo
func getUTXOsWithAmount(responUtxos []*utxopb.Utxo, amount *common.Amount, tip *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount) ([]*utxo.UTXO, error) {
	//得到Utxo集合
	var inputUtxos []*utxo.UTXO
	for _, u := range responUtxos {
		utxo := utxo.UTXO{}
		utxo.Value = common.NewAmountFromBytes(u.Amount)
		utxo.Txid = u.Txid
		utxo.PubKeyHash = account.PubKeyHash(u.PublicKeyHash)
		utxo.TxIndex = int(u.TxIndex)
		inputUtxos = append(inputUtxos, &utxo)
	}

	if tip != nil {
		amount = amount.Add(tip)
	}
	if gasLimit != nil {
		limitedFee := gasLimit.Mul(gasPrice)
		amount = amount.Add(limitedFee)
	}

	var retUtxos []*utxo.UTXO
	sum := common.NewAmount(0)
	for _, u := range inputUtxos {
		sum = sum.Add(u.Value)
		retUtxos = append(retUtxos, u)
		if sum.Cmp(amount) >= 0 {
			break
		}
	}

	if sum.Cmp(amount) < 0 {
		return nil, errors.New("cli: the balance is insufficient")
	}

	return retUtxos, nil
}

func UpdateUTXOs(pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, tx *transaction.Transaction) {
	adaptedTx := transaction.NewTxAdapter(tx)
	if adaptedTx.IsNormal() || adaptedTx.IsContract() || adaptedTx.IsContractSend() {
		//把这笔交易中所有的vin全都从utxo中删除
		for _, txin := range tx.Vin {
			//是否为智能合约
			isContract, _ := account.PubKeyHash(txin.PubKey).IsContract()
			// spent contract utxo
			//pubKeyHash := txin.PubKey
			if !isContract {
				// spent normal utxo
				//通过txin的公钥生成新的交易地址和公钥哈希
				//ta := account.NewTransactionAccountByPubKey(txin.PubKey)
				_, err := account.IsValidPubKey(txin.PubKey)
				if err != nil {
					logger.WithError(err).Warn("UTXOIndex: txin.pubKey error, discard update in utxo.")
					//return false
				}
				//pubKeyHash = ta.GetPubKeyHash()
			}
			//从utxos中把单笔utxo移除
			err := removeUTXO(utxos, txin.Txid, txin.Vout)
			if err != nil {
				logger.WithError(err).Warn("UTXOIndex: removeUTXO error, discard update in utxo.")
				//return false
			}
		}
	}
	//把这笔交易中所有的Vout添加到utxos中
	for i, txout := range tx.Vout {
		AddUTXO(pubkeyHash, utxos, txout, tx.ID, i)
	}
}
//// removeUTXO finds and removes a UTXO from UTXOIndex
//func (utxos *UTXOIndex) removeUTXO(pkh account.PubKeyHash, txid []byte, vout int) error {
//	originalUtxos := utxos.GetAllUTXOsByPubKeyHash(pkh)
//
//	u := originalUtxos.GetUtxo(txid, vout)
//	if u == nil {
//		return ErrUTXONotFound
//	}
//
//	utxos.mutex.Lock()
//	originalUtxos.RemoveUtxo(txid, vout)
//	utxos.index[pkh.String()] = originalUtxos
//	utxos.mutex.Unlock()
//
//	if u.UtxoType != utxo.UtxoCreateContract {
//		return nil
//	}
//	// remove contract utxos
//	isContract, _ := pkh.IsContract()
//	if isContract {
//		contractUtxos := utxos.GetAllUTXOsByPubKeyHash(contractUtxoKey)
//
//		contractUtxo := contractUtxos.GetUtxo(txid, vout)
//
//		if contractUtxo == nil {
//			return ErrUTXONotFound
//		}
//		utxos.mutex.Lock()
//		contractUtxos.RemoveUtxo(txid, vout)
//		utxos.index[hex.EncodeToString(contractUtxoKey)] = contractUtxos
//		delete(utxos.contractCreateIndex, contractUtxo.PubKeyHash.String())
//		utxos.mutex.Unlock()
//	}
//	return nil
//}

func removeUTXO(utxoTx *utxo.UTXOTx, txid []byte, vout int) error {
	u := utxoTx.GetUtxo(txid, vout)
	//检查出不存在，不存在就报错
	if u == nil {
		return errors.New("utxo not found when trying to remove from cache")
	}
	//移除utxos中的utxo
	utxoTx.RemoveUtxo(txid, vout)

	return nil
}

func AddUTXO(pubkeyHash account.PubKeyHash, utxoTx *utxo.UTXOTx, txout transactionbase.TXOutput, txid []byte, vout int) {
	//通过fromaccount 的哈希判断，如果不是他的vout就不加到utxo
	if !bytes.Equal(txout.PubKeyHash, pubkeyHash) {
		return
	}

	var u *utxo.UTXO
	//if it is a smart contract deployment utxo add it to contract utxos
	if isContract, _ := txout.PubKeyHash.IsContract(); isContract {
		//智能合约先不管
	} else {
		u = utxo.NewUTXO(txout, txid, vout, utxo.UtxoNormal)
	}
	utxoTx.PutUtxo(u)
}
