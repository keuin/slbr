package common

/*
Copied from https://ixday.github.io/post/golang-cancel-copy/
*/

import (
	"context"
	"io"
	"os"
)

// CopyToFileWithBuffer copies data from io.Reader to os.File with given buffer and read chunk size.
// The reader and file won't be closed.
// If syncFile is set, the file will be synced after every read.
func CopyToFileWithBuffer(
	ctx context.Context,
	out *os.File,
	in io.Reader,
	buffer []byte,
	chunkSize int,
	syncFile bool,
) (written int64, err error) {
	bufSize := len(buffer)
	off := 0   // offset to the end of data in buffer
	nRead := 0 // how many bytes were read in the last read
	defer func() {
		if off+nRead > 0 {
			// write unwritten data in buffer
			nWrite, _ := out.Write(buffer[:off+nRead])
			written += int64(nWrite)
			if syncFile {
				_ = out.Sync()
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			nRead, err = in.Read(buffer[off:Min[int](off+chunkSize, bufSize)])
			if err != nil {
				return
			}
			off += nRead
			if off == bufSize {
				// buffer is full
				var nWritten int
				nWritten, err = out.Write(buffer)
				if err != nil {
					return
				}
				if syncFile {
					err = out.Sync()
					if err != nil {
						return
					}
				}
				written += int64(nWritten)
				off = 0
			}
		}
	}
}
