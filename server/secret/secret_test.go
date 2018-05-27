package secret

import (
	"bytes"
	"io/ioutil"
	"log"
	"testing"
)

func TestSecret(*testing.T) {
	pubkey, err := ioutil.ReadFile("./id_rsa.pub") // emulate local got a remote pubkey.
	if err != nil {
		log.Fatalln(err)
	}
	if !bytes.Equal(Pubkey(), pubkey) { // local verified the remote poubkey.
		log.Fatalf("pubkey not match, pubkey: %x, expected: %x", Pubkey(), pubkey)
	}

	raw := "Hello RSA."
	pubPem := PubPem()                          // emulate remote got the pubPem after local verified succeed.
	cipher, err := Encrypt(pubPem, []byte(raw)) // remote encrypt data with the got pubPem.
	if err != nil {
		log.Fatalln(err)
	}
	data, err := Decrypt(cipher) // emulate the local got the encrypted data.
	if err != nil {
		log.Fatalln(err)
	}
	if string(data) != raw {
		log.Fatalln(err)
	}
}

func init() {
	if err := initKey("./", 1024); err != nil {
		log.Fatalln(err)
	}
}
