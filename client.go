package websocket_chat_client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn  *websocket.Conn
	Addr  string
	Proto string
	// SelfSigned Disables checking CA store for cert
	SelfSigned  bool
	wsPath      string
	RoomID      string
	RoomPass    string
	MessageKey  string
	Username    string
	MessageChan chan struct {
		Type int
		Data []byte
	}
}

type ConnectMessage struct {
	ID       string `json:"id"`
	Password string `json:"p"`
}

type Msg struct {
	Username string `json:"un"`
	Message  string `json:"msg"`
}

func (c *Client) Connect() error {
	var err error
	c.wsPath = "message_ws"
	dd := websocket.DefaultDialer
	dd.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: c.SelfSigned,
	}
	dd.HandshakeTimeout = 10 * time.Second
	c.Conn, _, err = dd.Dial(fmt.Sprintf("%s://%s/%s", c.Proto, c.Addr, c.wsPath), nil)
	if err != nil {
		return fmt.Errorf("dial: %v", err)
	}
	c.listener()
	loginJson, err := json.Marshal(ConnectMessage{
		ID:       c.RoomID,
		Password: c.RoomPass,
	})
	if err != nil {
		return fmt.Errorf("error: json marshal: %v", err)
	}
	err = c.Conn.WriteMessage(websocket.TextMessage, loginJson)
	if err != nil {
		return fmt.Errorf("write:", err)
	}
	if err := c.SendMsg(" <- has entered the room!"); err != nil {
		log.Printf("error: sendMsg: %v", err)
	}
	return err
}

func (c *Client) listener() {
	go func() {
		for {
			mt, message, err := c.Conn.ReadMessage()
			if err != nil {
				return
			}
			var jsonString []byte
			switch mt {
			case websocket.TextMessage:
				if jsonString, err = decrypt([]byte(c.MessageKey), message); err != nil {
					log.Println("decrypt:", err)
				}
			default:
				jsonString = message
			}
			c.MessageChan <- struct {
				Type int
				Data []byte
			}{Type: mt, Data: jsonString}
		}
	}()
}

func (c *Client) SendMsg(msg string) error {
	m := Msg{
		Username: c.Username,
		Message:  msg,
	}
	jm, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error: json marshal: %v", err)
	}
	encBytes, err := encrypt([]byte(c.MessageKey), jm)
	if err != nil {
		return fmt.Errorf("encrypt: %v", err)
	}
	return c.Conn.WriteMessage(websocket.TextMessage, encBytes)
	
}

func (c *Client) disconnect() {
	err := c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
	if err := c.Conn.Close(); err != nil {
		log.Printf("error: websocket Conn close: %v", err)
	}
}
