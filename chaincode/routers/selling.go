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

func CreateSelling(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("参数个数不满足")
	}
	objectOfSale := args[0]
	seller := args[1]
	price := args[2]
	salePeriod := args[3]
	if objectOfSale == "" || seller == "" || price == "" || salePeriod == "" {
		return shim.Error("参数存在空值")
	}
	var formattedPrice float64
	if val, err := strconv.ParseFloat(price, 64); err != nil {
		return shim.Error(fmt.Sprintf("price参数格式转换出错: %s", err))
	} else {
		formattedPrice = val
	}

	var formattedSalePeriod int
	if val, err := strconv.Atoi(salePeriod); err != nil {
		return shim.Error(fmt.Sprintf("salePeriod参数格式转换出错: %s", err))
	} else {
		formattedSalePeriod = val
	}

	resultsRealEstate, err := utils.GetStateByPartialCompositeKeys2(stub, lib.RealEstateKey, []string{seller, objectOfSale})
	if err != nil || len(resultsRealEstate) != 1 {
		return shim.Error(fmt.Sprintf("验证%s属于%s失败: %s", objectOfSale, seller, err))
	}

	var realEstate lib.RealEstate
	if err = json.Unmarshal(resultsRealEstate[0], &realEstate); err != nil {
		return shim.Error(fmt.Sprintf("CreateSelling-反序列化出错: %s", err))
	}

	if realEstate.Encumbrance {
		return shim.Error("此房地产已经作为担保状态，不能重复发起销售")
	}

	selling := &lib.Selling{
		ObjectOfSale:  objectOfSale,
		Seller:        seller,
		Buyer:         "",
		Price:         formattedPrice,
		CreateTime:    time.Now().Local().Format("2006-01-02 15:04:05"),
		SalePeriod:    formattedSalePeriod,
		SellingStatus: lib.SellingStatusConstant()["saleStart"],
	}

	if err := utils.WriteLedger(realEstate, stub, lib.RealEstateKey, []string{realEstate.Proprietor, realEstate.RealEstateID}); err != nil {
		return shim.Error(fmt.Sprintf("%s", err))
	}

	realEstate.Encumbrance = true
	if err := utils.WriteLedger(realEstate, stub, lib.RealEstateKey, []string{realEstate.Proprietor, realEstate.RealEstateID}); err != nil {
		return shim.Error(fmt.Sprintf("%s", err))
	}
	sellingByte, err := json.Marshal(selling)
	if err != nil {
		return shim.Error(fmt.Sprintf("序列化成功创建的信息出错: %s", err))
	}
	return shim.Success(sellingByte)

}

