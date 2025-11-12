package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type Message struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	Seq      int    `json:"seq"`
	ClientID string `json:"client_id"`
}

type Client struct {
	conn       *net.UDPConn
	serverAddr *net.UDPAddr
	clientID   string
	seq        int
	localPort  int
}

func NewClient(serverHost string, serverPort, localPort int) (*Client, error) {
	// 解析服务器地址
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverHost, serverPort))
	if err != nil {
		return nil, err
	}

	// 绑定到本地端口
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}

	clientID := fmt.Sprintf("client-%d", time.Now().Unix())

	return &Client{
		conn:       conn,
		serverAddr: serverAddr,
		clientID:   clientID,
		localPort:  localPort,
	}, nil
}

func (c *Client) sendMessage(msgType, data string) error {
	c.seq++
	msg := Message{
		Type:     msgType,
		Data:     data,
		Seq:      c.seq,
		ClientID: c.clientID,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = c.conn.WriteToUDP(jsonData, c.serverAddr)
	if err != nil {
		return err
	}

	log.Printf("发送消息: %s, 序列号: %d", msgType, c.seq)
	return nil
}

func (c *Client) startHeartbeat() {
	ticker := time.NewTicker(15 * time.Second) // 15秒心跳
	defer ticker.Stop()

	for range ticker.C {
		err := c.sendMessage("heartbeat", "keep-alive")
		if err != nil {
			log.Printf("心跳发送失败: %v", err)
		}
	}
}

func (c *Client) receiveMessages() {
	buffer := make([]byte, 1024)

	for {
		n, addr, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("读取数据失败: %v", err)
			continue
		}

		var msg Message
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			log.Printf("解析消息失败: %v", err)
			continue
		}

		log.Printf("收到来自 %s 的消息: %s, 序列号: %d",
			addr.String(), msg.Type, msg.Seq)

		// 如果是数据消息，发送确认
		if msg.Type == "data" {
			c.sendAck(msg.Seq)
		}
	}
}

func (c *Client) sendAck(seq int) error {
	ackMsg := Message{
		Type:     "ack",
		Data:     fmt.Sprintf("ack for %d", seq),
		Seq:      seq,
		ClientID: c.clientID,
	}

	jsonData, err := json.Marshal(ackMsg)
	if err != nil {
		return err
	}

	_, err = c.conn.WriteToUDP(jsonData, c.serverAddr)
	if err != nil {
		return err
	}

	log.Printf("发送ACK: %d", seq)
	return nil
}

func (c *Client) Start() {
	log.Printf("客户端启动, ID: %s, 本地端口: %d", c.clientID, c.localPort)

	// 先注册到服务器
	err := c.sendMessage("register", "client registration")
	if err != nil {
		log.Printf("注册失败: %v", err)
	}

	// 启动心跳
	go c.startHeartbeat()

	// 启动接收
	c.receiveMessages()
}

func main() {
	client, err := NewClient("xhzzf.huazyk.cn", 8899, 8898)
	if err != nil {
		log.Fatal(err)
	}

	client.Start()
}
