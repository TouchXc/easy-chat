package websocket

import (
	"time"
)

// FrameType 表示 WebSocket 消息的帧类型。
type FrameType uint8

const (
	FrameData  FrameType = 0x0 // 数据帧
	FramePing  FrameType = 0x1 // Ping 帧
	FrameAck   FrameType = 0x2 // Ack 帧
	FrameNoAck FrameType = 0x3 // 无 Ack 帧
	FrameErr   FrameType = 0x9 // 错误帧

	// 其他可能的帧类型（已注释）
	//FrameHeaders      FrameType = 0x1
	//FramePriority     FrameType = 0x2
	//FrameRSTStream    FrameType = 0x3
	//FrameSettings     FrameType = 0x4
	//FramePushPromise  FrameType = 0x5
	//FrameGoAway       FrameType = 0x7
	//FrameWindowUpdate FrameType = 0x8
	//FrameContinuation FrameType = 0x9
)

type Message struct {
	Id        string `json:"id"`
	FrameType `json:"frameType"`
	AckSeq    int         `json:"ackSeq"`
	ackTime   time.Time   `json:"ackTime"`
	errCount  int         `json:"errCount"`
	Method    string      `json:"method"`
	FormId    string      `json:"formId"`
	Data      interface{} `json:"data"`
}

func NewMessage(formId string, data interface{}) *Message {
	return &Message{
		FrameType: FrameData,
		FormId:    formId,
		Data:      data,
	}
}

func NewErrMessage(err error) *Message {
	return &Message{
		FrameType: FrameErr,
		Data:      err.Error(),
	}
}