func CreateSellingByBuy(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("参数个数不满足")
	}
	objectOfSale := args[0]
	seller := args[1]
	buyer := args[2]
	if objectOfSale == "" || seller == "" || buyer == "" {
		return shim.Error("参数存在空值")
	}
	if seller == buyer {
		return shim.Error("买家和卖家不能同一人")
	}

	resultsRealEstate, err := utils.GetStateByPartialCompositeKeys2(stub, lib.RealEstateKey, []string{seller, objectOfSale})
	if err != nil || len(resultsRealEstate) != 1 {
		return shim.Error(fmt.Sprintf("根据%s和%s获取想要购买的房产信息失败: %s", objectOfSale, seller, err))
	}
	resultsSelling, err := utils.GetStateByPartialCompositeKeys2(stub, lib.SellingKey, []string{seller, objectOfSale})
	if err != nil || len(resultsSelling) != 1 {
		return shim.Error(fmt.Sprintf("根据%s和%s获取销售信息失败: %s", objectOfSale, seller, err))
	}

	var selling lib.Selling
	if err := json.Unmarshal(resultsSelling[0], &selling); err != nil {
		return shim.Error(fmt.Sprintf("CreateSellingBuy-反序列化出错: %s", err))
	}

	if selling.SellingStatus != lib.SellingStatusConstant()["saleStart"] {
		return shim.Error("此交易不属于销售中状态，已经无法购买")
	}

	resultsAccount, err := utils.GetStateByPartialCompositeKeys(stub, lib.AccountKey, []string{buyer})
	if err != nil || len(resultsAccount) != 1 {
		return shim.Error(fmt.Sprintf("buyer买家信息验证失败%s", err))
	}

	var buyerAccount lib.Account
	if err = json.Unmarshal(resultsAccount[0], &buyerAccount); err != nil {
		return shim.Error(fmt.Sprintf("查询buyer买家信息-反序列化出错: %s", err))
	}
	if buyerAccount.UserName == "管理员" {
		return shim.Error(fmt.Sprintf("管理员不能购买%s", err))
	}

	if buyerAccount.Balance < selling.Price {
		return shim.Error(fmt.Sprintf("房产售价为%f,您的当前余额为%f,购买失败", selling.Price, buyerAccount.Balance))
	}

	selling.Buyer = buyer
	selling.SellingStatus = lib.SellingStatusConstant()["delivery"]
	if err := utils.WriteLedger(selling, stub, lib.AccountKey, []string{selling.Seller, selling.ObjectOfSale}); err != nil {
		return shim.Error(fmt.Sprintf("将buyer写入交易selling,修改交易状态 失败%s", err))
	}

	sellingBuy := &lib.SellingBuy{
		Buyer:      buyer,
		CreateTime: time.Now().Local().Format("2006-01-02 15:04:05"),
		Selling:    selling,
	}
	local, _ := time.LoadLocation("Local")
	createTimeUnixNano, _ := time.ParseInLocation("2006-01-02 15:04:05", sellingBuy.CreateTime, local)
	if err := utils.WriteLedger(sellingBuy, stub, lib.SellingBuyKey, []string{sellingBuy.Buyer, fmt.Sprintf("%d", createTimeUnixNano.UnixNano())}); err != nil {
		return shim.Error(fmt.Sprintf("将本次购买交易写入账本失败%s", err))
	}
	sellingBuyByte, err := json.Marshal(sellingBuy)
	if err != nil {
		return shim.Error(fmt.Sprintf("序列化成功创建的信息出错: %s", err))
	}
	buyerAccount.Balance -= selling.Price
	if err := utils.WriteLedger(buyerAccount, stub, lib.AccountKey, []string{buyerAccount.AccountId}); err != nil {
		return shim.Error(fmt.Sprintf("扣取买家余额失败%s", err))
	}
	// 成功返回
	return shim.Success(sellingBuyByte)
}

func QuerySellingList(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var sellingList []lib.Selling
	results, err := utils.GetStateByPartialCompositeKeys2(stub, lib.SellingKey, args)
	if err != nil {
		return shim.Error(fmt.Sprintf("%s", err))
	}
	for _, v := range results {
		if v != nil {
			var selling lib.Selling
			err := json.Unmarshal(v, &selling)
			if err != nil {
				return shim.Error(fmt.Sprintf("QuerySellingList-反序列化出错: %s", err))
			}
			sellingList = append(sellingList, selling)
		}
	}
	sellingListByte, err := json.Marshal(sellingList)
	if err != nil {
		return shim.Error(fmt.Sprintf("QuerySellingList-序列化出错: %s", err))
	}
	return shim.Success(sellingListByte)
}

func QuerySellingListByBuyer(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error(fmt.Sprintf("必须指定买家AccountId查询"))
	}
	var sellingBuyList []lib.SellingBuy
	results, err := utils.GetStateByPartialCompositeKeys2(stub, lib.SellingBuyKey, args)
	if err != nil {
		return shim.Error(fmt.Sprintf("%s", err))
	}
	for _, v := range results {
		if v != nil {
			var sellingBuy lib.SellingBuy
			err := json.Unmarshal(v, &sellingBuy)
			if err != nil {
				return shim.Error(fmt.Sprintf("QuerySellingListByBuyer-反序列化出错: %s", err))
			}
			sellingBuyList = append(sellingBuyList, sellingBuy)
		}
	}
	sellingBuyListByte, err := json.Marshal(sellingBuyList)
	if err != nil {
		return shim.Error(fmt.Sprintf("QuerySellingListByBuyer-序列化出错: %s", err))
	}
	return shim.Success(sellingBuyListByte)

}

