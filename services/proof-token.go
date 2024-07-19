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
		util.Client.SetProxy(proxy)
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
	timeout := time.Second * 6

	resultChan := make(chan string, 1)
	doneChan := make(chan struct{})
	closeOnce := &sync.Once{} // 创建一个sync.Once实例

	numWorkers := 8

	for i := 0; i < numWorkers; i++ {
		go calcPart(i*60000, (i+1)*60000, proof, resultChan, doneChan, closeOnce)
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
