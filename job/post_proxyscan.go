package job

import (
	"errors"
	"github.com/phpgao/proxy_pool/model"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type proxyScan struct {
	Spider
}

func (s *proxyScan) Fetch(proxyURL string, useProxy bool) (body string, err error) {
	if s.RandomDelay() {
		time.Sleep(time.Duration(rand.Intn(6)) * time.Second)
	}

	// 设置 Post 参数
	resp, err := http.Post("https://www.proxyscan.io/Home/FilterResult",
		"application/x-www-form-urlencoded",
		strings.NewReader("status=1&ping=&selectedType=HTTP&selectedType=HTTPS&sortPing=false&sortTime=true&sortUptime=false"))

	if err != nil {
		return
	}

	bodyByte, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	body = string(bodyByte)

	return
}

func (s *proxyScan) StartUrl() []string {
	return []string{
		"https://www.proxyscan.io/Home/FilterResult",
	}
}

func (s *proxyScan) Protocol() string {
	return "Fetch"
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