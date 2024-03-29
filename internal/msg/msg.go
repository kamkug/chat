package msg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

const hdrLength = 24

const (
	Init = uint8(iota)
	Message
	InCache
)

// MSG defines the message protocol data.
type MSG struct {
	Sender    string
	Recipient string
	Type      uint8
	Data      string
}

// String implements the fmt.Stringer interface.
func (m MSG) String() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("\n{\n\tSender: %s\n", m.Sender))
	b.WriteString(fmt.Sprintf("\tRecipient: %s\n", m.Recipient))
	b.WriteString(fmt.Sprintf("\tType: %d\n", m.Type))
	b.WriteString(fmt.Sprintf("\tData: %s\n}", m.Data))

	return b.String()
}

// Read waits on the network to receive a chat message.
func Read(r io.Reader) ([]byte, int, error) {

	// TODO: Not handling potential partial reads.

	// Read the first header length of bytes.
	buf := make([]byte, hdrLength)
	if _, err := io.ReadFull(r, buf); err != nil {
		errors.Wrap(err, "ReadFull header")
		return nil, 0, err
	}

	// Get the length for the remaining bytes.
	length := int(binary.BigEndian.Uint16(buf[20:22])) + hdrLength

	// Copy the header bytes into the final slice.
	data := make([]byte, length)
	copy(data, buf)

	// Read the remaining bytes.
	if _, err := io.ReadFull(r, data[hdrLength:]); err != nil {
		errors.Wrap(err, "ReadFull data")
		return nil, 0, err
	}

	return data, length, nil
}

// Decode will take the bytes and create a MSG value.
func Decode(data []byte) MSG {

	// Extract the bytes for the sender.
	var sender string
	if n := bytes.IndexByte(data[:10], 0); n != -1 {
		sender = string(data[:n])
	} else {
		sender = string(data[:10])
	}

	// Extract the bytes for the recipient.
	var recipient string
	if n := bytes.IndexByte(data[10:20], 0); n != -1 {
		recipient = string(data[10 : 10+n])
	} else {
		recipient = string(data[10:20])
	}

	// Return the full message.
	return MSG{
		Sender:    sender,
		Recipient: recipient,
		Type:      data[22],
		Data:      string(data[24:]),
	}
}

// Encode will take a message and produce byte slice.
func Encode(m MSG) []byte {

	// We can't have more than the first 10 bytes.
	ns := len(m.Sender)
	if ns > 10 {
		ns = 10
	}

	nr := len(m.Recipient)
	if nr > 10 {
		nr = 10
	}

	// Create a slice of the exact length we need.
	data := make([]byte, hdrLength+len(m.Data))

	// Copy the bytes into the slice for our protocol.

	copy(data, m.Sender[:ns])
	copy(data[10:], m.Recipient[:nr])
	binary.BigEndian.PutUint16(data[20:22], uint16(len(m.Data)))
	data[22] = m.Type
	copy(data[24:], m.Data)

	return data
}

// Gets Recipient from message
func GetRecipient(m string) string {
	// If the message starts with @ then we have a recipient.
	if strings.HasPrefix(m, "@") {
		data := strings.Fields(m)
		recipient := strings.TrimPrefix(data[0], "@")
		return recipient
	}

	return ""
}

// Gets data from message
func GetData(m string) string {
	// If the message starts with @ then we have a recipient.
	if strings.HasPrefix(m, "@") {
		data := strings.Fields(m)
		if len(data) > 1 {
			return strings.Join(data[1:], " ")
		}
		return ""
	}

	return m
}
