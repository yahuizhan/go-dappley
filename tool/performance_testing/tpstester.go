package main

import (
	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/core/account"
	performance_configpb "github.com/dappley/go-dappley/tool/performance_testing/pb"
	account_ron "github.com/dappley/go-dappley/tool/performance_testing/sdk"
	"github.com/dappley/go-dappley/tool/performance_testing/service"
	logger "github.com/sirupsen/logrus"
	"time"
)

//说明：default.conf: goCount设置为10，tps为2
func TPSTester() {

	//config
	runTime24 := 86400 //运行交易时间  1小时=3,600，24小时=86,400
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)

	logger.Info("持续测试开始，可使用 Ctrl+C 中断测试")
	logger.Info("在", runTime24, "秒内,向服务器持续发送交易请求")
	logger.Info("TPS为", float32(configs.GoCount)*configs.Tps,)
	logger.Info("")
	logger.Info("正在初始化...")

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())

	acInfo := account_ron.NewAccountInfo()
	stopChan := make(chan bool)
	startTest := false

	var err error

	RESTART:

	acInfo.Accounts, err = account_ron.ReadAccountFromFile()
	if err != nil {
		logger.Info("未找到account.dat，根据default启动测试，")
		logger.Info("正在向矿工获取token...")
		//交易生成
		for i := int32(0); i < configs.GetGoCount(); i++ {
			go StartTransactionGoroutine(
				serviceClient,
				acInfo,
				minerAccount.GetAddress().String(),
				configs,
				&startTest,
				stopChan)
			time.Sleep(100*time.Millisecond)
		}
		acInfo.WaitTillGetToken(configs.GetAmountFromMinner() * uint64(configs.GetGoCount()))
		account_ron.SaveAccountToFile(acInfo) //写入account.bat
	} else {
		lenAccount := len(acInfo.Accounts)
		if lenAccount%2 != 0 {
			logger.Error("account.dat出错，请删除重启程序")
			return
		}
		logger.Info("找到:",len(acInfo.Accounts)/2," 对账户。 TPS为", float32(len(acInfo.Accounts)/2)*configs.Tps)
		for i := 0; i < lenAccount; i = i + 2 {
			fromAccount := acInfo.Accounts[i]
			toAccount := acInfo.Accounts[i+1]
			acInfo.FromAddress = append(acInfo.FromAddress, fromAccount.GetAddress().String())
			acInfo.Balances[fromAccount.GetAddress().String()] = uint64(serviceClient.GetBalance(fromAccount.GetAddress().String()))
			acInfo.ToAddress = append(acInfo.ToAddress, toAccount.GetAddress().String())
			acInfo.Balances[toAccount.GetAddress().String()] = uint64(serviceClient.GetBalance(toAccount.GetAddress().String()))

			go StartTransactionFromFile(
				serviceClient,
				acInfo,
				minerAccount.GetAddress().String(),
				configs,
				&startTest,
				stopChan,
				fromAccount,
				toAccount)
		}
	}
	//等待所有账户拿到钱

	startTest = true
	logger.Info("开始发送交易...")
	logger.Info("当前时间为：", time.Now().Format("2006-01-02 15:04:05"))

	//日志刷新
	//stopLog := make(chan bool)
	//go LogPrinter(acInfo, serviceClient, stopLog)
	for {
		if !startTest{
			//停止日志和所有go程交易
			//stopLog <- true
			//
			for i := int32(0); i < configs.GetGoCount(); i++ {
				stopChan <- true
			}
			logger.Info("发现错误...")
			logger.Info("关闭所有进程，15秒后重启...")
			time.Sleep(15 * time.Second)
			logger.Info("开始重启...")

			goto RESTART
		}
	}
}