// +build !release

// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-dappley library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//
package main

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dappley/go-dappley/logic/ltransaction"

	"github.com/dappley/go-dappley/core/transaction"
	transactionpb "github.com/dappley/go-dappley/core/transaction/pb"
	"github.com/dappley/go-dappley/core/utxo"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/config"
	configpb "github.com/dappley/go-dappley/config/pb"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/crypto/keystore/secp256k1"
	"github.com/dappley/go-dappley/logic"
	rpcpb "github.com/dappley/go-dappley/rpc/pb"
	"github.com/dappley/go-dappley/util"
	"github.com/dappley/go-dappley/wallet"
	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

//command names
const (
	cliGetBlocks         = "getBlocks"
	cliGetBlockchainInfo = "getBlockchainInfo"
	cliGetBalance        = "getBalance"
	cliGetPeerInfo       = "getPeerInfo"
	cliSend              = "send"
	cliAddPeer           = "addPeer"
	clicreateAccount     = "createAccount"
	cliListAddresses     = "listAddresses"
	clisendFromMiner     = "sendFromMiner"
	cliaddProducer       = "addProducer"
	cliEstimateGas       = "estimateGas"
	cliGasPrice          = "gasPrice"
	cliContractQuery     = "contractQuery"
	cliHelp              = "help"
	cliGetMetricsInfo    = "getMetricsInfo"
	cliGetBlockByHeight  = "getBlockByHeight"
)

//flag names
const (
	flagStartBlockHashes = "startBlockHashes"
	flagBlockMaxCount    = "maxCount"
	flagAddress          = "address"
	flagAddressBalance   = "to"
	flagAmountBalance    = "amount"
	flagTip              = "tip"
	flagToAddress        = "to"
	flagFromAddress      = "from"
	flagAmount           = "amount"
	flagData             = "data"
	flagFilePath         = "file"
	flagPeerFullAddr     = "peerFullAddr"
	flagProducerAddr     = "address"
	flagListPrivateKey   = "privateKey"
	flagGasLimit         = "gasLimit"
	flagGasPrice         = "gasPrice"
	flagContractAddr     = "contractAddr"
	flagKey              = "key"
	flagValue            = "value"
	flagBlockHeight      = "height"
)

type valueType int

//type enum
const (
	valueTypeInt = iota
	valueTypeString
	boolType
	valueTypeUint64
)

type serviceType int

const (
	rpcService = iota
	adminRpcService
	metricsRpcService
)

//list of commands
var cmdList = []string{
	cliGetBlocks,
	cliGetBlockchainInfo,
	cliGetBalance,
	cliGetPeerInfo,
	cliSend,
	cliAddPeer,
	clicreateAccount,
	cliListAddresses,
	clisendFromMiner,
	cliaddProducer,
	cliEstimateGas,
	cliGasPrice,
	cliContractQuery,
	cliHelp,
	cliGetMetricsInfo,
	cliGetBlockByHeight,
}

var (
	ErrInsufficientFund = errors.New("cli: the balance is insufficient")
)

