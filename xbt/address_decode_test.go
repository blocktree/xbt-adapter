package xbt

import (
	"encoding/hex"
	"testing"
)

//最终答案 : xB1CE3Ff24Bbe10dc457320D0BB3602d5C79F844a5
func TestAddressDecoder_AddressEncode(t *testing.T) {
	pub, _ := hex.DecodeString("0265ff85a638b555ad5f15359ef0d80688452bd4dae3a29ecdf53e74b76862a6f2")
	dec := NewAddressDecoderV2(tw)
	addr, _ := dec.AddressEncode(pub, false)
	t.Logf("addr: %s", addr)

	check := dec.AddressVerify(addr)
	t.Logf("check: %v \n", check)
}


func TestAddressDecoder_VerifyAddress(t *testing.T) {
	dec := NewAddressDecoderV2(tw)
	check := false

	//======== secp256k1 ==============

	check = dec.AddressVerify("xB1CE3Ff24Bbe10dc457320D0BB3602d5C79F844a5")
	t.Logf("check: %v \n", check)

	check = dec.AddressVerify("xB1CE3Ff24Bbe10dc457320D0BB3602d5C79F844a4")
	t.Logf("check: %v \n", check)

	check = dec.AddressVerify("xB1CE3Ff24Bbe10dc457320D0BB3602d5C79F844a5123")
	t.Logf("check: %v \n", check)
}
