package backup

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/config"
)

type s3Client struct {
	cfg    config.BackupS3Config
	client *http.Client
}

type s3Object struct {
	Key          string
	Size         int64
	LastModified time.Time
}

func newS3Client(cfg config.BackupS3Config) *s3Client {
	return &s3Client{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *s3Client) list(ctx context.Context, maxKeys int) ([]s3Object, error) {
	var out []s3Object
	continuation := ""
	for {
		query := map[string]string{
			"list-type": "2",
			"prefix":    prefixWithSlash(c.cfg.Prefix),
		}
		if maxKeys > 0 {
			query["max-keys"] = strconv.Itoa(maxKeys)
		}
		if continuation != "" {
			query["continuation-token"] = continuation
		}
		resp, err := c.do(ctx, http.MethodGet, "", query, nil, 0, "")
		if err != nil {
			return nil, err
		}
		body, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		var parsed listBucketV2Result
		if err := xml.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("parse S3 list response: %w", err)
		}
		for _, item := range parsed.Contents {
			if !strings.HasSuffix(item.Key, ".tar.gz") {
				continue
			}
			out = append(out, s3Object{Key: item.Key, Size: item.Size, LastModified: item.LastModified.Time})
		}
		if maxKeys > 0 || !parsed.IsTruncated || strings.TrimSpace(parsed.NextContinuationToken) == "" {
			break
		}
		continuation = strings.TrimSpace(parsed.NextContinuationToken)
	}
	return out, nil
}

func (c *s3Client) putFile(ctx context.Context, key, path string) (s3Object, error) {
	file, err := os.Open(path)
	if err != nil {
		return s3Object{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return s3Object{}, err
	}
	hash, err := hashFile(file)
	if err != nil {
		return s3Object{}, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return s3Object{}, err
	}
	resp, err := c.do(ctx, http.MethodPut, key, nil, file, info.Size(), hash)
	if err != nil {
		return s3Object{}, err
	}
	io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return s3Object{Key: key, Size: info.Size(), LastModified: time.Now().UTC()}, nil
}

func (c *s3Client) getFile(ctx context.Context, key, path string) error {
	resp, err := c.do(ctx, http.MethodGet, key, nil, nil, 0, emptySHA256)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := os.MkdirAll(filepathDir(path), 0o755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func (c *s3Client) delete(ctx context.Context, key string) error {
	resp, err := c.do(ctx, http.MethodDelete, key, nil, nil, 0, emptySHA256)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

const emptySHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func hashFile(file *os.File) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (c *s3Client) do(ctx context.Context, method, key string, query map[string]string, body io.Reader, contentLength int64, payloadHash string) (*http.Response, error) {
	if payloadHash == "" {
		payloadHash = emptySHA256
	}
	u, canonicalURI, err := c.objectURL(key)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		values := url.Values{}
		for k, v := range query {
			values.Set(k, v)
		}
		u.RawQuery = values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if contentLength > 0 {
		req.ContentLength = contentLength
	}
	if method == http.MethodPut {
		req.Header.Set("Content-Type", "application/gzip")
	}
	if err := c.sign(req, canonicalURI, query, payloadHash); err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("S3 %s %s failed: status %d: %s", method, key, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return resp, nil
}

func (c *s3Client) objectURL(key string) (*url.URL, string, error) {
	endpoint := strings.TrimSpace(c.cfg.Endpoint)
	if endpoint == "" {
		return nil, "", fmt.Errorf("S3 endpoint is required")
	}
	if !strings.Contains(endpoint, "://") {
		scheme := "http"
		if c.cfg.UseSSL {
			scheme = "https"
		}
		endpoint = scheme + "://" + endpoint
	}
	base, err := url.Parse(endpoint)
	if err != nil {
		return nil, "", fmt.Errorf("parse S3 endpoint: %w", err)
	}
	bucket := strings.TrimSpace(c.cfg.Bucket)
	if bucket == "" {
		return nil, "", fmt.Errorf("S3 bucket is required")
	}
	key = strings.TrimLeft(key, "/")
	escapedKey := escapePath(key)
	if c.cfg.ForcePathStyle {
		canonical := "/" + awsEscape(bucket) + optionalSlash(escapedKey)
		base.Path = strings.TrimRight(base.Path, "/") + canonical
		return base, canonical, nil
	}
	base.Host = bucket + "." + base.Host
	canonical := optionalSlash(escapedKey)
	if canonical == "" {
		canonical = "/"
	}
	base.Path = strings.TrimRight(base.Path, "/") + canonical
	return base, canonical, nil
}

func (c *s3Client) sign(req *http.Request, canonicalURI string, rawQuery map[string]string, payloadHash string) error {
	now := time.Now().UTC()
	date := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	region := strings.TrimSpace(c.cfg.Region)
	if region == "" {
		region = "us-east-1"
	}
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if token := strings.TrimSpace(c.cfg.SessionToken); token != "" {
		req.Header.Set("X-Amz-Security-Token", token)
	}

	headers := map[string]string{
		"host":                 req.URL.Host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
	}
	if token := req.Header.Get("X-Amz-Security-Token"); token != "" {
		headers["x-amz-security-token"] = token
	}
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var canonicalHeaders strings.Builder
	for _, k := range keys {
		canonicalHeaders.WriteString(k)
		canonicalHeaders.WriteByte(':')
		canonicalHeaders.WriteString(strings.TrimSpace(headers[k]))
		canonicalHeaders.WriteByte('\n')
	}
	signedHeaders := strings.Join(keys, ";")
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQueryString(rawQuery),
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")

	scope := date + "/" + region + "/s3/aws4_request"
	requestHash := sha256Hex([]byte(canonicalRequest))
	stringToSign := strings.Join([]string{"AWS4-HMAC-SHA256", amzDate, scope, requestHash}, "\n")
	signingKey := deriveSigningKey(c.cfg.SecretKey, date, region, "s3")
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))
	credential := strings.TrimSpace(c.cfg.AccessKey) + "/" + scope
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+credential+", SignedHeaders="+signedHeaders+", Signature="+signature)
	return nil
}

func canonicalQueryString(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, awsEscape(k)+"="+awsEscape(values[k]))
	}
	return strings.Join(parts, "&")
}

func deriveSigningKey(secret, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func escapePath(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = awsEscape(part)
	}
	return strings.Join(parts, "/")
}

func awsEscape(s string) string {
	escaped := url.QueryEscape(s)
	escaped = strings.ReplaceAll(escaped, "+", "%20")
	escaped = strings.ReplaceAll(escaped, "%7E", "~")
	return escaped
}

func optionalSlash(path string) string {
	if path == "" {
		return ""
	}
	return "/" + path
}

func prefixWithSlash(prefix string) string {
	prefix = strings.Trim(strings.TrimSpace(prefix), "/")
	if prefix == "" {
		return ""
	}
	return prefix + "/"
}

type listBucketV2Result struct {
	XMLName               xml.Name        `xml:"ListBucketResult"`
	IsTruncated           bool            `xml:"IsTruncated"`
	NextContinuationToken string          `xml:"NextContinuationToken"`
	Contents              []listObjectXML `xml:"Contents"`
}

type listObjectXML struct {
	Key          string `xml:"Key"`
	LastModified s3Time `xml:"LastModified"`
	Size         int64  `xml:"Size"`
}

type s3Time struct {
	time.Time
}

func (t *s3Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		parsed, err = time.Parse(time.RFC1123, value)
	}
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

func filepathDir(path string) string {
	idx := strings.LastIndexAny(path, `/\\`)
	if idx < 0 {
		return "."
	}
	return path[:idx]
}