//configure input parameters/flags for each command
var cmdFlagsMap = map[string][]flagPars{
	cliGetBlocks: {
		flagPars{
			flagBlockMaxCount,
			0,
			valueTypeInt,
			"maxCount. Eg. 500",
		},
		flagPars{
			flagStartBlockHashes,
			"",
			valueTypeString,
			"startBlockHashes. Eg. \"8334b4c19091ae7582506eec5b84bfeb4a5e101042e40b403490c4ceb33897ba, " +
				"8334b4c19091ae7582506eec5b84bfeb4a5e101042e40b403490c4ceb33897bb\"(no space)",
		},
	},
	cliGetBalance: {flagPars{
		flagAddress,
		"",
		valueTypeString,
		"Address. Eg. 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7",
	}},
	cliaddProducer: {flagPars{
		flagProducerAddr,
		"",
		valueTypeString,
		"Producer's address. Eg. 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7",
	}},
	clisendFromMiner: {
		flagPars{
			flagAddressBalance,
			"",
			valueTypeString,
			"Reciever's address. Eg. 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7"},
		flagPars{
			flagAmountBalance,
			0,
			valueTypeInt,
			"The amount to be sent to the receiver.",
		},
	},
	cliSend: {
		flagPars{
			flagFromAddress,
			"",
			valueTypeString,
			"Sender's account address. Eg. 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7",
		},
		flagPars{
			flagToAddress,
			"",
			valueTypeString,
			"Receiver's account address. Eg. 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7",
		},
		flagPars{
			flagAmount,
			0,
			valueTypeInt,
			"The amount to send from the sender to the receiver.",
		},
		flagPars{
			flagTip,
			uint64(0),
			valueTypeUint64,
			"Tip to miner.",
		},
		flagPars{
			flagData,
			"",
			valueTypeString,
			"Smart contract in JavaScript. Eg. helloworld!",
		},
		flagPars{
			flagFilePath,
			"",
			valueTypeString,
			"Smart contract file path. Eg. contract/smart_contract.js",
		},
		flagPars{
			flagGasLimit,
			uint64(0),
			valueTypeUint64,
			"Gas limit count of smart contract execution.",
		},
		flagPars{
			flagGasPrice,
			uint64(0),
			valueTypeUint64,
			"Gas price of smart contract execution.",
		},
	},
	cliAddPeer: {flagPars{
		flagPeerFullAddr,
		"",
		valueTypeString,
		"Full Address. Eg. /ip4/127.0.0.1/tcp/12345/ipfs/QmT5oB6xHSunc64Aojoxa6zg9uH31ajiAVyNfCdBZiwFTV",
	}},
	cliListAddresses: {flagPars{
		flagListPrivateKey,
		false,
		boolType,
		"with/without this optional argument to display the private keys or not",
	}},
	cliEstimateGas: {
		flagPars{
			flagFromAddress,
			"",
			valueTypeString,
			"Sender's account address. Eg. 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7",
		},
		flagPars{
			flagToAddress,
			"",
			valueTypeString,
			"Receiver's account address. Eg. 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7",
		},
		flagPars{
			flagAmount,
			0,
			valueTypeInt,
			"The amount to send from the sender to the receiver.",
		},
		flagPars{
			flagTip,
			uint64(0),
			valueTypeUint64,
			"Tip to miner.",
		},
		flagPars{
			flagData,
			"",
			valueTypeString,
			"Smart contract in JavaScript. Eg. helloworld!",
		},
		flagPars{
			flagFilePath,
			"",
			valueTypeString,
			"Smart contract file path. Eg. contract/smart_contract.js",
		},
		flagPars{
			flagGasLimit,
			uint64(0),
			valueTypeUint64,
			"Gas limit count of smart contract execution.",
		},
		flagPars{
			flagGasPrice,
			uint64(0),
			valueTypeUint64,
			"Gas price of smart contract execution.",
		},
	},
	cliGasPrice:       {},
	cliGetMetricsInfo: {},
	cliContractQuery: {
		flagPars{
			flagContractAddr,
			"",
			valueTypeString,
			"Contract address. Eg. cd9N6MRsYxU1ToSZjLnqFhTb66PZcePnAD",
		},
		flagPars{
			flagKey,
			"",
			valueTypeString,
			"The data key storaged in contract address.",
		},
		flagPars{
			flagValue,
			"",
			valueTypeString,
			"The data value storaged in contract address.",
		},
	},
	cliGetBlockByHeight: {
		flagPars{
			flagBlockHeight,
			0,
			valueTypeInt,
			"height. Eg. 1",
		},
	},
}

//map the callback function to each command
var cmdHandlers = map[string]commandHandlersWithType{
	cliGetBlocks:         {rpcService, getBlocksCommandHandler},
	cliGetBlockchainInfo: {rpcService, getBlockchainInfoCommandHandler},
	cliGetBalance:        {rpcService, getBalanceCommandHandler},
	cliGetPeerInfo:       {adminRpcService, getPeerInfoCommandHandler},
	cliSend:              {rpcService, sendCommandHandler},
	cliAddPeer:           {adminRpcService, addPeerCommandHandler},
	clicreateAccount:     {adminRpcService, createAccountCommandHandler},
	cliListAddresses:     {adminRpcService, listAddressesCommandHandler},
	clisendFromMiner:     {adminRpcService, sendFromMinerCommandHandler},
	cliaddProducer:       {adminRpcService, cliAddProducerCommandHandler},
	cliEstimateGas:       {rpcService, estimateGasCommandHandler},
	cliGasPrice:          {rpcService, gasPriceCommandHandler},
	cliHelp:              {adminRpcService, helpCommandHandler},
	cliContractQuery:     {rpcService, contractQueryCommandHandler},
	cliGetMetricsInfo:    {metricsRpcService, getMetricsInfoCommandHandler},
	cliGetBlockByHeight:  {rpcService, getBlockByHeightCommandHandler},
}

type commandHandlersWithType struct {
	serviceType serviceType
	cmdHandler  commandHandler
}

type commandHandler func(ctx context.Context, account interface{}, flags cmdFlags)

type flagPars struct {
	name         string
	defaultValue interface{}
	valueType    valueType
	usage        string
}

//map key: flag name   map defaultValue: flag defaultValue
type cmdFlags map[string]interface{}

