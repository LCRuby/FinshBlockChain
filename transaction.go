package main

import (
	_ "base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/labstack/gommon/log"
	"math/big"
	"strings"
)

//1.定义一个交易结构
type Transaction struct {
	TXID      []byte     //交易ID
	TXInputs  []TXInput  //交易输入数组
	TXOutputs []TXOutput //交易输出数组
}

const reward = 50.0

//定义交易输入
type TXInput struct {
	//引用交易ID
	TXid []byte
	//应用的Output的索引值
	Index int64
	//解锁脚本,用地址来模拟
	//Sig string
	//真正的数字签名，是由r ,s 拼成的[]byte
	Signature []byte
	//约定，这里的PubKey不存储原始的公钥，而是存储X和Y拼接的字符串，在校验端重新拆分（参考r,s传递）
	//注意，是公钥，不是哈希，也不是地址
	PubKey []byte
}

//定义交易输出
type TXOutput struct {
	//转账金额
	Value float64

	//锁定脚本,用地址来模拟
	//PubKeyHash string

	//收款方的公钥哈希，注意，是哈希而不是公钥，也不是地址
	PubKeyHash []byte
}

//由于现在存储的字段是地址的公钥哈希，所以无法直接创建TXOutput，
//为了能够得到公钥哈希，我们需要处理一下，写一个Lock函数
func (output *TXOutput) Lock(address string) {
	/*封装成一个函数
	1,解码
	//2.截取出公钥哈希，出去version(1字节)，出去校验码（4字节）
	addressByte := base58.Decode(address)
	len := len(addressByte)
	pubKeyHash := addressByte[1 : len-4]
	*/
	pubKeyHash := GetPubKeyFromAddress(address)
	//真正的锁定动作
	output.PubKeyHash = pubKeyHash
}

//给TXoutput提供一个创建的方法，否则无法调用Lock（）
func NewTXOutPut(value float64, address string) *TXOutput {
	output := TXOutput{
		Value: value,
	}
	output.Lock(address)
	return &output
}

//设置交易ID
func (tx *Transaction) SetHash() {
	//转成字节流
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	data := buffer.Bytes()
	hash := sha256.Sum256(data)
	tx.TXID = hash[:]
}

//判断此交易是否是挖矿交易
func (tx *Transaction) IsCoinBase() bool {
	//1.交易input只有一个
	//if len(tx.TXInputs) == 1 {
	//	input := tx.TXInputs[0]
	//	//2.交易id为空
	//	//交易index为-1
	//	if !bytes.Equal(input.TXid, []byte{}) || input.Index != -1 {
	//		return false
	//	}
	if len(tx.TXInputs) == 1 && len(tx.TXInputs[0].TXid) == 0 && tx.TXInputs[0].Index == -1 {
		return true
	}

	return false

}

//提供创建交易方法(挖矿交易)
func NewCoinBaseTX(address string, data string) *Transaction {
	/*
		挖矿交易的特点
		1.只有一个input
		2.无需应用id
		3.无需引用index
		4.矿工由于挖矿时无需指定签名,所以这个sig字段可以有矿工自由填写,一般填写矿池的名字
	*/
	//签名先填写为空，后面创建完整交易后，最后做一次签名即可
	input := TXInput{[]byte{}, -1, nil, []byte(data)}
	//output := TXOutput{reward, address}
	output := NewTXOutPut(reward, address)

	//对于一个挖矿交易来说只有一个input和一个output
	tx := Transaction{[]byte{}, []TXInput{input}, []TXOutput{*output}}
	tx.SetHash()
	return &tx
}

//创建一个普通的转账交易
/*
1.找到合理UTXO集合
2.创建输入交易
3.创建outputs
4.如果有零钱,找零
*/
//从from转到to ,数量amount,在哪个区块上的交易
func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	//1. 创建交易之后要进行数字签名->所以需要私钥->打开钱包"NewWallets()"
	ws := NewWallets()

	//2. 找到自己的钱包，根据地址返回自己的wallet
	wallet := ws.WalletMap[from]
	if wallet == nil {
		fmt.Printf("没有找到该地址的钱包，交易创建失败!\n")
		return nil
	}

	//3. 得到对应的公钥，私钥
	pubKey := wallet.PublicKey
	privateKey := wallet.Private //稍后再用
	//传递公钥的哈希，而不是传递地址
	pubKeyHash := HashPubKey(pubKey)

	//1.找到合理UTXO集合map[string][]uint64
	utxos, resValue := bc.FindNeedUTXOs(pubKeyHash, amount) //返回一个UTXO集合,和总钱数
	if resValue < amount {
		fmt.Printf("余额不足,交易失败!!!\n")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput
	//2.创建交易输入,将UTXO逐一转成inputs
	//map[string][]uint64: key是这个output的交易id，value是这个交易中索引的数组
	for id, indexArray := range utxos {
		for _, i := range indexArray {
			input := TXInput{[]byte(id), int64(i), nil, pubKey}
			inputs = append(inputs, input)
		}
	}

	//创建输出交易
	//output := TXOutput{amount, to}
	output := NewTXOutPut(amount, to)
	outputs = append(outputs, *output)

	//找零
	if resValue > amount {
		//在创建一个输出交易,转给自己
		//outputs = append(outputs, TXOutput{resValue - amount, from})
		output = NewTXOutPut(resValue-amount, from)
		outputs = append(outputs, *output)
	}
	tx := Transaction{[]byte{}, inputs, outputs}
	tx.SetHash()

	//对交易进行签名
	bc.SignTransaction(&tx, privateKey)
	return &tx
}

