package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

const (
	ollamaAddr = "http://localhost"
)

func GenerateRawBase64URL(size int) (string, error) {
	// 1. 创建一个指定大小的字节切片 (byte slice)
	randomBytes := make([]byte, size)

	// 2. 使用加密学安全的随机数源填充切片
	// crypto/rand.Read 是生成安全随机数的标准方法
	_, err := rand.Read(randomBytes)
	if err != nil {
		// 如果生成随机数失败，返回一个包含具体错误的 error
		return "", fmt.Errorf("无法生成随机字节: %w", err)
	}

	// 3. 使用 RawURLEncoding 进行编码
	// RawURLEncoding 是无填充的 Base64URL 编码器
	encodedString := base64.URLEncoding.EncodeToString(randomBytes)

	return encodedString, nil
}

func main() {
	target, _ := url.Parse(ollamaAddr)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
	}
	nonce, err := GenerateRawBase64URL(64)
	if err != nil {
		panic(err)
	}
	log.Printf("nonce %s\n", nonce)

	Backend = NewModelBackend()
	ctx := context.Background()
	model := os.Getenv("MODEL")
	err = Backend.RunModel(ctx, model)
	if err != nil {
		panic(err)
	}
	select {
	case <-Backend.StartupDone.Done():

	case <-ctx.Done():
		panic(errors.New("ctx is calceled"))
	}
	log.Printf("model started %s", model)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/api/attest" {
			AttestHandler(w, r)
			return
		}
		proxy.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	log.Printf("Go proxy started on %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("http server quite: %v", err)
	}
}