func main() {

	var filePath string
	flag.StringVar(&filePath, "f", "default.conf", "CLI config file path")
	flag.Parse()

	cliConfig := &configpb.CliConfig{}
	config.LoadConfig(filePath, cliConfig)

	conn := initRpcClient(int(cliConfig.GetPort()))
	defer conn.Close()
	clients := map[serviceType]interface{}{
		rpcService:        rpcpb.NewRpcServiceClient(conn),
		adminRpcService:   rpcpb.NewAdminServiceClient(conn),
		metricsRpcService: rpcpb.NewMetricServiceClient(conn),
	}
	args := os.Args[1:]

	if len(args) < 1 {
		printUsage()
		return
	}

	if args[0] == "-f" {
		args = args[2:]
	}

	cmdFlagSetList := map[string]*flag.FlagSet{}
	//set up flagset for each command
	for _, cmd := range cmdList {
		fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
		cmdFlagSetList[cmd] = fs
	}

	cmdFlagValues := map[string]cmdFlags{}
	//set up flags for each command
	for cmd, pars := range cmdFlagsMap {
		cmdFlagValues[cmd] = cmdFlags{}
		for _, par := range pars {
			switch par.valueType {
			case valueTypeInt:
				cmdFlagValues[cmd][par.name] = cmdFlagSetList[cmd].Int(par.name, par.defaultValue.(int), par.usage)
			case valueTypeString:
				cmdFlagValues[cmd][par.name] = cmdFlagSetList[cmd].String(par.name, par.defaultValue.(string), par.usage)
			case boolType:
				cmdFlagValues[cmd][par.name] = cmdFlagSetList[cmd].Bool(par.name, par.defaultValue.(bool), par.usage)
			case valueTypeUint64:
				cmdFlagValues[cmd][par.name] = cmdFlagSetList[cmd].Uint64(par.name, par.defaultValue.(uint64), par.usage)
			}
		}
	}

	cmdName := args[0]

	cmd := cmdFlagSetList[cmdName]
	if cmd == nil {
		fmt.Println("\nError:", cmdName, "is an invalid command")
		printUsage()
	} else {
		err := cmd.Parse(args[1:])
		if err != nil {
			return
		}
		if cmd.Parsed() {
			md := metadata.Pairs("password", cliConfig.GetPassword())
			ctx := metadata.NewOutgoingContext(context.Background(), md)
			cmdHandlers[cmdName].cmdHandler(ctx, clients[cmdHandlers[cmdName].serviceType], cmdFlagValues[cmdName])
		}
	}

}

func printUsage() {
	fmt.Println("Usage:")
	for _, cmd := range cmdList {
		fmt.Println(" ", cmd)
	}
	fmt.Println("Note: Use the command 'cli help' to get the command usage in details")
}

func getMetricsInfoCommandHandler(ctx context.Context, c interface{}, flags cmdFlags) {
	var flag bool = true
	file, err := os.Create("metricsInfo_result.csv")
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	writer.Comma = ','
	tick := time.NewTicker(time.Duration(5000) * time.Millisecond)
	for {
		select {
		case <-tick.C:
			metricsServiceRequest := rpcpb.MetricsServiceRequest{}
			metricsInfoResponse, err := c.(rpcpb.MetricServiceClient).RpcGetMetricsInfo(ctx, &metricsServiceRequest)
			if err != nil {
				switch status.Code(err) {
				case codes.Unavailable:
					fmt.Println("Error: server is not reachable!")
				default:
					fmt.Println("Error:", err.Error())
				}
				return
			}
			fmt.Println("metricsInfo:", metricsInfoResponse.Data)

			m, ok := gjson.Parse(metricsInfoResponse.Data).Value().(map[string]interface{})
			if !ok {
				fmt.Println("parse data is not json")
				continue
			}
			var titleStr []string
			var metricsInfostr []string
			metricsInfoMap := make(map[string]string)
			for key, value := range m {
				if value != nil {
					childValue := value.(map[string]interface{})
					for cKey, cValue := range childValue {
						if cKey == "txRequestSend" || cKey == "txRequestSendFromMiner" {
							grandChildValue := cValue.(map[string]interface{})
							for gcKey, gcValue := range grandChildValue {
								if gcValue != nil {
									switch v := gcValue.(type) {
									case float64:
										metricsInfoMap[key+":"+cKey+":"+gcKey] = strconv.Itoa(int(v))
									}
								}
							}
						} else {
							switch v := cValue.(type) {
							case float64:
								metricsInfoMap[key+":"+cKey] = strconv.Itoa(int(v))
							}
						}
					}
				}
			}
			var keys []string
			for key := range metricsInfoMap {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for i := 0; i < len(keys); i++ {
				value := metricsInfoMap[keys[i]]
				metricsInfostr = append(metricsInfostr, value)
				titleStr = append(titleStr, keys[i])
			}
			var strArray [][]string
			if flag == true {
				strArray = append(strArray, titleStr)
			}
			strArray = append(strArray, metricsInfostr)
			flag = false
			writer.WriteAll(strArray)
			writer.Flush()
		}
	}
}

func getBlocksCommandHandler(ctx context.Context, account interface{}, flags cmdFlags) {
	maxCount := int32(*(flags[flagBlockMaxCount].(*int)))
	if maxCount <= 0 {
		fmt.Println("\n Example: cli getBlocks -startBlockHashes 10 -maxCount 5")
		fmt.Println()
		return
	}

	getBlocksRequest := &rpcpb.GetBlocksRequest{MaxCount: maxCount}

	// set startBlockHashes of getBlocksRequest if specified in flag
	startBlockHashesString := string(*(flags[flagStartBlockHashes].(*string)))
	if len(startBlockHashesString) > 0 {
		var startBlockHashes [][]byte
		for _, startBlockHash := range strings.Split(startBlockHashesString, ",") {
			startBlockHashInByte, err := hex.DecodeString(startBlockHash)
			if err != nil {
				fmt.Println("Error:", err.Error())
				return
			}
			startBlockHashes = append(startBlockHashes, startBlockHashInByte)
		}
		getBlocksRequest = &rpcpb.GetBlocksRequest{MaxCount: maxCount, StartBlockHashes: startBlockHashes}
	}

	response, err := account.(rpcpb.RpcServiceClient).RpcGetBlocks(ctx, getBlocksRequest)
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}

	var encodedBlocks []map[string]interface{}
	for _, block := range response.Blocks {

		var encodedTransactions []map[string]interface{}

		for _, transaction := range block.GetTransactions() {

			var encodedVin []map[string]interface{}
			for _, vin := range transaction.GetVin() {
				encodedVin = append(encodedVin, map[string]interface{}{
					"Vout":      vin.GetVout(),
					"Signature": hex.EncodeToString(vin.GetSignature()),
					"PubKey":    hex.EncodeToString(vin.GetPublicKey()),
				})
			}

			var encodedVout []map[string]interface{}
			for _, vout := range transaction.GetVout() {
				encodedVout = append(encodedVout, map[string]interface{}{
					"Value":      common.NewAmountFromBytes(vout.GetValue()),
					"PubKeyHash": hex.EncodeToString(vout.GetPublicKeyHash()),
					"Contract":   vout.GetContract(),
				})
			}

			encodedTransaction := map[string]interface{}{
				"ID":   hex.EncodeToString(transaction.GetId()),
				"Vin":  encodedVin,
				"Vout": encodedVout,
			}
			encodedTransactions = append(encodedTransactions, encodedTransaction)
		}

		encodedBlock := map[string]interface{}{
			"Header": map[string]interface{}{
				"Hash":      hex.EncodeToString(block.GetHeader().GetHash()),
				"Prevhash":  hex.EncodeToString(block.GetHeader().GetPreviousHash()),
				"Timestamp": time.Unix(block.GetHeader().GetTimestamp(), 0).String(),
				"Sign":      hex.EncodeToString(block.GetHeader().GetSignature()),
				"height":    block.GetHeader().GetHeight(),
			},
			"Transactions": encodedTransactions,
		}

		encodedBlocks = append(encodedBlocks, encodedBlock)
	}

	blocks, err := json.MarshalIndent(encodedBlocks, "", "  ")
	if err != nil {
		fmt.Println("Error:", err.Error())
	}

	fmt.Println(string(blocks))
}

