package bilibili

type BaseResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    T      `json:"data"`
}