func UpdateSelling(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("参数个数不满足")
	}
	objectOfSale := args[0]
	seller := args[1]
	buyer := args[2]
	status := args[3]
	if objectOfSale == "" || seller == "" || status == "" {
		return shim.Error("参数存在空值")
	}
	if buyer == seller {
		return shim.Error("买家和卖家不能同一人")
	}
	resultsRealEstate, err := utils.GetStateByPartialCompositeKeys2(stub, lib.RealEstateKey, []string{seller, objectOfSale})
	if err != nil || len(resultsRealEstate) != 1 {
		return shim.Error(fmt.Sprintf("根据%s和%s获取想要购买的房产信息失败: %s", objectOfSale, seller, err))
	}
	var realEstate lib.RealEstate
	if err := json.Unmarshal(resultsRealEstate[0], &realEstate); err != nil {
		return shim.Error(fmt.Sprintf("UpdateSellingBySeller-反序列化出错: %s", err))
	}
	resultsSelling, err := utils.GetStateByPartialCompositeKeys2(stub, lib.SellingKey, []string{seller, objectOfSale})
	if err != nil || len(resultsSelling) != 1 {
		return shim.Error(fmt.Sprintf("根据%s和%s获取销售信息失败: %s", objectOfSale, seller, err))
	}
	var selling lib.Selling
	if err := json.Unmarshal(resultsSelling[0], &selling); err != nil {
		return shim.Error(fmt.Sprintf("UpdateSellingBySeller-反序列化出错: %s", err))
	}

	var sellingBuy lib.SellingBuy
	if selling.SellingStatus != lib.SellingStatusConstant()["saleStart"] {
		resultsSellingByBuyer, err := utils.GetStateByPartialCompositeKeys2(stub, lib.SellingKey, []string{buyer})
		if err != nil || len(resultsSellingByBuyer) == 0 {
			return shim.Error(fmt.Sprintf("根据%s获取买家购买信息失败: %s", buyer, err))
		}
		for _, v := range resultsSellingByBuyer {
			if v != nil {
				var s lib.SellingBuy
				err := json.Unmarshal(v, &s)
				if err != nil {
					return shim.Error(fmt.Sprintf("UpdateSellingBySeller-反序列化出错: %s", err))
				}
				if s.Selling.ObjectOfSale == objectOfSale && s.Selling.Seller == seller && s.Buyer == buyer {
					if s.Selling.SellingStatus == lib.SellingStatusConstant()["delivery"] {
						sellingBuy = s
						break
					}
				}
			}
		}
	}
	var data []byte

	switch status {
	case "done":
		if selling.SellingStatus != lib.SellingStatusConstant()["delivery"] {
			return shim.Error("此交易并不处于交付中，确认收款失败")
		}
		resultsSellerAccount, err := utils.GetStateByPartialCompositeKeys(stub, lib.AccountKey, []string{seller})
		if err != nil || len(resultsSellerAccount) != 1 {
			return shim.Error(fmt.Sprintf("seller卖家信息验证失败%s", err))
		}
		var accountSeller lib.Account
		if err := json.Unmarshal(resultsSellerAccount[0], &accountSeller); err != nil {
			return shim.Error(fmt.Sprintf("查询seller卖家信息-反序列化出错: %s", err))
		}
		accountSeller.Balance += selling.Price
		if err := utils.WriteLedger(accountSeller, stub, lib.AccountKey, []string{accountSeller.AccountId}); err != nil {
			return shim.Error(fmt.Sprintf("卖家确认接收资金失败%s", err))
		}
		realEstate.Proprietor = buyer
		realEstate.Encumbrance = false
		realEstate.RealEstateID = fmt.Sprintf("%d", time.Now().Local().UnixNano())
		if err := utils.WriteLedger(realEstate, stub, lib.RealEstateKey, []string{realEstate.Proprietor, realEstate.RealEstateID}); err != nil {
			return shim.Error(fmt.Sprintf("%s", err))
		}

		if err := utils.DelLedger(stub, lib.RealEstateKey, []string{seller, objectOfSale}); err != nil {
			return shim.Error(fmt.Sprintf("%s", err))
		}

		sellingBuy.Selling = selling
		local, _ := time.LoadLocation("Local")
		sellingBuyCreateTimeUnixNano, _ := time.ParseInLocation("2006-01-02 15:04:05", sellingBuy.CreateTime, local)
		if err := utils.WriteLedger(sellingBuy, stub, lib.SellingBuyKey, []string{sellingBuy.Buyer, fmt.Sprintf("%d", sellingBuyCreateTimeUnixNano.UnixNano())}); err != nil {
			return shim.Error(fmt.Sprintf("将本次购买交易写入账本失败%s", err))
		}
		data, err = json.Marshal(sellingBuy)
		if err != nil {
			return shim.Error(fmt.Sprintf("序列化购买交易的信息出错: %s", err))
		}
		break
	case "cancelled":
		data, err = closeSelling("cancelled", selling, realEstate, sellingBuy, buyer, stub)
		if err != nil {
			return shim.Error(fmt.Sprintf("%s", err))
		}
		break
	case "expired":
		data, err = closeSelling("expired", selling, realEstate, sellingBuy, buyer, stub)
		if err != nil {
			return shim.Error(fmt.Sprintf("%s", err))
		}
		break
	default:
		return shim.Error(fmt.Sprintf("%s状态不支持", status))
	}
	return shim.Success(data)

}

