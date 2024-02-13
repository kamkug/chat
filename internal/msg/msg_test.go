package msg_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"chat/internal/msg"
)

const succeed = "\u2713"
const failed = "\u2717"

// multiReader is a mock io.Reader that simulates a partial read.
type multiReader struct {
	readers []io.Reader
	index   int
}

func (m *multiReader) Read(p []byte) (n int, err error) {
	// Simulate a short read
	if m.index >= len(m.readers) {
		return 0, io.EOF
	}
	// Simulate an unexpected error
	if m.index == 2 {
		return 0, io.ErrClosedPipe
	}
	// Simulate a full read
	n, err = m.readers[m.index].Read(p)
	m.index++
	return
}

func TestRead(t *testing.T) {
	tt := []struct {
		name    string
		msg     msg.MSG
		expData string
		expLen  int
		expErr  error
	}{
		{
			name: "successful read",
			msg: msg.MSG{
				Sender:    "Silvermoon",
				Recipient: "Lightbeam",
				Type:      msg.Message,
				Data:      "Wolverine",
			},
			expLen: 33,
			expErr: nil,
		},
	}

	for _, tst := range tt {
		buff := msg.Encode(tst.msg)
		reader := bytes.NewReader(buff)
		data, length, err := msg.Read(reader)
		if err != tst.expErr {
			t.Fatalf("\t%s\tShould return the expected error : exp[%v] got[%v]\n", failed, tst.expErr, err)
		}
		t.Logf("\t%s\tShould return the expected error.\n", succeed)
		if string(data) != string(buff) {
			t.Fatalf("\t%s\tShould return the expected data : exp[%s] got[%s]\n", failed, tst.expData, data)
		}
		t.Logf("\t%s\tShould return the expected data.\n", succeed)
		if length != tst.expLen {
			t.Fatalf("\t%s\tShould return the expected length : exp[%d] got[%d]\n", failed, tst.expLen, length)
		}
		t.Logf("\t%s\tShould return the expected length.\n", succeed)
	}
}

// TestReadHeader tests the ReadHeader function with various inputs.
func TestReadHeader(t *testing.T) {
	// Mock sleepFunc so it doesn't actually sleep
	msg.SleepFunc = func(d time.Duration) {}

	// Reset sleepFunc after the test
	defer func() {
		msg.SleepFunc = time.Sleep
	}()

	// Create a multiReader that simulates a partial read
	tt := []struct {
		name   string
		reader *multiReader
		length int
		expErr error
		expBuf []byte
		errMsg string
	}{
		{
			name: "short read",
			reader: &multiReader{
				readers: []io.Reader{
					strings.NewReader("Wolverine"),
				},
			},
			length: 20,
			expErr: io.EOF,
			expBuf: nil,
			errMsg: "",
		},
		{
			name: "full read",
			reader: &multiReader{
				readers: []io.Reader{
					strings.NewReader("Wolverine"),
				},
			},
			length: 9,
			expErr: nil,
			expBuf: []byte("Wolverine"),
			errMsg: "",
		},
		{
			name: "partial read",
			reader: &multiReader{
				readers: []io.Reader{
					strings.NewReader("Wolverine "),
					strings.NewReader("Silvermoon"),
				},
			},
			length: 20,
			expErr: nil,
			expBuf: []byte("Wolverine Silvermoon"),
			errMsg: "",
		},
		{
			name: "unexpected error",
			reader: &multiReader{
				readers: []io.Reader{
					strings.NewReader("Wolverine"),
					strings.NewReader("Silvermoon"),
					strings.NewReader("Lightbeam"),
				},
			},
			length: 25,
			expErr: io.ErrClosedPipe,
			expBuf: nil,
			errMsg: "Failed to read the full header: io: read/write on closed pipe",
		},
	}

	t.Log("Given the need to test reading the header of a message.")
	{
		for i, tst := range tt {
			t.Logf("\tTest %d:\t%s", i, tst.name)
			{
				buf, err := msg.ReadHeader(tst.reader, tst.length)
				if unwrappedErr := errors.Unwrap(err); unwrappedErr != nil {
					fmt.Println(unwrappedErr.Error())
					if unwrappedErr.Error() != "Failed to read the full header: io: read/write on closed pipe" {
						t.Fatalf("\t%s\tShould return an error : exp[%v] got[%+v]\n", failed, tst.expErr, err)
					}
					unwrappedErr = errors.Unwrap(unwrappedErr)
					if unwrappedErr != io.ErrClosedPipe {
						t.Fatalf("\t%s\tShould return an error : exp[%v] got[%+v]\n", failed, tst.expErr, err)
					}
				} else if err != tst.expErr {
					t.Fatalf("\t%s\tShould return the expected error : exp[%v] got[%v]\n", failed, tst.expErr, err)
				}
				t.Logf("\t%s\tShould return the expected error.\n", succeed)
				if string(buf) != string(tst.expBuf) {
					t.Fatalf("\t%s\tShould return the expected buffer : exp[%s] got[%s]\n", failed, tst.expBuf, buf)
				}
				t.Logf("\t%s\tShould return the expected buffer.\n", succeed)
			}
		}
	}
}

