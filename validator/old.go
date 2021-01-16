package validator

import (
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/queue"
	"github.com/phpgao/proxy_pool/util"
	"sync"
)

func OldValidator() {
	q := queue.GetOldChan()
	var wg sync.WaitGroup
	logger := util.GetLogger("validator_old")

	for i := 0; i < config.OldQueue; i++ {
		wg.Add(1)
		go func() {
			for {
				proxy := <-q
				func(p model.HttpProxy) {
					key := p.GetKey()
					if _, ok := lockMap.Load(key); ok {
						return
					}

					lockMap.Store(key, 1)
					defer func() {
						lockMap.Delete(key)
					}()
					if storeEngine.Exists(p) {
						var (
							score     int
							failHttp  bool // http 失败标识
							failHttps bool // https失败标识
						)

						// 先检测https，在检测http，并记录下成功与否
						err := p.TestProxy(true)
						if err != nil {
							logger.WithError(err).WithField(
								"proxy", p.GetProxyUrl()).Debug("error retest https proxy")
							failHttps = true

						} else {
							if !p.IsHttps() {
								// 以前是https的不用重新检测http
								err := p.TestProxy(false)
								if err != nil {
									logger.WithError(err).WithField(
										"proxy", p.GetProxyUrl()).Debug("error retest http proxy")
									failHttp = true
								}
							}
						}

						if p.IsHttps() {
							// 如果以前是https
							if failHttps {
								score = -20 // https失败
							} else {
								score = 10
							}
						} else {
							// 如果以前不是https
							if failHttp && failHttps {
								score = -30 // 两种协议都失败
							} else {
								score = 10 // 任一协议成功
							}

							// https 验证成功，把http改成https
							if !failHttps {
								p.Schema = "https"
								// save proxy to db
								err = storeEngine.UpdateSchema(p)
								if err != nil {
									logger.WithError(err).WithField("proxy", p.GetProxyWithSchema()).Info("error update schema")
								}
							}
						}

						//logger.WithFields(log.Fields{
						//	"score": score,
						//	"proxy": p.GetProxyWithSchema(),
						//}).Debug("set score")

						err = storeEngine.AddScore(p, score)
						if err != nil {
							logger.WithError(err).WithField(
								"proxy", p.GetProxyWithSchema()).Error("error setting score")
						}
					}
				}(*proxy)
			}

		}()

	}
	wg.Wait()
}
