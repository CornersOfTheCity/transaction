package routers

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"time"
	"transaction/chaincode/lib"
	"transaction/chaincode/utils"
)

func CreateRealEstate(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("参数个数不满足")
	}
	accountId := args[0]
	proprietor := args[1]
	totalArea := args[2]
	livingSpace := args[3]
	if accountId == "" || proprietor == "" || totalArea == "" || livingSpace == "" {
		return shim.Error("参数存在空值")
	}
	if accountId == proprietor {
		return shim.Error("操作人应为管理员且与所有人不能相同")
	}
	// 参数数据格式转换
	var formattedTotalArea float64
	if val, err := strconv.ParseFloat(totalArea, 64); err != nil {
		return shim.Error(fmt.Sprintf("totalArea参数格式转换出错: %s", err))
	} else {
		formattedTotalArea = val
	}
	var formattedLivingSpace float64
	if val, err := strconv.ParseFloat(livingSpace, 64); err != nil {
		return shim.Error(fmt.Sprintf("livingSpace参数格式转换出错: %s", err))
	} else {
		formattedLivingSpace = val
	}

	resultsAccount, err := utils.GetStateByPartialCompositeKeys(stub, lib.AccountKey, []string{accountId})
	if err != nil {
		return shim.Error(fmt.Sprintf("操作人权限验证失败%s", err))
	}
	var account lib.Account
	if err = json.Unmarshal(resultsAccount[0], &account); err != nil {
		return shim.Error(fmt.Sprintf("查询操作人信息-反序列化出错: %s", err))
	}
	if account.UserName != "管理员" {
		return shim.Error(fmt.Sprintf("操作人权限不足%s", err))
	}

	resultsProprietor, err := utils.GetStateByPartialCompositeKeys(stub, lib.AccountKey, []string{proprietor})
	if err != nil || len(resultsProprietor) != 1 {
		return shim.Error(fmt.Sprintf("业主proprietor信息验证失败%s", err))
	}
	realEstate := &lib.RealEstate{
		RealEstateID: fmt.Sprintf("%d", time.Now().Local().Unix()),
		Proprietor:   proprietor,
		Encumbrance:  false,
		TotalArea:    formattedTotalArea,
		LivingSpace:  formattedLivingSpace,
	}

	if err := utils.WriteLedger(realEstate, stub, lib.RealEstateKey, []string{realEstate.Proprietor, realEstate.RealEstateID}); err != nil {
		return shim.Error(fmt.Sprintf("%s", err))
	}
	realEstateByte, err := json.Marshal(realEstate)
	if err != nil {
		return shim.Error(fmt.Sprintf("序列化成功创建的信息出错: %s", err))
	}
	return shim.Success(realEstateByte)

}

func QueryRealEstateList(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var realEstateList []lib.RealEstate
	results, err := utils.GetStateByPartialCompositeKeys2(stub, lib.RealEstateKey, args)
	if err != nil {
		return shim.Error(fmt.Sprintf("%s", err))
	}
	for _, v := range results {
		if v != nil {
			var realEstate lib.RealEstate
			err := json.Unmarshal(v, &realEstate)
			if err != nil {
				return shim.Error(fmt.Sprintf("QueryRealEstateList-反序列化出错: %s", err))
			}
			realEstateList = append(realEstateList, realEstate)
		}
	}
	realEstateListByte, err := json.Marshal(realEstateList)
	if err != nil {
		return shim.Error(fmt.Sprintf("QueryRealEstateList-序列化出错: %s", err))
	}
	return shim.Success(realEstateListByte)
}
