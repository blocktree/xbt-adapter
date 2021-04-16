package xbtTransaction

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blocktree/go-owcrypt"
)

func SignTransaction(msgStr string, prikey []byte) ([]byte, error) {
	fmt.Println("msg :", msgStr, ", prikey : ", hex.EncodeToString(prikey) )
	msg, err := hex.DecodeString(msgStr)
	if err != nil || len(msg) == 0 {
		return nil, errors.New("invalid message to sign")
	}

	if prikey == nil || len(prikey) != 32 {
		return nil, errors.New("invalid private key")
	}

	//b2sum := blake2b.Sum256(msg)
	signature, _, retCode := owcrypt.Signature(prikey, nil, msg, owcrypt.ECC_CURVE_SECP256K1)
	if retCode != owcrypt.SUCCESS {
		return nil, errors.New("sign failed")
	}
	//signature = append(signature, v)

	return signature, nil
}

func VerifyAndCombineTransaction(emptyTrans, signature string, pubkey []byte) (string, bool) {
	ts, err := NewTxStructFromJSON(emptyTrans)
	if err != nil {
		return "", false
	}

	sig, err := hex.DecodeString(signature)
	if err != nil {
		return "", false
	}
	sig = serilizeS( sig )

	if len(pubkey) != 32 {
		//公钥hash处理
		pubkey = owcrypt.PointDecompress(pubkey, owcrypt.ECC_CURVE_SECP256K1)
	}

	sigPub := SignaturePubkey{
		Signature: sig,
		Pubkey:    pubkey,
	}
	derSig := sigPub.EncodeSignatureToScript()

	ts.Sig = hex.EncodeToString(derSig) + "@" + hex.EncodeToString(pubkey)

	return ts.ToJSONString(), true
}