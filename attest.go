package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
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

	nonce, err := hex.DecodeString(nonceStr)
	if err != nil {
		http.Error(w, "missing 'nonce' is not hex encoding", http.StatusBadRequest)
		return
	}

	if len(nonce) != 64 {
		http.Error(w, fmt.Sprintf("invalid 'nonce' length: expected 64 bytes, got %d", len(nonce)), http.StatusBadRequest)
		return
	}

	nonceStr = base64.URLEncoding.EncodeToString(nonce)
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
