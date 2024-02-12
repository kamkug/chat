# Chat Application

Chat is a CLI-based application that enables users to exchange text messages. The current version supports multiple users.

## Requirements:
- A Go environment to build two binaries: one for the client and one for the server.
- NATS, a messaging service by Golang. You can find the NATS documentation here.

## Architecture:

This project uses a simple client-server architecture. It requires a running NATS server to which our TCP server connects. Clients then reach this server with their messages, which are further processed and actioned by the TCP and NATS servers. Messages are then delivered to clients connected to a `msg` subject. This application utilizes an in-memory cache to manage user sessions. Upon login, user sessions are stored in the cache and promptly removed once the user logs out.

## Installation

1. Make sure you have a recent version of Golang installed. This project is based on the `chat` project by Ardan Labs, which uses a structure that may require Go 1.19 (for some support) and above. The project was developed in Go 1.21.6. Here is the link to install Golang for your specific OS: [Go install]( https://go.dev/doc/install).
2. Install the NATS server with the following command from outside of your module. This will install NATS under your `$GOPATH/bin` location. You can add it to your OS PATH with this command on Linux/MacOS: `export PATH=$GOPATH/bin:$PATH`. For Windows, look online for how to set up your PATH.

	```
	go install github.com/nats-io/gnatsd@latest
	```

	**Notes:**:
	- Since Go 1.17, go get has been deprecated for binary installation. Use go install instead. You can read more about this here [here](https://go.dev/doc/go-get-install-deprecation).
	- If you don't want to add this binary to your PATH, you can change into its directory and run it manually.
3. You can go ahead and clone the git project as long as you have Cisco credentials (this project is public at this moment):
	- Using SSH Key: `git clone git@wwwin-github.cisco.com:kkugler/chat.git`
	- Using HTTPS auth: `git clone https://wwwin-github.cisco.com/kkugler/chat.git`
4. Navigate to the `chat` directory in your cloned repository.
	```
	cd chat
	```
6. Fetch the required dependencies using the following commands:
	```
	go mod tidy
	go mod vendor
	```
6. Create binaries for the client and server with these commands:

	 For the client (this will create a binary named chatd):
	```
	cd cmd/chat
	go build
	```
	For the server (this will create a binary named **chat**):
	```
	cd cmd/chatd
	go build
	```

## Usage

1. Start your NATS server. You should see a message indicating that the server is ready.

	```
	terminal-user% gnatsd
	[96916] 2024/02/04 18:11:58.412758 [INF] Starting nats-server version 1.4.1
	[96916] 2024/02/04 18:11:58.413053 [INF] Git commit [not set]
	[96916] 2024/02/04 18:11:58.413578 [INF] Listening for client connections on 0.0.0.0:4222
	[96916] 2024/02/04 18:11:58.413591 [INF] Server is ready
	```

2. Start your server:

	```
	terminal-user% cd cmd/chatd
	terminal-user% ./chatd
	2024/02/04 18:14:13.829429 main.go:57: Configuration
	 HOST=:6000
	NATS_HOST=nats://localhost:4222

	2024/02/04 18:14:13.831340 main.go:106: main : Waiting for data on: 0.0.0.0:6000
	2024/02/04 18:14:13.831382 socket.go:17: ****> EVENT : IP[ <nil>:6000 ] : EVT[Accept] TYP[Info] : waiting
	2024/02/04 18:14:13.855746 nats.go:150: nats : subject subscribed : Subject[ msg ]
	2024/02/04 18:14:13.855796 nats.go:153: nats : service started : Host[ nats://localhost:4222 ]
	```

3. Connect at least two clients by starting them in separate terminals.

	```
	terminal-user% ./chat
	2024/02/04 18:18:47.593961 main.go:44: Configuration
	 HOST=:6000


	Name:>
	```
	```
	terminal-user% ./chat
	2024/02/04 18:18:47.593961 main.go:44: Configuration
	 HOST=:6000


	Name:>
	```


4. Enter a unique username. The names are stored in an in-memory cache as there is no database used at this stage.
	```
	terminal-user% ./chat
	2024/02/04 18:18:47.593961 main.go:44: Configuration
	 HOST=:6000


	Name:> user-1

	user-1#>
	```

	```
	kkugler@KKUGLER-M-4XWM chat % ./chat
	2024/02/04 18:21:08.654234 main.go:44: Configuration
	 HOST=:6000


	Name:> user-2

	user-2#>
	```
5. You will receive a login notification when other users join the `msg` subject.
6. Send messages. There are currently two ways to send messages:
	- Broadcast messages by just typing your message and pressing ENTER key:
		```

		user-1#> Hello all!

		user-1#>
		```

		This is what everyone connected to this subject will receive:
		```
		user-2#> 2024/02/04 18:22:34.662136 main.go:94:
		{
			Sender: user-1
			Recipient:
			Type: 1
			Data: Hello all!

		}

		user-2#>
		```
	- Targeted messages that would only land with your specified user (if he/she is available), u can do this by starting your message with an `@` symbol and not leaving any whitespace at the beginning of the line (might be improved in the future):
		```
		user-2#> @user-1 hello user-2

		user-2#>
		```

		This is what your peer (`user-1`) in this case will get to see it like:

		```
		user-1#> 2024/02/04 18:24:47.442179 main.go:94:
		{
			Sender: user-2
			Recipient: user-1
			Type: 1
			Data: hello user-1
		}

		user-1#>
		```

	**Notes:**
	- Broadcast messages have an empty Recipient field.
	- Targeted messages have the intended recipient's name in the Recipient field.


## Acknowledgments
I would like to thank `Ardan Labs` and the author of their chat application, which I used as a basis for this project. I added some features and tests based on the TODOs and my own initiative.

## Comments
This application does not yet have a full test suite. Most of the tests reside under the `internal` directory, specifically in the `cache` and `msg` packages. More tests will be added in the future.