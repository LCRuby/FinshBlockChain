package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type ProofOfWork struct {
	//区块
	block *Block
	//目标难度值,很大的值
	target *big.Int
}

//提供创建一个POW 的函数
func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}
	//指定难度值
	targetStr := "0000f1d0b2c458239d5d0f6a26f557596a82cdfc269331f823077c774e108260"
	tmpInt := big.Int{}
	//func (z *Int) SetString(s string, base int) (*Int, bool) { 把字符串转换成一个大数
	tmpInt.SetString(targetStr, 16) //可以自己赋值给自己
	pow.target = &tmpInt
	return &pow
}

//hash运算
func (pow *ProofOfWork) Run() ([]byte, uint64) {
	//1.拼装数据(区块的数据,还有不断变化的随机值)
	var nonce uint64
	var hash [32]byte
	block := pow.block
	for {
		tmp := [][]byte{
			Uint64ToByte(block.Version),
			block.PrevHash,
			block.MerkelRoot,
			Uint64ToByte(block.TimeStamp),
			Uint64ToByte(block.Difficulty),
			Uint64ToByte(nonce),
			//只对区块头做哈希值，区块体通过MerkelRoot产生影响
			//block.Data,
		}
		//将二维的切片数组链接起来，返回一个一维的切片
		blockInfo := bytes.Join(tmp, []byte{})

		//2.计算哈希值
		//func Sum256(data []byte) [Size]byte ,size是sha256的长度
		hash = sha256.Sum256(blockInfo)
		tmpInt := big.Int{}
		//func (z *Int) SetBytes(buf []byte) *Int {  把字节流转换成一个大数
		tmpInt.SetBytes(hash[:])

		//3.进行哈希比较(与pow.target进行比较)
		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y
		//func (x *Int) Cmp(y *Int) (r int) {
		if tmpInt.Cmp(pow.target) == -1 {
			//a,找打返回哈希值和nonce
			fmt.Printf("挖矿成功: hash=%x, nonce=%d\n", hash, nonce)
			break
		} else {
			//b,找不到,继续找
			nonce++
		}

	}
	return hash[:], nonce

}
