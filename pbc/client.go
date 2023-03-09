package pbc

import (
	"context"
	"github.com/google/uuid"
	"github.com/momentum-xyz/ubercontroller/logger"
	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
	"time"
)

const (
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	inMessageSizeLimit = 1024
	// maximal size of buffer in messages, after which we drop connection as not-working
	maxBufferSize = 10000
	// Negative Number to indicate closed chan, large enough to be less than any number of outstanding
	chanIsClosed = -0x3FFFFFFFFFFFFFFF
)

type Client struct {
	conn     *websocket.Conn
	log      *zap.SugaredLogger
	ctx      context.Context
	url      string
	send     chan []byte
	hs       posbus.HandShake
	callback func(msgType posbus.MsgType, data interface{}) error
}

func NewClient(ctx context.Context) *Client {
	c := &Client{}
	c.ctx = ctx

	c.log = logger.L()
	c.send = make(chan []byte)
	c.callback = c.defaultCallback
	return c
}

func (c *Client) Connect(url, token string, userId uuid.UUID) error {
	c.url = url
	c.hs.Token = token
	c.hs.UserId = userId
	c.hs.SessionId = uuid.New()
	c.hs.HandshakeVersion = 1
	c.hs.ProtocolVersion = 1
	c.doConnect()
	return nil
}
func (c *Client) Send(msg []byte) {
	c.send <- msg
}

func (c *Client) doConnect() error {
	var err error
	c.conn, _, err = websocket.Dial(c.ctx, c.url, nil)
	if err != nil {
		c.callback(posbus.TypeSignal, 6)
		return err
	}
	c.startIOPumps()

	c.Send(posbus.NewMessageFromData(posbus.TypeHandShake, c.hs).Buf())

	c.callback(posbus.TypeSignal, 7)
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

func (c *Client) SetCallback(f func(msgType posbus.MsgType, msg interface{}) error) error {
	c.callback = f
	return nil
}

func (c *Client) startIOPumps() {
	go c.readPump()
	go c.writePump()
}

func (c *Client) Close() error {
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

func (c *Client) readPump() {
	c.log.Infof("PBC: start of read pump")

	c.conn.SetReadLimit(inMessageSizeLimit)
	//c.conn.SetReadDeadline(time.Now().Add(pongWait))
	//c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		messageType, message, err := c.conn.Read(c.ctx)
		if err != nil {
			closedByClient := false
			if ce, ok := err.(*websocket.CloseError); ok {
				switch ce.Code {
				case websocket.StatusNormalClosure,
					websocket.StatusGoingAway,
					websocket.StatusNoStatusRcvd:
					closedByClient = true
				}
			}
			if closedByClient {
				c.log.Info(
					errors.WithMessagef(err, "PBC: read pump: websocket closed by server"),
				)
			} else {
				c.log.Debug(
					errors.WithMessagef(err, "PBC: read pump: failed to read message from connection"),
				)
			}
			break
		}
		if messageType != websocket.MessageBinary {
			c.log.Errorf("PBC: read pump: wrong incoming message type: %d", messageType)
		} else {
			if err := c.processMessage(posbus.BytesToMessage(message)); err != nil {
				c.log.Warn(errors.WithMessagef(err, "PBC: read pump: failed to handle message"))
			}
		}
	}
	c.conn.Close(websocket.StatusNormalClosure, "")
	c.callback(posbus.TypeSignal, 8)
	c.log.Infof("PBC: end of read pump")
}

func (c *Client) writePump() {
	c.log.Infof("PBC: start of write pump")

	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case message := <-c.send:
			c.log.Infof("Send message")
			if message == nil {
				return
			}

			if c.conn.Write(c.ctx, websocket.MessageBinary, message) != nil {
				return
			}

		case <-ticker.C:
			if c.conn.Ping(c.ctx) != nil {
				return
			}
		}
	}
}

func (c *Client) processMessage(msg *posbus.Message) error {
	var err error
	var data interface{}
	switch msg.Type() {
	case posbus.TypeSetUsersTransforms:
		upb := posbus.BytesToUserTransformBuffer(msg.Msg())
		if upb == nil {
			return nil
		}
		data = utils.GetPTR(upb.Decode())
	case posbus.TypeSendTransform:
		d := cmath.NewUserTransform()
		d.CopyFromBuffer(msg.Msg())
		data = &d
	case posbus.TypeGenericMessage:
		data = msg.Msg()
	default:
		//fmt.Println(string(msg.Buf()))
		data, err = msg.Decode()
	}

	if err != nil {
		return err
	}
	c.callback(msg.Type(), data)
	return nil
}

func (c *Client) defaultCallback(msgType posbus.MsgType, data interface{}) error {
	msgName := posbus.MessageNameById(msgType)
	c.log.Infof("PSB: got a message of type: %+v , data: %+v\n", msgName, data)
	return nil

}
