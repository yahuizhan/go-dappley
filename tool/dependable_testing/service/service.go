package service

import (
	"context"
	"fmt"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	transactionpb "github.com/dappley/go-dappley/core/transaction/pb"
	"github.com/dappley/go-dappley/core/utxo"
	rpcpb "github.com/dappley/go-dappley/rpc/pb"
	sdk_ron "github.com/dappley/go-dappley/tool/dependable_testing/sdk"
	util_ron "github.com/dappley/go-dappley/tool/dependable_testing/util"
	logger "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"time"
)

type Service struct {
	conn   rpcpb.RpcServiceClient
	connAd rpcpb.AdminServiceClient
}

func NewServiceClient(ip, port string) *Service {
	conn, err := grpc.Dial(fmt.Sprint(ip, ":", port), grpc.WithInsecure())
	if err != nil {
		logger.Error("网络异常", err)
		return nil
	}
	//defer conn.Close()
	return &Service{
		conn:   rpcpb.NewRpcServiceClient(conn),
		connAd: rpcpb.NewAdminServiceClient(conn),
	}
}

func (ser *Service) GetUtxo(accInfo *sdk_ron.AccountInfo, minnerAcc string, fromAcc string, amountFromminer uint64) *utxo.UTXOTx {
	//获取utxo，更新balance
	utxoTx,balance,err := ser.GetUTXOTxFromServer(fromAcc)
	if err != nil {
		logger.Error("Get UTXOTx error:", err)
	}
	accInfo.Lock()
	accInfo.Balances[fromAcc] = uint64(balance)
	accInfo.Unlock()
	//这里应该是通过计算utxo总和，更新,
	//accInfo.UpdateBalance(fromAcc, amountFromminer)
	return utxoTx

}

func (ser *Service) GetToken(accInfo *sdk_ron.AccountInfo, minnerAcc string, fromAcc string, amountFromminer uint64) *utxo.UTXOTx {
	var sum uint64
	for { //等待矿工有钱
		if sum > 500{ // 大于50s 就退出
			return nil
		}
		if ser.GetBalance(minnerAcc) > amountFromminer {
			break
		} else {
			time.Sleep(100 * time.Millisecond)
			sum ++
		}
	}
	ser.minerSendTokenToAccount(amountFromminer, fromAcc)

	for { //等待token到账
		if sum > 500{ // 大于50s 就退出
			return nil
		}
		if ser.GetBalance(fromAcc) >= amountFromminer {
			//这里更新账户余额
			//logger.Info(fromAcc,":",amountFromminer," token has been received.")
			break
		} else {
			time.Sleep(100 * time.Millisecond)
			sum ++
		} //这里应该再多一个判断，如果10秒还没到账，可能出错了，要再问矿工要钱
	}

	//获取utxo，更新balance
	utxoTx,balance,err := ser.GetUTXOTxFromServer(fromAcc)
	if err != nil {
		logger.Error("Get UTXOTx error:", err)
	}
	accInfo.Lock()
	accInfo.Balances[fromAcc] = uint64(balance)
	//fmt.Println("ser.GetBalance(fromAcc) amount:",ser.GetBalance(fromAcc),"fromAccount:",fromAcc)
	accInfo.Unlock()

	//这里应该是通过计算utxo总和，更新,
	//accInfo.UpdateBalance(fromAcc, amountFromminer)
	return utxoTx

}

//从矿工拿钱，得等到挖出一个快，矿工才有钱
func (ser *Service) minerSendTokenToAccount(amount uint64, account string) error{
	sendFromMinerRequest := &rpcpb.SendFromMinerRequest{To: account, Amount: common.NewAmount(amount).Bytes()}

	//通过句柄调用函数：rpc RpcSendFromMiner (SendFromMinerRequest) returns (SendFromMinerResponse) {}，
	_, err := ser.connAd.RpcSendFromMiner(context.Background(), sendFromMinerRequest) //SendFromMinerResponse里啥都没返，就不接收了
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Error:", err.Error())
		}
		return err
	}
	//logger.Info(amount, " has been sent to FromAddress.",time.Now().Format("2006-01-02 15:04:05"))
	logger.Info( "...")
	return nil
}

