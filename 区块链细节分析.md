### 一.每个.go文件里面的函数(总数)

![](.\image\文件函数1.png)

![](.\image\文件函数2.png)



![](.\image\文件函数3.png)

### 二.整体流程图

![](.\image\区块链流程图v4版.png)

### 三.根据整体流程来细化每个函数

![](.\image\创建区块链流程.PNG)

### 3.1main函数的执行

```
main.go:
bc := NewBlockChain("1DCm5Z3WhKVMqVTPFrBqhjDbjyF3ahtpgE")   //传入一个地址,创建一个区块链
```

### 3.2 区块链的创建

blockChain.go:

//定义一个区块链结构体,把区块链存储到数据库中(这里使用bolt轻量级数据库来存储),
type BlockChain struct {
​	//使用数据库来存书区块
​	db   *bolt.DB
​	tail []byte //存储最后一个区块的哈希值,方便遍历的时候查找
}

const blockChainDb = "blockChain.db" //存储区块的数据库
const blockBucket = "blockBucket"    //数据库中的桶

//接收一个地址,返回一个区块链
func NewBlockChain(address string) *BlockChain {

   //最后一个区块的hash从数据库中读取出来
   var lastHash []byte
   //1.打开数据库
   db, err := bolt.Open(blockChainDb, 0600, nil) //0600是读写操作权限
   //defer db.Close()
   if err != nil {
​      log.Panic("数据库打开失败")
   }
   //将数据写入数据库
   db.Update(func(tx *bolt.Tx) error {
​      //找抽屉,没有就创建
​      bucket := tx.Bucket([]byte(blockBucket))
​      if bucket == nil {
​         bucket, err = tx.CreateBucket([]byte(blockBucket))
​         if err != nil {
​            log.Panic("创建桶失败")
​         }
​         //写数据,创建一个创世快,并把其作为第一个区块添加到区块链中
​         **genesisBlock := GenesisBlock(address)</u>**
​         //hash作为key， block的字节流作为value，Serialize()

​	//因为bolt是Key-Value结构,把区块头哈希作为key,把数据内容序列化作为date

​         //这里的hash是工作量证明的hash,第二个参数是把区块序列化

​         err = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
​         if err != nil {
​            log.Panic("创世块写入数据失败")
​         }

​	//保存最后一个区块头的哈希,便于遍历查找

​         err = bucket.Put([]byte("LastHashKey"), genesisBlock.Hash)
​         if err != nil {
​            log.Panic("创世块写入数据失败")
​         }
​         lastHash = genesisBlock.Hash
​      } else {
​         lastHash = bucket.Get([]byte("LastHashKey"))
​      }
​      return nil
   })
   return &BlockChain{db, lastHash}
}

#### 3.2.1区块链创建中调用创世块函数

```
blockChain.go:

//定义一个创世快
func GenesisBlock(address string) *Block {
//创建基于挖矿交易
   coinbase := NewCoinBaseTX(address, "I am number one !!!")
   //返回一个新建的区块,参数是交易,数据是空
   return NewBlock([]*Transaction{coinbase}, []byte{})
}
```



##### 3.2.1.1  在创世块中调用基于挖矿交易函数

交易的流程

![](.\image\UTXO详解.PNG)

```
transaction.go

//交易的一些数据结构
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
	//引用的Output的索引值
	Index int64
	//真正的数字签名，是由r ,s 拼成的[]byte
	Signature []byte
	//约定，这里的PubKey不存储原始的公钥，而是存储X和Y拼接的字符串，在校验端重新拆分（参考r,s传递）
	//注意，是公钥，不是哈希，也不是地址(发送方的公钥)
	PubKey []byte
}

//定义交易输出
type TXOutput struct {
	//转账金额
	Value float64
	//收款方的公钥哈希，注意，是哈希而不是公钥，也不是地址
	PubKeyHash []byte
}

//提供创建交易方法(挖矿交易)
func NewCoinBaseTX(address string, data string) *Transaction {
   /*
      挖矿交易的特点
      1.只有一个input
      2.无需引用id
      3.无需引用index
      4.矿工由于挖矿时无需指定签名,所以这个sig字段可以有矿工自由填写,一般填写矿池的名字
   */
   //签名先填写为空，后面创建完整交易后，最后做一次签名即可
   input := TXInput{[]byte{}, -1, nil, []byte(data)}//挖矿相当于没有输入,即可以随便填写字段
   output := NewTXOutPut(reward, address)

   //对于一个挖矿交易来说只有一个input和一个output
   tx := Transaction{[]byte{}, []TXInput{input}, []TXOutput{*output}}
   tx.SetHash()
   return &tx
}
```

