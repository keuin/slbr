package bilibili

import (
	"context"
	"nhooyr.io/websocket"
)

func (b *Bilibili) DialWebSocket(ctx context.Context, url string) (*websocket.Conn, error) {
	ws, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		HTTPClient: b.http,
	})
	return ws, err
}
