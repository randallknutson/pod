package bluetooth

// Message is one CTwiPacket
type Message struct {
	Data []byte
}

func (m *Message) toByteArray() []byte {
	return m.Data
}

func fromByteArray(data []byte) (*Message, error) {
	return &Message{
		Data: data,
	}, nil
}