```
transaction.go

//给TXoutput提供一个创建的方法，否则无法调用Lock（）
func NewTXOutPut(value float64, address string) *TXOutput {
   output := TXOutput{
      Value: value,
   }
   output.Lock(address)
   return &output
}
```

![](.\image\生成地址的流程.PNG)

```
transaction.go

//由于现在存储的字段是地址的公钥哈希，所以无法直接创建TXOutput，
//为了能够得到公钥哈希，我们需要处理一下，写一个Lock函数(上面流程图即是)
func (output *TXOutput) Lock(address string) {
  //由地址的得到公钥哈希
   pubKeyHash := GetPubKeyFromAddress(address)
   //真正的锁定动作
   output.PubKeyHash = pubKeyHash
}
```

```
blockChain.go

//由地址得到公钥哈希
func GetPubKeyFromAddress(address string) []byte {
   //1. 解码
   //2. 截取出公钥哈希：去除version（1字节），去除校验码（4字节）
   addressByte := base58.Decode(address) //25字节
   len := len(addressByte)

   pubKeyHash := addressByte[1 : len-4]

   return pubKeyHash
}
```

//将交易整个打包,生成哈希值

```
block.go

//3.生成哈希
func (block *Block) SetHash() {
   var blockInfo []byte
   //1. 拼装数据
   tmp := [][]byte{
      Uint64ToByte(block.Version),
      block.PrevHash,
      block.MerkelRoot,
      Uint64ToByte(block.TimeStamp),
      Uint64ToByte(block.Difficulty),
      Uint64ToByte(block.Nonce),
      //block.Data,
   }
   //将二维的切片数组链接起来，返回一个一维的切片
   blockInfo = bytes.Join(tmp, []byte{})

   //2. sha256
   //func Sum256(data []byte) [Size]byte {
   hash := sha256.Sum256(blockInfo)
   block.Hash = hash[:]
}

//至此,基于挖矿交易结束
```

##### 3.2.1.2在创世块中调用创建区块函数

```
block.go

//创建区块
func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
   block := Block{
      Version:  00,
      PrevHash: prevBlockHash,
      //merkelRoot里面存放的是每个交易数据的hash,通过二叉树往上合并成root哈希
      MerkelRoot: []byte{},
      TimeStamp:  uint64(time.Now().Unix()),
      Difficulty: 0, //随便填写的无效值
      Nonce:      0, //同上
      Hash:       []byte{},
      //Data:       []byte(data),
      //交易的指针
      Transactions: txs,
   }
   //创建一个pow对象
   pow := NewProofOfWork(&block)
   //查找随机数,不停的进行hash运算
   hash, nonce := pow.Run()

   //根据挖矿结果对数据进行更新
   block.Hash = hash
   block.Nonce = nonce
   return &block
}


```

```
proofofwork.go

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
```

```
proofofwork.go


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
```

```
block.go

//将uint64转成byte类型
func Uint64ToByte(num uint64) []byte {
   var buffer bytes.Buffer

   err := binary.Write(&buffer, binary.BigEndian, num)
   if err != nil {
      log.Panic(err)
   }

   return buffer.Bytes()
}
```

#### 3.2.2区块链创建中,调用序列化函数

block.go

//1.序列化,将区块数据序列化,返回一个字节流存储到数据库中
func (block *Block) Serialize() []byte {
   var buffer bytes.Buffer

   //- 使用gob进行序列化（编码）得到字节流
   //1. 定义一个编码器
   //2. 使用编码器进行编码
   encoder := gob.NewEncoder(&buffer)
   err := encoder.Encode(&block)
   if err != nil {
​      log.Panic("编码出错!")
   }
   return buffer.Bytes()
}

![](.\image\挖矿流程2.PNG)

### 3.3执行命令行参数动作

