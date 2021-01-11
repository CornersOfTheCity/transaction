# transaction

## 概要

这是一个Fabric练习项目，使用GO语言SDK，用GO语言编写的链码。有五个成员，四个用户和一个管理员，可以出售、捐赠和购买房产。

**环境：**

- Go Version：1.11.1
- Fabric Version：1.4.0

共划分三个组织，使用solo共识机制。

## 运行步骤

- 项目运行目录：$GOPATH/src
- 未使用mod管理，即哪个go mod环境设置为go env GO111MODULE=auto
- 首先测试chaincode是否正常调用，运行`chaincode/chaincode_test.go`测试用例
- 在deploy目录下运行`./start.sh`,观察有无报错提示。运行成功后在终端执行`docker exec cli peer chaincode invoke -C assetschannel -n blockchain-real-estate -c '{"Args":["queryAccountList"]}'` 等cli命令，Args可以替换为Invoke中的funcName，先验证链码是否正确安装及区块链网络能否正常工作。建议`./start.sh`之前可以先运行`./stop.sh`清理一下环境。
- 执行`application/sdk_test.go`，测试是否可以成功使用SDK调用链码
- go run main.go运行。





