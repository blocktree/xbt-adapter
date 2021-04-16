package xbt

import (
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)
import "github.com/blocktree/openwallet/v2/log"

// A Client is a Elastos RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type XbtToolsClient struct {
	BaseURL     string
	AccessToken string
	Debug       bool
	client      *req.Req
	Symbol      string
	Decimal     int32
}

func NewXbtToolsClient(url string /*token string,*/, debug bool, symbol string, decimal int32) *XbtToolsClient {
	c := XbtToolsClient{
		BaseURL: url,
		Debug: debug,
	}

	log.Debug("Xbt Tools BaseURL : ", url)

	api := req.New()

	c.client = api
	c.Symbol = symbol
	c.Decimal = decimal

	return &c
}

func (c *XbtToolsClient) PostCall(path string, v map[string]interface{}) (*gjson.Result, error) {
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

// 获取当前最高区块
func (c *XbtToolsClient) getAddressByPublicKey(publicKey string) (string, error) {
	body := map[string]interface{}{
		"public" : publicKey,
	}

	resp, err := c.PostCall("/account/address/public", body)
	if err != nil {
		return "", err
	}

	address := gjson.Get(resp.Raw, "address").String()

	return address, nil
}