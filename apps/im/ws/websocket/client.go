package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/url"
)

type Client interface {
	Close() error
	Send(v any) error
	Read(v any) error
}
type client struct {
	*websocket.Conn
	host string
	opt  dailOption
}

//初始化客户端

func NewClient(host string, opts ...DailOptions) *client {
	opt := newDailOptions(opts...)
	c := client{
		Conn: nil,
		host: host,
		opt:  opt,
	}
	conn, err := c.dail()
	if err != nil {
		panic(err)
	}
	c.Conn = conn
	return &c
}

//建立客户端与websocket连接

func (c *client) dail() (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: c.host, Path: c.opt.pattern}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), c.opt.header)
	return conn, err
}

//发送消息

func (c *client) Send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	err = c.WriteMessage(websocket.TextMessage, data)
	if err == nil {
		return nil
	}
	//todo: 客户端增加重连发送
	conn, err := c.dail()
	if err != nil {
		panic(err)
	}
	c.Conn = conn
	return c.WriteMessage(websocket.TextMessage, data)
}

// 读取消息
func (c *client) Read(v any) error {
	_, msg, err := c.Conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(msg, v)
}

//关闭客户端

func (c *client) Close() error {
	return c.Conn.Close()
}
