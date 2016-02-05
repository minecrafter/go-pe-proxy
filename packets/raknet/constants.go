package raknet

const (
  ID_CONNECTED_PING byte = 0x00
  ID_UNCONNECTED_PING byte = 0x01
  ID_CONNECTED_PONG byte = 0x03
  ID_OPEN_CONNECTION_REQUEST_1 byte = 0x05
  ID_OPEN_CONNECTION_REPLY_1 byte = 0x06
  ID_OPEN_CONNECTION_REQUEST_2 byte = 0x07
  ID_OPEN_CONNECTION_REPLY_2 byte = 0x08
  ID_CONNECTION_REQUEST byte = 0x09
  ID_CONNECTION_REQUEST_ACCEPTED byte = 0x10
  ID_NEW_INCOMING_CONNECTION byte = 0x13
  ID_NO_FREE_INCOMING_CONNECTIONS byte = 0x14
  ID_DISCONNECT_NOTIFICATION byte = 0x15
  ID_CONNECTION_BANNED byte = 0x17
  ID_UNCONNECTED_PONG byte = 0x1c
  ID_USER_DATA_ENUM byte = 0x7E
  ID_DATA_4 byte = 0x84
  ID_DATA_C byte = 0x8c
  ID_ACK = 0xc0
  ID_NAK = 0xa0
)

var (
  MAGIC = []byte{ 0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78 }
)

type Reliability byte

const (
  Unreliable Reliability = 0
  UnreliableSequenced Reliability = 1
  Reliable Reliability = 2
  ReliableOrdered Reliability = 3
  ReliableSequenced Reliability = 4
  UnreliableWithAckReceipt Reliability = 5
  ReliableWithAckReceipt Reliability = 6
  ReliableOrderedWithAckReceipt Reliability = 7
)
