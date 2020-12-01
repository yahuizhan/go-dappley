package account_ron

import (
	"bufio"
	"fmt"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/wallet"
	laccountpb "github.com/dappley/go-dappley/wallet/pb"
	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"
	"io"
	"os"
)

const fileName = "account.dat"

func SaveAccountToFile(accInfo *AccountInfo) {
	acManager := &wallet.AccountManager{}
	acManager.Accounts = make([]*account.Account, 0)
	for i := 0; i < len(accInfo.FromAddress); i++ {
		acManager.Accounts = append(acManager.Accounts, accInfo.GetAccount(accInfo.FromAddress[i]))
		acManager.Accounts = append(acManager.Accounts, accInfo.GetAccount(accInfo.ToAddress[i]))
	}

	// 序列化
	dataMarshal, err := proto.Marshal(acManager.ToProto()) //newAccount.ToProto() 是先把newAccount转化成 在pd里定义的message格式 accountpb.Account
	if err != nil {
		fmt.Println("proto.Unmarshal.Err: ", err)
	}

	//写入文档
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	//使用NewWriter方法返回的io.Writer缓冲默认大小为4096，也可以使用NewWriterSize方法设置缓存的大小
	newWriter := bufio.NewWriter(file)
	//将文件写入缓存
	if _, err = newWriter.Write(dataMarshal); err != nil {
		fmt.Println(err)
	}
	//从缓存写入到文件中
	if err = newWriter.Flush(); err != nil {
		fmt.Println(err)
	}
	logger.Info(fileName," 写入成功")
}

func ReadAccountFromFile() ([]*account.Account, error) {
	//读取文档
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	//创建一个新的io.Reader，它实现了Read方法
	reader := bufio.NewReader(file)
	//设置读取的长度
	buf := make([]byte, 1024)
	var totalBuf []byte
	//读取文件
	for {
		// 循环读取文件
		n, err2 := reader.Read(buf)
		if err2 == io.EOF { // io.EOF表示文件末尾
			logger.Info("account.dat 文件读取成功，将使用保存账户继续测试")
			logger.Info("如果需重新测试，请删除此文件重新启动测试程序")
			break
		}
		//	fmt.Print(string(buf[:n]))
		totalBuf = append(totalBuf, buf[:n]...)
	}

	// 反序列化
	acManager := &laccountpb.AccountManager{}
	err = proto.Unmarshal(totalBuf, acManager)
	if err != nil {
		fmt.Println("proto.Unmarshal.Err: ", err)
		return nil, err
	}
	newAcManager := &wallet.AccountManager{}
	newAcManager.FromProto(acManager) //从接收到的pb message转成原来account格式
	//for k,v := range newAcManager.Accounts{
	//	fmt.Println("解码数据:key:",k,"value:",v)
	//}
	return newAcManager.Accounts, nil
}
