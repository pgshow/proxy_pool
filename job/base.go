package job

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/apex/log"
	"github.com/avast/retry-go"
	"github.com/parnurzeal/gorequest"
	"github.com/phpgao/proxy_pool/db"
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/util"
	"github.com/phpgao/proxy_pool/validator"
	"gitlab.com/NebulousLabs/fastrand"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	logger             = util.GetLogger("source")
	MaxProxyReachedErr = errors.New("max proxy reached")
	storeEngine        = db.GetDb()
)

func init() {
	htmlquery.DisableSelectorCache = true
}

type Crawler interface {
	Run()
	StartUrl() []string
	Profile() *Setting
	Cron() string
	Name() string
	Retry() uint
	NeedRetry() bool
	Enabled() bool
	// url , if use proxy
	Fetch(string, bool, Crawler) (string, string, error)
	SplashFetch(string, bool, Crawler) (string, string, error) // 使用 Splash 打开并渲染网页
	SetProxyChan(chan<- *model.HttpProxy)
	GetProxyChan() chan<- *model.HttpProxy
	Parse(string) ([]*model.HttpProxy, error)
}

type Spider struct {
	ch chan<- *model.HttpProxy
}

// 爬虫Job的设置
type Setting struct {
	Protocol    string //http 方法
	AlwaysProxy bool   //仅代理才能爬的站
	CnWebsite   bool   //网站在中国
}

func (s *Spider) StartUrl() []string {
	panic("implement me")
}

func (s *Spider) errAndStatus(errs []error, resp gorequest.Response) (err error) {
	if len(errs) > 0 {
		err = errs[0]
		return
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http code: %d", resp.StatusCode)
	}

	return
}

func (s *Spider) Cron() string {
	panic("implement me")
}

func (s *Spider) Enabled() bool {
	return true
}

func (s *Spider) NeedRetry() bool {
	return true
}

func (s *Spider) TimeOut() int {
	return util.ServerConf.Timeout
}

func (s *Spider) Name() string {
	panic("implement me")
}

func (s *Spider) Parse(string) ([]*model.HttpProxy, error) {
	panic("implement me")
}

func (s *Spider) GetReferer() string {
	return "https://www.baidu.com/"
}

func (s *Spider) SetProxyChan(ch chan<- *model.HttpProxy) {
	s.ch = ch
}

func (s *Spider) GetProxyChan() chan<- *model.HttpProxy {
	return s.ch
}

func (s *Spider) RandomDelay() bool {
	return true
}

func (s *Spider) Retry() uint {
	return uint(util.ServerConf.MaxRetry)
}

func (s *Spider) Fetch(proxyURL string, useProxy bool, c Crawler) (body string, spiderProxy string, err error) {

	if s.RandomDelay() {
		time.Sleep(time.Duration(fastrand.Intn(6)) * time.Second)
	}

	body, spiderProxy, err = FetchGet(proxyURL, useProxy, s, c)
	return
}

func getProxy(s Crawler) {
	logger.WithField("spider", s.Name()).Debug("spider begin")
	if !s.Enabled() {
		logger.WithField("spider", s.Name()).Debug("spider is not enabled")
	}
	for _, url := range s.StartUrl() {
		go func(proxySiteURL string, inputChan chan<- *model.HttpProxy) {
			defer func() {
				if r := recover(); r != nil {
					logger.WithFields(log.Fields{
						"url":   proxySiteURL,
						"fatal": r,
					}).Warn("Recovered")
				}
			}()

			var newProxies []*model.HttpProxy
			var spiderProxy string

			var attempts = 0
			err := retry.Do(
				func() error {
					attempts++
					logger.WithField("attempts", attempts).Debug(proxySiteURL)

					var err error
					if !validator.CanDo() {
						return MaxProxyReachedErr
					}

					var withProxy bool

					if attempts > 1 || s.Profile().AlwaysProxy {
						// 如果失败了一次，或者是只能使用代理才能爬的站
						withProxy = true
					}

					var resp string

					if s.Profile().Protocol == "RenderFetch" {
						// 需要 浏览器渲染 的网站
						resp, spiderProxy, err = s.SplashFetch(proxySiteURL, withProxy, s)
					} else {
						// go.http 就能爬取的网站
						resp, spiderProxy, err = s.Fetch(proxySiteURL, withProxy, s)
					}

					if err != nil {
						return err
					}

					if resp == "" {
						return errors.New("empty resp")
					}

					newProxies, err = s.Parse(resp)
					if err != nil {
						return err
					}

					if newProxies == nil {
						return errors.New("empty proxy")
					}

					return nil
				},
				retry.Attempts(s.Retry()),
				retry.RetryIf(func(err error) bool {
					// should give up
					if err.Error() == MaxProxyReachedErr.Error() || err.Error() == "empty proxy" {
						return false
					}

					return s.NeedRetry()
				}),
			)

			if err != nil {
				logger.WithError(err).WithField("url", proxySiteURL).Debug("error get new proxy")
			}

			logger.WithFields(log.Fields{
				"name":        s.Name(),
				"url":         proxySiteURL,
				"count":       len(newProxies),
				"spiderProxy": spiderProxy,
			}).Info("url proxy report")

			var tmpMap = map[string]int{}

			for _, newProxy := range newProxies {
				newProxy.Ip = strings.TrimSpace(newProxy.Ip)
				newProxy.Port = strings.TrimSpace(newProxy.Port)
				if _, found := tmpMap[newProxy.GetKey()]; found {
					continue
				}
				tmpMap[newProxy.GetKey()] = 1
				newProxy.From = s.Name()
				if newProxy.Score == 0 {
					newProxy.Score = util.ServerConf.Score
				}
				if model.FilterProxy(newProxy) {
					inputChan <- newProxy
				}
			}
		}(url, s.GetProxyChan())
	}

}