//签名的具体实现
//参数：私钥，inputs里面的所有引用的交易的结构map[string]Transaction
//map[2222]Transaction222  这里把索引的结构不是所用到的input，而是整个所引用里面的数据
//map[3333]Transaction333
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinBase() {
		return
	}
	//1.创建一个交易副本:txCopy,使用函数:TrimmedCopy:把Signature和PubKey字段设置为nil
	txCopy := tx.TrimmedCopy()
	for i, input := range txCopy.TXInputs {
		prevTx := prevTXs[string(input.TXid)]
		if len(prevTx.TXID) == 0 {
			log.Panic("引用的交易无效!!!")
		}
		//不要对input进行赋值，这是一个副本，要对txCopy.TXInputs[xx]进行操作，否则无法把pubKeyHash传进来
		txCopy.TXInputs[i].PubKey = prevTx.TXOutputs[input.Index].PubKeyHash
		//所需要的三个数据都具备了，开始做哈希处理
		//3. 生成要签名的数据。要签名的数据一定是哈希值
		//a. 我们对每一个input都要签名一次，签名的数据是由当前input引用的output的哈希+当前的outputs（都承载在当前这个txCopy里面）
		//b. 要对这个拼好的txCopy进行哈希处理，SetHash得到TXID，这个TXID就是我们要签名最终数据。
		txCopy.SetHash()
		//还原，以免影响后面input的签名
		txCopy.TXInputs[i].PubKey = nil
		//signDataHash认为是原始数据,这是setHash操作流程决定的
		signDataHash := txCopy.TXID
		//4. 执行签名动作得到r,s字节流
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, signDataHash)
		if err != nil {
			log.Panic(err)
		}

		//5. 放到我们所签名的input的Signature中
		signature := append(r.Bytes(), s.Bytes()...)
		tx.TXInputs[i].Signature = signature
	}
}

//copy一个副本,并把inout签名和pubKey设置为nil
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, input := range tx.TXInputs {
		inputs = append(inputs, TXInput{input.TXid, input.Index, nil, nil})
	}
	for _, output := range tx.TXOutputs {
		outputs = append(outputs, output)
	}
	return Transaction{tx.TXID, inputs, outputs}
}

//分析校验
//所需要的数据:公钥,数据(txCopy,生成哈希)签名
//对每一个签名过得input进行校验
func (tx *Transaction) Verfy(prevTXs map[string]Transaction) bool {
	if tx.IsCoinBase() {
		return true
	}
	//1.得到签名数据
	txCopy := tx.TrimmedCopy() //拷贝了一份数据

	for i, input := range tx.TXInputs {
		prevTX := prevTXs[string(input.TXid)]
		if len(prevTX.TXID) == 0 {
			log.Panic("引用交易无效!")

		}
		txCopy.TXInputs[i].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash
		txCopy.SetHash() //得到签名哈希
		dataHash := txCopy.TXID

		//2.得到Signature,反推回r,s
		signature := input.Signature
		//3.拆解PubKey,X,Y得到原生的公钥
		pubKey := input.PubKey

		//1.定义来年改革辅助的big.Int
		r := big.Int{}
		s := big.Int{}

		//拆分我们的signature,平均分,前半部分给r,后半部分给s
		r.SetBytes(signature[:len(signature)/2])
		s.SetBytes(signature[len(signature)/2:])

		//拆公钥
		X := big.Int{}
		Y := big.Int{}
		X.SetBytes(pubKey[:len(pubKey)/2])
		Y.SetBytes(pubKey[len(pubKey)/2:])

		//还原原始的公钥
		pubKeyOrigin := ecdsa.PublicKey{elliptic.P256(), &X, &Y}

		//4.Verify
		if !ecdsa.Verify(&pubKeyOrigin, dataHash, &r, &s) {
			return false
		}
	}
	return true
}

//打印区块格式化输出
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.TXID))

	for i, input := range tx.TXInputs {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TXid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Index))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.TXOutputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %f", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
