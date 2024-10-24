package meshtastic

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"google.golang.org/protobuf/proto"
)

const (
	START1    = 0x94
	START2    = 0xC3
	MAX_SIZE  = 512
	DEBUGGING = false
)

func wakeDevice(writer io.Writer) error {
	// Comments copied from Python implementation
	// https://github.com/meshtastic/python/blob/0bb4b31b6a147134c57fb720492c8719c037d195/meshtastic/stream_interface.py#L55-L75

	// Send some bogus UART characters to force a sleeping device to wake, and
	// if the reading statemachine was parsing a bad packet make sure
	// we write enough start bytes to force it to resync (we don't use START1
	// because we want to ensure it is looking for START1)
	bytes := make([]byte, 32)
	_, err := writer.Write(bytes)
	if err != nil {
		return err
	}

	// wait 100ms to give device time to start running
	time.Sleep(100 * time.Millisecond)
	return nil
}

func writeMessage(writer io.Writer, message *meshtastic.ToRadio) error {
	if DEBUGGING {
		log.Println("\033[90mSending: " + message.String() + "\033[0m")
	}

	bytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	header := [4]byte{START1, START2, byte(len(bytes) >> 8), byte(len(bytes) & 0xFF)}
	_, err = writer.Write(header[:])
	if err != nil {
		return err
	}

	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func readMessage(reader io.Reader) (*meshtastic.FromRadio, error) {
	buffer := make([]byte, 1)
	state := 0
	length := 0

searching:
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, errors.New("unexpected end of file")
		}

		switch state {
		case 0:
			if buffer[0] == START1 {
				state = 1
			} else if DEBUGGING {
				// Handle any other bytes as text debug output
				fmt.Print(buffer)
			}
		case 1:
			if buffer[0] == START2 {
				state = 2
			} else {
				state = 0
				if DEBUGGING {
					fmt.Print([]byte{START1})
					fmt.Print(buffer)
				}
			}
		case 2:
			length = int(buffer[0]) << 8
			state = 3
		case 3:
			length |= int(buffer[0]) & 0xFF
			if length > MAX_SIZE {
				log.Printf("Invalid packet size: %d\n", length)
				if DEBUGGING {
					fmt.Print([]byte{START1, START2, byte(length >> 8)})
					fmt.Print(buffer)
				}
				state = 0
			} else if length == 0 {
				state = 0
			} else {
				break searching
			}
		}
	}

	protobuffer := make([]byte, length)
	n, err := reader.Read(protobuffer)
	if err != nil {
		return nil, err
	}
	if n != length {
		return nil, errors.New("unexpected end of file")
	}

	result := meshtastic.FromRadio{}
	err = proto.Unmarshal(protobuffer, &result)
	if err != nil {
		return nil, err
	}
	if DEBUGGING {
		log.Println("\033[90mReceived: " + result.String() + "\033[0m")
	}
	return &result, nil
}
