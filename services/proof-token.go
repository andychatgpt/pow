package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bogdanfinn/fhttp"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"golang.org/x/crypto/sha3"
	"math/rand"
	"pow/models"
	"pow/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	userAgent       = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	cores           = []int{8, 12, 16, 24}
	screens         = []int{3000, 4000, 6000}
	timeLocation, _ = time.LoadLocation("Asia/Shanghai")
	timeLayout      = "Mon Jan 2 2006 15:04:05"
	startTime       = time.Now()
)

var (
	cachedSid      = uuid.NewString()
	cachedHardware = 0
	cachedScripts  []string
	cachedDpl      = ""
	configPool     = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	CachedRequireProof = ""
)

type Config []string

func getParseTime() string {
	now := time.Now()
	now = now.In(timeLocation)
	return now.Format(timeLayout) + " GMT+0800 (中国标准时间)"
}

func getConfig(userAgent string) []interface{} {
	getDpl("", userAgent)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	script := cachedScripts[rand.Intn(len(cachedScripts))]
	timeNum := (float64(time.Since(startTime).Nanoseconds()) + rand.Float64()) / 1e6
	return []interface{}{cachedHardware, getParseTime(), int64(4294705152), 0, userAgent, script, cachedDpl, "en-US", "en-US", 0, "webkitGetUserMedia−function webkitGetUserMedia() { [native code] }", "location", "ontransitionend", timeNum, cachedSid}
}

func getDpl(proxy, userAgent string) {
	if len(cachedScripts) > 0 {
		return
	}
	if proxy != "" {
		_ = util.Client.SetProxy(proxy)
	}
	cachedScripts = append(cachedScripts, "https://cdn.oaistatic.com/_next/static/chunks/9598-0150caea9526d55d.js?dpl=abad631f183104e6c8a323392d7bc30b933c5c7c")
	cachedDpl = "dpl=abad631f183104e6c8a323392d7bc30b933c5c7c"
	request, err := http.NewRequest(http.MethodGet, "https://chatgpt.com/?oai-dm=1", nil)
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "*/*")
	if err != nil {
		return
	}
	response, err := util.Client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(response.Body)
	scripts := []string{}
	inited := false
	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists {
			scripts = append(scripts, src)
			if !inited {
				idx := strings.Index(src, "dpl")
				if idx >= 0 {
					cachedDpl = src[idx:]
					inited = true
				}
			}
		}
	})
	if len(scripts) != 0 {
		cachedScripts = scripts
	}

}

func CalcProofToken(proof *models.ParamGetPow) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	result := make(chan string)

	// 启动1000个协程
	for i := 0; i < 500; i++ {
		go GenerateConfig(ctx, i*1000, 1000, proof.Seed, proof.Diff, result, proof.UserAgent)
	}

	// 等待结果或超时
	select {
	case a := <-result:
		cancel() // 一旦找到结果，取消其他协程
		return a, nil
	case <-ctx.Done():
		cancel()
		return "", fmt.Errorf("timeout")
	}
}

func GenerateConfig(ctx context.Context, startIndex, size int, seed string, diff string, result chan<- string, UserAgent string) {

	config := getConfig(UserAgent)

	for i := startIndex; i < startIndex+size; i++ {
		select {
		case <-ctx.Done(): // 如果上下文被取消，则退出
			return
		default:
			config[3] = strconv.Itoa(i)
			config[9] = strconv.Itoa((i + 2) / 2)

			buf := configPool.Get().(*bytes.Buffer)
			buf.Reset()
			enc := sonic.ConfigDefault.NewEncoder(buf)
			if err := enc.Encode(config); err != nil {
				configPool.Put(buf)
				continue
			}

			line := buf.Bytes()
			base := base64.StdEncoding.EncodeToString(line)
			digest := sha3.Sum512([]byte(seed + base))
			configPool.Put(buf)

			// 改为前3个字节的比较
			if hex.EncodeToString(digest[:3]) <= diff {
				select {
				case result <- base:
					return
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
