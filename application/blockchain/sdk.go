package blockchain

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var (
	SDK           *fabsdk.FabricSDK
	ChannelName   = "assetschannel"
	ChainCodeName = "blockchain-real-estate"
	Org           = "org1"
	User          = "Admin"
	ConfigPath    = "blockchain/config.yaml"
)

func Init() {
	var err error
	SDK, err = fabsdk.New(config.FromFile(ConfigPath))
	if err != nil {
		panic(err)
	}
}

func ChannelExecute(fcn string, args [][]byte) (channel.Response, error) {
	ctx := SDK.ChannelContext(ChannelName, fabsdk.WithOrg(Org), fabsdk.WithUser(User))
	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}

	resp, err := cli.Execute(channel.Request{
		ChaincodeID: ChainCodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints("peer0.org1.blockchainrealestate.com"))
	if err != nil {
		return channel.Response{}, err
	}
	return resp, nil
}

func ChannelQuery(fcn string, args [][]byte) (channel.Response, error) {
	ctx := SDK.ChannelContext(ChannelName, fabsdk.WithOrg(Org), fabsdk.WithUser(User))
	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}

	resp, err := cli.Query(channel.Request{
		ChaincodeID: ChainCodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints("peer0.org1.blockchainrealestate.com"))
	if err != nil {
		return channel.Response{}, err
	}
	return resp, nil
}
