package myurl

import (
	"errors"
	"net/url"
	"strings"
)

type Url string

func (o Url) Url() string {
	return string(o)
}

// FileExtension returns file extension of file name from this url.
// copied from https://elisegev.medium.com/get-a-file-extension-from-a-url-in-golang-5061d4a298a
func (o Url) FileExtension() (string, error) {
	u, err := url.Parse(o.Url())
	if err != nil {
		return "", err
	}
	pos := strings.LastIndex(u.Path, ".")
	if pos == -1 {
		return "", errors.New("couldn't find a period to indicate a file extension")
	}
	return u.Path[pos+1 : len(u.Path)], nil
}
