package msg_test

import (
	"testing"

	"chat/internal/msg"
)

const succeed = "\u2713"
const failed = "\u2717"

// TestEncode test that the encoding of a message works.
func TestEncode(t *testing.T) {
	tt := []struct {
		name   string
		m      msg.MSG
		length int
	}{
		{
			name: "length",
			m: msg.MSG{
				Sender:    "BillKenned",
				Recipient: "JillKenned",
				Type:      msg.Message,
				Data:      "hello",
			},
			length: 29,
		},
		{
			name: "shortname",
			m: msg.MSG{
				Sender:    "Bill",
				Recipient: "Cory",
				Type:      msg.Message,
				Data:      "helloworld",
			},
			length: 34,
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
				if m.Sender != tst.m.Sender {
					t.Fatalf("\t%s\tShould have the correct Sender : exp[%v] got[%v]\n", failed, tst.m.Sender, m.Sender)
				}
				t.Logf("\t%s\tShould have the correct Sender.\n", succeed)

				if m.Recipient != tst.m.Recipient {
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
