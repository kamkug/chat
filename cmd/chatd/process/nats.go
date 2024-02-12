package process

/*
https://github.com/nats-io/nats.go

# Server
go install github.com/nats-io/gnatsd@latest

# Run the server
gnatsd
*/

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"chat/internal/msg"
	"chat/internal/platform/cache"

	"github.com/ardanlabs/kit/tcp"
	nats "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// natsProcess handles the messages that are consumed from nats.
func natsProcess(cc *cache.Cache, nts *NATS, t *tcp.TCP, nm *nats.Msg) {
	switch nm.Subject {
	case natsSubject:

		// Decode the message received.
		id, m := natsDecode(nm.Data)
		log.Printf("Nats_Process : IP[ nats ] : Inbound : ID[ %s ]%v\n", id, m)

		d := msg.Encode(m)

		if m.Type == msg.InCache {
			tcpValues := strings.Split(m.Data, ":")
			ipv4 := net.ParseIP(tcpValues[0])
			port, err := strconv.Atoi(tcpValues[1])
			if err != nil {
				log.Printf("Nats_Process : IP[ %s ] : ERROR : Port : %s\n", ipv4, err)
			}
			log.Printf("Nats_Process : IP[ %s ] Port [%d] : InCache : Client [ %s ] already in cache\n", ipv4, port, m.Recipient)

			forwardTCPResponse(ipv4, port, d, t)
			return
		}

		// Select clients to send this message towards otherwise.
		for _, client := range cc.Get(m.Sender) {
			ipAddress := client.TCPAddr.IP.String()
			if m.Recipient != "" && m.Recipient != client.ID {
				log.Printf("Nats_Process : IP[ %s ] : Skip : client[ %s ]\n", ipAddress, client.ID)
				continue
			}

			log.Printf("Nats_Process : IP[ %s ] : Send : client[ %s ]\n", ipAddress, client.ID)
			forwardTCPResponse(client.TCPAddr.IP, client.TCPAddr.Port, d, t)
		}

	default:
		log.Printf("Nats_Process : IP[ nats ] : Inbound : Unknown Subject[ %s ]\n", nm.Subject)
	}
}

// Prepares and sends a TCP response.
func forwardTCPResponse(ipv4 net.IP, port int, d []byte, t *tcp.TCP) {
	resp := tcp.Response{
		TCPAddr: &net.TCPAddr{
			IP:   ipv4,
			Port: port,
		},
		Data:   d,
		Length: len(d),
	}
	if err := t.Send(context.TODO(), &resp); err != nil {
		log.Printf("Socket_Process : IP[ %s ] : ERROR : Send : %s\n", ipv4, err)
	}
}

// =============================================================================

// Nats subjects.
const (
	natsSubject = "msg" // Handling based communication.
)

// NATSConfig represents required configuration for the nats system.
type NATSConfig struct {
	Host string
	CC   *cache.Cache
	TCP  *tcp.TCP
}

// NATS represents a nats system from message handling.
type NATS struct {
	Config NATSConfig

	id   string
	conn *nats.Conn
	subs map[string]*nats.Subscription
}

// StartNATS initializes access to a nats system.
func StartNATS(cfg NATSConfig) (*NATS, error) {

	// Set nats options for connection.
	opts := nats.Options{
		Url:            cfg.Host,
		AllowReconnect: true,
		MaxReconnect:   -1,
		ReconnectWait:  time.Second,
		Timeout:        5 * time.Second,
	}

	// Connect to the specified nats server.
	conn, err := opts.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "connecting to NATS")
	}

	// Construct the nats value.
	nts := NATS{
		Config: cfg,
		id:     uuid.NewV1().String(),
		conn:   conn,
		subs:   make(map[string]*nats.Subscription),
	}

	// Declare the event handler for handling recieved messages.
	f := func(msg *nats.Msg) {
		natsProcess(cfg.CC, &nts, cfg.TCP, msg)
	}

	// Register the event handler for each known subject.
	for _, subject := range []string{natsSubject} {

		// Subscribe to receive messages for the specified subject.
		sub, err := conn.Subscribe(subject, f)
		if err != nil {
			return nil, errors.Wrapf(err, "subscribing to subject : %s", subject)
		}

		// Save the subscription with its associated subject.
		nts.subs[subject] = sub
		log.Printf("nats : subject subscribed : Subject[ %s ]\n", subject)
	}

	log.Printf("nats : service started : Host[ %s ]\n", cfg.Host)
	return &nts, nil
}

// Stop shutdowns access to the nats system.
func (nts *NATS) Stop() {
	if nts == nil {
		log.Println("nats : WARNING : nats was not initialized")
		return
	}

	if nts.subs != nil {

		// Go through each subscription and unsubscribe.
		for subject, subscription := range nts.subs {
			if err := subscription.Unsubscribe(); err != nil {
				log.Printf("nats : ERROR : unsubscribe : subject[ %s ] : %v\n", subject, err)
				continue
			}

			log.Printf("nats : unsubscribed : subject[ %s ]\n", subject)
		}
	}

	log.Printf("nats : service stoped : Host[ %s ]\n", nts.Config.Host)
}

// SendMsg publishes the nats  to other Tea services.
func (nts *NATS) SendMsg(m msg.MSG) error {
	log.Printf("Nats_Process : IP[ nats ] : Outbound : Sending To NATS : %v\n", m)
	return nts.conn.Publish(natsSubject, nts.natsEncode(m))
}

// ledID represents the length of the UUID based string we use for the id.
const lenID = 36

// natsEncode encodes the natsMsg so it can be sent to other Chat services.
func (nts *NATS) natsEncode(m msg.MSG) []byte {

	// Encode the message into bytes.
	mData := msg.Encode(m)

	// Create a slice large enough to hold all the data.
	ld := len(mData)
	data := make([]byte, lenID+ld)

	// Copy the all the data for delivery.
	copy(data, nts.id)
	copy(data[lenID:], mData)

	return data
}

// natsDecode decodes the byte data into a msg.MSG.
func natsDecode(data []byte) (string, msg.MSG) {

	// Decode the part of the data that represents the id.
	id := string(data[:lenID])

	return id, msg.Decode(data[lenID:])
}
