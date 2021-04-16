# xbt-adapter

本项目适配了openwallet.AssetsAdapter接口，给应用提供了底层的区块链协议支持。

## 项目依赖库

- [go-owcrypt](https://github.com/blocktree/go-owcrypt.git)
- [go-owcdrivers](https://github.com/blocktree/.git)

## 项目信息
- 官网 : https://www.xbt.wang/
- 区块浏览器 : https://www.xbt.wang/explorer
- 全节点 : TODO
- 全节点rpc接口 : https://www.yuque.com/docs/share/4fdb8f60-395a-4055-894f-ff1639615209
- curl -H 'Content-Type: application/json' -d '{"address":"xB666d7020F961D96cf99aFD440D010575C99b4e30"}' http://127.0.0.1:3000/account/balance
## 如何测试

openwtester包下的测试用例已经集成了openwallet钱包体系，创建conf文件，新建FIL.ini文件，编辑如下内容：

```ini
# node api url
serverAPI = "https://api.xbt.wang"

#xbt tools api
xbtToolsAPI = "http://127.0.0.1:3000"

# Cache data file directory, default = "", current directory: ./data
dataDir = ""

# min fee
fixedFee = "0.1"
```