func getBlockchainInfoCommandHandler(ctx context.Context, account interface{}, flags cmdFlags) {
	response, err := account.(rpcpb.RpcServiceClient).RpcGetBlockchainInfo(ctx, &rpcpb.GetBlockchainInfoRequest{})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}
	encodedResponse := map[string]interface{}{
		"TailBlockHash": hex.EncodeToString(response.TailBlockHash),
		"BlockHeight":   response.BlockHeight,
		"Producers":     response.Producers,
	}

	blockchainInfo, err := json.MarshalIndent(encodedResponse, "", "  ")
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	fmt.Println(string(blockchainInfo))
}

func getBalanceCommandHandler(ctx context.Context, c interface{}, flags cmdFlags) {
	if len(*(flags[flagAddress].(*string))) == 0 {
		printUsage()
		fmt.Println("\n Example: cli getBalance -address 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7")
		fmt.Println()
		return
	}

	address := *(flags[flagAddress].(*string))
	addressAccount := account.NewTransactionAccountByAddress(account.NewAddress(address))
	if !addressAccount.IsValid() {
		fmt.Println("Error: address is not valid")
		return
	}

	response, err := c.(rpcpb.RpcServiceClient).RpcGetBalance(ctx, &rpcpb.GetBalanceRequest{Address: address})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}
	fmt.Printf("The balance is: %d\n", response.GetAmount())
}

func createAccountCommandHandler(ctx context.Context, account interface{}, flags cmdFlags) {
	empty, err := logic.IsAccountEmpty()
	prompter := util.NewTerminalPrompter()
	passphrase := ""
	if empty {
		passphrase = prompter.GetPassPhrase("Please input the password for the new account: ", true)
		if passphrase == "" {
			fmt.Println("Error: password cannot be empty!")
			return
		}
		account, err := logic.CreateAccountWithPassphrase(passphrase)
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
		if account != nil {
			fmt.Printf("Account is created. The address is %s \n", account.GetAddress().String())
			return
		}
	}

	locked, err := logic.IsAccountLocked()
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	if locked {
		passphrase = prompter.GetPassPhrase("Please input the password: ", false)
		if passphrase == "" {
			fmt.Println("Error: password should not be empty!")
			return
		}
		acc, err := logic.CreateAccountWithPassphrase(passphrase)
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
		if account != nil {
			fmt.Printf("Account is created. The address is %s\n", acc.GetAddress().String())
		}
		//unlock the account
		_, err = account.(rpcpb.AdminServiceClient).RpcUnlockAccount(ctx, &rpcpb.UnlockAccountRequest{})

		if err != nil {
			switch status.Code(err) {
			case codes.Unavailable:
				fmt.Println("Error: server is not reachable!")
			default:
				fmt.Println("Error:", status.Convert(err).Message())
			}
			return
		}
	} else {
		account, err := logic.CreateAccount()
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
		if account != nil {
			fmt.Printf("Account is created. The address is %s\n", account.GetAddress().String())
		}
	}

	return
}

