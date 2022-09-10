package bilibili

import (
	"context"
	"errors"
	"fmt"
	"github.com/keuin/slbr/common"
	"net/http"
	"os"
	"strings"
)

// CopyLiveStream read data from a livestream video stream, copy them to a writer.
func (b Bilibili) CopyLiveStream(
	ctx context.Context,
	roomId common.RoomId,
	stream StreamingUrlInfo,
	out *os.File,
	buffer []byte,
	readChunkSize int,
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
		return ErrRoomIsClosed
	}

	err = validateHttpStatus(resp)
	if err != nil {
		return
	}

	defer func() { _ = resp.Body.Close() }()

	b.logger.Info("Copying live stream...")
	// blocking copy
	n, err := common.CopyToFileWithBuffer(ctx, out, resp.Body, buffer, readChunkSize, false)

	if err != nil && !errors.Is(err, context.Canceled) {
		b.logger.Error("Stream copying was interrupted unexpectedly: %v", err)
	}

	if err == nil {
		b.logger.Info("The live is ended. (room %v)", roomId)
	}

	b.logger.Info("Total downloaded: %v", common.PrettyBytes(uint64(n)))
	return err
}
