/*
This file implements the background WebSocket messaging channel in Bilibili webui.
Server send livestream start and stop messages via this channel.
Note: In this file we manage the concrete WebSocket connection.
The Bilibili WebSocket channel protocol is decoupled and implemented in package `dmpkg`.
*/
package danmaku

import (
	"context"
	"fmt"
	"github.com/keuin/slbr/common"
	"github.com/keuin/slbr/danmaku/dmpkg"

	"nhooyr.io/websocket"
)

// BilibiliWebSocketMessageType Bilibili uses only binary WebSocket messages
const BilibiliWebSocketMessageType = websocket.MessageBinary

type DanmakuClient struct {
	ws   *websocket.Conn
	wsio wsDatagramIO
}

type DanmakuMessageType int

// wsDatagramIO wraps websocket into a datagram I/O,
// since Bilibili uses only binary data,
// which is effectively a datagram communication.
type wsDatagramIO struct {
	ws  *websocket.Conn
	ctx context.Context
}

func (w *wsDatagramIO) Consume(data []byte) error {
	return w.ws.Write(w.ctx, BilibiliWebSocketMessageType, data)
}

func (w *wsDatagramIO) Get() (data []byte, err error) {
	typ, data, err := w.ws.Read(w.ctx)
	if err != nil {
		return
	}
	if typ != BilibiliWebSocketMessageType {
		err = fmt.Errorf("invalid message type: expected a binary WebSocket message, however got %v", typ.String())
	}
	return
}

func NewDanmakuClient() DanmakuClient {
	return DanmakuClient{
		ws: nil,
	}
}

func (d *DanmakuClient) Connect(ctx context.Context, url string) error {
	// thread unsafe

	// dial
	if d.ws != nil {
		return fmt.Errorf("already connected")
	}
	ws, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to establish WebSocket connection: %w", err)
	}
	d.ws = ws

	// init wsio
	d.wsio = wsDatagramIO{
		ws:  ws,
		ctx: ctx,
	}

	return nil
}

func (d *DanmakuClient) Disconnect() error {
	// thread unsafe
	ws := d.ws
	if ws == nil {
		return nil
	}
	d.ws = nil
	d.wsio = wsDatagramIO{}
	return ws.Close(websocket.StatusInternalError, "disconnected")
}

func (d *DanmakuClient) Authenticate(roomId common.RoomId, authKey string) error {
	pkg := dmpkg.NewAuth(dmpkg.ProtoPlainJson, roomId, authKey)
	data, err := pkg.Marshal()
	if err != nil {
		return fmt.Errorf("exchange marshal failed: %w", err)
	}
	err = d.wsio.Consume(data)
	if err != nil {
		return fmt.Errorf("channel write failed: %w", err)
	}
	// read server response
	resp, err := d.wsio.Get()
	if err != nil {
		return err
	}
	respEx, err := dmpkg.DecodeExchange(resp)
	if err != nil {
		return fmt.Errorf("server danmaku exchange decode error: %w", err)
	}
	ok, err := dmpkg.IsAuthOk(respEx)
	if !ok {
		return fmt.Errorf("danmaku auth failed: %w", err)
	}
	return nil
}

func (d *DanmakuClient) Heartbeat() error {
	pkg := dmpkg.NewPing()
	data, err := pkg.Marshal()
	if err != nil {
		return fmt.Errorf("exchange marshal failed: %w", err)
	}
	err = d.wsio.Consume(data)
	if err != nil {
		return fmt.Errorf("channel write failed: %w", err)
	}
	return nil
}

// ReadExchange read and decode some kind of exchanges which we are interested
func (d *DanmakuClient) ReadExchange() (dmpkg.DanmakuExchange, error) {
	data, err := d.wsio.Get()
	if err != nil {
		return dmpkg.DanmakuExchange{}, fmt.Errorf("failed to read danmaku datagram from server: %w", err)
	}
	return dmpkg.DecodeExchange(data)
}
