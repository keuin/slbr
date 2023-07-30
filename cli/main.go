package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/keuin/slbr/danmaku/dmpkg"
	"log"
	"os"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			panic(err)
		}
		var data []byte
		err = func() error {
			data, err = hex.DecodeString(string(line))
			if err == nil {
				return nil
			}
			data, err = base64.StdEncoding.DecodeString(string(line))
			if err != nil {
				log.Println("Failed to decode as Hex or Base64 string")
			}
			return err
		}()
		if err != nil {
			continue
		}
		ex, err := dmpkg.DecodeExchange(data)
		if err != nil {
			log.Println("Failed to decode Danmaku exchange: ", err)
			continue
		}
		ex, err = ex.Inflate()
		if err != nil {
			log.Println("Failed to decompress Danmaku exchange: ", err)
			continue
		}
		fmt.Println(ex.PrettyString())
	}
}
