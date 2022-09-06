package dmpkg

// NewPing construct a new PING exahange.
func NewPing() (exc DanmakuExchange) {
	// compliant with Bilibili webapp behavior
	exc, _ = NewPlainExchange(OpHeartbeat, "[object Object]")
	return
}
