package dmmsg

type BaseRawMessage[I any, D any] struct {
	Cmd  string `json:"cmd"`
	Info I      `json:"info"`
	Data D      `json:"data"`
}