func (ser *Service) GetUTXOTxFromServer(fromAccount string) (*utxo.UTXOTx, int64, error) {
	//从服务器得到响应，包含指定账户地址的utxo信息
	response, err := ser.conn.RpcGetUTXO(context.Background(), &rpcpb.GetUTXORequest{
		Address: fromAccount})

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Error:", status.Convert(err).Message())
		}
		return nil,0, err
	}
	balance := common.NewAmount(0)
	utxoTx := utxo.NewUTXOTx()
	for _, u := range response.GetUtxos() {
		utxo := utxo.UTXO{}
		utxo.Value = common.NewAmountFromBytes(u.Amount)
		balance = balance.Add(utxo.Value)
		utxo.Txid = u.Txid
		utxo.PubKeyHash = account.PubKeyHash(u.PublicKeyHash)
		utxo.TxIndex = int(u.TxIndex)
		utxoTx.PutUtxo(&utxo) //组装成UTXOTx
	}
	//fmt.Println("utxo amount:",balance,"fromAccount:",fromAccount)
	return &utxoTx,balance.BigInt().Int64(),nil
}

//付款
func (ser *Service) DeploySmartContract(pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo,
		amount uint64, fromAccount, toAccount string,tip uint64, gasLimit uint64, gasPrice uint64, contract string)string {
	//创建交易
	tx, err := util_ron.CreateTransactionByUTXOs(utxos, accInfo, amount, fromAccount, toAccount,tip,gasLimit,gasPrice,contract)
	if err != nil {
		logger.Error("The transaction was abandoned.", err)
		return ""
	}
	//发送交易请求
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err = ser.conn.RpcSendTransaction(context.Background(), sendTransactionRequest)

	contractAddr := tx.Vout[0].GetAddress().String()
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Other error:", status.Convert(err).Message())
		}
		return ""
	}else {
		//logger.Println("达到预期结果 正常交易发送成功")
	}

	//这里更新部分是对的，因为更新好以后还是可以继续交易。
	//util_ron.UpdateUTXOs(pubkeyHash, utxos, &tx) //更新utxo
	sum := common.NewAmount(0)
	for _, u := range utxos.Indices {
		sum = sum.Add(u.Value)
	}

	txAmount := (amount + tip)+(gasLimit*gasPrice)
	accInfo.UpdateBalance(fromAccount, -txAmount)
	//accInfo.UpdateBalance(toAccount, amount)
	//logger.Info("New transaction is sent! ")
	return contractAddr
}
//付款
func (ser *Service) InvokingSmartContract(pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo,
	amount uint64, fromAccount, toAccount string,tip uint64, gasLimit uint64, gasPrice uint64, contract string) {
	//创建交易
	tx, err := util_ron.CreateTransactionByUTXOs(utxos, accInfo, amount, fromAccount, toAccount,tip,gasLimit,gasPrice,contract)
	if err != nil {
		logger.Error("The transaction was abandoned.", err)
		return
	}
	//发送交易请求
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err = ser.conn.RpcSendTransaction(context.Background(), sendTransactionRequest)

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Other error:", status.Convert(err).Message())
		}
		return
	}else {
		//logger.Println("达到预期结果 正常交易发送成功")
	}

	//这里更新部分是对的，因为更新好以后还是可以继续交易。
	util_ron.UpdateUTXOs(pubkeyHash, utxos, &tx) //更新utxo
	txAmount := (amount + tip)+(gasLimit*gasPrice)
	accInfo.UpdateBalance(fromAccount, -txAmount)
	accInfo.UpdateBalance(toAccount, amount)
	//logger.Info("New transaction is sent! ")
}
//付款
func (ser *Service) SendFalsifyContractToken(pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo, amount uint64, fromAccount, toAccount string) {

	//创建交易
	tx, err := util_ron.CreateTransactionByUTXOs(utxos, accInfo, amount, fromAccount, toAccount,0,0,0,"")
	if err != nil {
		logger.Error("The transaction was abandoned.", err)
		return
	}
	//篡改交易
	tx.Vout[0].Contract = "00000"
	//发送交易请求
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err = ser.conn.RpcSendTransaction(context.Background(), sendTransactionRequest)

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Other error:", status.Convert(err).Message())
			logger.Println("达到预期结果 篡改交易发送失败")
		}
		return
	}else {
		logger.Println("篡改交易发送成功")
	}

	//这里更新部分是对的，因为更新好以后还是可以继续交易。
	util_ron.UpdateUTXOs(pubkeyHash, utxos, &tx) //更新utxo

	accInfo.UpdateBalance(fromAccount, -amount)
	accInfo.UpdateBalance(toAccount, amount)
	//logger.Info("New transaction is sent!")
}