// TestEncode tests the Encode function with various inputs.
// It checks that the function correctly encodes a message into a byte slice.
func TestEncode(t *testing.T) {
	tt := []struct {
		name         string
		m            msg.MSG
		length       int
		expSender    string
		expRecipient string
	}{
		{
			name: "caplength",
			m: msg.MSG{
				Sender:    "Silvermoon",
				Recipient: "Lightbeam",
				Type:      msg.Message,
				Data:      "hello",
			},
			length:       29,
			expSender:    "Silvermoon",
			expRecipient: "Lightbeam",
		},
		{
			name: "shortname",
			m: msg.MSG{
				Sender:    "John",
				Recipient: "Adam",
				Type:      msg.Message,
				Data:      "helloworld",
			},
			length:       34,
			expSender:    "John",
			expRecipient: "Adam",
		},
		{
			name: "toolong",
			m: msg.MSG{
				Sender:    "Swiftbreeze",
				Recipient: "Quickrunner",
				Type:      msg.Message,
				Data:      "hello",
			},
			length:       29,
			expSender:    "Swiftbreez",
			expRecipient: "Quickrunne",
		},
	}

	t.Log("Given the need to test encoding/decoding.")
	{
		for i, tst := range tt {
			t.Logf("\tTest %d:\t%s", i, tst.name)
			{
				data := msg.Encode(tst.m)
				if len(data) != tst.length {
					t.Fatalf("\t%s\tShould have the correct number of bytes : exp[%d] got[%d]\n", failed, tst.length, len(data))
				}
				t.Logf("\t%s\tShould have the correct number of bytes.\n", succeed)

				m := msg.Decode(data)
				// if m.Sender != tst.m.Sender {
				if tst.expSender != m.Sender {
					t.Fatalf("\t%s\tShould have the correct Sender : exp[%v] got[%v]\n", failed, tst.m.Sender, m.Sender)
				}
				t.Logf("\t%s\tShould have the correct Sender.\n", succeed)

				if tst.expRecipient != m.Recipient {
					// if m.Recipient != tst.m.Recipient {
					t.Fatalf("\t%s\tShould have the correct Recipient : exp[%v] got[%v]\n", failed, tst.m.Recipient, m.Recipient)
				}
				t.Logf("\t%s\tShould have the correct Recipient.\n", succeed)

				if m.Type != tst.m.Type {
					t.Fatalf("\t%s\tShould have the correct Type : exp[%v] got[%v]\n", failed, tst.m.Type, m.Type)
				}
				t.Logf("\t%s\tShould have the correct Type.\n", succeed)

				if m.Data != tst.m.Data {
					t.Fatalf("\t%s\tShould have the correct data : exp[%s] got[%s]\n", failed, tst.m.Data, m.Data)
				}
				t.Logf("\t%s\tShould have the correct data.\n", succeed)
			}
		}
	}
}

