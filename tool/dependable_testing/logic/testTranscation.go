package logic

import (
	"flag"
	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/utxo"
	dependable_configpb "github.com/dappley/go-dappley/tool/dependable_testing/pb"
	account_ron "github.com/dappley/go-dappley/tool/dependable_testing/sdk"
	"github.com/dappley/go-dappley/tool/dependable_testing/service"
	logger "github.com/sirupsen/logrus"
)

func TestTransaction() {
	////config
	configs := &dependable_configpb.Config{}
	config.LoadConfig(configFilePath, configs)
	var filePath string
	flag.StringVar(&filePath, "f", configFilePath, "Configuration File Path. Default to conf/default.conf")
	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())
	//存放创建的account
	acInfo := account_ron.NewAccountInfo()
	//发送普通的交易
	SendTransaction(serviceClient, acInfo, minerAccount.GetAddress().String(), configs,"normal")
	//发送篡改的交易
	SendTransaction(serviceClient, acInfo, minerAccount.GetAddress().String(), configs,"falsify")
}
//开始交易，from问矿工要钱，再给to,没钱了再问矿工要，一直重复
func SendTransaction(service *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *dependable_configpb.Config,txPtr string) {
	fromAccount,toAccount:=accInfo.CreateAccountPair()
	fromAcc := fromAccount.GetAddress().String()
	var utxoTx *utxo.UTXOTx
	if txPtr == "falsify"{
		//本地没钱了就问服务器要，如果使用服务器的余额判断，因为延迟关系，本地早没钱了，
		//还在发送交易，传到服务器，服务器会接受到很多不存在的交易。
		if accInfo.GetBalance(fromAcc) < config.AmountPerTx {
			utxoTx = service.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
		}
		// 发送一个篡改的交易
		logger.Info("正在发送篡改Contract的交易...")
		logger.Info("预期结果：交易校验失败")
		if accInfo.GetBalance(fromAcc) > config.AmountPerTx  {
			service.SendFalsifyContractToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, config.AmountPerTx, fromAcc, toAccount.GetAddress().String())
		}
		logger.Info("正在发送篡改vout的value的交易...")
		logger.Info("预期结果：交易校验失败")
		if accInfo.GetBalance(fromAcc) > config.AmountPerTx  {
			service.SendFalsifyVoutToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, config.AmountPerTx, fromAcc, toAccount.GetAddress().String())
		}
		logger.Info("正在发送篡改vin的txIndex的交易...")
		logger.Info("预期结果：交易校验失败")
		if accInfo.GetBalance(fromAcc) > config.AmountPerTx  {
			service.SendFalsifyTxIndexToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, config.AmountPerTx, fromAcc, toAccount.GetAddress().String())
		}
	}else {
		if accInfo.GetBalance(fromAcc) < config.AmountPerTx {
			utxoTx = service.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
		}
		// 发送一个正常的交易
		logger.Info("发送正常的交易...")
		logger.Info("预期结果：交易校验成功")
		if accInfo.GetBalance(fromAcc) > config.AmountPerTx  {
			service.SendToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, config.AmountPerTx, fromAcc, toAccount.GetAddress().String())
		}
	}
}
