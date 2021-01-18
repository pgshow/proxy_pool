package job

import (
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type rudnkh struct {
	Spider
}

func (s *rudnkh) StartUrl() []string {
	return []string{
		"https://proxy.rudnkh.me/txt",
		"https://raw.githubusercontent.com/a2u/free-proxy-list/master/free-proxy-list.txt",
	}
}

func (s *rudnkh) Profile() *Setting {
	return &Setting{
		Protocol:    "Fetch",
		AlwaysProxy: false,
		CnWebsite:   false,
	}
}

func (s *rudnkh) GetReferer() string {
	return "https://proxy.rudnkh.me/"
}

func (s *rudnkh) Cron() string {
	return "@every 2m"
}

func (s *rudnkh) Name() string {
	return "rudnkh"
}

func (s *rudnkh) Run() {
	getProxy(s)
}

func (s *rudnkh) Parse(body string) (proxies []*model.HttpProxy, err error) {

	proxyString := strings.Split(body, "\n")
	for _, proxy := range proxyString {
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