// 使用 Splash 爬取目标页面
func (s *Spider) SplashFetch(proxyURL string, useProxy bool, c Crawler) (body string, spiderProxy string, err error) {

	if s.RandomDelay() {
		time.Sleep(time.Duration(fastrand.Intn(6)) * time.Second)
	}

	body, spiderProxy, err = SplashGet(proxyURL, useProxy, c)

	return
}

func FetchGet(proxyURL string, useProxy bool, s *Spider, c Crawler) (body string, spiderProxy string, err error) {
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
		//var proxy model.HttpProxy
		//proxy, err = storeEngine.Random()

		// 根据条件提取一个代理给爬虫用
		filter := map[string]string{
			"schema":  "http",
			"score":   "60",
			"country": "-cn", // 国外的网站别使用中国代理
			"limit":   "1000",
		}
		if strings.HasPrefix(c.StartUrl()[0], "https") {
			filter["schema"] = "https"
		}
		if c.Profile().CnWebsite {
			filter["country"] = "cn" // 中国的网站使用中国代理
		}

		var list []model.HttpProxy
		list, err = storeEngine.Get(filter)

		if err != nil || list == nil {
			return
		}

		randP := list[fastrand.Intn(len(list))]

		spiderProxy = fmt.Sprintf("%s://%s:%s", filter["schema"], randP.Ip, randP.Port)

		logger.WithField("proxy", spiderProxy).Debug(c.Name() + " with proxy")
		resp, body, errs = superAgent.Proxy(spiderProxy).End()
	} else {
		resp, body, errs = superAgent.End()
	}
	if err = s.errAndStatus(errs, resp); err != nil {
		return
	}

	body = strings.TrimSpace(body)
	return
}

func FetchPost(proxyURL string, useProxy bool, s *Spider, c Crawler, postData interface{}) (body string, spiderProxy string, err error) {
	request := gorequest.New()
	contentType := "text/html; charset=utf-8"
	var superAgent *gorequest.SuperAgent
	var resp gorequest.Response
	var errs []error
	superAgent = request.Post(proxyURL).
		Set("User-Agent", util.GetRandomUA()).
		Set("Content-Type", contentType).
		Set("Referer", s.GetReferer()).
		Set("Pragma", `no-cache`).
		Timeout(time.Duration(s.TimeOut()) * time.Second).SetDebug(util.ServerConf.DumpHttp)

	if useProxy {

		// 根据条件提取一个代理给爬虫用
		filter := map[string]string{
			"schema":  "http",
			"score":   "60",
			"country": "-cn", // 国外的网站别使用中国代理
			"limit":   "1000",
		}
		if strings.HasPrefix(c.StartUrl()[0], "https") {
			filter["schema"] = "https"
		}
		if c.Profile().CnWebsite {
			filter["country"] = "cn" // 中国的网站使用中国代理
		}

		var list []model.HttpProxy
		list, err = storeEngine.Get(filter)

		if err != nil || list == nil {
			return
		}

		randP := list[fastrand.Intn(len(list))]

		spiderProxy = fmt.Sprintf("%s://%s:%s", filter["schema"], randP.Ip, randP.Port)

		logger.WithField("proxy", spiderProxy).Debug(c.Name() + " with proxy")
		resp, body, errs = superAgent.Proxy(spiderProxy).Send(postData).End()
	} else {
		resp, body, errs = superAgent.Send(postData).End()
	}
	if err = s.errAndStatus(errs, resp); err != nil {
		return
	}

	body = strings.TrimSpace(body)
	return
}

func SplashGet(proxyURL string, useProxy bool, c Crawler) (body string, spiderProxy string, err error) {

	var values interface{}

	if useProxy {

		// 根据条件提取一个代理给爬虫用
		filter := map[string]string{
			"schema":  "http",
			"score":   "60",
			"country": "-cn", // 国外的网站别使用中国代理
			"limit":   "1000",
		}
		if strings.HasPrefix(c.StartUrl()[0], "https") {
			filter["schema"] = "https"
		}
		if c.Profile().CnWebsite {
			filter["country"] = "cn" // 中国的网站使用中国代理
		}

		var list []model.HttpProxy
		list, err = storeEngine.Get(filter)

		if err != nil || list == nil {
			return
		}

		randP := list[fastrand.Intn(len(list))]

		spiderProxy = fmt.Sprintf("%s://%s:%s", filter["schema"], randP.Ip, randP.Port)

		// 设置 Splash 参数,使用代理
		type Post struct {
			Url     string            `json:"url"`
			Html    string            `json:"html"`
			Images  string            `json:"images"`
			Headers map[string]string `json:"headers"`
			Proxy   string            `json:"proxy"`
		}

		values = &Post{
			proxyURL,
			"1",
			"0",
			map[string]string{"User-Agent": util.GetRandomUA()},
			spiderProxy,
		}

		logger.WithField("proxy", spiderProxy).Debug(c.Name() + " with proxy")
	} else {
		// 设置 Splash 参数,不使用代理
		type Post struct {
			Url     string            `json:"url"`
			Html    string            `json:"html"`
			Images  string            `json:"images"`
			Headers map[string]string `json:"headers"`
		}

		values = &Post{
			proxyURL,
			"1",
			"0",
			map[string]string{"User-Agent": util.GetRandomUA()},
		}
	}

	jsonValue, _ := json.Marshal(values)

	resp, err := http.Post("http://localhost:8050/render.html", "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		return
	}

	bodyByte, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		return
	}

	body = string(bodyByte)

	return
}