func listAddressesCommandHandler(ctx context.Context, c interface{}, flags cmdFlags) {
	listPriv := false
	if flags[flagListPrivateKey] == nil {
		return
	} else if *(flags[flagListPrivateKey].(*bool)) {
		listPriv = true
	} else {
		listPriv = false
	}

	passphrase := ""
	prompter := util.NewTerminalPrompter()

	empty, err := logic.IsAccountEmpty()
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	if empty {
		fmt.Println("Please use cli createAccount to generate a account first!")
		return
	}

	locked, err := logic.IsAccountLocked()
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	if locked {
		passphrase = prompter.GetPassPhrase("Please input the password: ", false)
		if passphrase == "" {
			fmt.Println("Password should not be empty!")
			return
		}
		am, err := logic.GetAccountManager(wallet.GetAccountFilePath())
		addressList, err := am.GetAddressesWithPassphrase(passphrase)
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
		//unlock the account
		_, err = c.(rpcpb.AdminServiceClient).RpcUnlockAccount(ctx, &rpcpb.UnlockAccountRequest{})
		if err != nil {
			switch status.Code(err) {
			case codes.Unavailable:
				fmt.Println("Error: server is not reachable!")
			default:
				fmt.Println("Error:", status.Convert(err).Message())
			}
		}
		if !listPriv {
			if len(addressList) == 0 {
				fmt.Println("The addresses in the account is empty!")
			} else {
				i := 1
				fmt.Println("The address list:")
				for _, addr := range addressList {
					fmt.Printf("Address[%d]: %s\n", i, addr)
					i++
				}
				fmt.Println()
				fmt.Println("Use the command 'cli listAddresses -privateKey' to list the addresses with private keys")
			}
		} else {
			privateKeyList := []string{}
			for _, addr := range addressList {
				keyPair := am.GetKeyPairByAddress(account.NewAddress(addr))
				pvk := keyPair.GetPrivateKey()
				privateKey, err1 := secp256k1.FromECDSAPrivateKey(&pvk)
				if err1 != nil {
					err = err1
					return
				}
				privateKeyList = append(privateKeyList, hex.EncodeToString(privateKey))
				err = err1
			}
			if len(addressList) == 0 {
				fmt.Println("The addresses in the account is empty!")
			} else {
				i := 1
				fmt.Println("The address list with private keys:")
				for _, addr := range addressList {
					fmt.Println("--------------------------------------------------------------------------------")
					fmt.Printf("Address[%d]: %s \nPrivate Key[%d]: %s", i, addr, i, privateKeyList[i-1])
					fmt.Println()
					i++
				}
				fmt.Println("--------------------------------------------------------------------------------")
			}

		}
	} else {
		am, err := logic.GetAccountManager(wallet.GetAccountFilePath())
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
		addressList := am.GetAddresses()
		if !listPriv {
			if len(addressList) == 0 {
				fmt.Println("The addresses in the account is empty!")
			} else {
				i := 1
				fmt.Println("The address list:")
				for _, addr := range addressList {
					fmt.Printf("Address[%d]: %s\n", i, addr.String())
					i++
				}
				fmt.Println()
				fmt.Println("Use the command 'cli listAddresses -privateKey' to list the addresses with private keys")
			}
		} else {
			privateKeyList := []string{}
			for _, addr := range addressList {
				keyPair := am.GetKeyPairByAddress(addr)
				pvk := keyPair.GetPrivateKey()
				privateKey, err1 := secp256k1.FromECDSAPrivateKey(&pvk)
				if err1 != nil {
					err = err1
					return
				}
				privateKeyList = append(privateKeyList, hex.EncodeToString(privateKey))
				err = err1
			}
			if len(addressList) == 0 {
				fmt.Println("The addresses in the account is empty!")
			} else {
				i := 1
				fmt.Println("The address list with private keys:")
				for _, addr := range addressList {
					fmt.Println("--------------------------------------------------------------------------------")
					fmt.Printf("Address[%d]: %s \nPrivate Key[%d]: %s", i, addr.String(), i, privateKeyList[i-1])
					fmt.Println()
					i++
				}
				fmt.Println("--------------------------------------------------------------------------------")
			}

		}

	}
	return
}

