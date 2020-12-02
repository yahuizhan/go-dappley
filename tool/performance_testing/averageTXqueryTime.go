package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/utxo"
	performance_configpb "github.com/dappley/go-dappley/tool/performance_testing/pb"
	account_ron "github.com/dappley/go-dappley/tool/performance_testing/sdk"
	"github.com/dappley/go-dappley/tool/performance_testing/service"
	util_ron "github.com/dappley/go-dappley/tool/performance_testing/util"
	logger "github.com/sirupsen/logrus"
)

type txQueryTime struct {
	time  []int64
	mutex sync.Mutex
}

func AverageTxQue() {
	duration := 60 //测试持续时间，单位秒，发送交易的时间
	var goCount int32 = 10
	//config
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())
	//存放创建的aacount
	acInfo := account_ron.NewAccountInfo()

	logger.Info("4.1.9 基准性能测试")
	logger.Info("(28)平均查询时间测试, (29)最长查询时间测试")
	logger.Info("测试目的：")
	logger.Info("验证产品【平均】查询时间")
	logger.Info("验证产品【最长】查询时间")
	logger.Info("测试步骤：")
	logger.Info("发送查询请求，记录查询请求发送时间和信息返回时间,")
	logger.Info("获得平均及最长交易确认时间")
	logger.Info("")
	logger.Info("此次测试将会持续", duration, "秒")
	logger.Info("正在初始化...")

	txQueryTime := txQueryTime{}
	//交易生成
	for i := int32(0); i < goCount; i++ {
		go startTransactionQuery(
			serviceClient,
			acInfo,
			minerAccount.GetAddress().String(),
			configs,
			&txQueryTime)
	}
	//让交易发送一会
	time.Sleep(time.Duration(duration) * time.Second)
	txQueryTime.mutex.Lock()
	logger.Info("共耗时: ", duration, "秒,查询交易: ", len(txQueryTime.time), "笔")
	ave, max := util_ron.GetAvergeAndMax(txQueryTime.time)
	txQueryTime.mutex.Unlock()
	logger.Info("【平均】交易查询时间为: ", ave, "微秒，【最长】交易查询时间为: ", max, "微秒")
	logger.Info("测试结束")

}

//开始交易，from问矿工要钱，再给to,没钱了再问矿工要，一直重复
func startTransactionQuery(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string,
	config *performance_configpb.Config, txQueryTime *txQueryTime) {

	fromAccount, toAccount := accInfo.CreateAccountPair()
	fromAcc := fromAccount.GetAddress().String()

	var lastToBalace, nowBalance, spendTime int64
	var utxoTx *utxo.UTXOTx

	for {
		if accInfo.GetBalance(fromAcc) < 1 {
			utxoTx = ser.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
			for uint64(ser.GetBalance(fromAcc)) < 0 { //等待fromAcc token到账
				time.Sleep(100 * time.Millisecond)
			}
		}
		//rand sleep 1-5秒
		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
		//发送交易
		if accInfo.GetBalance(fromAcc) >= 1 {
			ser.SendToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, 0, fromAcc, toAccount.GetAddress().String())
		}

		//计算打印时间，并报错，后续求平均值和最大值
		for nowBalance <= lastToBalace { //等待token到账
			time.Sleep(100 * time.Millisecond)
			nowBalance, spendTime = ser.GetBalanceWithRespondTime(toAccount.GetAddress().String())
		}
		txQueryTime.mutex.Lock()
		txQueryTime.time = append(txQueryTime.time, spendTime)
		txQueryTime.mutex.Unlock()
		logger.Info("查询已落块交易,耗时: ", spendTime, "微秒")
		lastToBalace = nowBalance
	}

}
