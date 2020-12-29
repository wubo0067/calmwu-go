package main

import "math/big"
import "github.com/monnand/dhkx"
import "crypto/aes"
import "crypto/cipher"

// https://www.zhihu.com/topic/20030277/hot

const (
	KEY_BYTES_SIZE = 32
)

var (
	p_value                    = "8b79f180cbd3f282de92e8b8f2d092674ffda61f01ed961f8ef04a1b7a3709ff748c2abf6226cf0c4538e48838193da456e92ee530ef7aa703e741585e475b26cd64fa97819181cef27de2449cd385c49c9b030f89873b5b7eaf063a788f00db3cb670c73846bc4f76af062d672bde8f29806b81548411ab48b99aebfd9c2d09"
	g_value                    = "029843c81d0ea285c41a49b1a2f8e11a56a4b39040dfbc5ec040150c16f72f874152f9c44c659d86f7717b2425b62597e9a453b13da327a31cde2cced600915252d30262d1e54f4f864ace0e484f98abdbb37ebb0ba4106af5f0935b744677fa2f7f3826dcef3a1586956105ebea805d871f34c46c25bc30fc66b2db26cb0a93"
	g_dh_key     *dhkx.DHKey   = nil
	g_dh_group   *dhkx.DHGroup = nil
	g_crypto_key               = make([]byte, KEY_BYTES_SIZE, KEY_BYTES_SIZE)
	// 这个是接口
	g_cipher_block cipher.Block = nil
)

// 算出来的ka发送给对方
func get_dhka() *string {
	if g_dh_group == nil {
		p, _ := new(big.Int).SetString(p_value, 16)
		g, _ := new(big.Int).SetString(g_value, 16)
		// g_log.Debug("p[%s]", p.String())
		// g_log.Debug("g[%s]", g.String())
		g_dh_group = dhkx.CreateGroup(p, g)
	}

	if g_dh_key == nil {
		g_dh_key, _ = g_dh_group.GeneratePrivateKey(nil)
	}

	ka := g_dh_key.String()
	g_log.Debug("ka[%s]", ka)
	return &ka
}

func generate_key(kb *string) int {
	g_log.Debug("kb[%s]", *kb)
	//s := []byte(*kb)
	peer_pubkey := dhkx.NewPublicSKey(kb, 16)
	// 计算密钥
	key, err := g_dh_group.ComputeKey(peer_pubkey, g_dh_key)
	if err != nil {
		g_log.Error("generate key failed, reason[%s]", err.Error())
		return -1
	}

	key_b := key.Bytes()
	copy(g_crypto_key, key_b)

	g_log.Debug("g_crypto_key size:%d, content[%x]", len(g_crypto_key), g_crypto_key)
	// for index, _ := range g_crypto_key {
	// 	g_log.Debug("g_crypto_key[%x]", index, g_crypto_key[index])
	// }

	g_cipher_block, err = aes.NewCipher(g_crypto_key)
	if err != nil {
		g_log.Error("new cipher failed! reason[%s]", err.Error())
		return -1
	}
	return 0
}

// 传入明文，返回密文
func encrypt_plaintext(plaintext []byte) ([]byte, int) {
	if g_cipher_block == nil {
		g_log.Error("cipher block object uninitialized!")
		return nil, -1
	}

	plaintext_len := len(plaintext)
	plaintext_buff := plaintext

	// 明文数据必须是blocksize的倍数，aes.blocksize=16
	if plaintext_len%aes.BlockSize != 0 {
		// 重新生成
		plaintext_len = (((plaintext_len) + ((aes.BlockSize) - 1)) & ^((aes.BlockSize) - 1))
		plaintext_buff = make([]byte, plaintext_len)
		copy(plaintext_buff, plaintext)
	}

	// 生成加密缓冲区
	ciphertext_len := aes.BlockSize + plaintext_len
	ciphertext_buff := make([]byte, ciphertext_len)

	g_log.Debug("plaintext_len[%d] ciphertext_len[%d]", plaintext_len,
		ciphertext_len)

	iv := ciphertext_buff[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(g_cipher_block, iv)
	mode.CryptBlocks(ciphertext_buff[aes.BlockSize:], plaintext_buff)
	return ciphertext_buff, len(ciphertext_buff)
}

// in-place模式
func decrypt_ciphertext(ciphertext []byte) ([]byte, int) {
	if g_cipher_block == nil {
		g_log.Error("cipher block object uninitialized!")
		return nil, -1
	}

	ciphertext_len := len(ciphertext)

	if ciphertext_len < aes.BlockSize {
		g_log.Error("ciphertext too short! ciphertext_len[%d]", ciphertext_len)
		return nil, -1
	}

	iv := ciphertext[:aes.BlockSize]
	//g_log.Debug("ciphertext len[%d]", len(ciphertext))
	ciphertext = ciphertext[aes.BlockSize:]
	//g_log.Debug("ciphertext len[%d]", len(ciphertext))
	mode := cipher.NewCBCDecrypter(g_cipher_block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)
	//g_log.Debug("ciphertext len[%d]", len(ciphertext))
	return ciphertext, 0
}

func test_crypto() {
	plaintext := []byte(`exampleplaintextexampleplaintextexampleplaintextexampleplaintext1234`)

	g_log.Debug("plaintext len[%d]", len(plaintext))

	ciphertext, _ := encrypt_plaintext(plaintext)
	decrypt_ciphertext(ciphertext)
	g_log.Debug("ciphertext[%s] len[%d]", ciphertext, len(ciphertext))
}
