package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
	logger.Info("TPS为", float32(configs.GoCount)*configs.Tps)
	logger.Info("")
	logger.Info("正在初始化...")

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())

	acInfo := account_ron.NewAccountInfo()
	var sendTxSignals []bool

	var err error
	acInfo.Accounts, err = account_ron.ReadAccountFromFile()
	numRoutines := 0
	if err != nil {
		logger.Info("未找到account.dat，根据default启动测试，")
		logger.Info("正在向矿工获取token...")
		numRoutines = int(configs.GetGoCount())
		sendTxSignals = make([]bool, numRoutines)
		//交易生成
		for i := int32(0); i < configs.GetGoCount(); i++ {
			go startTxGoroutine(
				serviceClient,
				acInfo,
				minerAccount.GetAddress().String(),
				configs,
				&sendTxSignals[i])
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
		numRoutines = lenAccount / 2
		sendTxSignals = make([]bool, numRoutines)
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

			go startTxFromFile(
				serviceClient,
				acInfo,
				minerAccount.GetAddress().String(),
				configs,
				&sendTxSignals[i/2],
				fromAccount,
				toAccount)
		}
		acInfo.WaitTillGetToken(total)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("[Enter 'S' To Send Transactions; Enter 'B' To Check Local Balance; Enter 'E' To Exit] ")
		input, _ := reader.ReadString('\n')
		switch strings.TrimSpace(input) {
		case "s", "S":
			logger.Info("Sending transactions ...")
			for i := 0; i < numRoutines; i++ {
				sendTxSignals[i] = true
			}
			for i := 0; i < numRoutines; i++ {
				waitTillSent(&sendTxSignals[i])
			}
			break
		case "b", "B":
			printBalance(acInfo)
			break
		case "e", "E":
			logger.Info("Exiting All Go Routines ...")
			os.Exit(0)
		default:
			logger.Warn("Input is not valid")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func startTxGoroutine(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, sendTxSignal *bool) {
	fromAccount, toAccount := accInfo.CreateAccountPair()
	var utxoTx *utxo.UTXOTx
	startTx(ser, accInfo, minnerAcc, config, sendTxSignal, fromAccount, toAccount, utxoTx)
}

func startTxFromFile(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, sendTxSignal *bool, fromAccount, toAccount *account.Account) {
	fromAcc := fromAccount.GetAddress().String()
	utxoTx, err := ser.GetUTXOTxFromServer(fromAcc)
	if err != nil {
		logger.Error("Get UTXOTx error:", err)
		os.Exit(2)
	}
	startTx(ser, accInfo, minnerAcc, config, sendTxSignal, fromAccount, toAccount, utxoTx)
}

func startTx(ser *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *performance_configpb.Config, sendTxSignal *bool, fromAccount, toAccount *account.Account, utxoTx *utxo.UTXOTx) {
	fromAcc := fromAccount.GetAddress().String()
	toAcc := toAccount.GetAddress().String()

	ticker := time.NewTicker(time.Microsecond * time.Duration(1000000/config.Tps)) //定时1秒
	defer ticker.Stop()
	var err error
	for {
		select {
		case <-ticker.C:
			//本地没钱了就问服务器要，如果使用服务器的余额判断，因为延迟关系，本地早没钱了，
			//还在发送交易，传到服务器，服务器会接受到很多不存在的交易
			if accInfo.GetBalance(fromAcc) <= 1 {
				logger.Infof("Getting %v Tokens from Miner...\n", config.AmountFromMinner)
				utxoTx, err = ser.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
				if err != nil {
					logger.Error("GetToken failed! error: ", err)
					os.Exit(1)
				}
			}
			if accInfo.GetBalance(fromAcc) > 1 && *sendTxSignal { //每次交易就发1个token
				logger.Infof("Sending 1 Token from %s to %s ...\n", shortenAddress(fromAcc), shortenAddress(toAcc))
				err = ser.SendToken(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, 0, fromAcc, toAcc)
				if err != nil {
					logger.Error("SendToken failed! error: ", err)
					os.Exit(1)
				}
				*sendTxSignal = false
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

func waitTillSent(signal *bool) {
	for *signal {
		time.Sleep(100 * time.Millisecond)
	}
}
