package pretty

import (
	"fmt"
	"time"
)

func Duration(duration time.Duration) string {
	d := int64(duration.Seconds())
	h := d / 3600
	m := (d % 3600) / 60
	s := d % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
