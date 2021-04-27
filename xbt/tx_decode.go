/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package xbt

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/xbt-adapter/xbtTransaction"
	"github.com/shopspring/decimal"
	"sort"
	"strconv"
	"time"

	"github.com/blocktree/openwallet/v2/openwallet"
)

type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	openwallet.AddressDecoderV2
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.CreateXbtRawTransaction(wrapper, rawTx)
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.SignXbtRawTransaction(wrapper, rawTx)
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.VerifyXBTRawTransaction(wrapper, rawTx)
}

func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}

	from := rawTx.Signatures[rawTx.Account.AccountID][0].Address.Address

	decoder.wm.Log.Info("update from : ", from)

	txStruct, err := xbtTransaction.NewTxStructFromJSON( rawTx.RawHex )
	if err != nil {
		decoder.wm.Log.Error("Error Tx : ", rawTx.RawHex)
		return nil, err
	}

	txid, err := decoder.wm.ApiClient.sendTransaction( txStruct )
	if err != nil {
		decoder.wm.Log.Error("Error Tx to send: ", rawTx.RawHex)
		return nil, err
	}

	rawTx.TxID = txid
	rawTx.IsSubmit = true

	decimals := int32(6)

	tx := openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       rawTx.Fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(&tx)

	return &tx, nil
}

func (decoder *TransactionDecoder) CreateXbtRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)

	if err != nil {
		return err
	}

	if len(addresses) == 0 {
		return openwallet.Errorf(openwallet.ErrAccountNotAddress, "[%s] have not addresses", rawTx.Account.AccountID)
	}

	addressesBalanceList := make([]AddrBalance, 0, len(addresses))

	for i, addr := range addresses {
		balance, err := decoder.wm.ApiClient.getBalance(addr.Address)
		if err != nil {
			return err
		}

		balance.index = i
		addressesBalanceList = append(addressesBalanceList, *balance)
	}

	sort.Slice(addressesBalanceList, func(i int, j int) bool {
		return addressesBalanceList[i].Balance.Cmp(addressesBalanceList[j].Balance) >= 0
	})

	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}

	amount, err := decimal.NewFromString( amountStr )
	if err!=nil {
		return errors.New( "wrong amount : " + amountStr )
	}

	fee, err := decoder.GetTxFee(rawTx.FeeRate, &amount)
	if err!=nil {
		return err
	}

	from := ""
	nonce := uint64(0)
	for _, a := range addressesBalanceList {
		from = a.Address
		break
	}

	if from == "" {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "the balance: %s is not enough", amountStr)
	}

	rawTx.TxFrom = []string{from}
	rawTx.TxTo = []string{to}
	rawTx.TxAmount = amountStr
	rawTx.Fees = fee.String()
	rawTx.FeeRate = fee.String()

	emptyTrans, message, err := decoder.CreateEmptyRawTransactionAndMessage(to, &amount, &fee)
	if err != nil {
		return err
	}
	rawTx.RawHex = emptyTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	keySigs := make([]*openwallet.KeySignature, 0)

	addr, err := wrapper.GetAddress(from)
	if err != nil {
		return err
	}
	signature := openwallet.KeySignature{
		EccType: decoder.wm.Config.CurveType,
		Nonce:   "0x" + strconv.FormatUint(nonce, 16),
		Address: addr,
		Message: message,
	}

	keySigs = append(keySigs, &signature)

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs

	rawTx.FeeRate = fee.String()

	rawTx.IsBuilt = true

	return nil
}

func (decoder *TransactionDecoder) SignXbtRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	key, err := wrapper.HDKey()
	if err != nil {
		return nil
	}

	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]

	if keySignatures != nil {
		for _, keySignature := range keySignatures {

			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
			keyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				return err
			}

			//签名交易
			///////交易单哈希签名
			signature, err := xbtTransaction.SignTransaction(keySignature.Message, keyBytes)
			if err != nil {
				return fmt.Errorf("transaction hash sign failed, unexpected error: %v", err)
			}
			keySignature.Signature = hex.EncodeToString(signature)
		}
	}

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	return nil
}

func (decoder *TransactionDecoder) VerifyXBTRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	var (
		emptyTrans = rawTx.RawHex
		signature  = ""
		pub = ""
	)
	//
	for accountID, keySignatures := range rawTx.Signatures {
		log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature = keySignature.Signature
			pub = keySignature.Address.PublicKey

			log.Debug("Signature:", keySignature.Signature)
			log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	pubkey, err := hex.DecodeString(pub)
	if err!=nil {
		return errors.New("wrong public key")
	}

	signedTrans, pass := xbtTransaction.VerifyAndCombineTransaction( emptyTrans, signature, pubkey)

	if pass {
		log.Debug("transaction verify passed")
		rawTx.IsCompleted = true
		rawTx.RawHex = signedTrans
	} else {
		log.Debug("transaction verify failed")
		rawTx.IsCompleted = false
	}

	return nil
}

func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	return decoder.wm.Config.FixedFee, "TX", nil
}

//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	if sumRawTx.Coin.IsContract {
		return nil, nil
	} else {
		return decoder.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
	}
}

