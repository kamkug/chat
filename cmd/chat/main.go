package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"chat/internal/msg"

	"github.com/ardanlabs/kit/cfg"
)

/*
Start the Client:
CHAT_HOST=":6000" ./chat
*/

// Configuation settings.
const configKey = "CHAT"

func init() {

	// Setup default values that can be overridden in the env.
	if _, b := os.LookupEnv("CHAT_HOST"); !b {
		os.Setenv("CHAT_HOST", ":6000")
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)
}

func main() {

	// =========================================================================
	// Init the configuration system.

	if err := cfg.Init(cfg.EnvProvider{Namespace: configKey}); err != nil {
		log.Println("Error initalizing configuration system", err)
		os.Exit(1)
	}

	log.Println("Configuration\n", cfg.Log())

	// Get configuration.
	host := cfg.MustString("HOST")

	// =========================================================================
	// Connect and get going.

	// Let's connect back and send a TCP package
	conn, err := net.Dial("tcp4", host)
	if err != nil {
		log.Println("dial", err)
	}

	// Accept keyboard input.
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\nName:> ")
	name, _ := reader.ReadString('\n')
	name = name[:len(name)-1]

	// Show online.
	mSend := msg.MSG{
		Sender:    name,
		Recipient: "",
		Type:      msg.Init,
		Data:      fmt.Sprintf("%s is online", name),
	}
	data := msg.Encode(mSend)
	if _, err := conn.Write(data); err != nil {
		log.Println("write", err)
	}

	// Receiving goroutine.
	go func() {
		for {
			data, _, err := msg.Read(conn)
			if err != nil {
				log.Println("read", err)
				return
			}

			mRecv := msg.Decode(data)

			if mRecv.Type == msg.InCache {
				fmt.Printf("\nUsername '%s' is currently connected.\n", mRecv.Sender)
				fmt.Println("Please try a different username on next run.")
				os.Exit(1)
			}

			log.Println(mRecv)
			fmt.Printf("\n%s#> ", name)
		}
	}()

	// Process keyboard input.
	go func() {
		for {
			fmt.Printf("\n%s#> ", name)
			message, _ := reader.ReadString('\n')

			mSend := msg.MSG{
				Sender:    name,
				Recipient: msg.GetRecipient(message),
				Type:      msg.Message,
				Data:      msg.GetData(message),
			}

			data := msg.Encode(mSend)

			if _, err := conn.Write(data); err != nil {
				log.Println("write", err)
			}
		}
	}()

	// Listen for an interrupt signal from the OS.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	// Show offline.
	mSend = msg.MSG{
		Sender:    name,
		Recipient: "",
		Type:      msg.Message,
		Data:      fmt.Sprintf("%s is offline", name),
	}
	data = msg.Encode(mSend)
	if _, err := conn.Write(data); err != nil {
		log.Println("write", err)
	}
}
