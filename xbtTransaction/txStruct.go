package xbtTransaction

import (
	"encoding/hex"
	"encoding/json"
	"github.com/blocktree/go-owcrypt"
	"github.com/shopspring/decimal"
	"math/rand"
	"strconv"
	"time"
)

type TxStruct struct {
	Hash string `json:"hash"`
	To string `json:"to"`
	Amount *decimal.Decimal `json:"amount"`
	Fee *decimal.Decimal `json:"fee"`
	Nonce uint64 `json:"nonce"`
	Time uint64 `json:"time"`
	Sig string `json:"sig"`
}

func GetTxStruct(to string, amount, fee *decimal.Decimal) (TxStruct, []byte, error){
	nonce := uint64( rand.Int63n(2147483647) )
	txTime := uint64( time.Now().UnixNano() / 1e6 )

	ts := TxStruct{
		To     :    to,
		Amount :     amount,
		Nonce  :     nonce,
		Fee    :     fee,
		Time   :     txTime,
	}

	txString := `[`
	txString = txString + "\"" + ts.To + "\","
	txString = txString + ts.Amount.String() + `,`
	txString = txString + ts.Fee.String() + `,`
	txString = txString + strconv.FormatUint(ts.Nonce, 10) + `,`
	txString = txString + strconv.FormatUint(ts.Time, 10)
	txString = txString + `]`

	txStringBytes := []byte( txString )
	message := owcrypt.Hash(txStringBytes, 0, owcrypt.HASH_ALG_SHA3_256)

	messageBytes := []byte( hex.EncodeToString(message) )
	messageHash := owcrypt.Hash(messageBytes, 0, owcrypt.HASH_ALG_SHA3_256)

	ts.Hash = hex.EncodeToString(message)

	return ts, messageHash, nil
}

func (tx TxStruct) ToJSONString() string {
	j, _ := json.Marshal(tx)

	return string(j)
}

func NewTxStructFromJSON(j string) (*TxStruct, error) {

	ts := TxStruct{}

	err := json.Unmarshal([]byte(j), &ts)

	if err != nil {
		return nil, err
	}

	return &ts, nil
}