func (decoder *TransactionDecoder) CreateSimpleSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {

	var (
		rawTxArray      = make([]*openwallet.RawTransaction, 0)
		accountID       = sumRawTx.Account.AccountID
		zeroDec = decimal.NewFromInt(0)
	)
	minTransfer, err := decimal.NewFromString(sumRawTx.MinTransfer)
	if err!=nil{
		return nil, errors.New("wrong minTransfer : "+sumRawTx.MinTransfer)
	}
	retainedBalance, err := decimal.NewFromString(sumRawTx.RetainedBalance)
	if err!=nil{
		return nil, errors.New("wrong retainedBalance : "+sumRawTx.RetainedBalance)
	}

	if minTransfer.Cmp(retainedBalance) < 0 {
		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
	}

	//获取wallet
	addresses, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit,
		"AccountID", sumRawTx.Account.AccountID)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return nil, err
	}

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalanceDec, err := decimal.NewFromString( addrBalance.Balance )
		if err!=nil {
			decoder.wm.Log.Error("wrong addr balance : ", addrBalance.Balance, ", address : ", addrBalance.Address)
			continue
		}

		if addrBalanceDec.Cmp(minTransfer) < 0 {
			continue
		}
		//计算汇总数量 = 余额 - 保留余额
		sumAmount := addrBalanceDec.Sub( retainedBalance )

		fee, err := decoder.GetTxFee( sumRawTx.FeeRate, &sumAmount )
		if err!=nil {
			continue
		}

		//减去手续费
		sumAmount = sumAmount.Sub( fee )
		if sumAmount.Cmp( zeroDec )<=0 {
			continue
		}

		decoder.wm.Log.Debug(
			"address : ", addrBalance.Address,
			" balance : ", addrBalance.Balance,
			" fees : ", fee,
			" sumAmount : ", sumAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:     sumRawTx.Coin,
			Account:  sumRawTx.Account,
			ExtParam: sumRawTx.ExtParam,
			To: map[string]string{
				sumRawTx.SummaryAddress: sumAmount.String(),
			},
			Required: 1,
			FeeRate:  sumRawTx.FeeRate,
		}

		createErr := decoder.createRawTransaction(
			wrapper,
			rawTx,
			addrBalance)
		if createErr != nil {
			return nil, createErr
		}

		//创建成功，添加到队列
		rawTxArray = append(rawTxArray, rawTx)
	}
	return rawTxArray, nil
}

func (decoder *TransactionDecoder) createRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction, addrBalance *openwallet.Balance) error {

	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}

	amount, err := decimal.NewFromString( amountStr )
	if err!=nil {
		return errors.New( "wrong amount : " + amountStr )
	}

	fee, err := decoder.GetTxFee( rawTx.FeeRate, &amount )

	from := addrBalance.Address
	fromAddr, err := wrapper.GetAddress(from)
	if err != nil {
		return err
	}

	rawTx.TxFrom = []string{from}
	rawTx.TxTo = []string{to}
	rawTx.TxAmount = amountStr
	rawTx.Fees = fee.String()
	rawTx.FeeRate = fee.String()

	emptyTrans, hash, err := decoder.CreateEmptyRawTransactionAndMessage(to, &amount, &fee)

	rawTx.RawHex = emptyTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	keySigs := make([]*openwallet.KeySignature, 0)

	signature := openwallet.KeySignature{
		EccType: decoder.wm.Config.CurveType,
		Address: fromAddr,
		Message: hash,
	}

	keySigs = append(keySigs, &signature)

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs

	rawTx.FeeRate = fee.String()

	rawTx.IsBuilt = true

	return nil
}

//CreateSummaryRawTransactionWithError 创建汇总交易，返回能原始交易单数组（包含带错误的原始交易单）
func (decoder *TransactionDecoder) CreateSummaryRawTransactionWithError(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {
	raTxWithErr := make([]*openwallet.RawTransactionWithError, 0)
	rawTxs, err := decoder.CreateSummaryRawTransaction(wrapper, sumRawTx)
	if err != nil {
		return nil, err
	}
	for _, tx := range rawTxs {
		raTxWithErr = append(raTxWithErr, &openwallet.RawTransactionWithError{
			RawTx: tx,
			Error: nil,
		})
	}
	return raTxWithErr, nil
}

func (decoder *TransactionDecoder) CreateEmptyRawTransactionAndMessage(to string, amount, fee *decimal.Decimal) (string, string, error) {
	txStruct, hash, err := xbtTransaction.GetTxStruct(to, amount, fee)
	if err != nil {
		return "", "", err
	}

	//messageBytes := []byte( hex.EncodeToString(hash) )
	//messageHash := owcrypt.Hash(messageBytes, 0, owcrypt.HASH_ALG_SHA3_256)

	return txStruct.ToJSONString(), hex.EncodeToString(hash), nil
}

//通过转账金额，计算手续费，千分之2，最低0.1
func (decoder *TransactionDecoder) GetTxFee(feeRate string, amount *decimal.Decimal) (decimal.Decimal, error) {
	zeroFee := decimal.NewFromInt32(0)

	//最低0.1
	minFee, err := decimal.NewFromString( decoder.wm.Config.FixedFee )
	if err != nil{
		return zeroFee, errors.New("wrong minFee : " + decoder.wm.Config.FixedFee )
	}
	netFeeRate, _ := decimal.NewFromString("0.002")

	//计算出来的手续费
	fee := amount.Mul( netFeeRate ).Round( decoder.wm.Config.Decimal )

	if len(feeRate) > 0 {
		result, err := decimal.NewFromString( feeRate )
		if err != nil{
			return zeroFee, errors.New("wrong feeRate : " + feeRate )
		}
		fee = result
	}

	if fee.Cmp( minFee )<0 {	//低于0.1，就按照0.1收
		fee = minFee
	}
	return fee, nil
}