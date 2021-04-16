package xbt

import (
	"encoding/hex"
	"errors"
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/openwallet/v2/openwallet"
	"strconv"
	"strings"
)

var (
	Default = AddressDecoderV2{}
)

//AddressDecoderV2
type AddressDecoderV2 struct {
	*openwallet.AddressDecoderV2Base
	wm *WalletManager
}

//NewAddressDecoder 地址解析器
func NewAddressDecoderV2(wm *WalletManager) *AddressDecoderV2 {
	decoder := AddressDecoderV2{}
	decoder.wm = wm
	return &decoder
}

//AddressDecode 地址解析
func (dec *AddressDecoderV2) AddressDecode(addr string, opts ...interface{}) ([]byte, error) {
	addr = strings.TrimPrefix(addr, "0x")
	decodeAddr, err := hex.DecodeString(addr)
	if err != nil {
		return nil, err
	}
	return decodeAddr, err
}

//AddressEncode 地址编码
func (dec *AddressDecoderV2) AddressEncode(publicKey []byte, opts ...interface{}) (string, error) {
	if len(publicKey) != 32 {
		//公钥hash处理
		publicKey = owcrypt.PointDecompress(publicKey, owcrypt.ECC_CURVE_SECP256K1)
	}
	//fmt.Println("publicKey : ", hex.EncodeToString(publicKey) )

	hash := owcrypt.Hash(publicKey, 0, owcrypt.HASH_ALG_SHA3_256)
	//fmt.Println("publicHash : ", hex.EncodeToString(hash) )

	publicHashCode := hash[:20]
	publicHashCodeStr := hex.EncodeToString(publicHashCode)
	//fmt.Println("publicHashCode : ", publicHashCodeStr )

	address := "xB" + publicHashCodeStr
	returnAddress, err := dec.CheckAddress(address)
	if err!= nil {
		return "", err
	}

	return returnAddress, nil
}

func (dec *AddressDecoderV2) CheckAddress(address string) (string, error){
	if len(address) != 42 {
		return "", errors.New("wrong address")
	}

	content := strings.ReplaceAll( strings.ToLower( address ), "xb", "")
	//fmt.Println("content : ", content )

	publicHashCodeBytes := []byte( content )
	hash := owcrypt.Hash(publicHashCodeBytes, 0, owcrypt.HASH_ALG_SHA3_256)
	//fmt.Println("hash : ", hex.EncodeToString(hash), ",hash len :", len(hex.EncodeToString(hash)), ",content len :", len(content) )

	hashCode := hex.EncodeToString(hash)[len( hex.EncodeToString(hash) ) - len(content) : ]
	//fmt.Println("hashCode : ", hashCode )

	returnAddress := "xB"

	for i := 0; i < len(hashCode); i++ {
		n, _ := strconv.ParseUint(hashCode[i:i+1], 16, 32)
		codeInt := int64(n)

		if codeInt >= 8 {
			returnAddress = returnAddress + strings.ToUpper( content[i:i+1] )
		}else{
			returnAddress = returnAddress + content[i:i+1]
		}
	}
	//fmt.Println("returnAddress : ", returnAddress )

	return returnAddress, nil
}

// AddressVerify 地址校验
func (dec *AddressDecoderV2) AddressVerify(address string, opts ...interface{}) bool {
	if len(address)==0 {
		return false
	}

	checkAddress, err := dec.CheckAddress(address)
	if err!=nil {
		return false
	}

	return checkAddress==address
}
