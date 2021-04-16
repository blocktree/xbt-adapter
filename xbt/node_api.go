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
	"errors"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/xbt-adapter/xbtTransaction"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"math/big"
	"strconv"
	"time"
)

type ClientInterface interface {
	Call(path string, request []interface{}) (*gjson.Result, error)
}

// A Client is a Elastos RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type Client struct {
	BaseURL     string
	AccessToken string
	Debug       bool
	client      *req.Req
	Symbol      string
	Decimal     int32
}

type Response struct {
	Code    int         `json:"code,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
	Id      string      `json:"id,omitempty"`
}

func NewClient(url string /*token string,*/, debug bool, symbol string, decimal int32) *Client {
	c := Client{
		BaseURL: url,
		//	AccessToken: token,
		Debug: debug,
	}

	log.Debug("BaseURL : ", url)

	api := req.New()

	c.client = api
	c.Symbol = symbol
	c.Decimal = decimal

	return &c
}

// 用get方法获取内容
func (c *Client) PostCall(path string, v map[string]interface{}) (*gjson.Result, error) {
	if c.Debug {
		log.Debug("Start Request API, url : ", path, ", body : ", v)
	}

	r, err := req.Post(c.BaseURL+path, req.BodyJSON(&v))

	if c.Debug {
		log.Std.Info("Request API Completed")
	}

	if c.Debug {
		log.Debugf("%+v\n", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())

	result := resp

	return &result, nil
}

// 用get方法获取内容
func (c *Client) PostStringCall(path string, v string) (*gjson.Result, error) {
	if c.Debug {
		log.Debug("Start Request API, url : ", path, ", body : ", v)
	}

	header := req.Header{
		"Content-Type":        "application/json",
	}

	r, err := req.Post(c.BaseURL+path, v, header)

	if c.Debug {
		log.Std.Info("Request API Completed")
	}

	if c.Debug {
		log.Debugf("%+v\n", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())

	result := resp

	return &result, nil
}

//从接口返回的json结果，提取data，code一定要等于200，才能返回，不是200，一律视为错误
func (c *Client) getDataInJson(json *gjson.Result)(*gjson.Result, error){
	code := gjson.Get(json.Raw, "code").Int()
	if code!=200 {
		return nil, errors.New("getBlockHeight return wrong code : " + strconv.FormatInt(code,10) )
	}

	data := gjson.Get(json.Raw, "data")
	return &data, nil
}

func (c *Client) getDataArrInJson(json *gjson.Result)(*gjson.Result, error){
	code := gjson.Get(json.Raw, "code").Int()
	if code!=200 {
		return nil, errors.New("getBlockHeight return wrong code : " + strconv.FormatInt(code,10) )
	}

	data := gjson.Get(json.Raw, "data")

	return &data, nil
}

// 获取当前最高区块
func (c *Client) getBlockHeight() (uint64, error) {
	body := map[string]interface{}{
	}

	resp, err := c.PostCall("/open/block/height", body)
	if err != nil {
		return 0, err
	}

	data, err := c.getDataInJson(resp)
	if err != nil {
		return 0, err
	}

	result, err := strconv.ParseUint(data.Raw, 10, 64)
	if err!=nil {
		return 0, nil
	}else{
		return result, nil
	}
}

// 获取地址余额
func (c *Client) getBalance(address string) (*AddrBalance, error) {
	body := map[string]interface{}{
		"address" : address,
	}

	resp, err := c.PostCall("/open/balance", body)
	if err != nil {
		return nil, err
	}

	data, err := c.getDataInJson(resp)
	if err != nil {
		return nil, err
	}

	balanceStr := gjson.Get(data.Raw, "balance").String()
	balance := convertFromAmount( balanceStr, c.Decimal)

	balanceBigInt, _ := big.NewInt(0).SetString(strconv.FormatUint(balance,10), 10)
	feeFrozen := big.NewInt(0)

	return &AddrBalance{Address: address, Balance: balanceBigInt, Freeze: feeFrozen, Free: balanceBigInt, Actived: true, Nonce: 0}, nil
}

func (c *Client) getBlockByHeight(height uint64) (*Block, error) {
	body := map[string]interface{}{
		"start" : height,
		"end" : height,
	}

	resp, err := c.PostCall("/open/block/range", body)
	if err != nil {
		return nil, err
	}

	data, err := c.getDataArrInJson(resp)
	if err != nil {
		return nil, err
	}

	dataArr := data.Array()

	if len(dataArr)>0 {
		return NewBlock(&(dataArr[0]), c.Decimal), nil
	}else{
		return nil, errors.New("block not found, height : "+strconv.FormatUint(height, 10) )
	}
}

func (c *Client) sendTransaction(ts *xbtTransaction.TxStruct) (string, error) {
	tx := ""
	tx = tx + `{`
	tx = tx + "\"tx\":{"
	tx = tx + "\"hash\":\"" + ts.Hash + "\""
	tx = tx + ",\"to\":\"" + ts.To + "\""
	tx = tx + ",\"amount\":" + ts.Amount.String()
	tx = tx + ",\"fee\":" + ts.Fee.String()
	tx = tx + ",\"nonce\":" + strconv.FormatUint(ts.Nonce, 10)
	tx = tx + ",\"time\":" + strconv.FormatUint(ts.Time, 10)
	tx = tx + ",\"sig\":\"" + ts.Sig + "\""
	tx = tx + `}`
	tx = tx + `}`

	log.Debug("sendTransaction tx : ", tx)

	resp, err := c.PostStringCall("/open/tx/send", tx)
	if err != nil {
		return "", err
	}

	time.Sleep(time.Duration(1) * time.Second)

	log.Debug("sendTransaction result : ", resp)

	code := gjson.Get(resp.Raw, "code").Int()
	if code==200 {
		return ts.Hash, nil
	}else{
		return "", errors.New( "send tx error : " + resp.String() )
	}
}
