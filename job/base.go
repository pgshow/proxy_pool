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
	"io/ioutil"
	"math/rand"
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
	Protocol() string
	Cron() string
	Name() string
	Retry() uint
	NeedRetry() bool
	Enabled() bool
	// url , if use proxy
	Fetch(string, bool) (string, error)
	SplashFetch(string) (string, error) // 使用 Splash 打开并渲染网页
	SetProxyChan(chan<- *model.HttpProxy)
	GetProxyChan() chan<- *model.HttpProxy
	Parse(string) ([]*model.HttpProxy, error)
}

type Spider struct {
	ch chan<- *model.HttpProxy
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

func (s *Spider) Fetch(proxyURL string, useProxy bool) (body string, err error) {

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

	body = strings.TrimSpace(body)
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

					if attempts > 1 {
						withProxy = true
					}

					var resp string

					if s.Protocol() == "RenderFetch" {
						// 需要 浏览器渲染 的网站
						resp, err = s.SplashFetch(proxySiteURL)
					} else {
						// go.http 就能爬取的网站
						resp, err = s.Fetch(proxySiteURL, withProxy)
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
				"name":  s.Name(),
				"url":   proxySiteURL,
				"count": len(newProxies),
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
func (s *Spider) SplashFetch(proxyURL string) (body string, err error) {

	if s.RandomDelay() {
		time.Sleep(time.Duration(rand.Intn(6)) * time.Second)
	}

	// 设置 Splash 参数
	type Post struct {
		Url     string            `json:"url"`
		Html    string            `json:"html"`
		Images  string            `json:"images"`
		Headers map[string]string `json:"headers"`
	}

	values := &Post{
		proxyURL,
		"1",
		"0",
		map[string]string{"User-Agent": util.GetRandomUA()},
	}

	jsonValue, _ := json.Marshal(values)

	resp, err1 := http.Post("http://localhost:8050/render.html", "application/json", bytes.NewBuffer(jsonValue))

	err = err1

	if err1 != nil {
		return
	}

	bodyByte, err2 := ioutil.ReadAll(resp.Body)

	err = err2

	defer resp.Body.Close()

	if err2 != nil {
		return
	}

	body = string(bodyByte)

	return
}
