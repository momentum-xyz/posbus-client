package pbc

import (
	"context"
	"time"

	"github.com/momentum-xyz/ubercontroller/logger"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
)

const (
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	inMessageSizeLimit = 32768 // TODO: determine a sane value of in-browser user and impl batching in backend
	// maximal size of buffer in messages, after which we drop connection as not-working
	maxBufferSize = 10000
	// Negative Number to indicate closed chan, large enough to be less than any number of outstanding
	chanIsClosed = -0x3FFFFFFFFFFFFFFF
)

type Client struct {
	conn          *websocket.Conn
	log           *zap.SugaredLogger
	url           string
	hs            posbus.HandShake
	currentTarget umid.UMID
	callback      func(data posbus.Message)
	clientCtx     context.Context
	connectionCtx context.Context
	cancelConn    context.CancelFunc
}

func NewClient() *Client {
	c := &Client{}
	c.log = logger.L()
	c.callback = c.defaultCallback
	return c
}

func (c *Client) Connect(ctx context.Context, url, token string, userId umid.UMID) error {
	c.url = url
	c.hs.Token = token
	c.hs.UserId = userId
	c.hs.SessionId = umid.New()
	c.hs.HandshakeVersion = 1
	c.hs.ProtocolVersion = 1
	c.clientCtx = ctx
	c.connectionCtx, c.cancelConn = context.WithCancel(ctx)
	return c.doConnect(c.connectionCtx, false)
}

func (c *Client) Send(msg []byte) (err error) {
	//c.send <- msg
	if err := c.conn.Write(c.connectionCtx, websocket.MessageBinary, msg); err != nil {
		c.log.Debugf("write error: %v", err)
	}
	return
}

func (c *Client) doConnect(ctx context.Context, reconnect bool) error {
	var err error
	c.log.Infof("PBC: connecting to %s (re:%v)... ", c.url, reconnect)
	for ok := true; ok; ok = err != nil {
		conn, _, err := websocket.Dial(ctx, c.url, nil)
		if err != nil {
			c.log.Infof("websocker dail: %v", err)
			time.Sleep(time.Second)
		} else {
			c.conn = conn
		}
	}
	//if err != nil {
	//c.callback(posbus.TypeSignal, posbus.Signal{Value: posbus.SignalConnectionFailed})
	//	return err
	//}
	c.startIOPumps(ctx, c.cancelConn)
	c.Send(posbus.BinMessage(&c.hs))
	c.callback(&posbus.Signal{Value: posbus.SignalConnected})
	if reconnect {
		c.Send(posbus.BinMessage(&posbus.TeleportRequest{Target: c.currentTarget}))
	}
	return nil
}

func (c *Client) SetToken(token string) error {
	c.hs.Token = token
	return nil
}

func (c *Client) SetURL(url string) error {
	c.url = url
	return nil
}

func (c *Client) SetCallback(f func(msg posbus.Message)) {
	c.callback = f
}

func (c *Client) startIOPumps(ctx context.Context, cf context.CancelFunc) {
	go c.readPump(ctx, cf)
	//go c.writePump(ctx, cf)
}

func (c *Client) Close() error {
	c.log.Infof("PBC: disconnect")
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

func (c *Client) readPump(ctx context.Context, connectionCancel context.CancelFunc) {
	c.log.Infof("PBC: start of read pump")

	c.conn.SetReadLimit(inMessageSizeLimit)
	//c.conn.SetReadDeadline(time.Now().Add(pongWait))
	//c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	closeReason := ""
	for {
		messageType, message, err := c.conn.Read(ctx)
		if err != nil {
			if ce, ok := err.(*websocket.CloseError); ok {
				switch ce.Code {
				case websocket.StatusNormalClosure,
					websocket.StatusGoingAway,
					websocket.StatusNoStatusRcvd:
					c.log.Info(
						errors.WithMessagef(err, "PBC: read pump: websocket closed by server"),
					)
					closeReason = "server"
				}
			} else if errors.Is(err, context.Canceled) {
				c.log.Info(
					errors.WithMessagef(err, "PBC: read pump: cancelled by client"),
				)
				closeReason = "user"
			} else {
				c.log.Debug(
					errors.WithMessagef(err, "PBC: read pump: failed to read message from connection"),
				)
				closeReason = "client"
			}
			break
		}
		if messageType != websocket.MessageBinary {
			c.log.Errorf("PBC: read pump: wrong incoming message type: %d", messageType)
		} else {
			if err := c.processMessage(message); err != nil {
				c.log.Warn(errors.WithMessage(err, "PBC: read pump: failed to handle message"))
			}
		}
	}
	c.conn.Close(websocket.StatusNormalClosure, closeReason)
	c.callback(&posbus.Signal{Value: posbus.SignalConnectionClosed})
	c.log.Infof("PBC: end of read pump")
	if ctx.Err() == nil { // Only try reconnecting if it was not cancelled by us
		connectionCancel()                                              //stops the read/write goroutines for (previous) connection
		c.connectionCtx, c.cancelConn = context.WithCancel(c.clientCtx) // from original client context
		go c.doConnect(c.connectionCtx, true)
	}
}

func (c *Client) processMessage(buf []byte) error {
	msg, err := posbus.Decode(buf)
	if err != nil {
		l := len(buf)
		head := 8
		if l < head {
			head = l
		}
		return errors.WithMessagef(err, "PBC: read pump: failed to decode message, head=%#v (total len=%d)", buf[:head], l)
	}

	if msg.GetType() == posbus.TypeSetWorld {
		c.currentTarget = msg.(*posbus.SetWorld).ID
	}
	c.callback(msg)
	return nil
}

func (c *Client) defaultCallback(data posbus.Message) {
	msgName := posbus.MessageNameById(data.GetType())
	c.log.Infof("PSB: got a message of type: %+v , data: %+v\n", msgName, data)
}
