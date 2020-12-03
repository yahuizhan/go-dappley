package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/transaction"
	transactionpb "github.com/dappley/go-dappley/core/transaction/pb"
	"github.com/dappley/go-dappley/core/utxo"
	rpcpb "github.com/dappley/go-dappley/rpc/pb"
	sdk_ron "github.com/dappley/go-dappley/tool/performance_testing/sdk"
	util_ron "github.com/dappley/go-dappley/tool/performance_testing/util"
	logger "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (ser *Service) GetToken(accInfo *sdk_ron.AccountInfo, minnerAcc string, fromAcc string, amountFromminer uint64) *utxo.UTXOTx {
	count := 0
	for uint64(ser.GetBalance(minnerAcc)) <= amountFromminer { //等待矿工有钱
		time.Sleep(100 * time.Millisecond)
	}
	ser.minerSendTokenToAccount(amountFromminer, fromAcc)

	for uint64(ser.GetBalance(fromAcc)) < amountFromminer {
		time.Sleep(1000 * time.Millisecond)
		count++ //如果7秒还没到账，可能出错了，再问矿工要钱
		logger.Info("等待token到账...")
		if count > 7 {
			logger.Warn("获取token失败，请重启测试程序")
			os.Exit(0)
		}
	}

	//获取utxo，更新balance
	logger.Info("查询From账户UTXO...")
	utxoTx, err := ser.GetUTXOTxFromServer(fromAcc)
	if err != nil {
		logger.Error("Get UTXOTx error:", err)
	}
	logger.Info("更新本地From账户余额...")
	accInfo.UpdateBalance(fromAcc, amountFromminer) //自己维护账户
	return utxoTx

}

//从矿工拿钱，得等到挖出一个快，矿工才有钱
func (ser *Service) minerSendTokenToAccount(amount uint64, account string) {
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
		return
	}
	//logger.Info(amount, " has been sent to FromAddress.",time.Now().Format("2006-01-02 15:04:05"))
	logger.Info("已向矿工发送获取token请求...")
}

func (ser *Service) GetUTXOTxFromServer(fromAccount string) (*utxo.UTXOTx, error) {
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
		return nil, err
	}

	utxoTx := utxo.NewUTXOTx()
	for _, u := range response.GetUtxos() {
		utxo := utxo.UTXO{}
		utxo.Value = common.NewAmountFromBytes(u.Amount)
		utxo.Txid = u.Txid
		utxo.PubKeyHash = account.PubKeyHash(u.PublicKeyHash)
		utxo.TxIndex = int(u.TxIndex)
		utxoTx.PutUtxo(&utxo) //组装成UTXOTx
	}
	logger.Info("收到UTXO,个数: ", len(utxoTx.Indices))
	return &utxoTx, nil
}

//付款
func (ser *Service) SendToken(pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo, amount, tip uint64, fromAccount, toAccount string) {
	//创建交易
	tx, err := util_ron.CreateTransactionByUTXOs(utxos, accInfo, amount, tip, fromAccount, toAccount)
	if err != nil {
		logger.Error("The transaction was abandoned.", err)
		return
	}
	//发送交易请求
	ser.sendToken(tx, pubkeyHash, utxos, accInfo, amount, tip, fromAccount, toAccount)
}

func (ser *Service) sendToken(tx transaction.Transaction, pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo, amount, tip uint64, fromAccount, toAccount string) {
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err := ser.conn.RpcSendTransaction(context.Background(), sendTransactionRequest)

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Other error:", status.Convert(err).Message())
		}
		return
	}

	//这里更新部分是对的，因为更新好以后还是可以继续交易。
	util_ron.UpdateUTXOs(pubkeyHash, utxos, &tx) //更新utxo

	accInfo.UpdateBalance(fromAccount, -amount-tip)
	accInfo.UpdateBalance(toAccount, amount)
	//logger.Info("New transaction is sent! ")
}

//付款，并报错
func (ser *Service) SendTokenWithError(pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo, amount, tip uint64, fromAccount, toAccount string) error {
	//创建交易
	tx, err := util_ron.CreateTransactionByUTXOs(utxos, accInfo, amount, tip, fromAccount, toAccount)
	if err != nil {
		logger.Error("The transaction was abandoned.", err)
		return err
	}
	//发送交易请求
	return ser.sendTokenWithError(tx, pubkeyHash, utxos, accInfo, amount, tip, fromAccount, toAccount)
}

func (ser *Service) sendTokenWithError(tx transaction.Transaction, pubkeyHash account.PubKeyHash, utxos *utxo.UTXOTx, accInfo *sdk_ron.AccountInfo, amount, tip uint64, fromAccount, toAccount string) error {
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err := ser.conn.RpcSendTransaction(context.Background(), sendTransactionRequest)

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Other error:", status.Convert(err).Message())
		}
		return err
	}

	//这里更新部分是对的，因为更新好以后还是可以继续交易。
	util_ron.UpdateUTXOs(pubkeyHash, utxos, &tx) //更新utxo

	accInfo.UpdateBalance(fromAccount, -amount-tip)
	accInfo.UpdateBalance(toAccount, amount)
	//logger.Info("New transaction is sent! ")
	return nil
}

//得到指定账户的余额
func (ser *Service) GetBalance(account string) int64 {
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
	return response.GetAmount()
}

//得到指定账户的余额，报错不退出程序
func (ser *Service) GetBalanceNotExit(account string) int64 {
	response, err := ser.conn.RpcGetBalance(context.Background(), &rpcpb.GetBalanceRequest{Address: account})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Error:", status.Convert(err).Message())
		}
		return -1
	}
	return response.GetAmount()
}

//得到指定账户的余额
func (ser *Service) GetBalanceWithRespondTime(account string) (int64, int64) {
	startTime := time.Now()
	response, err := ser.conn.RpcGetBalance(context.Background(), &rpcpb.GetBalanceRequest{Address: account})
	totalSpendTime := time.Now().Sub(startTime).Microseconds()
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Error:", status.Convert(err).Message())
		}
		os.Exit(1)
	}
	return response.GetAmount(), totalSpendTime
}