func sendFromMinerCommandHandler(ctx context.Context, c interface{}, flags cmdFlags) {
	toAddr := *(flags[flagAddressBalance].(*string))
	if len(toAddr) == 0 {
		printUsage()
		fmt.Println("\n Example: cli sendFromMiner -to 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7 -amount 15")
		fmt.Println()
		return
	}
	amount := int64(*(flags[flagAmountBalance].(*int)))
	if amount <= 0 {
		fmt.Println("Error: amount must be greater than zero!")
		return
	}

	addressAccount := account.NewTransactionAccountByAddress(account.NewAddress(toAddr))
	if !addressAccount.IsValid() {
		fmt.Println("Error: address is invalid!")
		return
	}

	amountBytes := common.NewAmount(uint64(*(flags[flagAmountBalance].(*int)))).Bytes()
	sendFromMinerRequest := rpcpb.SendFromMinerRequest{To: toAddr, Amount: amountBytes}

	_, err := c.(rpcpb.AdminServiceClient).RpcSendFromMiner(ctx, &sendFromMinerRequest)
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", err.Error())
		}
		return
	}
	fmt.Println("Requested amount is sent. Pending approval from network.")
}

func getPeerInfoCommandHandler(ctx context.Context, account interface{}, flags cmdFlags) {
	response, err := account.(rpcpb.AdminServiceClient).RpcGetPeerInfo(ctx, &rpcpb.GetPeerInfoRequest{})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", err.Error())
		}
		return
	}
	fmt.Println(proto.MarshalTextString(response))
	fmt.Println("00000000")
}

func cliAddProducerCommandHandler(ctx context.Context, c interface{}, flags cmdFlags) {
	producerAddress := *(flags[flagProducerAddr].(*string))
	if len(producerAddress) == 0 {
		printUsage()
		fmt.Println("\n Example: cli addProducer -address 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7")
		fmt.Println()
		return
	}
	addressAccount := account.NewTransactionAccountByAddress(account.NewAddress(producerAddress))

	if !addressAccount.IsValid() {
		fmt.Println("Error: address is invalid")
		return
	}

	_, err := c.(rpcpb.AdminServiceClient).RpcAddProducer(ctx, &rpcpb.AddProducerRequest{
		Address: producerAddress,
	})

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}
	fmt.Println("Producer is added.")
}

func sendCommandHandler(ctx context.Context, c interface{}, flags cmdFlags) {
	var data string
	fromAddress := *(flags[flagFromAddress].(*string))
	addressAccount := account.NewTransactionAccountByAddress(account.NewAddress(fromAddress))
	path := *(flags[flagFilePath].(*string))
	if path == "" {
		data = *(flags[flagData].(*string))
	} else {
		script, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error: smart contract path \"%s\" is invalid.\n", path)
			return
		}
		data = string(script)
	}

	if !addressAccount.IsValid() {
		fmt.Println("Error: 'from' address is not valid!")
		return
	}

	//Contract deployment transaction does not need to validate to address
	if data == "" && !addressAccount.IsValid() {
		fmt.Println("Error: 'to' address is not valid!")
		return
	}

	response, err := c.(rpcpb.RpcServiceClient).RpcGetUTXO(ctx, &rpcpb.GetUTXORequest{
		Address: account.NewAddress(*(flags[flagFromAddress].(*string))).String(),
	})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}
	utxos := response.GetUtxos()
	var inputUtxos []*utxo.UTXO
	for _, u := range utxos {
		uu := utxo.UTXO{}
		uu.Value = common.NewAmountFromBytes(u.Amount)
		uu.Txid = u.Txid
		uu.PubKeyHash = account.PubKeyHash(u.PublicKeyHash)
		uu.TxIndex = int(u.TxIndex)
		inputUtxos = append(inputUtxos, &uu)
	}
	tip := common.NewAmount(0)
	gasLimit := common.NewAmount(0)
	gasPrice := common.NewAmount(0)
	if flags[flagTip] != nil {
		tip = common.NewAmount(*(flags[flagTip].(*uint64)))
	}
	if flags[flagGasLimit] != nil {
		gasLimit = common.NewAmount(*(flags[flagGasLimit].(*uint64)))
	}
	if flags[flagGasPrice] != nil {
		gasPrice = common.NewAmount(*(flags[flagGasPrice].(*uint64)))
	}
	tx_utxos, err := GetUTXOsfromAmount(inputUtxos, common.NewAmount(uint64(*(flags[flagAmount].(*int)))), tip, gasLimit, gasPrice)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	am, err := logic.GetAccountManager(wallet.GetAccountFilePath())
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	senderAccount := am.GetAccountByAddress(account.NewAddress(*(flags[flagFromAddress].(*string))))

	if senderAccount == nil {
		fmt.Println("Error: invalid account address.")
		return
	}
	sendTxParam := transaction.NewSendTxParam(account.NewAddress(*(flags[flagFromAddress].(*string))), senderAccount.GetKeyPair(),
		account.NewAddress(*(flags[flagToAddress].(*string))), common.NewAmount(uint64(*(flags[flagAmount].(*int)))), tip, gasLimit, gasPrice, data)
	tx, err := ltransaction.NewUTXOTransaction(tx_utxos, sendTxParam)
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err = c.(rpcpb.RpcServiceClient).RpcSendTransaction(ctx, sendTransactionRequest)

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}

	if *(flags[flagToAddress].(*string)) == "" {
		fmt.Println("Contract address:", tx.Vout[0].GetAddress().String())
	}

	fmt.Println("Transaction is sent! Pending approval from network.")
}

