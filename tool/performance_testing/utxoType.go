package main

import (
	"time"

	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/core/account"
	performance_configpb "github.com/dappley/go-dappley/tool/performance_testing/pb"
	account_ron "github.com/dappley/go-dappley/tool/performance_testing/sdk"
	"github.com/dappley/go-dappley/tool/performance_testing/service"
	logger "github.com/sirupsen/logrus"
)

func OneToOneUtxo() {
	//config
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())
	//存放创建的aacount
	acInfo := account_ron.NewAccountInfo()

	logger.Info("UTXO 功能测试")
	logger.Info("类别 1对1")
	logger.Info("交易input 1个UTXO")
	logger.Info("交易output 1个UTXO")
	logger.Info("开始测试")

	txQueryTime := txQueryTime{}
	fromAccount, toAccount := acInfo.CreateAccountPair()
	fromAccPubkeyHash := fromAccount.GetPubKeyHash()
	var lastToBalace, nowBalance, spendTime int64

	go serviceClient.StartTestUTXO(
		fromAccPubkeyHash,
		acInfo,
		minerAccount.GetAddress().String(),
		fromAccount.GetAddress().String(),
		toAccount.GetAddress().String(),
		1,
		configs.GetAmountFromMinner(),
		"oneToOne")
	//计算打印时间，并报错，后续求平均值和最大值
	for nowBalance <= lastToBalace { //等待token到账
		time.Sleep(100 * time.Millisecond)
		nowBalance, spendTime = serviceClient.GetBalanceWithRespondTime(toAccount.GetAddress().String())
	}
	txQueryTime.mutex.Lock()
	txQueryTime.time = append(txQueryTime.time, spendTime)
	txQueryTime.mutex.Unlock()
	logger.Info("查询已落块交易,耗时: ", spendTime, "微秒")
	lastToBalace = nowBalance
	logger.Info("测试结束")

}

func OneToAllUtxo() {
	//config
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())
	//存放创建的aacount
	acInfo := account_ron.NewAccountInfo()

	logger.Info("UTXO 功能测试")
	logger.Info("类别 1对多")
	logger.Info("交易input 1个UTXO")
	logger.Info("交易output 多个UTXO")
	logger.Info("开始测试")

	txQueryTime := txQueryTime{}
	fromAccount, toAccount := acInfo.CreateAccountPair()
	fromAccPubkeyHash := fromAccount.GetPubKeyHash()
	var lastToBalace, nowBalance, spendTime int64

	go serviceClient.StartTestUTXO(
		fromAccPubkeyHash,
		acInfo,
		minerAccount.GetAddress().String(),
		fromAccount.GetAddress().String(),
		toAccount.GetAddress().String(),
		1,
		configs.GetAmountFromMinner(),
		"oneToAll")
	//计算打印时间，并报错，后续求平均值和最大值
	for nowBalance <= lastToBalace { //等待token到账
		time.Sleep(100 * time.Millisecond)
		nowBalance, spendTime = serviceClient.GetBalanceWithRespondTime(toAccount.GetAddress().String())
	}
	txQueryTime.mutex.Lock()
	txQueryTime.time = append(txQueryTime.time, spendTime)
	txQueryTime.mutex.Unlock()
	logger.Info("查询已落块交易,耗时: ", spendTime, "微秒")
	lastToBalace = nowBalance
	logger.Info("测试结束")
}

func AllToAllUtxo() {
	//config
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())
	//存放创建的aacount
	acInfo := account_ron.NewAccountInfo()

	logger.Info("UTXO 功能测试")
	logger.Info("类别 多对对")
	logger.Info("交易input 多个UTXO")
	logger.Info("交易output 多个UTXO")
	logger.Info("开始测试")

	txQueryTime := txQueryTime{}
	fromAccount, toAccount := acInfo.CreateAccountPair()
	fromAccPubkeyHash := fromAccount.GetPubKeyHash()
	var lastToBalace, nowBalance, spendTime int64

	go serviceClient.StartTestUTXO(
		fromAccPubkeyHash,
		acInfo,
		minerAccount.GetAddress().String(),
		fromAccount.GetAddress().String(),
		toAccount.GetAddress().String(),
		1,
		configs.GetAmountFromMinner(),
		"allToAll")

	//计算打印时间，并报错，后续求平均值和最大值
	for nowBalance <= lastToBalace { //等待token到账
		time.Sleep(100 * time.Millisecond)
		nowBalance, spendTime = serviceClient.GetBalanceWithRespondTime(toAccount.GetAddress().String())
	}
	txQueryTime.mutex.Lock()
	txQueryTime.time = append(txQueryTime.time, spendTime)
	txQueryTime.mutex.Unlock()
	logger.Info("查询已落块交易,耗时: ", spendTime, "微秒")
	lastToBalace = nowBalance
	logger.Info("测试结束")
}
