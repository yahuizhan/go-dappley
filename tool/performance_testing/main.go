package main

import (
	"os"
	"runtime"
	"time"

	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/utxo"
	performance_configpb "github.com/dappley/go-dappley/tool/performance_testing/pb"
	account_ron "github.com/dappley/go-dappley/tool/performance_testing/sdk"
	"github.com/dappley/go-dappley/tool/performance_testing/service"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	logger "github.com/sirupsen/logrus"
)

const (
	configFilePath = "default.conf"
	logFreshTime   = 5 //刷新日志间隔
)

func main() {

	if len(os.Args) > 1 {
		switch os.Args[1] { //判断第二个命令
		case "1":
			AverageTxCon()
		case "2":
			AverageTxQue()
		case "3":
			HighConcurrency()
		case "4":
			Lowload()
		case "5":
			TPSTester()
		case "9":
			ManualTPSTester()
		case "balance":
			printServerBalance()
		default:
			testMenu()
			return
		}
	} else {
		testMenu()
	}

}

func testMenu() {
	logger.Info("可信区块链测试目录")
	logger.Info("1.平均交易确认时间 及 最长交易确认时间")
	logger.Info("2.平均查询时间 及 最长查询时间")
	logger.Info("3.高并发场景")
	logger.Info("4.低负载运行场景")
	logger.Info("命令说明:")
	logger.Info("启动第一个测试,即:《1.平均交易确认时间 及 最长交易确认时间》命令: ./performance_testing 1")
}

func CheckTransactionNumber(acInfo *account_ron.AccountInfo, serviceClient *service.Service) (int64, int64) {
	timeOut := 25 //交易结束发送后，等待交易打包最长时间,单位秒
	var toSum, localToSum int64
	for _, address := range acInfo.ToAddress {
		if address == "" {
			continue
		}
		localToSum = localToSum + int64(acInfo.GetBalance(address))

	} //这里只验证交易数量，丢失数量，交易准确性不是这在这个测试中

	count := 0
	for {
		toSum = 0 //重新计算服务器接收到的交易量
		for _, address := range acInfo.ToAddress {
			if address == "" {
				continue
			}
			toSum = toSum + serviceClient.GetBalance(address)
		}
		if toSum != localToSum {
			logger.Info("等待服务器打包交易...")
			count++
			time.Sleep(time.Second * 5)
		} else {
			break
		}

		if count == timeOut/5 {
			logger.Info("等待超时...")
			break
		}

	}
	return toSum, localToSum
}

func LogPrinter(acInfo *account_ron.AccountInfo, serviceClient *service.Service, out chan bool) {
	ticker := time.NewTicker(time.Second * logFreshTime)
	defer ticker.Stop()
	var toSum, lastToSum int64
	for {
		select {
		case <-ticker.C:
			for _, address := range acInfo.ToAddress {
				if address == "" {
					continue
				}
				toSum = toSum + serviceClient.GetBalance(address)
			}
			if toSum >= lastToSum {
				logger.Info("交易发送中，当前TPS：", (toSum-lastToSum)/int64(logFreshTime))
			} else {
				//logger.Info("toSum: ", toSum,",lastToSum: ",lastToSum)
			}
			lastToSum = toSum
			toSum = 0
		case <-out:
			runtime.Goexit()
		}
	}
}

//开始交易，from问矿工要钱，再给to,没钱了再问矿工要，一直重复
func StartTransactionGoroutine(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, start *bool, stop chan bool) {
	fromAccount, toAccount := accInfo.CreateAccountPair()
	fromAcc := fromAccount.GetAddress().String()
	var utxoTx *utxo.UTXOTx
	ticker := time.NewTicker(time.Microsecond * time.Duration(1000000/config.Tps)) //定时1秒
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//本地没钱了就问服务器要，如果使用服务器的余额判断，因为延迟关系，本地早没钱了，
			//还在发送交易，传到服务器，服务器会接受到很多不存在的交易
			if accInfo.GetBalance(fromAcc) <= 1 { //每次交易就发1个token和1个tip
				utxoTx = ser.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
			}
			if accInfo.GetBalance(fromAcc) > 1 && *start {
				ser.SendToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, 1, fromAcc, toAccount.GetAddress().String())

			}
		case <-stop:
			time.Sleep(2)
			runtime.Goexit() //退出go线程
		}
	}
}

