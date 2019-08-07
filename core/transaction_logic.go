package core

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	logger "github.com/sirupsen/logrus"
)

// VerifyInEstimate returns whether the current tx in estimate mode is valid.
func VerifyInEstimate(utxoIndex *UTXOIndex, ctx *ContractTx) error {
	if ctx.IsExecutionContract() && !ctx.IsContractDeployed(utxoIndex) {
		return errors.New("Transaction: contract state check failed")
	}

	_, err := verify(&ctx.Transaction, utxoIndex)
	if err != nil {
		return err
	}
	return nil
}

// VerifyContractTx ensures signature of transactions is correct or verifies against blockHeight if it's a coinbase transactions
func VerifyContractTx(utxoIndex *UTXOIndex, ctx *ContractTx) (bool, error) {
	if ctx.IsExecutionContract() && !ctx.IsContractDeployed(utxoIndex) {
		return false, errors.New("Transaction: contract state check failed")
	}

	totalBalance, err := verify(&ctx.Transaction, utxoIndex)
	if err != nil {
		return false, err
	}
	return ctx.verifyGas(totalBalance)
}

// VerifyTransaction ensures signature of transactions is correct or verifies against blockHeight if it's a coinbase transactions
func VerifyTransaction(utxoIndex *UTXOIndex, tx *Transaction, blockHeight uint64) (bool, error) {
	ctx := tx.ToContractTx()
	if ctx != nil {
		return VerifyContractTx(utxoIndex, ctx)
	}
	if tx.IsCoinbase() {
		//TODO coinbase vout check need add tip
		if tx.Vout[0].Value.Cmp(subsidy) < 0 {
			return false, errors.New("Transaction: subsidy check failed")
		}
		bh := binary.BigEndian.Uint64(tx.Vin[0].Signature)
		if blockHeight != bh {
			return false, fmt.Errorf("Transaction: block height check failed expected=%v actual=%v", blockHeight, bh)
		}
		return true, nil
	}
	if tx.IsRewardTx() || tx.IsGasRewardTx() || tx.IsGasChangeTx() {
		//TODO: verify reward tx here
		return true, nil
	}

	_, err := verify(tx, utxoIndex)
	if err != nil {
		return false, err
	}
	return true, nil
}

func verify(tx *Transaction, utxoIndex *UTXOIndex) (*common.Amount, error) {
	prevUtxos := getPrevUTXOs(tx, utxoIndex)
	if prevUtxos == nil {
		return nil, errors.New("Transaction: prevUtxos not found")
	}
	result, err := tx.verifyID()
	if !result {
		return nil, err
	}

	result, err = verifyPublicKeyHash(prevUtxos, tx)
	if !result {
		return nil, err
	}

	totalPrev := calculateUtxoSum(prevUtxos)
	totalVoutValue, ok := tx.calculateTotalVoutValue()
	if !ok {
		return nil, errors.New("Transaction: vout is invalid")
	}
	result, err = tx.verifyAmount(totalPrev, totalVoutValue)
	if !result {
		return nil, err
	}
	result, err = tx.verifyTip(totalPrev, totalVoutValue)
	if !result {
		logger.WithFields(logger.Fields{
			"tx_id": hex.EncodeToString(tx.ID),
		}).Warn("Transaction: tip is invalid.")
		return nil, err
	}
	result, err = tx.verifySignatures(prevUtxos)
	if !result {
		return nil, err
	}
	totalBalance, _ := totalPrev.Sub(totalVoutValue)
	totalBalance, _ = totalBalance.Sub(tx.Tip)
	return totalBalance, nil
}

