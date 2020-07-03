/**
* Created by GoLand.
* User: link1st
* Date: 2019-08-21
* Time: 15:42
 */

package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go-stress-testing/AlphaProto"
	"go-stress-testing/model"
	"go-stress-testing/server/golink"
	"go-stress-testing/server/statistics"
	"go-stress-testing/server/verify"
	"math/rand"
	"sync"
	"time"
	pitayaClient "github.com/woshihaomei/pitaya/client"
)

const (
	connectionMode = 1 // 1:顺序建立长链接 2:并发建立长链接
)

var productLine = "dev"

// 注册验证器
func init() {

	// http
	model.RegisterVerifyHttp("statusCode", verify.HttpStatusCode)
	model.RegisterVerifyHttp("json", verify.HttpJson)

	// webSocket
	model.RegisterVerifyWebSocket("json", verify.WebSocketJson)
}

// 处理函数
func Dispose(concurrency, totalNumber uint64, request *model.Request) {

	// 设置接收数据缓存
	ch := make(chan *model.RequestResults, 1000)
	var (
		wg          sync.WaitGroup // 发送数据完成
		wgReceiving sync.WaitGroup // 数据处理完成
	)

	wgReceiving.Add(1)
	go statistics.ReceivingResults(concurrency, ch, &wgReceiving)

	for i := uint64(0); i < concurrency; i++ {
		wg.Add(1)
		switch request.Form {
		case model.FormTypeHttp:

			go golink.Http(i, ch, totalNumber, &wg, request)

		case model.FormTypeWebSocket:

			switch connectionMode {
			case 1:
				// 连接以后再启动协程
				ws := pitayaClient.New(logrus.InfoLevel)
				var err error

				if productLine == "10000" {
					err = ws.ConnectToWS("47.96.161.48:3251", "")
				} else {
					err = ws.ConnectToWS(":3251", "")
				}

				if err != nil {
					fmt.Println("连接失败:", i, err)
					continue
				}else{
					accountId := rand.Int63n(655658)
					req := &AlphaProto.Req_ConnectorEntry{AccountId: accountId}
					_, _ = ws.SendRequest("connector.connector.entry", AlphaProto.Serializer(req))
					time.Sleep(time.Millisecond * 10)
				}

				go golink.WebSocket(i, ch, totalNumber, &wg, request, ws)
			case 2:
				// 并发建立长链接
				go func(i uint64) {
					// 连接以后再启动协程
					ws := pitayaClient.New(logrus.InfoLevel)
					var err error

					if productLine == "10000" {
						fmt.Println("DialConnector=47.96.161.48:3251")
						err = ws.ConnectToWS("47.96.161.48:3251", "")
					} else {
						fmt.Println("DialConnector=:3251")
						err = ws.ConnectToWS(":3251", "")
					}

					if err != nil {
						fmt.Println("连接失败:", i, err)
						return
					}

					golink.WebSocket(i, ch, totalNumber, &wg, request, ws)
				}(i)

				// 注意:时间间隔太短会出现连接失败的报错 默认连接时长:20毫秒(公网连接)
				time.Sleep(5 * time.Millisecond)
			default:

				data := fmt.Sprintf("不支持的类型:%d", connectionMode)
				panic(data)
			}

		default:
			// 类型不支持
			wg.Done()
		}
	}

	// 等待所有的数据都发送完成
	wg.Wait()

	// 延时1毫秒 确保数据都处理完成了
	time.Sleep(1 * time.Millisecond)
	close(ch)

	// 数据全部处理完成了
	wgReceiving.Wait()

	return
}
