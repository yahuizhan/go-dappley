package logic

import (
	"flag"
	"fmt"
	"github.com/dappley/go-dappley/config"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/utxo"
	account_ron "github.com/dappley/go-dappley/tool/dependable_testing/sdk"
	"github.com/dappley/go-dappley/tool/dependable_testing/service"
	dependable_configpb "github.com/dappley/go-dappley/tool/dependable_testing/pb"
	logger "github.com/sirupsen/logrus"
	"math"
	"runtime"
	"sync"
	"time"
)

const (
	configFilePath = "../dependable_testing/default.conf"
	logFreshTime   = 5  //刷新日志间隔
)
//说明：default.conf: goCount设置为10，tps为2
func TestSmartContract() {
	configs := &dependable_configpb.Config{}
	config.LoadConfig(configFilePath, configs)
	var filePath string
	flag.StringVar(&filePath, "f", configFilePath, "Configuration File Path. Default to conf/default.conf")
	//config
	runTime24 := 60 //运行交易时间  1小时=3,600，24小时=86,400

	//网络服务
	serviceClient := service.NewServiceClient(configs.GetIp(), configs.GetPort())
	minerAccount := account.NewAccountByPrivateKey(configs.GetMinerPrivKey())
	//存放创建的aacount
	acInfo := account_ron.NewAccountInfo()

	logger.Info("智能合约测试场景...")
	logger.Info("测试目的：")
	logger.Info("验证系统智能合约场景下的表现")
	logger.Info("测试步骤：")
	logger.Info("在", runTime24, "秒内,向服务器持续发送交易请求,TPS为",configs.GoCount*configs.Tps,",然后验证交易成功率")
	logger.Info("正在初始化...")

	stopChan := make(chan bool,6000)
	startTest := false
	var goGoroutineMap sync.Map
	//交易生成
	for i := int32(0); i < configs.GetGoCount(); i++ {
		go StartTransactionGoroutine(
			i,
			&goGoroutineMap,
			serviceClient,
			acInfo,
			minerAccount.GetAddress().String(),
			configs,
			&startTest,
			stopChan)
	}

	//等待所有账户拿到钱
	var sum uint64
	for {
		goGoroutineMap.Range(func(k, v interface{}) bool {
			sum ++
			return true
		})
		if sum == uint64(configs.GetGoCount()) {
			logger.Info("测试工具初始化完成")  //部署合约完成
			break
		}
		sum = 0
		time.Sleep(100 * time.Millisecond)
		//for i := int32(0); i < configs.GetGoCount(); i++ {
		//	_,okay:= goGoroutineMap.Load(i)
		//	if !okay{
		//		fmt.Println("go runtime not exit:",i)
		//	}
		//}
	}

	//acInfo.WaitTillGetToken(serviceClient,configs.GetAmountFromMinner()*uint64(configs.GetGoCount()))

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


//开始交易，from问矿工要钱，再给to,没钱了再问矿工要，一直重复
func StartTransactionGoroutine(i int32,goGoroutineMap *sync.Map,service *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *dependable_configpb.Config, start *bool, stop chan bool) {
	fromAccount,_ := accInfo.CreateFromAccount()
	fromAcc := fromAccount.GetAddress().String()

	var utxoTx *utxo.UTXOTx
	if accInfo.GetBalance(fromAcc) < config.AmountPerTx {
		utxoTx = service.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner+1)
	}
	if utxoTx == nil{  // 并发太多会出现转账失败
		goGoroutineMap.Store(i,1)
		return
	}
	time.Sleep(1*time.Second)
	contract := `'use strict';

			var StepRecorder = function () {
			};

			StepRecorder.prototype = {
   				 record: function (addr, steps) {
       			 var originalSteps = LocalStorage.get(addr);
        		 LocalStorage.set(addr, originalSteps + steps);
       			 return _native_reward.record(addr, steps);
    		},
			dapp_schedule: function () {
			}
		};
		module.exports = new StepRecorder();
	`
	var contractAddr string
	if accInfo.GetBalance(fromAcc) > config.AmountPerTx  {
		contractAddr = service.DeploySmartContract(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, fromAcc, "",1,30000,1,contract)
	}
	accInfo.Lock()
	accInfo.ToAddress = append(accInfo.ToAddress, contractAddr)
	accInfo.Balances[contractAddr] = 1
	accInfo.Unlock()
	utxoTx = service.GetUtxo(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
	goGoroutineMap.Store(i,1)
	ticker := time.NewTicker(time.Microsecond * time.Duration(1000000/config.Tps)) //定时1秒
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//本地没钱了就问服务器要，如果使用服务器的余额判断，因为延迟关系，本地早没钱了，
			//还在发送交易，传到服务器，服务器会接受到很多不存在的交易。
			if accInfo.GetBalance(fromAcc) <= config.AmountPerTx { //每次交易就发1个token
				utxoTx = service.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
			}
			if utxoTx == nil{  // 并发太多会出现转账失败
				continue
			}
			InvokingContract := `{"function":"record","args":["dYgmFyXLg5jSfbysWoZF7Zimnx95xg77Qo","2000"]}`
			if accInfo.GetBalance(fromAcc) > config.AmountPerTx && *start {
				service.InvokingSmartContract(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, fromAcc, contractAddr,1,30000,1,InvokingContract)
			}
		case <-stop:
			time.Sleep(2)
			runtime.Goexit() //退出go线程
		}
	}
}

