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
	"fmt"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"math/big"
	"strconv"
	"time"
)

type Block struct {
	Hash          string        `json:"block"`         // actually block signature in XBT chain
	PrevBlockHash string        `json:"previousBlock"` // actually block signature in XBT chain
	Timestamp     uint64        `json:"timestamp"`
	Height        uint64        `json:"height"`
	Transactions  []Transaction `json:"transactions"`
}

type Transaction struct {
	TxID        string
	Fee         uint64
	TimeStamp   uint64
	From        string
	To          string
	Amount      uint64
	BlockHeight uint64
	BlockHash   string
	Status      string
	ToArr       []string //@required 格式："地址":"数量"
	ToDecArr    []string //@required 格式："地址":"数量(带小数)"
}

func GetTransactionInBlock(json *gjson.Result, decimal int32) []Transaction {
	blockHash := gjson.Get(json.Raw, "hash").String()
	blockHeight := gjson.Get(json.Raw, "height").Uint()
	transactions := make([]Transaction, 0)

	blockTime := gjson.Get(json.Raw, "time").Uint()

	for _, txItem := range gjson.Get(json.Raw, "tx").Array() {
		txid := gjson.Get(txItem.Raw, "hash").String()
		from := gjson.Get(txItem.Raw, "send_address").String()          //来源地址
		to := gjson.Get(txItem.Raw, "receive_address").String()            //目标地址
		amountStr := gjson.Get(txItem.Raw, "amount").String()      //金额
		feeStr := gjson.Get(txItem.Raw, "fee").String() 			//手续费

		amount := convertFromAmount(amountStr, decimal)
		fee := convertFromAmount(feeStr, decimal)
		transaction := Transaction{
			TxID:        txid,
			Fee:         fee,
			TimeStamp:   blockTime,
			From:        from,
			To:          to,
			Amount:      amount,
			BlockHeight: blockHeight,
			BlockHash:   blockHash,
			Status:      "1",
		}

		transactions = append(transactions, transaction)

	}

	return transactions
}

func NewBlock(json *gjson.Result, decimal int32) *Block {
	obj := &Block{}
	// 解析
	obj.Hash = gjson.Get(json.Raw, "hash").String()
	obj.PrevBlockHash = gjson.Get(json.Raw, "prev_hash").String()
	obj.Height = gjson.Get(json.Raw, "height").Uint()
	obj.Transactions = GetTransactionInBlock(json, decimal)

	if obj.Hash == "" {
		time.Sleep(5 * time.Second)
	}
	return obj
}

//BlockHeader 区块链头
func (b *Block) BlockHeader() *openwallet.BlockHeader {

	obj := openwallet.BlockHeader{}
	//解析json
	obj.Hash = b.Hash
	//obj.Confirmations = b.Confirmations
	obj.Previousblockhash = b.PrevBlockHash
	obj.Height = b.Height
	//obj.Symbol = Symbol

	return &obj
}

type AddrBalance struct {
	Address string
	Balance *big.Int
	Free    *big.Int
	Freeze  *big.Int
	Nonce   uint64
	index   int
	Actived bool
}

// 从最小单位的 amount 转为带小数点的表示
func convertToAmount(amount uint64, amountDecimal int32) string {
	amountStr := fmt.Sprintf("%d", amount)
	d, _ := decimal.NewFromString(amountStr)
	ten := math.BigPow(10, int64(amountDecimal) )
	w, _ := decimal.NewFromString(ten.String())

	d = d.Div(w)
	return d.String()
}

// amount 字符串转为最小单位的表示
func convertFromAmount(amountStr string, amountDecimal int32) uint64 {
	d, _ := decimal.NewFromString(amountStr)
	ten := math.BigPow(10, int64(amountDecimal) )
	w, _ := decimal.NewFromString(ten.String())
	d = d.Mul(w)
	r, _ := strconv.ParseInt(d.String(), 10, 64)
	return uint64(r)
}