func GetUTXOsfromAmount(inputUTXOs []*utxo.UTXO, amount *common.Amount, tip *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount) ([]*utxo.UTXO, error) {
	if tip != nil {
		amount = amount.Add(tip)
	}
	if gasLimit != nil {
		limitedFee := gasLimit.Mul(gasPrice)
		amount = amount.Add(limitedFee)
	}
	var retUtxos []*utxo.UTXO
	sum := common.NewAmount(0)
	for _, u := range inputUTXOs {
		sum = sum.Add(u.Value)
		retUtxos = append(retUtxos, u)
		if sum.Cmp(amount) >= 0 {
			break
		}
	}

	if sum.Cmp(amount) < 0 {
		return nil, ErrInsufficientFund
	}

	return retUtxos, nil
}

func helpCommandHandler(ctx context.Context, account interface{}, flags cmdFlags) {
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("Command: cli ", "createAccount")
	fmt.Println("Usage Example: cli createAccount")
	for cmd, pars := range cmdFlagsMap {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("Command: cli ", cmd)
		fmt.Printf("Usage Example: cli %s", cmd)
		for _, par := range pars {
			fmt.Printf(" -%s", par.name)
			if par.name == flagFromAddress {
				fmt.Printf(" dWRFRFyientRqAbAmo6bskp9sBCTyFHLqF ")
				continue
			}
			if par.name == flagData {
				fmt.Printf(" helloworld! ")
				continue
			}
			if par.name == flagStartBlockHashes {

				fmt.Printf(" 8334b4c19091ae7582506eec5b84bfeb4a5e101042e40b403490c4ceb33897ba, 8334b4c19091ae7582506eec5b84bfeb4a5e101042e40b403490c4ceb33897bb ")
				continue
			}
			if par.name == flagPeerFullAddr {
				fmt.Printf(" /ip4/127.0.0.1/tcp/12345/ipfs/QmT5oB6xHSunc64Aojoxa6zg9uH31ajiAVyNfCdBZiwFTV ")
				continue
			}
			switch par.valueType {
			case valueTypeInt:
				fmt.Printf(" 10 ")
			case valueTypeString:
				fmt.Printf(" 1MeSBgufmzwpiJNLemUe1emxAussBnz7a7 ")
			case valueTypeUint64:
				fmt.Printf(" 50 ")
			}

		}
		fmt.Println()
		fmt.Println("Arguments:")
		for _, par := range pars {
			fmt.Println(par.name, "\t", par.usage)
		}
		fmt.Println()
	}
}

func addPeerCommandHandler(ctx context.Context, account interface{}, flags cmdFlags) {
	req := &rpcpb.AddPeerRequest{
		FullAddress: *(flags[flagPeerFullAddr].(*string)),
	}
	response, err := account.(rpcpb.AdminServiceClient).RpcAddPeer(ctx, req)
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}
	fmt.Println(proto.MarshalTextString(response))
}

func initRpcClient(port int) *grpc.ClientConn {
	//prepare grpc account
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprint(":", port), grpc.WithInsecure())
	if err != nil {
		logger.Panic("Error:", err.Error())
	}
	return conn
}

func estimateGasCommandHandler(ctx context.Context, c interface{}, flags cmdFlags) {
	var data string
	path := *(flags[flagFilePath].(*string))
	fromAddress := *(flags[flagFromAddress].(*string))
	fromAccount := account.NewTransactionAccountByAddress(account.NewAddress(fromAddress))
	toAddress := *(flags[flagToAddress].(*string))
	toAccount := account.NewTransactionAccountByAddress(account.NewAddress(toAddress))
	if path == "" {
		data = *(flags[flagData].(*string))
	} else {
		script, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error: smart contract path \"%s\" is invalid.\n", path)
			return
		}
		data = string(script)
	}

	if !fromAccount.IsValid() {
		fmt.Println("Error: 'from' address is not valid!")
		return
	}

	//Contract deployment transaction does not need to validate to address
	if data == "" && !toAccount.IsValid() {
		fmt.Println("Error: 'to' address is not valid!")
		return
	}

	response, err := c.(rpcpb.RpcServiceClient).RpcGetUTXO(ctx, &rpcpb.GetUTXORequest{
		Address: account.NewAddress(*(flags[flagFromAddress].(*string))).String(),
	})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}
	utxos := response.GetUtxos()
	var InputUtxos []*utxo.UTXO
	for _, u := range utxos {
		uu := utxo.UTXO{}
		uu.Value = common.NewAmountFromBytes(u.Amount)
		uu.Txid = u.Txid
		uu.PubKeyHash = account.PubKeyHash(u.PublicKeyHash)
		uu.TxIndex = int(u.TxIndex)
		InputUtxos = append(InputUtxos, &uu)
	}
	tip := common.NewAmount(0)
	gasLimit := common.NewAmount(0)
	gasPrice := common.NewAmount(0)
	tx_utxos, err := GetUTXOsfromAmount(InputUtxos, common.NewAmount(uint64(*(flags[flagAmount].(*int)))), tip, gasLimit, gasPrice)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	am, err := logic.GetAccountManager(wallet.GetAccountFilePath())
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	senderAccount := am.GetAccountByAddress(account.NewAddress(*(flags[flagFromAddress].(*string))))

	if senderAccount == nil {
		fmt.Println("Error: invalid account address.")
		return
	}
	sendTxParam := transaction.NewSendTxParam(account.NewAddress(*(flags[flagFromAddress].(*string))), senderAccount.GetKeyPair(),
		account.NewAddress(*(flags[flagToAddress].(*string))), common.NewAmount(uint64(*(flags[flagAmount].(*int)))), tip, gasLimit, gasPrice, data)
	tx, err := ltransaction.NewUTXOTransaction(tx_utxos, sendTxParam)
	estimateGasRequest := &rpcpb.EstimateGasRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	gasResponse, err := c.(rpcpb.RpcServiceClient).RpcEstimateGas(ctx, estimateGasRequest)

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}

	gasCount := gasResponse.GasCount

	fmt.Println("Gas estimiated num: ", common.NewAmountFromBytes(gasCount).String())
}