// TestGetRecipient tests the GetRecipient function with various inputs.
// It checks that the function correctly extracts the recipient from a message.
func TestGetRecipient(t *testing.T) {
	tt := []struct {
		Data      string
		Recipient string
	}{
		{
			Data:      "@bill hello",
			Recipient: "bill",
		},
		{
			Data:      "@bill hello world",
			Recipient: "bill",
		},
		{
			Data:      "@billhello",
			Recipient: "billhello",
		},
		{
			Data:      "hello",
			Recipient: "",
		},
		{
			Data:      "",
			Recipient: "",
		},
	}

	t.Log("Given the need to test getting the recipient from a message.")
	{
		for i, tst := range tt {
			t.Logf("\tTest %d:\t'%s'", i, tst.Data)
			{
				recipient := msg.GetRecipient(tst.Data)
				if recipient != tst.Recipient {
					t.Fatalf("\t%s\tShould have the correct recipient : exp[%s] got[%s]\n", failed, tst.Recipient, recipient)
				}
				t.Logf("\t%s\tShould have the correct recipient.\n", succeed)
			}
		}
	}
}

// TestGetData tests the GetData function with various inputs.
// It checks that the function correctly extracts the data from a message.
func TestGetData(t *testing.T) {
	tt := []struct {
		Data string
		Msg  string
	}{
		{
			Data: "@bill hello",
			Msg:  "hello",
		},
		{
			Data: "@bill hello world",
			Msg:  "hello world",
		},
		{
			Data: "@billhello",
			Msg:  "",
		},
		{
			Data: "hello",
			Msg:  "hello",
		},
		{
			Data: "",
			Msg:  "",
		},
	}

	t.Log("Given the need to test getting the data from a message.")
	{
		for i, tst := range tt {
			t.Logf("\tTest %d:\t'%s'", i, tst.Data)
			{
				msg := msg.GetData(tst.Data)
				if msg != tst.Msg {
					t.Fatalf("\t%s\tShould have the correct message : exp[%s] got[%s]\n", failed, tst.Msg, msg)
				}
				t.Logf("\t%s\tShould have the correct message.\n", succeed)
			}
		}
	}
}

// TestString tests the String method of the msg.MSG type.
// It checks that the method correctly converts a message into a string.
func TestString(t *testing.T) {
	tt := []struct {
		msg       msg.MSG
		expString string
	}{
		{
			msg: msg.MSG{
				Sender:    "Johnny",
				Recipient: "Ronny",
				Type:      msg.Message,
				Data:      "Hello Ronny!",
			},
			expString: "\n{\n\tSender: Johnny\n\tRecipient: Ronny\n\tType: 1\n\tData: Hello Ronny!\n}",
		},
		{
			msg: msg.MSG{
				Sender:    "Silvermoon",
				Recipient: "Lordbottom",
				Type:      msg.Message,
				Data:      "Hello!",
			},
			expString: "\n{\n\tSender: Silvermoon\n\tRecipient: Lordbottom\n\tType: 1\n\tData: Hello!\n}",
		},
		{
			msg: msg.MSG{
				Sender:    "JohnnyBravo",
				Recipient: "RockyBalboa",
				Type:      msg.Message,
				Data:      "Hello Rocky!",
			},
			expString: "\n{\n\tSender: JohnnyBravo\n\tRecipient: RockyBalboa\n\tType: 1\n\tData: Hello Rocky!\n}",
		},
	}

	t.Log("Given the need to test format for displaying MSG struct.")
	for i, tst := range tt {
		t.Logf("\tTest %d:\t", i)
		{
			got := fmt.Sprintf("%s", tst.msg)
			if got != tst.expString {
				t.Fatalf("\tShould have the correct display format : exp[%s] got[%s]\n", tst.expString, got)
			}
			t.Logf("\tShould have the correct display format")
		}

	}

}