func CheckTransactionNumber(acInfo *account_ron.AccountInfo, serviceClient *service.Service) (int64, int64) {
	timeOut:= 25 //交易结束发送后，等待交易打包最长时间,单位秒
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
			toSum = toSum + int64(serviceClient.GetBalance(address))
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
				toSum = toSum + int64(serviceClient.GetBalance(address))
			}
			if toSum>=lastToSum{
				logger.Info("交易发送中，当前TPS：", (toSum-lastToSum)/int64(logFreshTime))
			}else{
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
func StartMinerTransactionGoroutine(i int32,goGoroutineMap *sync.Map,service *service.Service, accInfo *account_ron.AccountInfo, minnerAcc string, config *dependable_configpb.Config, start *bool, stop chan bool) {
	fromAccount,_ := accInfo.CreateFromAccount()
	fromAcc := fromAccount.GetAddress().String()
	var utxoTx *utxo.UTXOTx
	if accInfo.GetBalance(fromAcc) < config.AmountPerTx {
		utxoTx = service.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner+1)
	}
	if utxoTx == nil{  // 并发太多会出现转账失败
		goGoroutineMap.Store(i,1)
		return
	}
	//time.Sleep(1*time.Second)
	//contract := `'use strict';
	//
	//		var StepRecorder = function () {
	//		};
	//
	//		StepRecorder.prototype = {
	//			 record: function (addr, steps) {
	//   			 var originalSteps = LocalStorage.get(addr);
	//    		 LocalStorage.set(addr, originalSteps + steps);
	//   			 return _native_reward.record(addr, steps);
	//		},
	//		dapp_schedule: function () {
	//		}
	//	};
	//	module.exports = new StepRecorder();
	//`
	//fmt.Println("StartTransactionGoroutine DeploySmartContract:",i)
	//var contractAddr string
	//if accInfo.GetBalance(fromAcc) > config.AmountPerTx  {
	//	contractAddr = service.DeploySmartContract(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, fromAcc, "",1,30000,1,contract)
	//}
	//fmt.Println("fromAcc:",fromAcc)
	//fmt.Println("contractAddr:",contractAddr)
	//accInfo.Lock()
	//accInfo.ToAddress = append(accInfo.ToAddress, contractAddr)
	//fmt.Println("goRoutine：",i)
	//fmt.Println("goRoutine toaddress count:",len(accInfo.ToAddress))
	//accInfo.Balances[contractAddr] = 1
	//accInfo.Unlock()

	goGoroutineMap.Store(i,1)
	ticker := time.NewTicker(time.Microsecond * time.Duration(1000000/config.Tps)) //定时1秒
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//本地没钱了就问服务器要，如果使用服务器的余额判断，因为延迟关系，本地早没钱了，
			//还在发送交易，传到服务器，服务器会接受到很多不存在的交易。
			//if accInfo.GetBalance(fromAcc) <= config.AmountPerTx { //每次交易就发1个token
			//fmt.Println("fromAcc getBalance 222222:",accInfo.GetBalance(fromAcc)," fromAcc:",fromAcc)
			utxoTx = service.GetToken(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
			//} else {
			//	utxoTx = service.GetUtxo(accInfo, minnerAcc, fromAcc, config.AmountFromMinner)
			//}
			//if utxoTx == nil{  // 并发太多会出现转账失败
			//continue
			//}
			//fmt.Println("fromAcc getBalance 33333333:",accInfo.GetBalance(fromAcc)," fromAcc:",fromAcc)
			//InvokingContract := `{"function":"record","args":["dYgmFyXLg5jSfbysWoZF7Zimnx95xg77Qo","2000"]}`
			//if accInfo.GetBalance(fromAcc) > config.AmountPerTx && *start {
			//	fmt.Println("fromAcc getBalance 4444444:",accInfo.GetBalance(fromAcc)," fromAcc:",fromAcc)
			//	service.InvokingSmartContract(fromAccount.GetPubKeyHash(), utxoTx, accInfo, 1, fromAcc, contractAddr,1,30000,1,InvokingContract)
			//}
		case <-stop:
			runtime.Goexit() //退出go线程
		}
	}
}