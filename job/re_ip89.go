package job

import (
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/util"
	"regexp"
	"strings"
)

func (s *ip89) StartUrl() []string {
	return []string{
		"http://api.89ip.cn/tqdl.html?api=1&num=60&port=&address=&isp=",
	}
}

func (s *ip89) Protocol() string {
	return "GET"
}

func (s *ip89) GetReferer() string {
	return "http://api.89ip.cn/"
}

type ip89 struct {
	Spider
}

func (s *ip89) Cron() string {
	return "@every 1m"
}

func (s *ip89) Name() string {
	return "ip89"
}

func (s *ip89) Run() {
	getProxy(s)
}

func (s *ip89) Parse(body string) (proxies []*model.HttpProxy, err error) {
	reg := regexp.MustCompile(util.RegProxy)
	rs := reg.FindAllString(body, -1)

	for _, proxy := range rs {
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
