package raknet

import "io"

type RakNetDisconnectNotification struct {
}

func (RakNetDisconnectNotification) Id() byte {
	return ID_DISCONNECT_NOTIFICATION
}

func (*RakNetDisconnectNotification) Decode(reader io.Reader) error {
	return nil
}

func (RakNetDisconnectNotification) Encode(writer io.Writer) error {
	return WriteByte(writer, ID_DISCONNECT_NOTIFICATION)
}
