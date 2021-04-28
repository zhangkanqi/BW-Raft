package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type Rsa struct {

}


/*
 * RSA公钥加密
 */
func (r *Rsa) RSAEncrypt(src []byte, filename string) ([]byte, error)  {
	// 根据文件名读出文件内容
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)

	// 从数据中找出pem格式的块
	block, _ := pem.Decode(buf)
	if block == nil {
		return nil, err
	}

	// 解析一个der编码的公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	publicKey := pubInterface.(*rsa.PublicKey)

	// 公钥加密
	result, _ := rsa.EncryptPKCS1v15(rand.Reader, publicKey, src)
	return result, nil

}

/*
 * RSA私钥解密
 */
func (r *Rsa) RSADecrypt(src []byte, filename string) ([]byte, error) {
	// 根据文件名读出内容
	file, err := os.Open(filename)
	if err != nil {
		return nil,err
	}
	defer file.Close()

	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)

	// 从数据中解析出pem块
	block, _ := pem.Decode(buf)
	if block == nil {
		return nil,err
	}

	// 解析出一个der编码的私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	// 私钥解密
	result, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, src)
	if err != nil {
		return nil,err
	}
	return result,nil
}

func (r *Rsa) testRSA()  {
	msg := "This is a girl."
	cipherText, _:= r.RSAEncrypt([]byte(msg), "publicKey.pem")
	fmt.Println("加密后的字符串：", string(cipherText))

	plainText, _:= r.RSADecrypt(cipherText, "privateKey.pem")
	fmt.Println("解密后的字符串：", string(plainText))
}

func main() {
	r := &Rsa{}
	r.testRSA()
}