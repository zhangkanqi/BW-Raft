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
 * 生成RSA公钥和私钥并保存在对应的目录文件下
 * 参数bits: 指定生成的秘钥的长度, 单位: bit
 */
func (r *Rsa) RsaGenKey(bits int, privatePath, pubulicPath string) error {
	// 1. 生成私钥文件
	// GenerateKey函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	// 2. MarshalPKCS1PrivateKey将rsa私钥序列化为ASN.1 PKCS#1 DER编码
	derPrivateStream := x509.MarshalPKCS1PrivateKey(privateKey)

	// 3. Block代表PEM编码的结构, 对其进行设置
	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derPrivateStream,
	}

	// 4. 创建文件
	privateFile, err := os.Create(privatePath)
	defer privateFile.Close()
	if err != nil {
		return err
	}

	// 5. 使用pem编码, 并将数据写入文件中
	err = pem.Encode(privateFile, &block)
	if err != nil {
		return err
	}

	// 1. 生成公钥文件
	publicKey := privateKey.PublicKey
	derPublicStream, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}

	block = pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: derPublicStream,
	}

	publicFile, err := os.Create(pubulicPath)
	defer publicFile.Close()
	if err != nil {
		return err
	}

	// 2. 编码公钥, 写入文件
	err = pem.Encode(publicFile, &block)
	if err != nil {
		panic(err)
		return err
	}
	return nil
}


func (r *Rsa) testGenRSA()  {
	r.RsaGenKey(2048, "privateKey.pem","publicKey.pem")
}


func main() {
	r := &Rsa{}
	r.testGenRSA()
	fmt.Println("成功生成公钥和私钥")
}