```
cli.go

//这是一个用来接收命令行参数并且控制区块链操作的文件
//CLI的成员是区块链
type CLI struct {
   bc *BlockChain
}

//反引号可以多行写,usge相当于一个命令使用说明
const Usage = `
   printChain               "正向打印区块链"
   printChainR          "反向打印区块链"
   getBalcance  --address ADDRESS "获取地址的余额"
    send FROM TO AMOUNT MINER DATA "由FROM转AMOUNT给TO，由MINER挖矿，同时写入DATA"
   newWallet "创建一个钱包"
    listAddresses "列举所有的钱包地址"
`

//接受参数的动作，我们放到一个函数中

func (cli *CLI) Run() {

   //./block printChain
   //./block addBlock --data "HelloWorld"
   //1. 得到所有的命令
   args := os.Args
   if len(args) < 2 {
      fmt.Printf(Usage)
      return
   }

   //2. 分析命令
   cmd := args[1]
   switch cmd {

   case "printChain":
      fmt.Printf("打印区块\n")
      cli.PrintBlockChain()
   case "printChainR":
      fmt.Printf("反向打印区块链\n")
      cli.PrintBlockChainReverse()
   case "getBalance":
      fmt.Printf("h获取余额\n")
      if len(args) == 4 && args[2] == "--address" {
         adress := args[3]
         cli.GetBalance(adress)
      }
   case "send":
      fmt.Printf("转账开始...\n")
      if len(args) != 7 {
         fmt.Printf("参数个数错误，请检查！\n")
         fmt.Printf(Usage)
         return
      }
      //./block send FROM TO AMOUNT MINER DATA "由FROM转AMOUNT给TO，由MINER挖矿，同时写入DATA"
      from := args[2]
      to := args[3]
      amount, _ := strconv.ParseFloat(args[4], 64) //知识点，请注意
      miner := args[5]
      data := args[6]
      cli.Send(from, to, amount, miner, data)
   case "newWallet":
      fmt.Printf("创建新的钱包....\n")
      cli.NewWallet()
   case "listAddresses":
      fmt.Printf("列举所有地址...\n")
      cli.ListAddresses()
   default:
      fmt.Printf("无效的命令，请检查!\n")
      fmt.Printf(Usage)
   }
}
```

#### 3.3.1打印区块链

```
commandLine.go

//正向打印区块链
func (cli *CLI) PrintBlockChain() {
   cli.bc.PrintChain()
   fmt.Printf("打印区块链完成\n")
}
```

##### 3.3.1.1打印区块链

```
blockChain.go

//正向打印区块链
func (bc *BlockChain) PrintChain() {
   blockHeight := 0
   bc.db.View(func(tx *bolt.Tx) error {
      //假设桶已经存在,
      b := tx.Bucket([]byte("blockBucket"))

      //数据库中,hash作为key,block的字节流作为value
      //从第一个key->value 进行遍历,到最后一个固定的key 时直接返回
      b.ForEach(func(k, v []byte) error {
         if bytes.Equal(k, []byte("LastHashKey")) {
            return nil
         }
         block := Deserialize(v)
         fmt.Printf("===========区块高度=============")
         blockHeight++
         fmt.Printf("版本号: %d\n", block.Version)
         fmt.Printf("前区块哈希值: %x\n", block.PrevHash)
         fmt.Printf("梅克尔根: %x\n", block.MerkelRoot)
         fmt.Printf("时间戳: %d\n", block.TimeStamp)
         fmt.Printf("难度值(随便写的）: %d\n", block.Difficulty)
         fmt.Printf("随机数 : %d\n", block.Nonce)
         fmt.Printf("当前区块哈希值: %x\n", block.Hash)
         //只有创世快的挖矿交易里面有数据,且挖矿交易是第一个交易
         fmt.Printf("区块数据 :%s\n", block.Transactions[0].TXInputs[0].PubKey)
         return nil
      })
      return nil
   })
}
```

##### 3.3.1.2反序列化data

```
block.go

//2.反序列化
func Deserialize(data []byte) Block {

   decoder := gob.NewDecoder(bytes.NewReader(data))

   var block Block
   //2. 使用解码器进行解码
   err := decoder.Decode(&block)
   if err != nil {
      log.Panic("解码出错!")
   }

   return block
}
```

#### 3.3.2反向打印区块链

