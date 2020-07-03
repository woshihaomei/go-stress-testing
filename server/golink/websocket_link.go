/**
* Created by GoLand.
* User: link1st
* Date: 2019-08-21
* Time: 15:43
 */

package golink

import (
	"github.com/woshihaomei/pitaya/conn/message"
	"github.com/woshihaomei/pitaya/helpers"
	"go-stress-testing/heper"
	"go-stress-testing/model"
	"sync"
	"testing"
	"time"
	pitayaClient "github.com/woshihaomei/pitaya/client"
	"go-stress-testing/AlphaProto"
)

const (
	firstTime    = 1 * time.Second // 连接以后首次请求数据的时间
	intervalTime = 1 * time.Second // 发送数据的时间间隔
)

var (
	// 请求完成以后是否保持连接
	keepAlive bool
)

func init() {
	keepAlive = false
}

// web socket go link
func WebSocket(chanId uint64, ch chan<- *model.RequestResults, totalNumber uint64, wg *sync.WaitGroup, request *model.Request, ws *pitayaClient.Client) {

	defer func() {
		wg.Done()
	}()

	// fmt.Printf("启动协程 编号:%05d \n", chanId)

	defer func() {
		ws.Disconnect()
	}()

	var (
		i uint64
	)

	// 暂停60秒
	t := time.NewTimer(firstTime)
	for {
		select {
		case <-t.C:
			t.Reset(intervalTime)

			// 请求
			webSocketRequest(chanId, ch, i, request, ws)

			// 结束条件
			i = i + 1

			if i >= totalNumber {
				goto end
			}
		}
	}

end:
	t.Stop()

	if keepAlive {
		// 保持连接
		chWaitFor := make(chan int, 0)
		<-chWaitFor
	}

	return
}

// 请求
func webSocketRequest(chanId uint64, ch chan<- *model.RequestResults, i uint64, request *model.Request, ws *pitayaClient.Client) {

	var (
		startTime = time.Now()
		isSucceed = true
		errCode   = model.HttpOk
	)

	// 需要发送的数据
	reqLogin := &AlphaProto.Req_RoleLogin{}
	_, err := ws.SendRequest("game.role.login", AlphaProto.Serializer(reqLogin))
	if err != nil {
		isSucceed = false
		errCode = model.RequestErr // 请求错误
	} else {

		// time.Sleep(1 * time.Second)
		msg1 := helpers.ShouldEventuallyReceive(&testing.T{}, ws.IncomingMsgChan).(*message.Message)

		resp := &AlphaProto.Resp_ConnectorEntry{}
		AlphaProto.Deserializer(msg1.Data, resp)
		if resp.ErrorCode != AlphaProto.E_ErrorCode_SUCCESS{
			isSucceed = false
			errCode = model.RequestErr
		}
	}

	requestTime := uint64(heper.DiffNano(startTime))

	requestResults := &model.RequestResults{
		Time:      requestTime,
		IsSucceed: isSucceed,
		ErrCode:   errCode,
	}

	requestResults.SetId(chanId, i)

	ch <- requestResults

}
