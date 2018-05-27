package secret

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/zhengxiaoyao0716/zmodule"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zmodule/event"
	"github.com/zhengxiaoyao0716/zmodule/file"
)

var privateKey *rsa.PrivateKey

// Encrypt .
func Encrypt(pubPem, data []byte) ([]byte, error) {
	block, _ := pem.Decode(append(pubPem, []byte("qwe")...))
	if block == nil {
		return nil, errors.New("invalid public key, PEM data not found")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), data)
}

// Decrypt .
func Decrypt(cipher []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipher)
}

var cache = struct {
	pubkey []byte
	pubPem []byte

	title  string
	finger string
}{}

// Pubkey .
func Pubkey() []byte { return cache.pubkey }

// PubPem .
func PubPem() []byte { return cache.pubPem }

// KeyTitle .
func KeyTitle() string {
	return cache.title
}

// Fingerprint .
func Fingerprint() string {
	return cache.finger
}

func init() {
	zmodule.Args["keystore"] = zmodule.Argument{
		Default: "",
		Usage:   "Directory where the keys are stored.",
	}
	zmodule.Args["key-generate-size"] = zmodule.Argument{
		Default: 2048,
		Usage:   "Size of the new secret key to generate.",
	}
	event.OnInit(func(event.Event) error {
		event.On(event.KeyStart, func(event.Event) error {
			err := initKey(config.GetString("keystore"), config.GetInt("key-generate-size"))
			if err == nil {
				return nil
			}
			// clean.
			privateKey = nil
			cache.pubkey = nil
			cache.pubPem = nil
			cache.title = ""
			cache.finger = ""
			return err
		})
		return nil
	})
}

func initKey(keystore string, keysize int) error {
	keystore = file.AbsPath(keystore)
	privatePath := file.AbsPath(keystore, "./id_rsa")
	publicPath := file.AbsPath(keystore, "./id_rsa.pub")

	_, privateErr := os.Lstat(privatePath)
	_, publicErr := os.Lstat(publicPath)

	var err error
	switch [2]bool{
		privateErr == nil || !os.IsNotExist(privateErr),
		publicErr == nil || !os.IsNotExist(publicErr),
	} {
	case [2]bool{false, false}: // both file not exist.
		err = geneKey(keystore, keysize, privatePath, publicPath)
	case [2]bool{false, true}: // missing private key.
		err = fmt.Errorf("missing private key, path: %s", privatePath)
	case [2]bool{true, false}: // missing public key.
		err = loadKey(privatePath, "")
	case [2]bool{true, true}: // both file exist
		err = loadKey(privatePath, publicPath)
	}

	if err != nil {
		return err
	}
	if privateKey == nil { // assert privateKey != nil
		return errors.New("unexpected error, initial privateKey falied")
	}

	if cache.pubkey == nil {
		pubkey, err := pubkey(privateKey)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(publicPath, pubkey, 0666); err != nil {
			return err
		}
		cache.pubkey = pubkey
	}
	if cache.pubPem == nil {
		pubPem, err := pubPem(privateKey)
		if err != nil {
			return err
		}
		cache.pubPem = pubPem
	}
	if cache.finger == "" {
		sshPublicKey, comment, _, _, err := ssh.ParseAuthorizedKey(cache.pubkey)
		if err != nil {
			return err
		}
		cache.finger = ssh.FingerprintLegacyMD5(sshPublicKey)
		cache.title = comment
	}

	return nil
}

func geneKey(keystore string, keysize int, privatePath, publicPath string) error {
	// prepare files.
	if err := os.MkdirAll(keystore, 0600); err != nil {
		return err
	}
	privateFile, err := os.Create(privatePath)
	if err != nil {
		return err
	}
	defer privateFile.Close()
	publicFile, err := os.Create(publicPath)
	if err != nil {
		return err
	}
	defer publicFile.Close()

	// generate and write RSA private key.
	privKey, err := rsa.GenerateKey(rand.Reader, keysize)
	if err != nil {
		return err
	}
	privPem := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)}
	if err := pem.Encode(privateFile, privPem); err != nil {
		return err
	}

	// generate and write ssh-rsa public key.
	pubkey, err := pubkey(privKey)
	if err != nil {
		return err
	}
	publicFile.Write(pubkey)
	cache.pubkey = pubkey

	privateKey = privKey
	return nil
}

func loadKey(privatePath, publicPath string) error {
	// load private pem.
	privPem, err := ioutil.ReadFile(privatePath)
	if err != nil {
		return err
	}

	// parse to private key.
	block, rest := pem.Decode(privPem)
	if block == nil {
		return errors.New("invalid private key, PEM data not found")
	}
	if len(rest) > 0 {
		return errors.New("invalid private key, extra data included")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	if err := privKey.Validate(); err != nil {
		return err
	}
	if publicPath == "" { // public key not exist.
		privateKey = privKey
		return nil
	}

	// validate exist public key.
	sshPubkey, err := ioutil.ReadFile(publicPath)
	if err != nil {
		return err
	}
	pubkey, err := pubkey(privKey)
	if err != nil {
		return err
	}
	if !pubkeyEqual(pubkey, sshPubkey) {
		return fmt.Errorf("existed public key not match, pubkey %x, expected: %x", sshPubkey, pubkey)
	}
	cache.pubkey = sshPubkey // notice that sshPubkey != pubkey

	privateKey = privKey
	return nil
}

// Pubkey return bytes of the ssh-rsa public key.
func pubkey(privateKey *rsa.PrivateKey) ([]byte, error) {
	sshPubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(sshPubKey), nil
}

// PubPem return the PEM data of RSA public key.
func pubPem(privateKey *rsa.PrivateKey) ([]byte, error) {
	pubPemBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	pubPem := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubPemBytes}
	return pem.EncodeToMemory(pubPem), nil
}

func pubkeyEqual(pubkey, compareTo []byte) bool { // compare ingore comments, rest, etc.
	pk, _, _, _, err := ssh.ParseAuthorizedKey(pubkey)
	if err != nil {
		return false
	}
	ck, _, _, _, err := ssh.ParseAuthorizedKey(compareTo)
	if err != nil {
		return false
	}
	return bytes.Equal(pk.Marshal(), ck.Marshal())
}