```
commandLine.go

//反向打印区块链
func (cli *CLI) PrintBlockChainReverse() {
   bc := cli.bc
   //创建迭代器
   it := bc.NewIterator()

   //调用迭代器，返回我们的每一个区块数据
   for {
      //返回区块，左移
      block := it.Next()
      for _, tx := range block.Transactions {
         fmt.Println(tx)
      }
      if len(block.PrevHash) == 0 {
         fmt.Printf("区块链遍历结束！")
         break
      }
   }
}
```

##### 3.3.2.1迭代器

```
blockChainIterator.go

type BlockChainIterator struct {
   db *bolt.DB //数据库
   //定义一个游标,用于不断的索引
   currentHashPointer []byte
}

//创建一个迭代器,区块链来调用
func (bc *BlockChain) NewIterator() *BlockChainIterator {
   return &BlockChainIterator{
      bc.db,
      //最初指向最后一个区块,随着next()的调用不断变化
      bc.tail,
   }
}
```

##### 3.3.2.2遍历区块

```
blockChainIterator.go

//迭代器是属于区块链的:即迭代器由区块链调用
//Next方式是属于迭代器的:即Next()由迭代器来调用
//1. 返回当前的区块
//2. 指针前移
func (it *BlockChainIterator) Next() *Block {
   var block Block
   it.db.View(func(tx *bolt.Tx) error {
      bucket := tx.Bucket([]byte(blockBucket))
      if bucket == nil {
         log.Panic("迭代器遍历时bucket不应该为空，请检查!")
      }
      blockTmp := bucket.Get(it.currentHashPointer)

      //解码动作
      block = Deserialize(blockTmp)
      //游标哈希左移
      it.currentHashPointer = block.PrevHash
      return nil
   })
   return &block
}
```



#### 3.3.3获取指定地址的余额

```
commmandLine.go

//查找出指定地址的所有余额
func (cli *CLI) GetBalance(address string) {
   //1.验证地址是否有效
   if !IsValidAddress(address) {
      fmt.Printf("地址无效 : %s\n", address)
      return
   }
   //2.从地址反向得到公钥哈希
   pubKeyHash := GetPubKeyFromAddress(address)
   utxos := cli.bc.FindUTXOs(pubKeyHash)

   total := 0.0
   for _, utxo := range utxos {
      total += utxo.Value
   }

   fmt.Printf("\"%s\"的余额为：%f\n", address, total)
}
```

##### 3.3.3.1 验证地址是否有效

![](C:\Users\Steven\Desktop\go资料\image\生成地址的流程.PNG)

```
wallet.go

//地址校验
func IsValidAddress(address string) bool {
   //1. 解码
   addressByte := base58.Decode(address)

   if len(addressByte) < 4 {
      return false
   }

   //2. 取数据
   payload := addressByte[:len(addressByte)-4]
   checksum1 := addressByte[len(addressByte)-4:]

   //3. 做checksum函数
   checksum2 := CheckSum(payload)

   fmt.Printf("checksum1 : %x\n", checksum1)
   fmt.Printf("checksum2 : %x\n", checksum2)

   //4. 比较
   return bytes.Equal(checksum1, checksum2)
}
```

```
wallet.go

//获取checksum字节段
func CheckSum(data []byte) []byte {
   //两次sha256
   hash1 := sha256.Sum256(data)
   hash2 := sha256.Sum256(hash1[:])

   //前4字节校验码
   checkCode := hash2[:4]
   return checkCode
}
```

##### 3.3.3.2 从地址中得到公钥哈希

```
wallets.go

//从地址得到公钥哈希
func GetPubKeyFromAddress(address string) []byte {
   //1. 解码
   //2. 截取出公钥哈希：去除version（1字节），去除校验码（4字节）
   addressByte := base58.Decode(address) //25字节
   len := len(addressByte)

   pubKeyHash := addressByte[1 : len-4]

   return pubKeyHash
}
```

##### 3.3.3.3找到指定地址的UTXO

```
blockChain.go

//查找指定公钥哈希的UTXO集合
func (bc *BlockChain) FindUTXOs(pubKeyHash []byte) []TXOutput {
   var UTXO []TXOutput
   txs := bc.FindUTXOTransactions(pubKeyHash)
   for _, tx := range txs {
      for _, output := range tx.TXOutputs {
         if bytes.Equal(pubKeyHash, output.PubKeyHash) {
            UTXO = append(UTXO, output)
         }
      }
   }

   return UTXO
}
```

