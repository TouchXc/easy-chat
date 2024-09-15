package mqclient

import (
	"context"
	"easy-chat/apps/task/mq/mq"
	"encoding/json"
	"github.com/zeromicro/go-queue/kq"
)

//定义消息聊天转发客户端

type MsgChatTransferClient interface {
	Push(msg *mq.MsgChatTransfer) error
}

func (m *msgChatTransferClient) Push(msg *mq.MsgChatTransfer) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return m.pusher.Push(context.Background(), string(body))
}

type msgChatTransferClient struct {
	pusher *kq.Pusher
}

func NewMsgChatTransferClient(addr []string, topic string, opts ...kq.PushOption) MsgChatTransferClient {
	return &msgChatTransferClient{
		pusher: kq.NewPusher(addr, topic),
	}
}

//消息已读处理的客户端

type MsgReadTransferClient interface {
	Push(msg *mq.MsgMarkRead) error
}
type msgReadTransferClient struct {
	pusher *kq.Pusher
}

func (m *msgReadTransferClient) Push(msg *mq.MsgMarkRead) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return m.pusher.Push(context.Background(), string(body))
}

func NewMsgReadTransferClient(addr []string, topic string, opts ...kq.PushOption) MsgReadTransferClient {
	return &msgReadTransferClient{
		pusher: kq.NewPusher(addr, topic),
	}
}