func closeSelling(closeStart string, selling lib.Selling, realEstate lib.RealEstate, sellingBuy lib.SellingBuy, buyer string, stub shim.ChaincodeStubInterface) ([]byte, error) {
	switch selling.SellingStatus {
	case lib.SellingStatusConstant()["saleStart"]:
		selling.SellingStatus = lib.SellingStatusConstant()[closeStart]
		realEstate.Encumbrance = false
		if err := utils.WriteLedger(realEstate, stub, lib.RealEstateKey, []string{realEstate.Proprietor, realEstate.RealEstateID}); err != nil {
			return nil, err
		}
		if err := utils.WriteLedger(selling, stub, lib.SellingKey, []string{selling.Seller, selling.ObjectOfSale}); err != nil {
			return nil, err
		}
		data, err := json.Marshal(selling)
		if err != nil {
			return nil, err
		}
		return data, nil
	case lib.SellingStatusConstant()["delivery"]:
		resultsBuyerAccount, err := utils.GetStateByPartialCompositeKeys(stub, lib.AccountKey, []string{buyer})
		if err != nil || len(resultsBuyerAccount) != 1 {
			return nil, err
		}
		var accountBuyer lib.Account
		if err = json.Unmarshal(resultsBuyerAccount[0], &accountBuyer); err != nil {
			return nil, err
		}

		accountBuyer.Balance += selling.Price
		if err := utils.WriteLedger(accountBuyer, stub, lib.AccountKey, []string{accountBuyer.AccountId}); err != nil {
			return nil, err
		}
		realEstate.Encumbrance = false
		if err := utils.WriteLedger(realEstate, stub, lib.RealEstateKey, []string{realEstate.Proprietor, realEstate.RealEstateID}); err != nil {
			return nil, err
		}
		selling.SellingStatus = lib.SellingStatusConstant()[closeStart]
		if err := utils.WriteLedger(selling, stub, lib.SellingKey, []string{selling.Seller, selling.ObjectOfSale}); err != nil {
			return nil, err
		}
		sellingBuy.Selling = selling
		local, _ := time.LoadLocation("Local")
		sellingBuyCreateTimeUnixNano, _ := time.ParseInLocation("2006-01-02 15:04:05", sellingBuy.CreateTime, local)
		if err := utils.WriteLedger(sellingBuy, stub, lib.SellingBuyKey, []string{sellingBuy.Buyer, fmt.Sprintf("%d", sellingBuyCreateTimeUnixNano.UnixNano())}); err != nil {
			return nil, err
		}
		data, err := json.Marshal(sellingBuy)
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		return nil, nil

	}
}
