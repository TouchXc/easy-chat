package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"net/http"
	"sync"
	"time"
)

type AckType int

const (
	NoAck AckType = iota
	OnlyAck
	RigorAck
)

func (t AckType) ToString() string {
	switch t {
	case OnlyAck:
		return "OnlyAck"
	case RigorAck:
		return "RigorAck"
	}
	return "NoAck"
}

// Server 表示 WebSocket 服务器的实现。
//
// 字段:
//   - routes: map[string]HandlerFunc
//     存储与请求方法对应的处理函数的路由表，每个请求方法都映射到一个特定的 `HandlerFunc`。
//   - addr: string
//     服务器监听的地址，表示 WebSocket 服务器将在哪个地址和端口上监听连接。
//   - patten: string
//     WebSocket 连接的路径模式，用于匹配客户端连接的路径。
//   - opt: *websocketOption
//     WebSocket 连接的配置选项，定义了各种 WebSocket 相关的配置参数。
//   - upgrader: websocket.Upgrader
//     WebSocket 协议升级器，用于将 HTTP 连接升级为 WebSocket 连接。
//   - Logger: logx.Logger
//     日志记录器，用于记录服务器的日志信息，包括错误、信息和调试日志。
//   - connToUser: map[*Conn]string
//     连接到用户映射表，将每个 WebSocket 连接映射到其对应的用户 ID。
//   - userToConn: map[string]*Conn
//     用户到连接映射表，将每个用户 ID 映射到其当前的 WebSocket 连接。
//   - TaskRunner: *threading.TaskRunner
//     任务运行器，用于管理和执行异步任务。
//   - RWMutex: sync.RWMutex
//     读写互斥锁，用于保护连接和用户映射表的并发读写操作。
//   - authentication: Authentication
//     鉴权接口，负责处理 WebSocket 连接的鉴权逻辑。
type Server struct {
	sync.RWMutex
	*threading.TaskRunner
	opt    *serverOption
	routes map[string]HandlerFunc

	authentication Authentication

	addr   string
	patten string
	//websocket连接对象存储
	connToUser map[*Conn]string //从连接找到用户

	userToConn map[string]*Conn //从用户找到连接对象
	upgradee   websocket.Upgrader
	logx.Logger
}

// 服务初始化

func NewServer(addr string, opts ...ServerOptions) *Server {
	opt := newServerOptions(opts...)
	return &Server{
		routes:         make(map[string]HandlerFunc),
		opt:            &opt,
		addr:           addr,
		authentication: opt.Authentication,
		patten:         opt.patten,
		connToUser:     make(map[*Conn]string),
		userToConn:     make(map[string]*Conn),
		upgradee: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Logger:     logx.WithContext(context.Background()),
		TaskRunner: threading.NewTaskRunner(opt.concurrency),
	}
}

// 为服务添加具体接受请求方法

func (s *Server) ServerWs(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			s.Errorf("server handler ws recover err %v", r)
		}
	}()

	conn := NewConn(s, w, r)
	if conn == nil {
		return
	}
	//conn, err := s.upgradee.Upgrade(w, r, nil)
	//if err != nil {
	//	s.Errorf("upgrade fail err %v", err)
	//	return
	//}

	if !s.authentication.Auth(w, r) {
		s.Send(&Message{FrameType: FrameData, Data: fmt.Sprint("不具备访问权限")}, conn)
		//conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("不具备访问权限")))
		conn.Close()
		return
	}
	//记录连接
	s.addConn(conn, r)
	//处理连接
	go s.handlerConn(conn)
}

// 根据连接对象进行任务处理
func (s *Server) handlerConn(conn *Conn) {
	uids := s.GetUsers(conn)
	conn.Uid = uids[0]
	//处理任务
	go s.handlerWrite(conn)

	if s.isAck(nil) {
		go s.readAck(conn)
	}
	for {
		//获取请求消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.Errorf("websocket conn read message err %v", err)
			s.Close(conn)
			return
		}
		//解析消息
		var message Message
		if err = json.Unmarshal(msg, &message); err != nil {
			s.Errorf("websocket unmarshal err %v,msg %v", err, msg)
			//s.Close(conn)
			//return
			continue
		}
		//给客户端回复一个ack

		//根据消息进行处理
		if s.isAck(&message) {
			s.Infof("conn message read ack msg %v", msg)
			conn.appendMsgMq(&message)
		} else {
			conn.message <- &message
		}
	}
}

func (s *Server) isAck(message *Message) bool {
	if message == nil {
		return s.opt.ack != NoAck
	}
	return s.opt.ack != NoAck && message.FrameType != FrameNoAck
}