func gasPriceCommandHandler(ctx context.Context, account interface{}, flags cmdFlags) {
	gasPriceRequest := &rpcpb.GasPriceRequest{}
	gasPriceResponse, err := account.(rpcpb.RpcServiceClient).RpcGasPrice(ctx, gasPriceRequest)
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}
	gasPrice := gasPriceResponse.GasPrice
	fmt.Println("Gas price: ", common.NewAmountFromBytes(gasPrice).String())
}

func contractQueryCommandHandler(ctx context.Context, c interface{}, flags cmdFlags) {
	contractAddr := *(flags[flagContractAddr].(*string))
	queryKey := *(flags[flagKey].(*string))
	queryValue := *(flags[flagValue].(*string))
	contractAccount := account.NewTransactionAccountByAddress(account.NewAddress(contractAddr))

	if !contractAccount.IsValid() {
		fmt.Println("Error: contract address is not valid!")
		return
	}
	if queryKey == "" && queryValue == "" {
		fmt.Println("Error: query key and value cannot be null at the same time!")
		return
	}
	response, err := c.(rpcpb.RpcServiceClient).RpcContractQuery(ctx, &rpcpb.ContractQueryRequest{
		ContractAddr: contractAddr,
		Key:          queryKey,
		Value:        queryValue,
	})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}
	resultKey := response.GetKey()
	resultValue := response.GetValue()

	fmt.Println("Contract query result: key=", resultKey, ", value=", resultValue)
}

func getBlockByHeightCommandHandler(ctx context.Context, c interface{}, flags cmdFlags) {
	blkHeight := uint64(*(flags[flagBlockHeight].(*int)))
	if blkHeight <= 0 {
		fmt.Println("\n Example: cli getBlocksByHeight -height 5")
		fmt.Println()
		return
	}

	getBlockByHeightRequest := &rpcpb.GetBlockByHeightRequest{Height: blkHeight}

	response, err := c.(rpcpb.RpcServiceClient).RpcGetBlockByHeight(ctx, getBlockByHeightRequest)
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}

	block := response.Block
	var encodedTransactions []map[string]interface{}
	for _, transaction := range block.GetTransactions() {
		var encodedVin []map[string]interface{}
		for _, vin := range transaction.GetVin() {
			encodedVin = append(encodedVin, map[string]interface{}{
				"Vout":      vin.GetVout(),
				"Signature": hex.EncodeToString(vin.GetSignature()),
				"PubKey":    hex.EncodeToString(vin.GetPublicKey()),
			})
		}

		var encodedVout []map[string]interface{}
		for _, vout := range transaction.GetVout() {
			encodedVout = append(encodedVout, map[string]interface{}{
				"Value":      common.NewAmountFromBytes(vout.GetValue()),
				"PubKeyHash": hex.EncodeToString(vout.GetPublicKeyHash()),
				"Contract":   vout.GetContract(),
			})
		}

		encodedTransaction := map[string]interface{}{
			"ID":   hex.EncodeToString(transaction.GetId()),
			"Vin":  encodedVin,
			"Vout": encodedVout,
		}
		encodedTransactions = append(encodedTransactions, encodedTransaction)
	}

	encodedBlock := map[string]interface{}{
		"Header": map[string]interface{}{
			"Hash":      hex.EncodeToString(block.GetHeader().GetHash()),
			"Prevhash":  hex.EncodeToString(block.GetHeader().GetPreviousHash()),
			"Timestamp": time.Unix(block.GetHeader().GetTimestamp(), 0).String(),
			"Sign":      hex.EncodeToString(block.GetHeader().GetSignature()),
			"height":    block.GetHeader().GetHeight(),
		},
		"Transactions": encodedTransactions,
	}

	blockJSON, err := json.MarshalIndent(encodedBlock, "", "  ")
	if err != nil {
		fmt.Println("Error:", err.Error())
	}

	fmt.Println(string(blockJSON))
}
