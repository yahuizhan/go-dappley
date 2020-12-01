package main

import (
	"bufio"
	"fmt"
	"os"
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

// ManualTPSTester makes one transaction per pair of accounts every time user presses Enter
func ManualTPSTester() {
	configs := &performance_configpb.Config{}
	config.LoadConfig(configFilePath, configs)
	buildLog(configs)

	logger.Info("手动持续测试开始，可使用 Ctrl+C 中断测试")
	logger.Info("TPS为", configs.GoCount*configs.Tps)
	logger.Info("")
	logger.Info("正在初始化...")

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())

	acInfo := account_ron.NewAccountInfo()
	runChan := make(chan bool)

	var err error
	acInfo.Accounts, err = account_ron.ReadAccountFromFile()
	numRoutines := 0
	if err != nil {
		logger.Info("未找到account.dat，根据default启动测试，")
		logger.Info("正在向矿工获取token...")
		//交易生成
		for i := int32(0); i < configs.GetGoCount(); i++ {
			go startTxGoroutine(
				serviceClient,
				acInfo,
				minerAccount.GetAddress().String(),
				configs,
				runChan)
			time.Sleep(100 * time.Millisecond)
		}
		numRoutines = int(configs.GetGoCount())
		acInfo.WaitTillGetToken(configs.GetAmountFromMinner() * uint64(configs.GetGoCount()))
		account_ron.SaveAccountToFile(acInfo) //写入account.bat
	} else {
		lenAccount := len(acInfo.Accounts)
		if lenAccount%2 != 0 {
			logger.Error("account.dat出错，请删除重启程序")
			return
		}
		for i := 0; i < lenAccount; i = i + 2 {
			fromAccount := acInfo.Accounts[i]
			toAccount := acInfo.Accounts[i+1]
			acInfo.FromAddress = append(acInfo.FromAddress, fromAccount.GetAddress().String())
			acInfo.Balances[fromAccount.GetAddress().String()] = uint64(serviceClient.GetBalance(fromAccount.GetAddress().String()))
			acInfo.ToAddress = append(acInfo.ToAddress, toAccount.GetAddress().String())
			acInfo.Balances[toAccount.GetAddress().String()] = uint64(serviceClient.GetBalance(toAccount.GetAddress().String()))

			go startTxFromFile(
				serviceClient,
				acInfo,
				minerAccount.GetAddress().String(),
				configs,
				runChan,
				fromAccount,
				toAccount)
		}
		numRoutines = lenAccount / 2
		acInfo.WaitTillGetToken(configs.GetAmountFromMinner() * uint64(lenAccount/2))
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("[Enter 'S' To Send Transactions; Enter 'B' To Check Local Balance; Enter 'E' To Exit] ")
		input, _ := reader.ReadString('\n')
		if input == "s\n" || input == "S\n" {
			logger.Info("Sending transactions ...")
			for i := 0; i < numRoutines; i++ {
				runChan <- true
			}
		} else if input == "b\n" || input == "B\n" {
			printBalance(acInfo)
		} else if input == "e\n" || input == "E\n" {
			logger.Info("Exiting All Go Routines ...")
			for i := 0; i < numRoutines; i++ {
				runChan <- false
			}
			time.Sleep(2)
			break
		} else {
			logger.Warn("Input is not valid")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func startTxGoroutine(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, run chan bool) {
	fromAccount, toAccount := accInfo.CreateAccountPair()
	var utxoTx *utxo.UTXOTx
	startTx(ser, accInfo, minnerAcc, config, run, fromAccount, toAccount, utxoTx)
}

func startTxFromFile(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, run chan bool, fromAccount, toAccount *account.Account) {
	fromAcc := fromAccount.GetAddress().String()
	utxoTx, err := ser.GetUTXOTxFromServer(fromAcc)
	if err != nil {
		logger.Error("Get UTXOTx error:", err)
	}
	startTx(ser, accInfo, minnerAcc, config, run, fromAccount, toAccount, utxoTx)
}

func startTx(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, run chan bool, fromAccount, toAccount *account.Account, utxoTx *utxo.UTXOTx) {
	fromAcc := fromAccount.GetAddress().String()
	toAcc := toAccount.GetAddress().String()
	for {
		select {
		case canRun := <-run:
			//本地没钱了就问服务器要，如果使用服务器的余额判断，因为延迟关系，本地早没钱了，
			//还在发送交易，传到服务器，服务器会接受到很多不存在的交易
			if accInfo.GetBalance(fromAcc) <= 1 { //每次交易就发1个token
				logger.Infof("Getting %v Tokens from Miner...\n", config.AmountFromMinner)
				utxoTx = ser.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
			}
			if canRun && accInfo.GetBalance(fromAcc) > 1 {
				logger.Infof("Sending 1 Token from %s to %s...\n", shortenAddress(fromAcc), shortenAddress(toAcc))
				ser.SendToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, fromAcc, toAcc)
			} else if !canRun {
				time.Sleep(2)
				runtime.Goexit() //退出go线程
			}
		}
	}
}

func printBalance(accInfo *account_ron.AccountInfo) {
	if accInfo != nil {
		fromAddrs := accInfo.FromAddress
		for i, from := range fromAddrs {
			to := accInfo.ToAddress[i]
			logger.WithFields(logger.Fields{
				("From " + shortenAddress(from)): accInfo.GetBalance(from),
				("To " + shortenAddress(to)):     accInfo.GetBalance(to),
			}).Infof("Balance of Account Pair %v", i)
		}
	}
}

func shortenAddress(address string) string {
	if len(address) > 6 {
		return address[:6] + "..."
	}
	return address
}
