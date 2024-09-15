package websocket

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type Conn struct {
	idleMu sync.Mutex
	Uid    string
	*websocket.Conn
	s *Server
	//当前空闲时间
	idle time.Time
	//最大空闲时间
	maxConnectionIdle time.Duration

	messageMu      sync.Mutex
	readMessage    []*Message
	readMessageSeq map[string]*Message
	message        chan *Message

	//关闭通道
	done chan struct{}
}

func NewConn(s *Server, w http.ResponseWriter, r *http.Request) *Conn {

	var responseHeader http.Header
	if protocol := r.Header.Get("Sec-Websocket-Protocol"); protocol != "" {
		responseHeader = http.Header{
			"Sec-Websocket-Protocol": []string{protocol},
		}
	}

	//根据请求升级为ws服务连接
	c, err := s.upgradee.Upgrade(w, r, responseHeader)
	if err != nil {
		s.Errorf("upgrade fail err %v", err)
		return nil
	}
	conn := &Conn{
		Conn:              c,
		s:                 s,
		idle:              time.Now(),
		maxConnectionIdle: s.opt.maxConnectionIdle,
		readMessage:       make([]*Message, 0, 2),
		readMessageSeq:    make(map[string]*Message, 2),
		message:           make(chan *Message, 1),
		done:              make(chan struct{}),
	}
	go conn.keepalive()
	return conn
}
func (c *Conn) appendMsgMq(msg *Message) {
	c.messageMu.Lock()
	defer c.messageMu.Unlock()
	//读队列中
	if m, ok := c.readMessageSeq[msg.Id]; ok {
		//已经有消息的记录，已经存在ack的确认
		if len(c.readMessage) == 0 {
			//队列中无消息
			return
		}
		// msg.AckSeq > m.AckSeq  消息的Ack序号大于已存储的Ack序号
		if m.AckSeq >= msg.AckSeq {
			//没有进行ack确认，重复
			return
		}
		c.readMessageSeq[msg.Id] = msg
		return
	}
	//还没有进行ack的确认,避免客户端重复发送多余ack消息
	if msg.FrameType == FrameAck {
		return
	}

	c.readMessage = append(c.readMessage, msg)
	c.readMessageSeq[msg.Id] = msg
}
func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	//方法并发不安全 加锁
	messageType, p, err = c.Conn.ReadMessage()
	c.idleMu.Lock()
	defer c.idleMu.Unlock()
	c.idle = time.Time{}
	return
}
func (c *Conn) WriteMessage(messageType int, data []byte) error {
	c.idleMu.Lock()
	defer c.idleMu.Unlock()
	//方法并发不安全 加锁
	err := c.Conn.WriteMessage(messageType, data)
	c.idle = time.Now()
	return err
}

// 关闭连接

func (c *Conn) Close() error {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	return c.Conn.Close()
}
func (c *Conn) keepalive() {
	//初始化空计时器
	idleTimer := time.NewTimer(c.maxConnectionIdle)
	defer func() {
		idleTimer.Stop()
	}()
	for {
		select {
		case <-idleTimer.C:
			c.idleMu.Lock()
			idle := c.idle
			if idle.IsZero() { // The connection is non-idle.
				c.idleMu.Unlock()
				idleTimer.Reset(c.maxConnectionIdle)
				continue
			}
			val := c.maxConnectionIdle - time.Since(idle)
			c.idleMu.Unlock()
			if val <= 0 {
				// The connection has been idle for a duration of keepalive.MaxConnectionIdle or more.
				// Gracefully close the connection.
				c.s.Close(c)
				return
			}
			idleTimer.Reset(val)
		case <-c.done:
			return
		}
	}
}