![](.\image\查找指定地址utxo的流程.png)

```
blockChain.go

func (bc *BlockChain) FindUTXOTransactions(SenderPubKeyHash []byte) []*Transaction {
   //存储所有包含utxo的集合
   var txs []*Transaction
   //定义一个map,来保存所有消费过的output
   spentOutputs := make(map[string][]int64)

   //创建一个迭代器
   it := bc.NewIterator()
   for {
      //遍历区块
      block := it.Next()
      //遍历交易
      for _, tx := range block.Transactions {
      OUTPUT:
         //遍历output,找到和自己相关的utxo(添加之前检查是否已经消耗过)
         for i, output := range tx.TXOutputs {
            if spentOutputs[string(tx.TXID)] != nil {
               for _, j := range spentOutputs[string(tx.TXID)] {
                  if int64(i) == j {
                     //当前准备添加的output已经消耗过,不添加
                     continue OUTPUT
                  }
               }
            }
            //这里的output和我们所找的地址相同,满足条件,加utxo数组中
            //if output.PubKeyHash == address {
            if bytes.Equal(output.PubKeyHash, SenderPubKeyHash) {
               //!!!!!重点
               //返回所有包含我的outx的交易的集合
               txs = append(txs, tx)
            }
         }
         //如果当前交易是挖矿交易的话，那么不做遍历，直接跳过
         if !tx.IsCoinBase() {
            //4. 遍历input，找到自己花费过的utxo的集合(把自己消耗过的标示出来)
            for _, input := range tx.TXInputs {
               //判断一下当前这个input和目标（李四）是否一致，如果相同，说明这个是李四消耗过的output,就加进来
               //if input.Sig == address {
               pubkeyHash := HashPubKey(input.PubKey) //把公钥转换成公钥哈希
               if bytes.Equal(pubkeyHash, SenderPubKeyHash) {
                  //spentOutputs := make(map[string][]int64)
                  //indexArray := spentOutputs[string(input.TXid)]
                  //indexArray = append(indexArray, input.Index)
                  spentOutputs[string(input.TXid)] = append(spentOutputs[string(input.TXid)], input.Index)
                  //map[2222] = []int64{0}
                  //map[3333] = []int64{0, 1}
               }
            }
         } else {
            //fmt.Printf("这是coinbase，不做input遍历！")
         }
      }

      if len(block.PrevHash) == 0 {
         break
         fmt.Printf("区块遍历完成退出!\n")
      }
   }

   return txs

}
//在一个区块中.TXinputs生成TXOutputs,TXinputs与前面的区块的TXoutputs有联系,所有遍历的前后关系不影响结果
```

```
transaction.go

//判断是否是挖矿交易
func (tx *Transaction) IsCoinBase() bool {
   if len(tx.TXInputs) == 1 && len(tx.TXInputs[0].TXid) == 0 && tx.TXInputs[0].Index == -1 {
      return true
   }

   return false

}
```

```
wallet.go



//将公约转化成公钥哈希
func HashPubKey(data []byte) []byte {
   hash := sha256.Sum256(data)
   //编码
   rip160hasher := ripemd160.New()
   _, err := rip160hasher.Write(hash[:])
   if err != nil {
      log.Panic(err)
   }

   //返回rip160的哈希结果
   rip160HashValue := rip160hasher.Sum(nil)
   return rip160HashValue
}
```

#### 3.3.4转账交易

```
//转账交易
func (cli *CLI) Send(from, to string, amount float64, miner, data string) {
   //地址校验
   if !IsValidAddress(from) {
      fmt.Printf("地址无效 from: %s\n", from)
      return
   }
   if !IsValidAddress(to) {
      fmt.Printf("地址无效 to: %s\n", to)
      return
   }
   if !IsValidAddress(miner) {
      fmt.Printf("地址无效 miner: %s\n", miner)
      return
   }
   //1. 创建挖矿交易
   coinbase := NewCoinBaseTX(miner, data)
   //2. 创建一个普通交易
   tx := NewTransaction(from, to, amount, cli.bc)
   if tx == nil {
      fmt.Printf("无效的交易\n")
      return
   }
   //3. 添加到区块

   cli.bc.AddBlock([]*Transaction{coinbase, tx})
   fmt.Printf("转账成功！\n")
}
```

