package bilibili

import (
	"bilibili-livestream-archiver/common"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CopyLiveStream read data from a livestream video stream, copy them to a writer.
func (b Bilibili) CopyLiveStream(
	ctx context.Context,
	roomId common.RoomId,
	stream StreamingUrlInfo,
	out io.Writer,
) (err error) {
	url := stream.URL
	if !strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "http://") {
		return fmt.Errorf("invalid URL: %v", url)
	}

	r, err := b.newGet(url)
	if err != nil {
		b.error.Printf("Cannot create HTTP GET instance on %v: %v", url, err)
		return err
	}

	r.Header.Set("Referer",
		fmt.Sprintf("https://live.bilibili.com/blanc/%d?liteVersion=true", roomId))

	resp, err := b.Do(r)
	if err != nil {
		b.error.Printf("Cannot make HTTP GET request on %v: %v\n", url, err)
		return
	}

	// 404 when not streaming
	if resp.StatusCode == http.StatusNotFound {
		return ErrRoomIsClosed
	}

	err = validateHttpStatus(resp)
	if err != nil {
		return
	}

	defer func() { _ = resp.Body.Close() }()

	// guard the following copy loop
	// if the context is cancelled, stop it by closing the reader
	guardianCtx, cancelGuardian := context.WithCancel(ctx)
	go func() {
		<-guardianCtx.Done()
		_ = resp.Body.Close()
	}()
	defer cancelGuardian()

	// blocking copy
	n, err := common.Copy(ctx, out, resp.Body)

	if err != nil && !errors.Is(err, context.Canceled) {
		// real error happens
		b.error.Printf("Stream copying was interrupted unexpectedly: %v", err)
	}

	if err == nil {
		b.info.Printf("The live is ended. (room %v)", roomId)
	}

	b.info.Printf("Bytes copied: %v", n)
	return err
}
