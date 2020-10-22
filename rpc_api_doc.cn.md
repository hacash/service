Hacash 全节点 RPC API 文档
===

此文档包含区块扫描、交易转账查询、账户余额查询、区块钻石信息查询、创建新账户、创建转账交易等接口的调用规范和示例，是开发Hacash区块链浏览器及对接交易所等功能所必须的接口支持。

本文档内将提供示例测试接口（临时可用，但随时可能关闭），可以帮助你即时查看接口返回的内容，或者临时做测试、调试使用。在生产环境中，请**务必**启用自己的服务器并搭建全节点，才能确保接口的稳定可用和安全性。全节点搭建教程见 [hacash.org](https://hacash.org/)。

下载最新版本的全节点程序，并启动程序同步完所有区块之后，要启用本 RPC API 服务，需要在配置文件（hacash.config.ini）内加上如下配置：

```ini

[service]
enable = true
rpc_listen_port = 8083

```

以上配置 `enable = true` 表示启用 RPC 接口服务，`rpc_listen_port = 8083` 表示监听的 http 服务端口为 8083 。

此时访问 `http://127.0.0.1:8083/` ，正常情况下你将看到如下返回：

```json
{
    "ret": 0,
    "service": "hacash node rpc"
}
```

此时表示 RPC 服务已经正常运行。

### 简单示例、快速入门

在文档正式开始之前，我们来测试一个简单的示例。在全节点保持运行的情况下，访问 [http://127.0.0.1:8083/query?action=balances&address_list=1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS](http://127.0.0.1:8083/query?action=balances&address_list=1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS)， 这是一个查询账户地址余额的接口，正常情况下将会返回例如：

```json
{
    "list": [
        {
            "diamond": 1016,
            "hacash": "ㄜ1,474,845:244",
            "satoshi": 0
        }
    ],
    "ret": 0
}
```

可以看到，Hacash 的 RPC API 服务采用标准的 http 接口方式，你可以简单的在浏览器中或你开发的程序中很方便、简单的使用它。

下面的文档将使用 hacash.org 提供的测试接口作为示例，访问 [http://rpcapi.hacash.org/query?action=balances&address_list=1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS](http://rpcapi.hacash.org/query?action=balances&address_list=1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS) 应该返回和上面相同的接口数据内容。

示例二：[http://rpcapi.hacash.org/query?action=balances&address_list=1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS,1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9&unitmei=true](http://rpcapi.hacash.org/query?action=balances&address_list=1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS,1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9&unitmei=true) 表示同时查询两个地址的余额，且返回的余额单位为“枚”。

### 接口格式、通用参数

采用标准的 HTTP 接口请求和应答方式，包含 4 个路径：

   1. /create  生成数据 ， GET 方式 ， 如创建账户、生成转账交易等
   2. /submit  提交数据 ， POST 方式 ， 如提交交易进待确认池等
   3. /query   查询数据 ， GET 方式 ， 如查询账户余额等
   4. /operate 修改数据 ， GET/POST 方式 ， 如修改系统运行参数等
   
多个接口通用的参数介绍如下：

| 参数名 | 类型 | 默认值 | 示例值 | 功能简介 |
| ----  | ----  | ----  | ----  | ----  |
| action | string | -    | balances, diamonds, block_intro, accounts ... | 要查询、生成的数据类型，定义接口的功能 |
| unitmei | bool  | false | true, false, 1, 0 | 是否采用浮点数的“枚”为单位传递或返回数额，否则采用Hacash标准化单位方式。 例如使用“枚”为单位："12.5086"，而标准化单位："ㄜ125086:244" |
| kind | menu  | - | h, s, d, hs, hd, hsd | 筛选返回的账户、交易信息类型。h: hacash, s: satoshi, d: diamond。用途：比如在扫描区块时只需要返回 HAC 转账内容而忽略其他两者，则传递 `kind=h` 即可。 |
| hexbody | bool  | false | true, false, 1, 0 | `/submit` 在提交数据时，是否使用 hex 字符串的形式为 Http Body。默认为原生 bytes 形式。 |


### 返回值、公共字段

采用标准 JSON 格式应答所有请求。公共字段如下：

```js

{
  "ret": 0,  // 表示返回类型， 0 为正确，  >= 1 则表示发生错误或者查询不存在
  "list": [...] // 某些返回列表数据的接口将使用
}

```

下面将介绍每一个具体接口的参数传递和返回值细节。

---

## 1. /create 创建

#### 1.1 创建账户 `GET: /create ? action=accounts`

随机批量批量创建账户，返回包含私钥、公钥和地址的账户信息列表。可传递参数：

1. number [int] 表示批量创建账户的数量，默认为 1， 最大值 200

示例接口： [http://rpcapi.hacash.org/create?action=accounts&number=5](http://rpcapi.hacash.org/create?action=accounts&number=5)

返回值：

```js
{
    list: [
        {
            address: "1NLEYVmmUkhAH18WfCUDc5CHnbr7Bv5TaS",  // 账户地址
            prikey: "2e50243243abc2e41f3b2ae90029640e235d884a88cfb5ea3e4d0e9efbae6710",  // 私钥
            pubkey: "03e22fc27a0d7ae325fa024875febd58266b8b6adbfb966116c9ba958ff5bad7e6"  // 公钥
        },
        ...
    ],
    ret: 0
}
```

特别注意：系统采用随机生成算法创建账户，系统并不会保留或记录本接口创建的账户私钥，请务必备份储存好创建的私钥，并做好安全防护。

#### 1.2 创建转账交易 `GET: /create ? action=value_transfer_tx`

本接口用于创建HAC、单向转移的比特币和区块钻石的转账交易，基本参数如下：

1. unitmei [bool] 是否采用单位“枚”浮点数形式，去解析传递的数额参数
2. main_prikey [hex string] 主地址/手续费地址的私钥hex字符串
3. fee [string] 交易将要给出的手续费值，例如 "0.0001" 或 "ㄜ1:244"
4. timestamp [int] 交易的时间戳；可选传递；不传时则自动设为当前时间戳
5. transfer_kind [menu] (hacash, satoshi, diamond) 要创建的交易类型，HAC转账、BTC转账还是区块钻石转账

【注意一】：当只有传递相同的 `timestamp` 时间戳参数，而且保持其它参数始终相同时，则每次创建的交易则具备相同的 hash 值，被视为同一笔交易。

【注意二】：由于 Hacash 系统支持对同一笔交易进行重复签名手续费竞价，所以仅改变 `fee` 字段并不会更改交易的 `hash` 值， 而只会改变其 `hash_with_fee` 值。

【注意三】：手续费 `fee` 字段不能设置得过小，否则将无法被整个系统接受，目前费用最小值为 0.0001 枚（即 ㄜ1:244 ）。请不要设置手续费低于这个值。

##### 1.2.1 创建HAC普通转账交易

传递参数 `transfer_kind=hacash`，且增加参数如下：

 - amount [string] 转账数额；单位格式视 `unitmei` 参数而定； 例如 "0.1" 或 "ㄜ1:247"。
 - to_address [string] 对方（收款）账户地址
 
调用接口示例： [http://rpcapi.hacash.org/create?action=value_transfer_tx&main_prikey=8D969EEF6ECAD3C29A3A629280E686CF0C3F5D5A86AFF3CA12020C923ADC6C92&fee=0.0001&unitmei=true&timestamp=1603284999&transfer_kind=hacash&amount=12.45&to_address=1NLEYVmmUkhAH18WfCUDc5CHnbr7Bv5TaS](http://rpcapi.hacash.org/create?action=value_transfer_tx&main_prikey=8D969EEF6ECAD3C29A3A629280E686CF0C3F5D5A86AFF3CA12020C923ADC6C92&fee=0.0001&unitmei=true&timestamp=1603284999&transfer_kind=hacash&amount=12.45&to_address=1NLEYVmmUkhAH18WfCUDc5CHnbr7Bv5TaS)

返回值如下：

```js
{
    // 公共参数
    ret: 0,
    // 交易的 hash 值
    hash: "6066cef4fe51669aec5b5596375dba11dafaf2560c4bd8c0432ac4ea98ff3ad1",
    // 交易包含 fee 的 hash 值
    hash_with_fee: "9cbc4821d0d921b429dbe4d6b67d6412aa5cae4b724f1e7dab9f870646cb1bb6",
    // 交易体、内容 的 hex 值
    body: "02005f90300700e63c33a796b3032ce6b856f68fccf06608d9ed18f401010001000100e9fdd992667de1734f0ef758bafcd517179e6f1bf60204dd00010231745adae24044ff09c3541537160abb8d5d720275bbaeed0b3d035b1e8b263cb73b724218f13c09c16e7065212128cf0c037ebb9e588754eb073886486d950607d59bef462d2731e15b667c6ff1f0badd6259c6f58d5ca7a5f75856b8cae8e80000",
    // 交易使用的时间戳
    timestamp: 1603284999
}
```

在生产环境中，请在数据库中保存以上的返回值，以便进行对账，或者在区块链网络延迟时重新提交未被确认的交易。上面的内容并不会泄露你的私钥，而仅仅是签名后的交易数据，请放心储存。

创建BTC转账和区块钻石的转账，调用接口的的返回值与上者相同。

##### 1.2.2 创建单向转移的 Bitcoin 普通转账交易

传递参数 `transfer_kind=satoshi`，且增加参数如下：

 - amount [int] 要支付的比特币数额，单位为“聪”、“satoshi” (0.00000001枚比特币)；例如转账10枚比特币则传递 "1000000000"，转账0.01枚则传递"1000000"；系统不支持低于 1 聪的比特币单位。
 - to_address [string] 对方（收款）账户地址

例如给某个地址转账一枚比特币，示例接口如：[http://rpcapi.hacash.org/create?action=value_transfer_tx&main_prikey=8D969EEF6ECAD3C29A3A629280E686CF0C3F5D5A86AFF3CA12020C923ADC6C92&fee=0.0001&unitmei=true&timestamp=1603284999&transfer_kind=satoshi&amount=100000000&to_address=1NLEYVmmUkhAH18WfCUDc5CHnbr7Bv5TaS](http://rpcapi.hacash.org/create?action=value_transfer_tx&main_prikey=8D969EEF6ECAD3C29A3A629280E686CF0C3F5D5A86AFF3CA12020C923ADC6C92&fee=0.0001&unitmei=true&timestamp=1603284999&transfer_kind=satoshi&amount=100000000&to_address=1NLEYVmmUkhAH18WfCUDc5CHnbr7Bv5TaS)

##### 1.2.3 创建区块钻石转账交易

传递参数 `transfer_kind=diamond`，且增加参数如下：

 - diamonds [string] 逗号分割的钻石字面值，例如 "EVUNXZ,BVVTSI"，可传一个或多个，最多一次批量转移200枚钻石
 - diamond_owner_prikey [hex string] 选填，付款（支付钻石、钻石所有者）的账户私钥；如果不传则默认为 `main_prikey`
 - to_address [string] 对方（收取钻石）账户地址

示例接口调用：[http://rpcapi.hacash.org/create?action=value_transfer_tx&main_prikey=8D969EEF6ECAD3C29A3A629280E686CF0C3F5D5A86AFF3CA12020C923ADC6C92&fee=0.0003&unitmei=true&timestamp=1603284999&transfer_kind=diamond&diamonds=EVUNXZ,BVVTSI&diamond_owner_prikey=EF797C8118F02DFB649607DD5D3F8C7623048C9C063D532CC95C5ED7A898A64F&to_address=1NLEYVmmUkhAH18WfCUDc5CHnbr7Bv5TaS](http://rpcapi.hacash.org/create?action=value_transfer_tx&main_prikey=8D969EEF6ECAD3C29A3A629280E686CF0C3F5D5A86AFF3CA12020C923ADC6C92&fee=0.0003&unitmei=true&timestamp=1603284999&transfer_kind=diamond&diamonds=EVUNXZ,BVVTSI&diamond_owner_prikey=EF797C8118F02DFB649607DD5D3F8C7623048C9C063D532CC95C5ED7A898A64F&to_address=1NLEYVmmUkhAH18WfCUDc5CHnbr7Bv5TaS)


---

## 2. /submit 提交

#### 2.1 提交交易至交易池 `POST: /submit ? action=transaction`

提交一笔交易至全网的内存池内。

url 参数如下：

 - hexbody [bool] 以 hex 字符串的形式传递 txbody 值；否则以原生 bytes 形式传递
 
post body 参数如下

 - txbody [string / bytes] 交易数据
 
调用后返回值如下

```js
{
    ret: 0 // ret = 0 表示返回成功
}
```

或者返回错误

```js
{
    ret: 1 // ret = 1 表示提交交易错误
    // 错误消息如下：
    errmsg: "address 1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9 balance ㄜ0:0 not enough， need ㄜ1,245:246."
}
```
curl 命令行测试调用示例：

```shell script
curl "http://127.0.0.1:8083/submit?action=transaction&hexbody=1" -X POST -d "txbody=02005f90300700e63c33a796b3032ce6b856f68fccf06608d9ed18f401010001000100e9fdd992667de1734f0ef758bafcd517179e6f1bf60204dd00010231745adae24044ff09c3541537160abb8d5d720275bbaeed0b3d035b1e8b263cb73b724218f13c09c16e7065212128cf0c037ebb9e588754eb073886486d950607d59bef462d2731e15b667c6ff1f0badd6259c6f58d5ca7a5f75856b8cae8e80000"
```
