package media

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func hmacBytes(key []byte, value string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(value))
	return h.Sum(nil)
}

// Upload stores real image/video bytes in an S3-compatible MinIO bucket using AWS Signature V4.
// Set MINIO_ENDPOINT (localhost:9000), MINIO_ACCESS_KEY and MINIO_SECRET_KEY.
func Upload(file *multipart.FileHeader, kind string) (string, error) {
	if file == nil {
		return "", errors.New("missing media")
	}
	limit := int64(20 << 20)
	ext := ""
	contentType := ""
	if kind == "image" {
		limit = 8 << 20
		ext = ".png"
		contentType = "image/png"
	} else if kind == "video" {
		ext = ".mp4"
		contentType = "video/mp4"
	} else {
		return "", errors.New("unsupported media kind")
	}
	if file.Size <= 0 || file.Size > limit {
		return "", fmt.Errorf("%s must be between 1 byte and %d bytes", kind, limit)
	}
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return "", err
	}
	if int64(len(body)) > limit {
		return "", errors.New("media exceeds size limit")
	}
	if kind == "image" && !(len(body) >= 8 && bytes.Equal(body[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10})) {
		return "", errors.New("only PNG images allowed")
	}
	if kind == "video" && !(len(body) >= 12 && string(body[4:8]) == "ftyp") {
		return "", errors.New("only MP4 videos allowed")
	}
	random := make([]byte, 16)
	if _, err = rand.Read(random); err != nil {
		return "", err
	}
	name := kind + "_" + hex.EncodeToString(random) + ext
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}
	base, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return "", errors.New("invalid MinIO endpoint")
	}
	access := os.Getenv("MINIO_ACCESS_KEY")
	secret := os.Getenv("MINIO_SECRET_KEY")
	if access == "" || secret == "" {
		return "", errors.New("MinIO credentials not configured")
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "hypertension-media"
	}
	uri := "/" + path.Join(bucket, name)
	target := base.Scheme + "://" + base.Host + uri
	now := time.Now().UTC()
	date := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	payloadHash := hash(body)
	region := "us-east-1"
	canonicalHeaders := "content-type:" + contentType + "\nhost:" + base.Host + "\nx-amz-content-sha256:" + payloadHash + "\nx-amz-date:" + amzDate + "\n"
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"
	canonicalRequest := "PUT\n" + uri + "\n\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + payloadHash
	scope := date + "/" + region + "/s3/aws4_request"
	signingString := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hash([]byte(canonicalRequest))
	signingKey := hmacBytes(hmacBytes(hmacBytes(hmacBytes([]byte("AWS4"+secret), date), region), "s3"), "aws4_request")
	signature := hex.EncodeToString(hmacBytes(signingKey, signingString))
	req, err := http.NewRequest(http.MethodPut, target, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+access+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+signature)
	client := &http.Client{Timeout: 25 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return "", fmt.Errorf("MinIO upload status %d: %s", resp.StatusCode, string(message))
	}
	return name, nil
}
