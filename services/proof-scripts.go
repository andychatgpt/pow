package services

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/bogdanfinn/fhttp"
	"net/http"
	"pow/util"
	"strings"
)

func GetScripts() ([]string, error) {
	if len(cachedScripts) > 0 {
		return cachedScripts, nil
	}
	request, err := http.NewRequest(http.MethodGet, "https://chatgpt.com/?oai-dm=1", nil)
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Cookie", "oai-dm-tgt-c-240329=2024-04-02")
	if err != nil {
		return nil, err
	}

	//util.Client.SetProxy("http://127.0.0.1:7890")
	response, err := util.Client.Do(request)
	if err != nil {
		return nil, err
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

	return cachedScripts, nil
}