##### 3.3.4.1 验证地址是否有效

```
wallet.go

//地址校验
func IsValidAddress(address string) bool {
   //1. 解码
   addressByte := base58.Decode(address)

   if len(addressByte) < 4 {
      return false
   }

   //2. 取数据
   payload := addressByte[:len(addressByte)-4]
   checksum1 := addressByte[len(addressByte)-4:]

   //3. 做checksum函数
   checksum2 := CheckSum(payload)

   //4. 比较
   return bytes.Equal(checksum1, checksum2)
}
```

##### 3.3.4.2 创建挖矿交易

```
transaction.go

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
```

##### 3.3.4.3 创建普通交易(重点)

交易的整个流程:

1. 创建交易之后要进行数字签名->所以需要私钥->打开钱包"NewWallets()"  得倒所有的钱包

2.  找到自己的钱包:根据地址返回自己的wallet

3. 从自己的钱包中得到对应的公钥，私钥

4. 找到合理UTXO集合

   ​	-找到指定地址的所有的utxos 

   ​		*创建迭代器

   ​		*遍历交易

   ​		*遍历output,找到和自己相关的utxo(没有消费过的)

   ​		*返回交易

   ​	-判断所有被的utxos的余额,是否满足交易

   ​	-满足交易,则创建交易输入,创建交易输出

   ​	

5. 创建输入

6. 创建交易outputs

7. 如果有零钱,找零	

8. 封装交易

9. 对交易进行签名

   ```
   找到所有引用的交易
   1. 根据inputs来找，有多少input, 就遍历多少次（重点）
   	-根据TXid查找交易本身
   	
   2. 找到目标交易，（根据TXid来找）
   3. 添加到prevTXs里面
   
   签名(有多少input签名多少次)
   1.判断是否是挖矿交易
   2.不是挖矿交易,则复制一份副本,把input 里面的签名,公钥字段置为nil
   3.把引用输入的公钥哈希,赋给这次交易输入的公钥字段
   4.对交易副本进行SetHash,存储在交易的TXID字段里
   5.把副本的公钥字段进行复制nil,以免后面的input签名
   6.签名r, s, err := ecdsa.Sign(rand.Reader, privateKey, signDataHash)
   signature := append(r.Bytes(), s.Bytes()...)
   ```

   ![](.\image\每次交易签名的信息.PNG)

![](.\image\签名的具体逻辑详解.jpg)

```
transaction.go

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
```

```
wallets.go

//定义一个wallets数据结构
type Wallets struct {
	//map[地址]钱包
	WalletMap map[string]*Wallet
}
//显示钱包的所有地址
func NewWallets() *Wallets {

   var ws Wallets
   ws.WalletMap = make(map[string]*Wallet)

   ws.LoadFile()
   return &ws
}
```

```
wallets.go

//读取文件方法，把所有的wallet读出来
func (ws *Wallets) LoadFile() {
   //在读取之前，要先确认文件是否在，如果不存在，直接退出
   _, err := os.Stat(walletFile)
   if os.IsNotExist(err) {
      //ws.WalletsMap = make(map[string]*Wallet)
      return
   }
```

![](.\image\找合适TUXO的整体流程.PNG)

