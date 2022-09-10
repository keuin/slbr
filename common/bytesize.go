package common

import "fmt"

func PrettyBytes(b uint64) string {
	if b < 1000 {
		return fmt.Sprintf("%d Byte", b)
	}
	if b < 1000_000 {
		return fmt.Sprintf("%.2f KiB", float64(b)/1024)
	}
	if b < 1000_000_000 {
		return fmt.Sprintf("%.2f MiB", float64(b)/1024/1024)
	}
	if b < 1000_000_000_000 {
		return fmt.Sprintf("%.2f GiB", float64(b)/1024/1024/1024)
	}
	return fmt.Sprintf("%.2f TiB", float64(b)/1024/1024/1024/1024)
}
