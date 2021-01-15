package job

import (
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/util"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

type openProxy struct {
	Spider
}

func (s *openProxy) Fetch(proxyURL string, useProxy bool) (body string, err error) {
	// 第一次爬取网站的随机目录名称
	if s.RandomDelay() {
		time.Sleep(time.Duration(rand.Intn(6)) * time.Second)
	}

	request := gorequest.New()
	contentType := "text/html; charset=utf-8"
	var superAgent *gorequest.SuperAgent
	var resp gorequest.Response
	var errs []error
	superAgent = request.Get(proxyURL).
		Set("User-Agent", util.GetRandomUA()).
		Set("Content-Type", contentType).
		Set("Referer", s.GetReferer()).
		Set("Pragma", `no-cache`).
		Timeout(time.Duration(s.TimeOut()) * time.Second).SetDebug(util.ServerConf.DumpHttp)

	if useProxy {
		var proxy model.HttpProxy
		proxy, err = storeEngine.Random()
		if err != nil {
			return
		}
		p := fmt.Sprintf("http://%s:%s", proxy.Ip, proxy.Port)
		logger.WithField("proxy", p).Debug("with proxy")
		resp, body, errs = superAgent.Proxy(p).End()
	} else {
		resp, body, errs = superAgent.End()
	}
	if err = s.errAndStatus(errs, resp); err != nil {
		return
	}

	scriptRe := regexp.MustCompile(`FRESH HTTP/S","code":"(\w+)"`) // 提取随机目录名
	scriptRs := scriptRe.FindAllStringSubmatch(body, 1)
	if scriptRs == nil {
		err = errors.New("random page name not found")
		return
	}

	pageUrl := "https://openproxy.space/list/" + scriptRs[0][1]

	// 第二次 爬取 代理列表页
	if s.RandomDelay() {
		time.Sleep(time.Duration(rand.Intn(6)) * time.Second)
	}

	superAgent = request.Get(pageUrl).
		Set("User-Agent", util.GetRandomUA()).
		Set("Content-Type", contentType).
		Set("Referer", s.GetReferer()).
		Set("Pragma", `no-cache`).
		Timeout(time.Duration(s.TimeOut()) * time.Second).SetDebug(util.ServerConf.DumpHttp)

	if useProxy {
		var proxy model.HttpProxy
		proxy, err = storeEngine.Random()
		if err != nil {
			return
		}
		p := fmt.Sprintf("http://%s:%s", proxy.Ip, proxy.Port)
		logger.WithField("proxy", p).Debug("with proxy")
		resp, body, errs = superAgent.Proxy(p).End()
	} else {
		resp, body, errs = superAgent.End()
	}
	if err = s.errAndStatus(errs, resp); err != nil {
		return
	}

	return

}

func (s *openProxy) StartUrl() []string {
	random := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	return []string{
		"https://api.openproxy.space/list?skip=0&ts=16107" + random,
	}
}

func (s *openProxy) Protocol() string {
	return "Fetch"
}

func (s *openProxy) Cron() string {
	return "@every 6h"
}

func (s *openProxy) GetReferer() string {
	return "https://openproxy.space"
}

func (s *openProxy) Run() {
	getProxy(s)
}

func (s *openProxy) Name() string {
	return "openProxy"
}

func (s *openProxy) Parse(body string) (proxies []*model.HttpProxy, err error) {
	rs := util.FindIpPort(body)
	if rs == nil {
		err = errors.New("random data not found")
		return
	}

	for _, match := range rs {
		ip := strings.TrimSpace(match[1])
		port := strings.TrimSpace(match[2])

		proxies = append(proxies, &model.HttpProxy{
			Ip:   ip,
			Port: port,
		})
	}
	return
}
