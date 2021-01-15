package job

import (
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type proxyScrape struct {
	Spider
}

func (s *proxyScrape) StartUrl() []string {
	return []string{
		"https://api.proxyscrape.com/v2/?request=getproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all",
	}
}

func (s *proxyScrape) Protocol() string {
	return "GET"
}

func (s *proxyScrape) GetReferer() string {
	return "https://proxyscrape.com/free-proxy-list"
}

func (s *proxyScrape) Cron() string {
	return "@every 5m"
}

func (s *proxyScrape) Name() string {
	return "proxyScrape"
}

func (s *proxyScrape) Run() {
	getProxy(s)
}

func (s *proxyScrape) Parse(body string) (proxies []*model.HttpProxy, err error) {
	proxyStrings := strings.Split(body, "\r\n")
	for _, proxy := range proxyStrings {
		if strings.Contains(proxy, ":") {
			proxyInfo := strings.Split(proxy, ":")
			proxies = append(proxies, &model.HttpProxy{
				Ip:   proxyInfo[0],
				Port: proxyInfo[1],
			})
		}
	}
	return
}
