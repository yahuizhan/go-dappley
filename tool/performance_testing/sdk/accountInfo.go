package account_ron

import (
	"sync"
	"time"

	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/logic"
	logger "github.com/sirupsen/logrus"
)

//保存在本地的账户余额，账户余额是本地计算，并非链上的
type AccountInfo struct {
	FromAddress []string          //account address
	ToAddress   []string          //account address
	Balances    map[string]uint64 //和addrs 对应的balance
	Accounts    []*account.Account
	sync.RWMutex
}

func NewAccountInfo() *AccountInfo {
	return &AccountInfo{
		[]string{},
		[]string{},
		make(map[string]uint64),
		[]*account.Account{},
		sync.RWMutex{},
	}
}

//本地创建账户
func (acc *AccountInfo) CreateAccount() (*account.Account, error) {
	newAccount, err := logic.CreateAccountWithPassphrase("123")
	if err != nil {
		logger.WithError(err).Error("Cannot create new account.")
	}
	//logger.WithFields(logger.Fields{
	//	"address": newAccount.GetAddress(),
	//}).Info("Account is created")
	return newAccount, err
	//return  ReadAccountFromFile(),err
}

func (acc *AccountInfo) AddFromAccountInfo(acct *account.Account) {
	acc.Lock()
	acc.FromAddress = append(acc.FromAddress, acct.GetAddress().String())
	acc.Balances[acct.GetAddress().String()] = 0
	acc.Accounts = append(acc.Accounts, acct)
	acc.Unlock()
}

func (acc *AccountInfo) AddToAccountInfo(acct *account.Account) {
	acc.Lock()
	acc.ToAddress = append(acc.ToAddress, acct.GetAddress().String())
	acc.Balances[acct.GetAddress().String()] = 0
	acc.Accounts = append(acc.Accounts, acct)
	acc.Unlock()
}

//本地自己更新的账户余额，并非链上
func (acc *AccountInfo) UpdateBalance(address string, balance uint64) {
	acc.Lock()
	acc.Balances[address] = acc.Balances[address] + balance
	acc.Unlock()
}

func (acc *AccountInfo) GetBalance(address string) uint64 {
	acc.RLock()
	balance := acc.Balances[address]
	acc.RUnlock()
	return balance

}

//通过地址得到Account
func (acc *AccountInfo) GetAccount(address string) *account.Account {
	for key, acct := range acc.Accounts {
		if acct.GetAddress().String() == address {
			return acc.Accounts[key]
		}
	}
	return nil
}

func (acc *AccountInfo) CreateAccountPair() (*account.Account, *account.Account) {
	fromAccount, _ := acc.CreateAccount()
	acc.AddFromAccountInfo(fromAccount)
	toAccount, _ := acc.CreateAccount()
	acc.AddToAccountInfo(toAccount)
	return fromAccount, toAccount
}

func (acc *AccountInfo) WaitTillGetToken(total uint64) {
	var sum uint64
	for sum < total {
		sum = 0
		for _, value := range acc.Balances {
			sum = sum + value
		}
		time.Sleep(100 * time.Millisecond)
	}
	logger.Info("测试工具初始化完成")
}
