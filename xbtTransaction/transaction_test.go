package xbtTransaction

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/go-owcrypt"
	"github.com/shopspring/decimal"
	"math/rand"
	"strconv"
	"testing"
)

func Test_Transaction(t *testing.T)  {
	prikey, _ := hex.DecodeString("xxx")
	pubkey, _ := hex.DecodeString("xxx")

	to := "xBa3F47458Fe70704ebD5061809fE2d390F6342D17"
	amount, _ := decimal.NewFromString("0.01")
	fee, _ := decimal.NewFromString("0.1")

	txStruct, hash, err := GetTxStruct(to, &amount, &fee)
	if err != nil {
		return
	}
	emptyTrans := txStruct.ToJSONString()

	signature, _, retCode := owcrypt.Signature(prikey, nil, hash, owcrypt.ECC_CURVE_SECP256K1)
	if retCode!=owcrypt.SUCCESS {
		fmt.Println("error")
		return
	}

	signedTrans, _ := VerifyAndCombineTransaction(emptyTrans, hex.EncodeToString(signature), pubkey)

	ts, err := NewTxStructFromJSON(signedTrans)

	tx := `{"tx":{`
	tx = tx + "\"hash\":\"" + ts.Hash + "\","
	tx = tx + "\"to\":\"" + ts.To + "\","
	tx = tx + "\"amount\":" + ts.Amount.String() + `,`
	tx = tx + "\"fee\":" + ts.Fee.String() + `,`
	tx = tx + "\"nonce\":" + strconv.FormatUint(ts.Nonce, 10) + `,`
	tx = tx + "\"time\":" + strconv.FormatUint(ts.Time, 10) + `,`
	tx = tx + "\"sig\":\"" + ts.Sig + "\""
	tx = tx + `}}`

	fmt.Println("curl -H 'Content-Type: application/json' -d'", tx, "' https://api.xbt.wang/transaction/emit")
}

func Test_json(t *testing.T)  {
	prikey, _ := hex.DecodeString("xxx")

	amount, _ := decimal.NewFromString("0.01234")
	fee, _ := decimal.NewFromString("0.1")
	nonce := uint64( 583532949 )
	txTime := uint64( 1618532597129 )

	//nonce := uint64( rand.Int63n(999999999) )
	//txTime := uint64( time.Now().Unix() )

	ts := TxStruct{
		To     :     "xB52c55E62d708CdE25Cec9B576F5bFDEcFB5C328B",
		Amount :     &amount,
		Nonce  :     nonce,
		Fee    :     &fee,
		Time   :     txTime,
	}

	txString := `[`
	txString = txString + "\"" + ts.To + "\","
	txString = txString + ts.Amount.String() + `,`
	txString = txString + ts.Fee.String() + `,`
	txString = txString + strconv.FormatUint(ts.Nonce, 10) + `,`
	txString = txString + strconv.FormatUint(ts.Time, 10)
	txString = txString + `]`

	fmt.Println( "txString : ", txString )

	txStringBytes := []byte( txString )
	message := owcrypt.Hash(txStringBytes, 0, owcrypt.HASH_ALG_SHA3_256)
	fmt.Println( "message : ", hex.EncodeToString(message) )

	messageBytes := []byte( hex.EncodeToString(message) )
	messageHash := owcrypt.Hash(messageBytes, 0, owcrypt.HASH_ALG_SHA3_256)
	fmt.Println( "messageHash : ", hex.EncodeToString(messageHash) )

	sig, _, retCode := owcrypt.Signature(prikey, nil, messageHash, owcrypt.ECC_CURVE_SECP256K1)
	if retCode!=owcrypt.SUCCESS {
		fmt.Println("error")
		return
	}
	sig = serilizeS( sig )
	fmt.Println("secp256k1 : ", hex.EncodeToString(sig) )
	
	sigPub := SignaturePubkey{
		Signature: sig,
		Pubkey:    nil,
	}
	signature := sigPub.EncodeSignatureToScript()
	fmt.Println("sig : ", hex.EncodeToString(signature) )

	pub, err := owcrypt.GenPubkey(prikey, owcrypt.ECC_CURVE_SECP256K1)
	if err != owcrypt.SUCCESS {
		return
	}
	result := hex.EncodeToString(signature) + "@04" + hex.EncodeToString(pub)
	fmt.Println("result : ", result )

	ts.Hash = hex.EncodeToString(message)
	ts.Sig = result

	tx := `{"tx":{`
	tx = tx + "\"hash\":\"" + ts.Hash + "\","
	tx = tx + "\"to\":\"" + ts.To + "\","
	tx = tx + "\"amount\":" + ts.Amount.String() + `,`
	tx = tx + "\"fee\":" + ts.Fee.String() + `,`
	tx = tx + "\"nonce\":" + strconv.FormatUint(ts.Nonce, 10) + `,`
	tx = tx + "\"time\":" + strconv.FormatUint(ts.Time, 10) + `,`
	tx = tx + "\"sig\":\"" + ts.Sig + "\""
	tx = tx + `}}`

	fmt.Println("curl -H 'Content-Type: application/json' -d'", tx, "' http://127.0.0.1:3000/transaction/emit")
}

func Test_Random(t *testing.T){
	nonce := uint64( rand.Int63n(2147483647) )
	for i := 0; i < 20; i++{
		nonce = uint64( rand.Int63n(2147483647) )
		fmt.Println( nonce )
	}
}