```
blockChain.go

//根据需求找到合理的utxo
func (bc *BlockChain) FindNeedUTXOs(senderPubKeyHash []byte, amount float64) (map[string][]uint64, float64) {
   //找到的合理的utxos集合
   //var utxos map[string][]uint64   //不能这样赋值,这样的话utxos是零值,向零值map在中赋值,会宕机,必须初始化(p73)
   utxos := make(map[string][]uint64)
   //找到的utxos里面包含前的总数
   var calc float64 //初始值为0
   txs := bc.FindUTXOTransactions(senderPubKeyHash)

   for _, tx := range txs {
      for i, output := range tx.TXOutputs {
         //if from == output.PubKeyHash {
         if bytes.Equal(senderPubKeyHash, output.PubKeyHash) {
            //fmt.Printf("222222")
            //UTXO = append(UTXO, output)
            //fmt.Printf("333333 : %f\n", UTXO[0].Value)
            //我们要实现的逻辑就在这里，找到自己需要的最少的utxo
            //3. 比较一下是否满足转账需求
            //   a. 满足的话，直接返回 utxos, calc
            //   b. 不满足继续统计

            if calc < amount {
               //1. 把utxo加进来，
               //utxos := make(map[string][]uint64)
               //array := utxos[string(tx.TXID)] //确认一下是否可行！！
               //array = append(array, uint64(i))
               utxos[string(tx.TXID)] = append(utxos[string(tx.TXID)], uint64(i))
               //2. 统计一下当前utxo的总额
               //第一次进来: calc =3,  map[3333] = []uint64{0}
               //第二次进来: calc =3 + 2,  map[3333] = []uint64{0, 1}
               //第三次进来：calc = 3 + 2 + 10， map[222] = []uint64{0}
               calc += output.Value

               //加完之后满足条件了，
               if calc >= amount {
                  //break
                  fmt.Printf("找到了满足的金额：%f\n", calc)
                  return utxos, calc
               }
            } else {
               fmt.Printf("不满足转账金额,当前总额：%f， 目标金额: %f\n", calc, amount)
            }
         }
      }
   }
   	return utxos, calc
}
```

**(难点)**

```
blockChain.go


//对交易进行签名
func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) {
   //签名，交易创建的最后进行签名
   //map[交易地址]所有交易内容
   prevTXs := make(map[string]Transaction) //找到所有与签名有关的交易

   //找到所有引用的交易
   //1. 根据inputs来找，有多少input, 就遍历多少次（重点）
   //2. 找到目标交易，（根据TXid来找）
   //3. 添加到prevTXs里面
   for _, input := range tx.TXInputs {
      //更具id查找交易本身,需要遍历整个区块链
      tx, err := bc.FindTransactionByTXid(input.TXid)
      if err != nil {
         log.Panic(err)
      }
      prevTXs[string(input.TXid)] = tx
      /*
         //第一个input查找之后：prevTXs：
         // map[2222]Transaction222

         //第二个input查找之后：prevTXs：
         // map[2222]Transaction222
         // map[3333]Transaction333

         //第三个input查找之后：prevTXs：
         // map[2222]Transaction222
         // map[3333]Transaction333(只不过是重新写了一次)

      */
   }
   tx.Sign(privateKey, prevTXs)
}
```

```
transaction.go

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
      //signDataHash认为是原始数据
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
```

##### 3.3.4.4 添加到区块

```
//6.添加区块
func (bc *BlockChain) AddBlock(txs []*Transaction) {
   //V3版本
   //获取前一个区块的哈希值
   db := bc.db
   lastHash := bc.tail
   db.Update(func(tx *bolt.Tx) error {
      //完成添加数据
      bucket := tx.Bucket([]byte(blockBucket))
      if bucket == nil {
         log.Panic("bucket不应该为空")
      }
      //a,创建新的区块
      block := NewBlock(txs, lastHash)

      //添加到数据库中
      //hash作为key,block的字节流作为value
      bucket.Put(block.Hash, block.Serialize())
      bucket.Put([]byte("LastHashKey"), block.Hash)
      //更新一下最后一个区块的hash值
      bc.tail = block.Hash

      return nil
   })
}
```

#### 3.3.5创建新的钱包

找到钱包所有的地址,把新创建的地址,加进钱包

![](.\image\钱包的创建.png)

##### 

```
commandLine.go

//创建钱包
func (cli *CLI) NewWallet() {
   //ws := NewWallet()
   //addr := ws.NewAddress()
   ws := NewWallets()
   address := ws.CreateWallet()
   fmt.Printf("地址：%s\n", address)
   //fmt.Printf("公钥 :%v\n", ws.PublicKey)
   //fmt.Printf("adress:=%s", addr)

}
```

##### 3.3.5.1 列出所有的地址

```
wallets.go


//显示钱包的所有地址
func NewWallets() *Wallets {

   var ws Wallets
   ws.WalletMap = make(map[string]*Wallet)

   ws.LoadFile()
   return &ws
}
```

