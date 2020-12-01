package main

import (
	"fmt"
	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/core/account"
	performance_configpb "github.com/dappley/go-dappley/tool/performance_testing/pb"
	account_ron "github.com/dappley/go-dappley/tool/performance_testing/sdk"
	"github.com/dappley/go-dappley/tool/performance_testing/service"
	logger "github.com/sirupsen/logrus"
	"math"
	"time"
)

//说明：default.conf: goCount设置为10，tps为2
func Lowload() {
	//config
	runTime24 := 86400 //运行交易时间  1小时=3,600，24小时=86,400
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())
	//存放创建的aacount
	acInfo := account_ron.NewAccountInfo()

	logger.Info("3.3低负载运行场景...")
	logger.Info("测试目的：")
	logger.Info("验证系统低负载运行场景下的表现")
	logger.Info("测试步骤：")
	logger.Info("在", runTime24, "秒内,向服务器持续发送交易请求,TPS为",configs.GoCount*configs.Tps,",然后验证交易成功率")
	logger.Info("")
	logger.Info("正在初始化...")

	stopChan := make(chan bool)
	startTest := false
	//交易生成
	for i := int32(0); i < configs.GetGoCount(); i++ {
		go StartTransactionGoroutine(
			serviceClient,
			acInfo,
			minerAccount.GetAddress().String(),
			configs,
			&startTest,
			stopChan)
	}

	//等待所有账户拿到钱
	acInfo.WaitTillGetToken(configs.GetAmountFromMinner()*uint64(configs.GetGoCount()))

	startTest = true
	logger.Info("开始发送交易...")
	logger.Info("当前时间为：",time.Now().Format("2006-01-02 15:04:05"))

	//日志刷新
	stopLog := make(chan bool)
	go LogPrinter(acInfo, serviceClient, stopLog)

	//让交易发送一段时间
	time.Sleep(time.Duration(runTime24) * time.Second)

	//停止日志和所有go程交易
	stopLog <- true
	startTest = false
	for i := int32(0); i < configs.GetGoCount(); i++ {
		stopChan <- true
	}
	logger.Info("交易发送停止,已用时间测试：", runTime24, "秒.")

	logger.Info("验证开始...")
	//计算发交易双方的balance
	toSum, localToSum := CheckTransactionNumber(acInfo, serviceClient)
	logger.Info("发送交易：", localToSum, "笔，成功接收交易:", toSum, "笔.")
	logger.Info("交易成功率：", fmt.Sprintf("%.2f", float64(toSum)/float64(localToSum)*100), "%")
	logger.Info("平均TPS：", math.Round(float64(toSum)/float64(runTime24)))
	logger.Info("测试结束")

}