// 读取消息ack确认
func (s *Server) readAck(conn *Conn) {
	for {
		select {
		case <-conn.done:
			s.Infof("close message ack uid %v ", conn.Uid)
			return
		default:
		}

		//从队列中读取新的消息
		conn.messageMu.Lock()
		if len(conn.readMessage) == 0 {
			conn.messageMu.Unlock()
			//增加睡眠
			time.Sleep(100 * time.Microsecond)
			continue
		}

		// 读取第一条消息
		message := conn.readMessage[0]
		//判断ack的方式
		switch s.opt.ack {
		case OnlyAck:
			//直接给客户端回复
			s.Send(&Message{
				FrameType: FrameAck,
				Id:        message.Id,
				AckSeq:    message.AckSeq + 1,
			}, conn)
			//进行业务处理
			//把消息从队列移除
			conn.readMessage = conn.readMessage[1:]
			conn.messageMu.Unlock()
			conn.message <- message
		case RigorAck:
			//先回
			if message.AckSeq == 0 {
				//还未确认
				conn.readMessage[0].AckSeq++
				conn.readMessage[0].ackTime = time.Now()
				s.Send(&Message{
					FrameType: FrameAck,
					Id:        message.Id,
					AckSeq:    message.AckSeq,
				}, conn)
				s.Infof("message ack RigorAck send mid %v,seq %v, time %v", message.Id, message.AckSeq, message.ackTime)
				conn.messageMu.Unlock()
				continue
			}
			//再验证

			//1.客户端返回结果，再一次确认
			//得到客户端序号
			msgSeq := conn.readMessageSeq[message.Id]
			if msgSeq.AckSeq > message.AckSeq {
				//确认
				conn.readMessage = conn.readMessage[1:]
				conn.messageMu.Unlock()
				conn.message <- message
				s.Infof("message ack RigroAck sucess mid %v", message.Id)
				continue
			}

			//2.客户端没有确认，考虑是否超过了ack机制确认时间
			val := s.opt.ackTimeout - time.Since(message.ackTime)
			if !message.ackTime.IsZero() && val <= 0 {
				//2.2 超过结束确认
				delete(conn.readMessageSeq, message.Id)
				conn.readMessage = conn.readMessage[1:]
				conn.messageMu.Unlock()
				continue
			}
			//2.1 未超过，重新发送
			conn.messageMu.Unlock()
			s.Send(&Message{
				FrameType: FrameAck,
				Id:        message.Id,
				AckSeq:    message.AckSeq,
			}, conn)
			//睡眠一定时间
			time.Sleep(3 * time.Second)
		}
	}
}

// 任务处理
func (s *Server) handlerWrite(conn *Conn) {
	for {
		select {
		case <-conn.done:
			//连接关闭
			return
		case message := <-conn.message:
			//根据消息进行处理
			switch message.FrameType {
			case FramePing:
				s.Send(&Message{FrameType: FramePing}, conn)
			case FrameData:
				//根据请求的method方法分发路由并执行
				if handler, ok := s.routes[message.Method]; ok {
					handler(s, conn, message)
				} else {
					s.Send(&Message{FrameType: FrameData, Data: fmt.Sprintf("不存在的执行方法 %v 请检查",
						message.Method)}, conn)
					//conn.WriteMessage(websocket.TextMessage,
					//	[]byte(fmt.Sprintf("不存在的执行方法 %v 请检查", message.Method)))
				}
			}
			if s.isAck(message) {
				conn.messageMu.Lock()
				delete(conn.readMessageSeq, message.Id)
				conn.messageMu.Unlock()
			}
		}
	}
}

func (s *Server) addConn(conn *Conn, req *http.Request) {
	//这里解析请求中的userId，原方法中中如果没有就根据时间戳生成id
	uid := s.authentication.UserId(req)
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	// 验证用户是否之前登入过
	if c := s.userToConn[uid]; c != nil {
		//关闭之前的连接
		c.Close()
	}
	s.connToUser[conn] = uid
	s.userToConn[uid] = conn
}

//根据uid获取ws连接

func (s *Server) GetConn(uid string) *Conn {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()
	return s.userToConn[uid]
}

//根据uid组获取ws连接组

func (s *Server) GetConns(uids ...string) []*Conn {
	if len(uids) == 0 {
		return nil
	}

	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	res := make([]*Conn, 0, len(uids))
	for _, uid := range uids {
		res = append(res, s.userToConn[uid])
	}
	return res
}

//根据ws连接获取用户组

func (s *Server) GetUsers(conns ...*Conn) []string {

	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	var res []string
	if len(conns) == 0 {
		// 获取全部
		res = make([]string, 0, len(s.connToUser))
		for _, uid := range s.connToUser {
			res = append(res, uid)
		}
	} else {
		// 获取部分
		res = make([]string, 0, len(conns))
		for _, conn := range conns {
			res = append(res, s.connToUser[conn])
		}
	}

	return res
}

//关闭ws连接

func (s *Server) Close(conn *Conn) {

	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	uid := s.connToUser[conn]
	//防止重复关闭
	if uid == "" {
		// 已经被关闭
		return
	}

	conn.Close()

	delete(s.connToUser, conn)
	delete(s.userToConn, uid)

}

//根据用户id发送消息

func (s *Server) SendByUserId(msg interface{}, sendIds ...string) error {
	if len(sendIds) == 0 {
		return nil
	}
	return s.Send(msg, s.GetConns(sendIds...)...)
}

//根据ws连接发送消息

func (s *Server) Send(msg interface{}, conns ...*Conn) error {
	if len(conns) == 0 {
		return nil
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	for _, conn := range conns {
		if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return err
		}
	}
	return nil
}

//添加路由方法

func (s *Server) AddRoutes(rs []Route) {
	for _, r := range rs {
		s.routes[r.Method] = r.Handler
	}
}

//服务启动方法

func (s *Server) Start() {
	http.HandleFunc(s.patten, s.ServerWs)
	s.Info(http.ListenAndServe(s.addr, nil))
}

// 停止服务

func (s *Server) Stop() {
	fmt.Println("停止服务")
}
