package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
)

func AttestHandler(w http.ResponseWriter, r *http.Request) {
	httpClient := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/run/container_launcher/teeserver.sock")
			},
		},
	}
	nonceStr := r.URL.Query().Get("nonce")
	if nonceStr == "" {
		http.Error(w, "missing 'nonce' query parameter", http.StatusBadRequest)
		return
	}

	decodedNonce, err := base64.URLEncoding.DecodeString(nonceStr)
	if err != nil {
		http.Error(w, "invalid 'nonce' format: not a base64 string", http.StatusBadRequest)
		return
	}

	if len(decodedNonce) != 64 {
		http.Error(w, fmt.Sprintf("invalid 'nonce' length: expected 64 bytes, got %d", len(decodedNonce)), http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("http://localhost/v1/attest?nonce=%s", nonceStr)

	resp, err := httpClient.Get(url)
	if err != nil {
		http.Error(w, "faile to get attestation report", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	tokenbytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "faile to get attestation report", http.StatusBadRequest)
		return
	}

	w.Write(tokenbytes)
}
