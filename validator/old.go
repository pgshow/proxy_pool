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

						var score, httpSuccess, httpsSuccess = 0, false, false // http 成功标识, https成功标识

						if p.IsHttps() {
							// 如果以前是https,则只检测https
							err := p.TestProxy(true)
							if err != nil {
								logger.WithError(err).WithField(
									"proxy", p.GetProxyUrl()).Debug("error retest https proxy")
							} else {
								httpsSuccess = true
							}
						} else {
							// 如果以前不是https,则检测两种协议
							err := p.TestProxy(false)
							if err != nil {
								logger.WithError(err).WithField(
									"proxy", p.GetProxyUrl()).Debug("error retest http proxy")
							} else {
								httpSuccess = true

								err := p.TestProxy(true)
								if err != nil {
									logger.WithError(err).WithField(
										"proxy", p.GetProxyUrl()).Debug("error retest https proxy")
								} else {
									httpsSuccess = true
								}
							}
						}

						if p.IsHttps() {
							// 如果以前是https

							if httpsSuccess {
								score = 10
							} else {
								score = -20 // https失败
							}
						} else {
							// 如果以前不是https

							if httpsSuccess {

								// https协议成功,更新协议
								score = 10
								p.Schema = "https"
								// save proxy to db
								err := storeEngine.UpdateSchema(p)
								if err != nil {
									logger.WithError(err).WithField("proxy", p.GetProxyWithSchema()).Info("error update schema")
								}

							} else if httpSuccess {
								// 只有http协议成功
								score = 10

							} else {
								score = -30 // 两种协议都失败
							}
						}

						//logger.WithFields(log.Fields{
						//	"score": score,
						//	"proxy": p.GetProxyWithSchema(),
						//}).Debug("set score")

						err := storeEngine.AddScore(p, score)
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
