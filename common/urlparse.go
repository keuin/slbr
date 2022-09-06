package common

import (
	"errors"
	"net/url"
	"strings"
)

// GetFileExtensionFromUrl
// copied from https://elisegev.medium.com/get-a-file-extension-from-a-url-in-golang-5061d4a298a
func GetFileExtensionFromUrl(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}
	pos := strings.LastIndex(u.Path, ".")
	if pos == -1 {
		return "", errors.New("couldn't find a period to indicate a file extension")
	}
	return u.Path[pos+1 : len(u.Path)], nil
}