```
wallets.go

//读取文件方法，把所有的wallet读出来
func (ws *Wallets) LoadFile() {
   //在读取之前，要先确认文件是否在，如果不存在，直接退出
   _, err := os.Stat(walletFile)
   if os.IsNotExist(err) {
      //ws.WalletsMap = make(map[string]*Wallet)
      return
   }

   //读取内容
   content, err := ioutil.ReadFile(walletFile)
   if err != nil {
      log.Panic(err)
   }

   //解码
   //panic: gob: type not registered for interface: elliptic.p256Curve
   gob.Register(elliptic.P256())

   decoder := gob.NewDecoder(bytes.NewReader(content))

   var wsLocal Wallets

   err = decoder.Decode(&wsLocal)
   if err != nil {
      log.Panic(err)
   }

   //ws = &wsLocal
   //对于结构来说，里面有map的，要指定赋值，不要再最外层直接赋值
   ws.WalletMap = wsLocal.WalletMap
}
```

##### 3.3.5.2 创建新的钱包

```
wallets.go

//创建钱包地址
func (ws *Wallets) CreateWallet() string {
   wallet := NewWallet()
   address := wallet.NewAddress()

   ws.WalletMap[address] = wallet
   //把地址保存到文件中
   ws.SaveToFile()
   return address
}
```

```
wallet.go

//创建钱包
func NewWallet() *Wallet {
   //创建地址看着流程图做
   //创建一个椭圆曲线
   curve := elliptic.P256()
   //生成私钥
   privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
   if err != nil {
      log.Panic()
   }
   //生成公钥
   pubKeyOrig := privateKey.PublicKey

   //拼接 x,y
   pubKey := append(pubKeyOrig.X.Bytes(), pubKeyOrig.Y.Bytes()...)
   return &Wallet{Private: privateKey, PublicKey: pubKey}
}
```

```
wallet.go

//生成地址
func (w *Wallet) NewAddress() string {
   pubKey := w.PublicKey
   rip160HashValue := HashPubKey(pubKey)
   version := byte(00)
   //拼接version
   payload := append([]byte{version}, rip160HashValue...)

   //checksum
   checkCode := CheckSum(payload)

   //25字节数据
   payload = append(payload, checkCode...)

   //go语言有一个库，叫做btcd,这个是go语言实现的比特币全节点源码
   address := base58.Encode(payload)

   return address
}
```

```
//保存方法,把新建的wallet添加进去
func (ws *Wallets) SaveToFile() {
   var buffer bytes.Buffer

   //使用gob是先注册,否则会报错
   //panic: gob: type not registered for interface: elliptic.p256Curve
   gob.Register(elliptic.P256())
   encoder := gob.NewEncoder(&buffer)
   err := encoder.Encode(ws)
   if err != nil {
      log.Panic(err)
   }

   err = ioutil.WriteFile(walletFile, buffer.Bytes(), 0600)
   if err != nil {
      log.Panic(err)
   }
}
```



#### 3.3.6列举所有的地址

```
func (cli *CLI) ListAddresses() {
   ws := NewWallets()
   addresses := ws.ListAllAddresses()
   for _, address := range addresses {
      fmt.Printf("地址：%s\n", address)
   }
}
```

##### 3.3.6.1 列举出所有的钱包地址

```
//显示钱包的所有地址
func NewWallets() *Wallets {

   var ws Wallets
   ws.WalletMap = make(map[string]*Wallet)

   ws.LoadFile()
   return &ws
}
```

##### 3.3.6.2 遍历钱包,列出所有的地址

```
func (ws *Wallets) ListAllAddresses() []string {
   var addresses []string
   //遍历钱包，将所有的key取出来返回
   for address := range ws.WalletMap {
      addresses = append(addresses, address)
   }

   return addresses
}
```



### 3.4其他一些技术细节

#### 3.4.1梅克尔树的拼装

```
//模拟梅克尔根，只是对交易的数据做简单的拼接，而不做二叉树处理！
func (block *Block) MakeMerkelRoot() []byte {

   var info []byte
   //var finalInfo [][]byte
   for _, tx := range block.Transactions {
      //将交易的哈希值拼接起来，再整体做哈希处理
      info = append(info, tx.TXID...)
      //finalInfo = [][]byte{tx.TXID}
   }

   hash := sha256.Sum256(info)
   return hash[:]
}
```

#### 3.4.2 交易打印格式化

```
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
```