// DescribeTransaction reverse-engineers the high-level description of a transaction
func DescribeTransaction(utxoIndex *UTXOIndex, tx *Transaction) (sender, recipient *account.Address, amount, tip *common.Amount, error error) {
	var receiverAddress account.Address
	vinPubKey := tx.Vin[0].PubKey
	pubKeyHash := account.PubKeyHash([]byte(""))
	inputAmount := common.NewAmount(0)
	outputAmount := common.NewAmount(0)
	payoutAmount := common.NewAmount(0)
	for _, vin := range tx.Vin {
		if bytes.Compare(vin.PubKey, vinPubKey) == 0 {
			switch {
			case tx.IsRewardTx():
				pubKeyHash = account.PubKeyHash(rewardTxData)
				continue
			case tx.IsFromContract(utxoIndex):
				// vinPubKey is the pubKeyHash if it is a sc generated tx
				pubKeyHash = account.PubKeyHash(vinPubKey)
			default:
				pkh, err := account.NewUserPubKeyHash(vin.PubKey)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				pubKeyHash = pkh
			}
			usedUTXO := utxoIndex.FindUTXOByVin([]byte(pubKeyHash), vin.Txid, vin.Vout)
			inputAmount = inputAmount.Add(usedUTXO.Value)
		} else {
			logger.Debug("Transaction: using UTXO from multiple accounts.")
		}
	}
	for _, vout := range tx.Vout {
		if bytes.Compare([]byte(vout.PubKeyHash), vinPubKey) == 0 {
			outputAmount = outputAmount.Add(vout.Value)
		} else {
			receiverAddress = vout.PubKeyHash.GenerateAddress()
			payoutAmount = payoutAmount.Add(vout.Value)
		}
	}
	tip, err := inputAmount.Sub(outputAmount)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	senderAddress := pubKeyHash.GenerateAddress()

	return &senderAddress, &receiverAddress, payoutAmount, tip, nil
}

// Returns related previous UTXO for current transaction
func getPrevUTXOs(tx *Transaction, utxoIndex *UTXOIndex) []*UTXO {
	var prevUtxos []*UTXO
	tempUtxoTxMap := make(map[string]*UTXOTx)
	for _, vin := range tx.Vin {
		pubKeyHash, err := account.NewUserPubKeyHash(vin.PubKey)
		if err != nil {
			logger.WithFields(logger.Fields{
				"tx_id":          hex.EncodeToString(tx.ID),
				"vin_tx_id":      hex.EncodeToString(vin.Txid),
				"vin_public_key": hex.EncodeToString(vin.PubKey),
			}).Warn("Transaction: failed to get PubKeyHash of vin.")
			return nil
		}
		tempUtxoTx, ok := tempUtxoTxMap[string(pubKeyHash)]
		if !ok {
			tempUtxoTx = utxoIndex.GetAllUTXOsByPubKeyHash(pubKeyHash)
			tempUtxoTxMap[string(pubKeyHash)] = tempUtxoTx
		}
		utxo := tempUtxoTx.GetUtxo(vin.Txid, vin.Vout)
		if utxo == nil {
			logger.WithFields(logger.Fields{
				"tx_id":      hex.EncodeToString(tx.ID),
				"vin_tx_id":  hex.EncodeToString(vin.Txid),
				"vin_index":  vin.Vout,
				"pubKeyHash": pubKeyHash.String(),
			}).Warn("Transaction: cannot find vin.")
			return nil
		}
		prevUtxos = append(prevUtxos, utxo)
	}
	return prevUtxos
}

//verifyPublicKeyHash verifies if the public key in Vin is the original key for the public
//key hash in utxo
func verifyPublicKeyHash(prevUtxos []*UTXO, tx *Transaction) (bool, error) {

	for i, vin := range tx.Vin {
		if prevUtxos[i].PubKeyHash == nil {
			logger.Error("Transaction: previous transaction is not correct.")
			return false, errors.New("Transaction: prevUtxos not found")
		}

		isContract, err := prevUtxos[i].PubKeyHash.IsContract()
		if err != nil {
			return false, err
		}
		//if the utxo belongs to a Contract, the utxo is not verified through
		//public key hash. It will be verified through consensus
		if isContract {
			continue
		}

		pubKeyHash, err := account.NewUserPubKeyHash(vin.PubKey)
		if err != nil {
			return false, err
		}
		if !bytes.Equal([]byte(pubKeyHash), []byte(prevUtxos[i].PubKeyHash)) {
			return false, errors.New("Transaction: ID is invalid")
		}
	}
	return true, nil
}
