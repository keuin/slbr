package common

/*
Copied from https://ixday.github.io/post/golang-cancel-copy/
*/

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
)

// CopyToFileWithBuffer copies data from io.Reader to os.File with given buffer and read chunk size.
// The reader and writer may not be synchronized if any error occurs.
// A block contains one or more chunks.
// Every syscall reads at most one chunk.
// Every disk-write writes at most one block.
// The buffer is a ring contains one or more blocks.
// The reader and file won't be closed.
// If syncFile is set, the file will be synced after every read.
// ringSize: how many blocks are in the buffer ring.
func CopyToFileWithBuffer(
	ctx context.Context,
	out *os.File,
	in io.Reader,
	buffer []byte,
	syncFile bool,
	ringSize uint,
) (written int64, err error) {
	var blkSize uint
	bufSize := uint(len(buffer))
	if bufSize%ringSize != 0 {
		err = fmt.Errorf("len(buffer) %% ringSize != 0")
	} else {
		blkSize = bufSize / ringSize
	}

	chWriteQue := make(chan uint, ringSize) // buffer write task queue
	chReadQue := make(chan uint, ringSize)  // buffer read task queue

	// when reader and writer are stopped, this channel must have 2 elements
	chResult := make(chan error, 2)

	type task struct {
		Offset uint
		Length uint
	}

	chLastWrite := make(chan task, 1)

	rwCtx, cancelReadWrite := context.WithCancel(ctx)

	// wait reader and writer to finish
	wg := sync.WaitGroup{}

	// buffer reader
	// buffer -> file
	go func() {
		var err error
		wg.Add(1)
		defer wg.Done()
		defer func() {
			chResult <- err
		}()
		for {
			select {
			case <-rwCtx.Done():
				return
			default:
				// get the next available block to read from
				// this block is fully written in the buffer
				off := <-chReadQue

				// read the entire block to file
				n, err2 := out.Write(buffer[off : off+blkSize])
				if err2 != nil {
					// failed to write to the file
					// we can do nothing more than stop reading and writing
					err = fmt.Errorf("write error (%v byte written in this call): %w", n, err2)
					cancelReadWrite()
					return
				}
			}
		}
	}()

	// buffer writer
	// reader -> buffer
	go func() {
		var err error
		wg.Add(1)
		defer wg.Done()
		defer func() {
			chResult <- err
		}()
		for {
			select {
			case <-rwCtx.Done():
				chLastWrite <- task{
					Offset: 0,
					Length: 0,
				}
				return
			default:
				// get the next free block for writing
				off := <-chWriteQue // byte offset

				// fully fill the block
				nWritten := uint(0) // bytes written in this block currently
				for nWritten < blkSize {
					n, err := in.Read(buffer[off+nWritten : off+blkSize])
					nWritten += uint(n)
					if err != nil {
						// if we can't fully fill current block (e.g. EOF or IO error),
						// set nLeft to make the main goroutine handle that
						cancelReadWrite()
						err = fmt.Errorf("reader failed: %w", err)
						chLastWrite <- task{
							Offset: off,
							Length: nWritten, // bytes of valid data in the last incomplete block
						}
						return
					}
				}
			}
		}
	}()

	// init write tasks, all buffer blocks are available for writing initially
	for i := uint(0); i < bufSize; i += blkSize {
		chWriteQue <- i
	}

	wg.Wait()

	for i := 0; i < 2; i++ {
		if err2 := <-chResult; err2 != nil {
			err = err2
			return
		}
	}

	// write the remaining data
	last := <-chLastWrite
	if last.Length > 0 {
		var n int
		n, err = out.Write(buffer[last.Offset : last.Offset+last.Length])
		written += int64(n)
	}
	return
}
