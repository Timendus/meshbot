package meshtastic

import (
	"errors"
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

func wakeDevice(writer io.ReadWriteCloser) error {
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

func writeMessage(writer io.ReadWriteCloser, message *meshtastic.ToRadio) error {
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

func readMessage(reader io.ReadWriteCloser, debugSink io.Writer) (*meshtastic.FromRadio, error) {
	buffer := make([]byte, 1)
	protobuffer := make([]byte, 0)
	status := 0
	length := 0
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, errors.New("unexpected end of file")
		}

		switch status {
		case 0:
			if buffer[0] == START1 {
				status = 1
			} else {
				// Handle any other bytes as text debug output
				debugSink.Write(buffer)
			}
		case 1:
			if buffer[0] == START2 {
				status = 2
			} else {
				status = 0
				debugSink.Write(buffer)
			}
		case 2:
			length = int(buffer[0]) << 8
			status = 3
		case 3:
			length |= int(buffer[0])
			if length > MAX_SIZE {
				log.Printf("Invalid packet size: %d\n", length)
				msb := make([]byte, 1)
				msb[0] = byte(length >> 8)
				debugSink.Write(msb)
				debugSink.Write(buffer)
				status = 0
			} else {
				status = 4
			}
		default:
			if length > 0 {
				protobuffer = append(protobuffer, buffer[0])
				length--
				if length == 0 {
					status = 0
					result := meshtastic.FromRadio{}
					proto.Unmarshal(protobuffer, &result)
					if DEBUGGING {
						log.Println("\033[90mReceived: " + result.String() + "\033[0m")
					}
					return &result, nil
				}
			}
		}
	}
}
