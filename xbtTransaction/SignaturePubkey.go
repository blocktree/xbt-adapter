package xbtTransaction

import (
	"math/big"
)

type SignaturePubkey struct {
	Signature []byte
	Pubkey    []byte
}

func serilizeS(sig []byte) []byte {
	s := sig[32:]
	numS := new(big.Int).SetBytes(s)
	numHalfOrder := new(big.Int).SetBytes(HalfCurveOrder)
	if numS.Cmp(numHalfOrder) > 0 {
		numOrder := new(big.Int).SetBytes(CurveOrder)
		numS.Sub(numOrder, numS)

		s = numS.Bytes()
		if len(s) < 32 {
			for i := 0; i < 32-len(s); i++ {
				s = append([]byte{0x00}, s...)
			}
		}
		return append(sig[:32], s...)
	}
	return sig
}

// wrong : 3044022050c0154dd41fba1ad6b1f6eee1bfda30eda7b2c333ab0fb85574cb833a43558e022043a1ec03d5875346e0a11a1f1caf746f751723c6513006cd742196b29f30dd12

// right secp256k1 : 651e3db75f5ad609c77d5621a39c6861e5f555163a3b0e1a3eafaaecdd1b45bf4e0def9386065bb9267bad15cc3af5f33e829f57b2a8c416a0168c1b0b412ffb
// right secp256k1 der : 30440220651e3db75f5ad609c77d5621a39c6861e5f555163a3b0e1a3eafaaecdd1b45bf02204e0def9386065bb9267bad15cc3af5f33e829f57b2a8c416a0168c1b0b412ffb
// right : 3045022100abb6c566da04adf446471e39a2d43cc79e0667e8aae756ce672bda4ffd0d311a022050fd797bd0ffb42b6a333c7c89910e5bb4bb93758d0b5f277529dc56ffcada66
func (sp SignaturePubkey) EncodeSignatureToScript() []byte {
	r := sp.Signature[:32]
	s := sp.Signature[32:]
	if r[0]&0x80 == 0x80 {
		r = append([]byte{0x00}, r...)
	} else {
		for i := 0; i < 32; i++ {
			if r[0] == 0 && r[1]&0x80 != 0x80 {
				r = r[1:]
			} else {
				break
			}
		}
	}
	if s[0]&0x80 == 0x80 {
		s = append([]byte{0}, s...)
	} else {
		for i := 0; i < 32; i++ {
			if s[0] == 0 && s[1]&0x80 != 0x80 {
				s = s[1:]
			} else {
				break
			}
		}
	}

	r = append([]byte{byte(len(r))}, r...)
	r = append([]byte{0x02}, r...)
	s = append([]byte{byte(len(s))}, s...)
	s = append([]byte{0x02}, s...)

	rs := append(r, s...)

	rs = append([]byte{byte(len(rs))}, rs...)
	//rs = append(rs, sigType)
	rs = append([]byte{0x30}, rs...)
	//rs = append([]byte{byte(len(rs))}, rs...)

	return rs
}
