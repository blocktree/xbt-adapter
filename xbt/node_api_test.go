package xbt

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/xbt-adapter/xbtTransaction"
	"github.com/shopspring/decimal"
	"strings"
	"testing"
)

const (
	testNodeAPI = "https://api.xbt.wang" //xbt-api
	symbol = "XBT"
	currencyDecimal = 6
)

func PrintJsonLog(t *testing.T, logCont string){
	if strings.HasPrefix(logCont, "{") {
		var str bytes.Buffer
		_ = json.Indent(&str, []byte(logCont), "", "    ")
		t.Logf("Get Call Result return: \n\t%+v\n", str.String())
	}else{
		t.Logf("Get Call Result return: \n\t%+v\n", logCont)
	}
}

func TestPostCall(t *testing.T) {
	tw := NewClient(testNodeAPI, true, symbol, currencyDecimal)

	body := map[string]interface{}{
	}

	if r, err := tw.PostCall("/open/block/height", body); err != nil {
		t.Errorf("Post Call Result failed: %v\n", err)
	} else {
		PrintJsonLog(t, r.String())
	}
}

func Test_getBlockHeight(t *testing.T) {
	c := NewClient(testNodeAPI, true, symbol, currencyDecimal)

	r, err := c.getBlockHeight()

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("height:", r)
	}
}

func Test_getBalance(t *testing.T) {

	c := NewClient(testNodeAPI, true, symbol, currencyDecimal)

	address := "xB8d4fDbe476Db5F1961Db61fFB39786bF383f0ABE"

	r, err := c.getBalance(address)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}

func Test_sendTransaction(t *testing.T){

	prikey, _ := hex.DecodeString("xxx")
	pubkey, _ := hex.DecodeString("xxx")

	to := "xB52c55E62d708CdE25Cec9B576F5bFDEcFB5C328B"
	amount, _ := decimal.NewFromString("0.01234")
	fee, _ := decimal.NewFromString("0.1")

	txStruct, hash, err := xbtTransaction.GetTxStruct(to, &amount, &fee)
	if err != nil {
		return
	}
	emptyTrans := txStruct.ToJSONString()

	signature, _, retCode := owcrypt.Signature(prikey, nil, hash, owcrypt.ECC_CURVE_SECP256K1)
	if retCode!=owcrypt.SUCCESS {
		fmt.Println("error")
		return
	}

	signedTrans, _ := xbtTransaction.VerifyAndCombineTransaction(emptyTrans, hex.EncodeToString(signature), pubkey)

	ts, err := xbtTransaction.NewTxStructFromJSON(signedTrans)

	c := NewClient(testNodeAPI, true, symbol, currencyDecimal)
	r, err := c.sendTransaction( ts )
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}

func Test_getBlockByHeight(t *testing.T) {
	c := NewClient(testNodeAPI, true, symbol, currencyDecimal)
	r, err := c.getBlockByHeight(307757)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}