package command

type Command struct {
	Data []byte // keep it simple for now
}

func Unmarshal(data []byte) (*Command, error) {
	return &Command{
		Data: data,
	}, nil
}

func Marshal(cmd *Command) ([]byte, error) {
	return cmd.Data, nil
}
