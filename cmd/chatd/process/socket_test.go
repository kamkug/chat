package process_test

import (
	"bytes"
	"errors"
	"net"
	"testing"

	"chat/cmd/chatd/process"

	"github.com/ardanlabs/kit/tcp"
)

var success = "\u2713"
var failure = "\u2717"

// mockWriter is a mock implementation of the io.Writer interface.
type mockWriter struct {
	Buffer bytes.Buffer
	Err    error
}

func (m *mockWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, m.Err
	}
	return m.Buffer.Write(data)
}

func TestWrite(t *testing.T) {

	// Create a mock response
	tt := []struct {
		name        string
		resp        tcp.Response
		data        []byte
		length      int
		wantErr     error
		wantMessage string
	}{
		{
			name: "Writes",
			resp: tcp.Response{
				TCPAddr: &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 50000,
					Zone: "",
				},
				Data:   []byte("Hello, World!"),
				Length: 0,
			},
			wantErr:     nil,
			wantMessage: "Hello, World!",
		},
		{
			name: "Errors",
			resp: tcp.Response{
				TCPAddr: &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 50000,
					Zone: "",
				},
				Data:   []byte(""),
				Length: 0,
			},
			wantErr:     errors.New("Test Error"),
			wantMessage: "",
		},
	}
	handler := process.RespHandler{}

	t.Logf("\tGiven the need to test writing a response.")
	for i, tst := range tt {
		// Create a mock writer
		var mw mockWriter
		t.Logf("\tTest %d:\t%s", i, tst.name)
		{
			t.Logf("\tWhen handling a response: %s", tst.wantMessage)
			if tst.wantErr != nil {
				mw.Err = tst.wantErr
				if err := handler.Write(&tst.resp, &mw); err != nil {
					t.Logf("\t%s\tShould not be able to write a response\n", success)
				} else {
					t.Fatalf("\t%s\tShould not be able to write a response\n", failure)
				}
			} else {
				if err := handler.Write(&tst.resp, &mw); err != nil {
					t.Fatalf("\t%s\tShould be able to write a response: '%v' exp[%s] got[%s]\n",
						failure,
						err,
						tst.wantMessage,
						mw.Buffer.String())
				}
				t.Logf("\t%s\tShould be able to write a response", success)
				got := mw.Buffer.String()
				if got != tst.wantMessage {
					t.Fatalf(
						"*\t%s\tShould be able to write the correct response: [exp] %s, [got] %s*",
						failure,
						tst.wantMessage,
						got,
					)
				}
				t.Logf("\t%s\tShould be able to write the correct response.", success)
			}
		}
	}
}
