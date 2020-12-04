package main

import (
	"runtime"
	"time"

	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/utxo"
	performance_configpb "github.com/dappley/go-dappley/tool/performance_testing/pb"
	account_ron "github.com/dappley/go-dappley/tool/performance_testing/sdk"
	"github.com/dappley/go-dappley/tool/performance_testing/service"
	logger "github.com/sirupsen/logrus"
)

var restartWaitTimeSeconds int = 5

//说明：default.conf: goCount设置为10，tps为2
func TPSTester() {

	//config
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)
	buildLog(configs)

	for {
		runTPSTester(configs)
		logger.Infof("Restarting TPS Tester in %v seconds ... ", restartWaitTimeSeconds)
		time.Sleep(time.Duration(restartWaitTimeSeconds) * time.Second)
	}

}

func runTPSTester(configs *performance_configpb.Config) {
	logger.Info("持续测试开始，可使用 Ctrl+C 中断测试")
	logger.Info("TPS为", float32(configs.GoCount)*configs.Tps)
	logger.Info("")
	logger.Info("正在初始化...")

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())

	acInfo := account_ron.NewAccountInfo()
	stopChan := make(chan bool)
	startTest := false
	restart := false

	var err error
	acInfo.Accounts, err = account_ron.ReadAccountFromFile()
	if err != nil {
		logger.Info("未找到account.dat，根据default启动测试，")
		logger.Info("正在向矿工获取token...")
		//交易生成
		for i := int32(0); i < configs.GetGoCount(); i++ {
			go startTransactionTPSGoroutine(
				serviceClient,
				acInfo,
				minerAccount.GetAddress().String(),
				configs,
				&startTest,
				&restart,
				stopChan)
			time.Sleep(100 * time.Millisecond)
		}
		acInfo.WaitTillGetToken(configs.GetAmountFromMinner() * uint64(configs.GetGoCount()))
		account_ron.SaveAccountToFile(acInfo) //写入account.bat
	} else {
		lenAccount := len(acInfo.Accounts)
		if lenAccount%2 != 0 {
			logger.Error("account.dat出错，请删除重启程序")
			return
		}
		logger.Info("找到:", len(acInfo.Accounts)/2, " 对账户。 TPS为", float32(len(acInfo.Accounts)/2)*configs.Tps)
		var total uint64 = 0
		for i := 0; i < lenAccount; i = i + 2 {
			fromAccount := acInfo.Accounts[i]
			toAccount := acInfo.Accounts[i+1]

			acInfo.FromAddress = append(acInfo.FromAddress, fromAccount.GetAddress().String())
			fromBalance, err := serviceClient.GetBalanceWithError(fromAccount.GetAddress().String())
			if err != nil {
				return
			}
			if fromBalance <= 1 {
				acInfo.Balances[fromAccount.GetAddress().String()] = uint64(fromBalance)
				total = total + configs.GetAmountFromMinner()
			} else {
				acInfo.Balances[fromAccount.GetAddress().String()] = uint64(fromBalance)
				total = total + uint64(fromBalance)
			}

			acInfo.ToAddress = append(acInfo.ToAddress, toAccount.GetAddress().String())
			toBalance, err := serviceClient.GetBalanceWithError(toAccount.GetAddress().String())
			if err != nil {
				return
			}
			acInfo.Balances[toAccount.GetAddress().String()] = uint64(toBalance)

			go startTransactionTPSFromFile(
				serviceClient,
				acInfo,
				minerAccount.GetAddress().String(),
				configs,
				&startTest,
				&restart,
				stopChan,
				fromAccount,
				toAccount)
		}
		acInfo.WaitTillGetToken(total)

	}
	//等待所有账户拿到钱

	startTest = true
	logger.Info("开始发送交易...")
	logger.Info("当前时间为：", time.Now().Format("2006-01-02 15:04:05"))

	//日志刷新
	stopLog := make(chan bool)
	go LogPrinter(acInfo, serviceClient, stopLog)

	for !restart {
	}
	//停止日志和所有go程交易
	startTest = false
	stopLog <- true
	for i := int32(0); i < configs.GetGoCount(); i++ {
		stopChan <- true
	}
	logger.Info("Exiting All StartTransaction Routines ... ")
}

//开始交易，from问矿工要钱，再给to,没钱了再问矿工要，一直重复
func startTransactionTPSGoroutine(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, start, restart *bool, stop chan bool) {
	fromAccount, toAccount := accInfo.CreateAccountPair()
	var utxoTx *utxo.UTXOTx
	startTransactionTPS(utxoTx, ser, accInfo, minnerAcc, config, start, restart, stop, fromAccount, toAccount)
}

//开始交易，from问矿工要钱，再给to,没钱了再问矿工要，一直重复
func startTransactionTPSFromFile(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, start, restart *bool, stop chan bool, fromAccount, toAccount *account.Account) {
	fromAcc := fromAccount.GetAddress().String()
	utxoTx, err := ser.GetUTXOTxFromServer(fromAcc)
	if err != nil {
		logger.Error("Get UTXOTx error:", err)
		*restart = true
		runtime.Goexit()
	}
	startTransactionTPS(utxoTx, ser, accInfo, minnerAcc, config, start, restart, stop, fromAccount, toAccount)
}

func startTransactionTPS(utxoTx *utxo.UTXOTx, ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, start, restart *bool, stop chan bool, fromAccount, toAccount *account.Account) {
	fromAcc := fromAccount.GetAddress().String()
	ticker := time.NewTicker(time.Microsecond * time.Duration(1000000/config.Tps)) //定时1秒
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//本地没钱了就问服务器要，如果使用服务器的余额判断，因为延迟关系，本地早没钱了，
			//还在发送交易，传到服务器，服务器会接受到很多不存在的交易
			var err error
			if accInfo.GetBalance(fromAcc) <= 1 { //每次交易就发1个token和1个tip
				utxoTx, err = ser.GetTokenWithError(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
				if err != nil {
					logger.Warn("GetToken failed!")
					*restart = true
				}
			}
			if accInfo.GetBalance(fromAcc) > 1 && *start {
				err = ser.SendTokenWithError(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, 1, fromAcc, toAccount.GetAddress().String())
				if err != nil {
					logger.Warn("SendToken failed!")
					*restart = true
				}
			}
		case <-stop:
			time.Sleep(2)
			runtime.Goexit() //退出go线程
		}
	}
}
