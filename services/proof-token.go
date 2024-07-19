package services

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bogdanfinn/fhttp"
	"github.com/google/uuid"
	"golang.org/x/crypto/sha3"
	"math/rand"
	"pow/models"
	"pow/util"
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
	cachedSid          = uuid.NewString()
	cachedHardware     = 0
	cachedScripts      []string
	cachedDpl          = ""
	CachedRequireProof = ""
)

func getParseTime() string {
	now := time.Now()
	now = now.In(timeLocation)
	return now.Format(timeLayout) + " GMT+0800 (中国标准时间)"
}

func getConfig(userAgent string) []interface{} {
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

	request, err := http.NewRequest(http.MethodGet, "http://test.haodongxi.site/?oai-dm=1", nil)
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Cookie", "oai-dm-tgt-c-240329=2024-04-02")
	if err != nil {
		return
	}

	// util.Client.SetProxy("http://127.0.0.1:7890")
	response, err := util.Client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(response.Body)
	cachedScripts = nil
	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists {
			cachedScripts = append(cachedScripts, src)
			if cachedDpl == "" {
				idx := strings.Index(src, "dpl")
				if idx >= 0 {
					cachedDpl = src[idx:]
				}
			}
		}
	})

	if len(cachedScripts) == 0 {
		cachedScripts = append(cachedScripts, "https://cdn.oaistatic.com/_next/static/chunks/polyfills-78c92fac7aa8fdd8.js?dpl=baf36960d05dde6d8b941194fa4093fb5cb78c6a")
		cachedDpl = "dpl=baf36960d05dde6d8b941194fa4093fb5cb78c6a"
	}
}

func calcPart(startIndex, endIndex int, proof *models.ParamGetPow, resultChan chan<- string, doneChan chan struct{}, closeOnce *sync.Once) {
	hasher := sha3.New512()
	diffLen := len(proof.Diff)
	config := getConfig(proof.UserAgent)

	loopCount := 0

	for i := startIndex; i < endIndex; i++ {
		loopCount++
		select {
		case <-doneChan:
			return
		default:
			config[3] = i
			config[9] = (i + 2) / 2
			json, _ := json.Marshal(config)
			base := base64.StdEncoding.EncodeToString(json)
			hasher.Write([]byte(proof.Seed + base))
			hash := hasher.Sum(nil)
			hasher.Reset()
			if hex.EncodeToString(hash[:diffLen]) <= proof.Diff {
				resultChan <- base
				return
			}
			if loopCount >= 30000 {
				closeOnce.Do(func() {
					close(doneChan) // 使用sync.Once确保只关闭一次
				})
				return
			}
		}
	}
}

func CalcProofToken(proof *models.ParamGetPow) string {
	start := time.Now()
	getDpl(proof.Proxy, proof.UserAgent)
	timeout := time.Second * 10

	resultChan := make(chan string, 1)
	doneChan := make(chan struct{})
	closeOnce := &sync.Once{} // 创建一个sync.Once实例

	numWorkers := 8
	for i := 0; i < numWorkers; i++ {
		go calcPart(i*50000, (i+1)*50000, proof, resultChan, doneChan, closeOnce)
	}

	select {
	case result := <-resultChan:
		elapsed := time.Since(start)
		fmt.Println("time: ", elapsed, "pow", proof.Seed, proof.Diff)
		return result
	case <-time.After(timeout):
		return ""
	case <-doneChan:
		return ""
	}
}
