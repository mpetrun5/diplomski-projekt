package crypto

type KeyType = string

const Secp256k1Type KeyType = "secp256k1"

type Keypair interface {
	Encode() []byte
	Decode([]byte) error
	Address() string
	PublicKey() string
}
