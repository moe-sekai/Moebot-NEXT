package gallery

import (
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"math"
	"math/bits"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

// calcHashes 计算图片的感知哈希 (hash1) 和灰度像素哈希 (hash2)。
func calcHashes(imgPath string) (hash1, hash2 string, err error) {
	f, err := os.Open(imgPath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return "", "", fmt.Errorf("decode image: %w", err)
	}

	// alpha blend 到白色背景
	bounds := src.Bounds()
	white := image.NewRGBA(bounds)
	draw.Draw(white, bounds, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(white, bounds, src, bounds.Min, draw.Over)

	// 缩放到 16x16 灰度 → hash2
	gray16 := imaging.Resize(white, 16, 16, imaging.Lanczos)
	grayImg := imaging.Grayscale(gray16)

	pixels16 := make([]byte, 256)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			r, _, _, _ := grayImg.At(grayImg.Bounds().Min.X+x, grayImg.Bounds().Min.Y+y).RGBA()
			pixels16[y*16+x] = byte(r >> 8)
		}
	}
	hash2 = hex.EncodeToString(pixels16)

	// 缩放到 8x8 → 64 位感知哈希 → hash1
	gray8 := imaging.Resize(grayImg, 8, 8, imaging.Lanczos)
	pixels8 := make([]byte, 64)
	var sum float64
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			r, _, _, _ := gray8.At(gray8.Bounds().Min.X+x, gray8.Bounds().Min.Y+y).RGBA()
			v := byte(r >> 8)
			pixels8[y*8+x] = v
			sum += float64(v)
		}
	}
	avg := sum / 64.0
	var hashBits uint64
	for i, p := range pixels8 {
		if float64(p) >= avg {
			hashBits |= 1 << (63 - i)
		}
	}
	hash1 = fmt.Sprintf("%016x", hashBits)

	return hash1, hash2, nil
}

// isSame 判断两张图片是否相似。
func isSame(h1a, h2a, h1b, h2b string, cfg *Config) bool {
	// 快速比较：hash1 汉明距离
	a, err1 := strconv.ParseUint(h1a, 16, 64)
	b, err2 := strconv.ParseUint(h1b, 16, 64)
	if err1 != nil || err2 != nil {
		return false
	}
	if bits.OnesCount64(a^b) > cfg.Hash1DifferenceThreshold {
		return false
	}

	// 精确比较：hash2 MAE
	ba, err1 := hex.DecodeString(h2a)
	bb, err2 := hex.DecodeString(h2b)
	if err1 != nil || err2 != nil || len(ba) != len(bb) {
		return false
	}
	var diff float64
	for i := range ba {
		diff += math.Abs(float64(ba[i]) - float64(bb[i]))
	}
	return int(diff) <= cfg.Hash2DifferenceThreshold
}

const thumbSize = 64

// ensureThumb 生成缩略图，返回缩略图路径。
func ensureThumb(imgPath string) (string, error) {
	name := filepath.Base(imgPath)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	thumbPath := filepath.Join(filepath.Dir(imgPath), base+"_thumb.jpg")

	if _, err := os.Stat(thumbPath); err == nil {
		return thumbPath, nil
	}

	f, err := os.Open(imgPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}

	thumb := imaging.Thumbnail(src, thumbSize, thumbSize, imaging.Lanczos)

	// alpha blend 到浅蓝色背景
	bg := imaging.New(thumb.Bounds().Dx(), thumb.Bounds().Dy(), color.NRGBA{230, 240, 255, 255})
	dst := imaging.Overlay(bg, thumb, image.Pt(0, 0), 1.0)

	out, err := os.Create(thumbPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if err := jpeg.Encode(out, dst, &jpeg.Options{Quality: 85}); err != nil {
		return "", err
	}
	return thumbPath, nil
}
