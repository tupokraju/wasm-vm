package secp256k1

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEthereumSig(t *testing.T) {
	msg, _ := hex.DecodeString("ce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008")
	r, _ := hex.DecodeString("90f27b8b488db00b00606796d2987f6a5f59ae62ea05effe84fef5b8b0e54998")
	s, _ := hex.DecodeString("4a691139ad57a3f0b906637673aa2f63d1f55cb1a69199d4009eea23ceaddc93")
	key, _ := hex.DecodeString("04e32df42865e97135acfb65f3bae71bdc86f4d49150ad6a440b6f15878109880a0a2b2667f7e725ceea70c673093bf67663e0312623c8e091b13cf2c0f11ef652")

	verifier := NewSecp256k1()
	sig := verifier.EncodeSecp256k1DERSignature(r, s)
	err := verifier.VerifySecp256k1(key, msg, sig, byte(ECDSAPlainMsg))

	assert.Nil(t, err)
}

func TestBitcoinSig(t *testing.T) {
	pubKey, _ := hex.DecodeString("04d2e670a19c6d753d1a6d8b20bd045df8a08fb162cf508956c31268c6d81ffdabab65528eefbb8057aa85d597258a3fbd481a24633bc9b47a9aa045c91371de52")
	msg, _ := hex.DecodeString("01020304")
	r, _ := hex.DecodeString("fef45d2892953aa5bbcdb057b5e98b208f1617a7498af7eb765574e29b5d9c2c")
	s, _ := hex.DecodeString("d47563f52aac6b04b55de236b7c515eb9311757db01e02cff079c3ca6efb063f")

	verifier := NewSecp256k1()
	sig := verifier.EncodeSecp256k1DERSignature(r, s)
	err := verifier.VerifySecp256k1(pubKey, msg, sig, byte(ECDSADoubleSha256))

	assert.Nil(t, err)
}
