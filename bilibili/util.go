package bilibili

import (
	"fmt"
	"net/http"
)

func validateHttpStatus(r *http.Response) (err error) {
	if code := r.StatusCode; code != http.StatusOK {
		err = fmt.Errorf("unsuccessful HTTP status on API %v: %v", r.Request.URL, code)
	}
	return
}
