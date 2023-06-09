package nsqx

type BaseMessage interface {
	Encode() ([]byte, error)
	GetTopic() string
}