//开始交易，from问矿工要钱，再给to,没钱了再问矿工要，一直重复
func StartTransactionFromFile(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, start *bool, stop chan bool, fromAccount, toAccount *account.Account) {
	fromAcc := fromAccount.GetAddress().String()
	utxoTx, err := ser.GetUTXOTxFromServer(fromAcc)
	if err != nil {
		logger.Error("Get UTXOTx error:", err)
	}
	ticker := time.NewTicker(time.Microsecond * time.Duration(1000000/config.Tps)) //定时1秒
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//本地没钱了就问服务器要，如果使用服务器的余额判断，因为延迟关系，本地早没钱了，
			//还在发送交易，传到服务器，服务器会接受到很多不存在的交易
			if accInfo.GetBalance(fromAcc) <= 1 { //每次交易就发1个token和1个tip
				utxoTx = ser.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
			}
			if accInfo.GetBalance(fromAcc) > 1 && *start {
				ser.SendToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, 1, fromAcc, toAccount.GetAddress().String())
			}
		case <-stop:
			time.Sleep(2)
			runtime.Goexit() //退出go线程
		}
	}
}

func buildLog(configs *performance_configpb.Config) {
	if configs.GetLogOpen() {
		name := configs.GetLogName()
		if name == "" {
			name = "log"
		}
		rotateTime := int(configs.GetLogRotateTime())
		if rotateTime == 0 {
			rotateTime = 86400
		}
		count := int(configs.GetLogCount())
		if count == 0 {
			count = 7
		}
		writeToLog(configs.GetLogLevel(), name, rotateTime, count)
	}
}

func writeToLog(logLevel, logName string, rotateTime, logCount int) {
	var level logger.Level
	switch logLevel {
	case "info":
		level = logger.InfoLevel
	case "debug":
		level = logger.DebugLevel
	case "error":
		level = logger.ErrorLevel
	case "warn":
		level = logger.WarnLevel
	default:
		level = logger.InfoLevel
	}

	logger.SetReportCaller(true)
	logger.SetLevel(level)
	logger.SetFormatter(&logger.TextFormatter{
		FullTimestamp:             false,
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		//TimestampFormat:           time.RFC3339Nano,
	})

	writer, err := rotatelogs.New(
		logName+".%Y%m%d%H%M%S",
		rotatelogs.WithLinkName(logName),
		rotatelogs.WithRotationTime(time.Second*time.Duration(rotateTime)),
		rotatelogs.WithMaxAge(-1),
		rotatelogs.WithRotationCount(uint(logCount)),
	)
	if err != nil {
		logger.Errorf("config local file system for logger error: %v", err)
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logger.DebugLevel: writer,
		logger.InfoLevel:  writer,
		logger.WarnLevel:  writer,
		logger.ErrorLevel: writer,
		logger.FatalLevel: writer,
		logger.PanicLevel: writer,
	}, &logger.TextFormatter{DisableColors: true})

	logger.AddHook(lfsHook)
}

func printServerBalance() {
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)

	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	accounts, err := account_ron.ReadAccountFromFile()

	if err != nil {
		logger.Info("未找到account.dat ... 没有账户，无法获取余额")
	} else {
		lenAccount := len(accounts)
		if lenAccount%2 != 0 {
			logger.Error("account.dat出错，请删除重启程序")
			return
		}

		for _, acc := range accounts {
			addr := acc.GetAddress().String()
			balance := uint64(serviceClient.GetBalance(addr))
			logger.Infof("Balance of Account %s : %v\n", shortenAddress(addr), balance)
		}
	}
}
