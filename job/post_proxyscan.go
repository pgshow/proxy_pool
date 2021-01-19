package job

import (
	"errors"
	"github.com/phpgao/proxy_pool/model"
	"gitlab.com/NebulousLabs/fastrand"
	"regexp"
	"strings"
	"time"
)

type proxyScan struct {
	Spider
}

func (s *proxyScan) Fetch(proxyURL string, useProxy bool, c Crawler) (body string, spiderProxy string, err error) {
	if s.RandomDelay() {
		time.Sleep(time.Duration(fastrand.Intn(6)) * time.Second)
	}

	body, spiderProxy, err = FetchPost(proxyURL, useProxy, &s.Spider, c,
		"status=1&ping=&selectedType=HTTP&selectedType=HTTPS&sortPing=false&sortTime=true&sortUptime=false")

	return
}

func (s *proxyScan) StartUrl() []string {
	return []string{
		"https://www.proxyscan.io/Home/FilterResult",
	}
}

func (s *proxyScan) Profile() *Setting {
	return &Setting{
		Protocol:    "Fetch",
		AlwaysProxy: false,
		CnWebsite:   false,
	}
}

func (s *proxyScan) Cron() string {
	return "@every 2m"
}

func (s *proxyScan) GetReferer() string {
	return "https://www.proxyscan.io/"
}

func (s *proxyScan) Run() {
	getProxy(s)
}

func (s *proxyScan) Name() string {
	return "proxyScan"
}

func (s *proxyScan) Parse(body string) (proxies []*model.HttpProxy, err error) {
	scriptRe := regexp.MustCompile(`<th scope="row">(\d+\.\d+\.\d+\.\d+)</th>[\s]+<td>(\d{2,5})</td>`)
	scriptRs := scriptRe.FindAllStringSubmatch(body, -1)
	if scriptRs == nil {
		err = errors.New("random data not found")
		return
	}

	for _, match := range scriptRs {
		ip := strings.TrimSpace(match[1])
		port := strings.TrimSpace(match[2])

		proxies = append(proxies, &model.HttpProxy{
			Ip:   ip,
			Port: port,
		})
	}
	return
}
