/*
Since Bilibili only uses WebSocket binary message,
the transportation is effectively a datagram channel.
So we use a consumer-supplier abstraction to decouple
the real protocol with WebSocket stuff.
*/
package dmpkg

type Consumer[T any] interface {
	Consume(value T) error
}

type Supplier[T any] interface {
	Get() (T, error)
}
