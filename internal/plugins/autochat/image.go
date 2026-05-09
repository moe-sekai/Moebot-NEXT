package autochat

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"time"
)

// DownloadImageToBase64 下载远端图片，缩放到 1024 边长以内并以 JPEG 重新编码后返回
// `data:image/jpeg;base64,...` 的 data URI。下载/解码失败则降级为返回原始 base64。
func DownloadImageToBase64(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		mimeType := http.DetectContentType(data)
		if !strings.HasPrefix(mimeType, "image/") {
			mimeType = "image/jpeg"
		}
		return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	const maxDim = 1024
	finalImg := img
	if w > maxDim || h > maxDim {
		var nw, nh int
		if w > h {
			nw = maxDim
			nh = int(float64(h) * float64(maxDim) / float64(w))
		} else {
			nh = maxDim
			nw = int(float64(w) * float64(maxDim) / float64(h))
		}
		dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
		xRatio := float64(w) / float64(nw)
		yRatio := float64(h) / float64(nh)
		for y := 0; y < nh; y++ {
			for x := 0; x < nw; x++ {
				dst.Set(x, y, img.At(bounds.Min.X+int(float64(x)*xRatio), bounds.Min.Y+int(float64(y)*yRatio)))
			}
		}
		finalImg = dst
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, finalImg, &jpeg.Options{Quality: 60}); err != nil {
		return "", err
	}
	return fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes())), nil
}

// splitDataURI 将 `data:image/...;base64,XXX` 拆为 (mimeType, rawBase64)。
// 如果 b64 不带 data URI 前缀，返回 ("image/jpeg", b64)。
func splitDataURI(b64 string) (mimeType, raw string) {
	mimeType = "image/jpeg"
	raw = b64
	if !strings.HasPrefix(b64, "data:") {
		return
	}
	parts := strings.SplitN(b64, ",", 2)
	if len(parts) != 2 {
		return
	}
	header := parts[0]
	switch {
	case strings.Contains(header, "image/png"):
		mimeType = "image/png"
	case strings.Contains(header, "image/gif"):
		mimeType = "image/gif"
	case strings.Contains(header, "image/webp"):
		mimeType = "image/webp"
	}
	raw = parts[1]
	return
}
