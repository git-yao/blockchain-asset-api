package util

import (
	"context"
	_ "encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

// ERC20 合约 ABI（仅包含 balanceOf 方法，用于查询代币余额）
const erc20ABI = `[
    {
        "constant": true,
        "inputs": [{"name": "_owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "balance", "type": "uint256"}],
        "type": "function"
    }
]`

// EthClient 全局以太坊客户端
var EthClient *ethclient.Client

// 初始化以太坊客户端
func InitEthClient(nodeURL string) error {
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return fmt.Errorf("连接以太坊节点失败: %v", err)
	}
	EthClient = client
	Log.Info("以太坊客户端初始化成功")
	return nil
}

// 转换余额单位（Wei -> ETH）
func WeiToEth(wei *big.Int) string {
	if wei == nil {
		return "0"
	}
	weiFloat := new(big.Float).SetInt(wei)
	ethFloat := new(big.Float).Quo(weiFloat, big.NewFloat(1e18))
	// 格式化为字符串，避免科学计数法
	// 使用 'f' 格式，-1 表示使用最少的精度，18 表示小数点后18位
	return ethFloat.Text('f', -1)
}

// 查询 ETH 余额
func GetEthBalance(address string) (string, error) {
	if !common.IsHexAddress(address) {
		return "", fmt.Errorf("无效的以太坊地址: %s", address)
	}
	addr := common.HexToAddress(address)
	balance, err := EthClient.BalanceAt(context.Background(), addr, nil) // nil 表示最新区块
	if err != nil {
		return "", fmt.Errorf("查询 ETH 余额失败: %v", err)
	}

	return WeiToEth(balance), nil
}

// 查询 ERC20 代币余额
func GetErc20Balance(address, contractAddress string) (string, error) {

	if !common.IsHexAddress(address) || !common.IsHexAddress(contractAddress) {
		return "", fmt.Errorf("无效的地址: address=%s, contract=%s", address, contractAddress)
	}
	addr := common.HexToAddress(address)
	contractAddr := common.HexToAddress(contractAddress)

	parsedAbi, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", fmt.Errorf("解析 ERC20 ABI 失败: %v", err)
	}

	// 构造 balanceOf 方法调用数据
	data, err := parsedAbi.Pack("balanceOf", addr)
	if err != nil {
		return "", err
	}

	// 调用合约方法（静态调用，无需发送交易）
	result, err := EthClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("调用 ERC20 balanceOf 失败: %v", err)
	}

	// 解析返回结果（uint256 -> 字符串）
	var balance *big.Int
	err = parsedAbi.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

// 查询交易详情
func GetTransactionByHash(txHash string) (*types.Transaction, *types.Receipt, error) {
	hash := common.HexToHash(txHash)
	if !strings.HasPrefix(txHash, "0x") {
		hash = common.HexToHash("0x" + txHash)
	}

	// 查询交易
	tx, isPending, err := EthClient.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, nil, fmt.Errorf("查询交易失败: %v", err)
	}
	if isPending {
		return nil, nil, fmt.Errorf("交易处于pending状态，未上链")
	}

	// 查询交易收据
	receipt, err := EthClient.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, nil, fmt.Errorf("查询交易收据失败: %v", err)
	}

	return tx, receipt, nil
}

// 查询区块信息
func GetBlockByNumber(blockNum string) (*types.Block, error) {
	// 支持 "latest"（最新区块）或数字区块号
	var number *big.Int
	if blockNum == "latest" {
		number = nil
	} else {
		num, ok := new(big.Int).SetString(blockNum, 10)
		if !ok {
			return nil, fmt.Errorf("无效的区块号: %s", blockNum)
		}
		number = num
	}

	block, err := EthClient.BlockByNumber(context.Background(), number)
	if err != nil {
		return nil, fmt.Errorf("查询区块失败: %v", err)
	}

	return block, nil
}

func LogDtaUnpack(start, end int, val interface{}, data []byte) (err error) {
	length := len(data)
	fmt.Println("call---- LogDataUnpack begin", reflect.TypeOf(val).String(), length)
	if start >= length || end > length {
		return errors.New("LogDataUnpack: start or end out of rang e")
	}
	pdata := data[start:end]

	fmt.Println(string(data), string(pdata))
	if reflect.TypeOf(val).String() == "int64" || reflect.TypeOf(val).String() == "*int64" {
		var tmpval *int64 = val.(*int64)
		*tmpval, err = strconv.ParseInt(string(pdata), 16, 32)
		fmt.Println("call ParseInt", val)
	} else if reflect.TypeOf(val).String() == "string" || reflect.TypeOf(val).String() == "*string" {
		var tmpval *string = val.(*string)
		*tmpval = string(pdata)
		fmt.Println("call ParseInt", val)
	}
	fmt.Println("call---- LogDtaUnpack end", val)
	return nil
}
