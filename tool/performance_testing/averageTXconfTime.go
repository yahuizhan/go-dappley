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

type txTime struct {
	time  []int64
	mutex sync.Mutex
}

func AverageTxCon() {
	duration := 60         //测试持续时间，单位秒，发送交易的时间
	var goCount int32 = 10 //默认开10个线程
	//config
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())
	//存放创建的aacount
	acInfo := account_ron.NewAccountInfo()

	logger.Info("4.1.9 基准性能测试")
	logger.Info("(26) 平均交易确认时间测试， (27) 最长交易确认时间测试")
	logger.Info("测试目的：")
	logger.Info("验证产品【平均】交易确认时间")
	logger.Info("验证产品【最长】交易确认时间")
	logger.Info("测试步骤：")
	logger.Info("发送交易，记录交易发送时间，查询区块链系统中交易的落块时间,")
	logger.Info("获得平均及最长交易确认时间")
	logger.Info("")
	logger.Info("此次测试将会持续", duration, "秒")
	logger.Info("正在初始化...")

	txTime := txTime{}
	//交易生成
	for i := int32(0); i < goCount; i++ {
		go startTransaction(
			serviceClient,
			acInfo,
			minerAccount.GetAddress().String(),
			configs,
			&txTime)
	}
	//让交易发送一会
	time.Sleep(time.Duration(duration) * time.Second)
	txTime.mutex.Lock()
	logger.Info("共耗时: ", duration, "秒,发送交易: ", len(txTime.time), "笔")
	ave, max := util_ron.GetAvergeAndMax(txTime.time)
	txTime.mutex.Unlock()
	logger.Info("【平均】交易确认时间为: ", ave, "毫秒，【最长】交易确认时间为: ", max, "毫秒")
	logger.Info("测试结束")

}

//开始交易，from问矿工要钱，再给to,没钱了再问矿工要，一直重复
func startTransaction(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string,
	config *performance_configpb.Config, txTime *txTime) {

	fromAccount, toAccount := accInfo.CreateAccountPair()
	fromAcc := fromAccount.GetAddress().String()

	var sendTime time.Time
	var lastToBalace, nowBalance int64
	var utxoTx *utxo.UTXOTx

	for {
		if accInfo.GetBalance(fromAcc) < 1 {
			utxoTx = ser.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
			for ser.GetBalance(fromAcc) < 0 { //等待fromAcc token到账
				time.Sleep(100 * time.Millisecond)
			}
		}
		//rand sleep 1-5秒
		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
		//发送交易
		if accInfo.GetBalance(fromAcc) >= 1 {
			ser.SendToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, 0, fromAcc, toAccount.GetAddress().String())
			sendTime = time.Now()
		}

		for nowBalance <= lastToBalace { //等待token到账
			time.Sleep(100 * time.Millisecond)
			nowBalance = ser.GetBalance(toAccount.GetAddress().String())
		}
		reciveTime := time.Now()
		timeDifference := reciveTime.Sub(sendTime).Milliseconds()
		txTime.mutex.Lock()
		txTime.time = append(txTime.time, timeDifference)
		txTime.mutex.Unlock()
		logger.Info("交易已落块,耗时: ", timeDifference, "毫秒")
		lastToBalace = nowBalance
	}

}
