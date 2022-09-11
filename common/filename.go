package common

import "fmt"

func CombineFileName(base string, ext string) string {
	return fmt.Sprintf("%s.%s", base, ext)
}
