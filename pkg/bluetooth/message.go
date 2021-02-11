package bluetooth

// Message is one CTwiPacket
type Message struct {
	Data []byte
}

func (m *Message) toByteArray() []byte {
	return nil
}

func fromByteArray(data []byte) (*Message, error) {
	return &Message{
		Data: data,
	}, nil
}
