package bilibili

import (
	"context"
	"errors"
	"fmt"
	errors2 "github.com/keuin/slbr/bilibili/errors"
	"github.com/keuin/slbr/common"
	"io"
	"net/http"
	"os"
	"strings"
)

const InitReadBytes = 4096 // 4KiB

// CopyLiveStream read data from a livestream video stream, copy them to a writer.
func (b Bilibili) CopyLiveStream(
	ctx context.Context,
	roomId RoomId,
	stream StreamingUrlInfo,
	fileCreator func() (*os.File, error),
	bufSize int64,
) (err error) {
	url := stream.URL
	if !strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "http://") {
		return fmt.Errorf("invalid URL: %v", url)
	}

	r, err := b.newGet(url)
	if err != nil {
		b.logger.Error("Cannot create HTTP GET instance on %v: %v", url, err)
		return err
	}

	r.Header.Set("Referer",
		fmt.Sprintf("https://live.bilibili.com/blanc/%d?liteVersion=true", roomId))

	resp, err := b.Do(r)
	if err != nil {
		b.logger.Error("Cannot make HTTP GET request on %v: %v\n", url, err)
		return
	}

	// 404 when not streaming
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("live is not started or the room does not exist")
	}

	err = validateHttpStatus(resp)
	if err != nil {
		return
	}

	defer func() { _ = resp.Body.Close() }()

	b.logger.Info("Waiting for stream initial bytes...")
	// read some first bytes to ensure that the live is really started,
	// so we don't create blank files if the live room is open
	// but the live hasn't started yet
	initBytes := make([]byte, InitReadBytes)
	_, err = io.ReadFull(resp.Body, initBytes)
	if err != nil {
		b.logger.Error("Failed to read stream initial bytes: %v", err)
		return
	}
	b.logger.Info("Stream is started. Receiving live stream...")
	// write initial bytes
	var out *os.File
	out, err = fileCreator()
	if err != nil {
		b.logger.Error("Cannot open file for writing: %v", err)
		err = errors2.NewUnrecoverableTaskError("failed to create file", err)
		return
	}
	_, err = out.Write(initBytes)
	if err != nil {
		b.logger.Error("Failed to write to file: %v", err)
		return
	}
	initBytes = nil // discard that buffer

	var n int64

	// blocking copy
copyLoop:
	for err == nil {
		select {
		case <-ctx.Done():
			// cancelled
			err = ctx.Err()
			break copyLoop
		default:
			var sz int64
			sz, err = io.CopyN(out, resp.Body, bufSize)
			n += sz
		}
	}

	if errors.Is(err, context.Canceled) {
		b.logger.Info("Stop copying...")
	} else if errors.Is(err, io.EOF) {
		b.logger.Info("The live is ended. (room %v)", roomId)
	} else {
		b.logger.Error("Stream copying was interrupted unexpectedly: %v", err)
	}

	b.logger.Info("Total downloaded: %v", common.PrettyBytes(uint64(n)))
	return err
}