//付款
func (ser *Service) SendFalsifyVoutToken(pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo, amount uint64, fromAccount, toAccount string) {

	//创建交易
	tx, err := util_ron.CreateTransactionByUTXOs(utxos, accInfo, amount, fromAccount, toAccount,0,0,0,"")
	if err != nil {
		logger.Error("The transaction was abandoned.", err)
		return
	}
	//篡改交易
	tx.Vout[0].Value = common.NewAmount(10)
	//发送交易请求
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err = ser.conn.RpcSendTransaction(context.Background(), sendTransactionRequest)

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Other error:", status.Convert(err).Message())
			logger.Println("达到预期结果 篡改交易发送失败")
		}
		return
	}else {
		logger.Println("篡改交易发送成功")
	}

	//这里更新部分是对的，因为更新好以后还是可以继续交易。
	util_ron.UpdateUTXOs(pubkeyHash, utxos, &tx) //更新utxo

	accInfo.UpdateBalance(fromAccount, -amount)
	accInfo.UpdateBalance(toAccount, amount)
	//logger.Info("New transaction is sent!")
}

//付款
func (ser *Service) SendFalsifyTxIndexToken(pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo, amount uint64, fromAccount, toAccount string) {
	for _, u := range utxos.Indices {
		u.TxIndex ++
	}
	//创建交易
	tx, err := util_ron.CreateTransactionByUTXOs(utxos, accInfo, amount, fromAccount, toAccount,0,0,0,"")
	if err != nil {
		logger.Error("The transaction was abandoned.", err)
		return
	}
	logger.Println("篡改utxo的txIndex")
	//篡改交易
	//tx.Vin[0].Vout ++
	//发送交易请求
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err = ser.conn.RpcSendTransaction(context.Background(), sendTransactionRequest)

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Other error:", status.Convert(err).Message())
			logger.Println("达到预期结果 篡改交易发送失败")
		}
		return
	}else {
		logger.Println("篡改交易发送成功")
	}

	//这里更新部分是对的，因为更新好以后还是可以继续交易。
	util_ron.UpdateUTXOs(pubkeyHash, utxos, &tx) //更新utxo

	accInfo.UpdateBalance(fromAccount, -amount)
	accInfo.UpdateBalance(toAccount, amount)
	//logger.Info("New transaction is sent! ")
}

//付款
func (ser *Service) SendToken(pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo, amount uint64, fromAccount, toAccount string) {
	//创建交易
	tx, err := util_ron.CreateTransactionByUTXOs(utxos, accInfo, amount, fromAccount, toAccount,0,0,0,"")
	if err != nil {
		logger.Error("The transaction was abandoned.", err)
		return
	}
	//发送交易请求
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err = ser.conn.RpcSendTransaction(context.Background(), sendTransactionRequest)

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Other error:", status.Convert(err).Message())
		}
		return
	}else {
		logger.Println("达到预期结果 正常交易发送成功")
	}

	//这里更新部分是对的，因为更新好以后还是可以继续交易。
	util_ron.UpdateUTXOs(pubkeyHash, utxos, &tx) //更新utxo

	accInfo.UpdateBalance(fromAccount, -amount)
	accInfo.UpdateBalance(toAccount, amount)
	//logger.Info("New transaction is sent! ")
}

//得到指定账户的余额
func (ser *Service) GetBalance(account string) uint64 {
	response, err := ser.conn.RpcGetBalance(context.Background(), &rpcpb.GetBalanceRequest{Address: account})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Error:", status.Convert(err).Message())
		}
		os.Exit(1)
	}
	return uint64(response.GetAmount())
}
