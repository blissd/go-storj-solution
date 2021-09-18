# golang-storj-solution

Solution for the storj file send interview problem defined [here](https://gist.githubusercontent.com/jtolds/0cde4aa3e07b20d6a42686ad3bc9cb53).

## Package Structure

The project uses the package structure recommended by [Go Best Practices 2016](https://peter.bourgon.org/go-best-practices-2016/#repository-structure).
Command line tools are in their own directories under `cmd/`, and libraries are under `pkg/`.

## Protocol
Each client must talk to two entities--the relay server and a peer client. A client first talks to a relay server
to establish a transfer session, and then talks to its peer client to transfer the files.

Clients send and receive messages that represent single fields. Each message starts with a single byte indicating
the field type, followed by data for the field. There are three data types supported:

1. 'b' for sending a single byte.
2. 'B' for sending a short string of up to 255 bytes.
3. 's' for sending a stream of bytes.

After clients have been connected via the relay server the sender will send both the file name and file size to the
receiver. The file size is sent so the receiver can determine if the full file has been received from the sender.
Without the file size a partial send by the sender would not be detected by the receiver because the relay server 
doesn't inform clients of any error conditions.

## The `wire` Package
The `wire` package defines functions for encoding and decoding data types into frames. The package defines
an `Encoder` and a `Decoder` which are intended to wrap standard Golang `io.Reader`s and `io.Writer`s.
The use of encoders is inspired by the JSON and XML encoders already present in Golang.

## The `client` Package
The sender and receiver clients use the `client` package to communicate with the relay server. The `client` package
is a higher-level thin wrapper around the `wire` package to provide a more client friendly API. 

## Relay Server
The relay server defines a `relay` struct type that is used to handle session establishment and transfers. The
`relay` is in effect an actor because it has an `actions` channel, defined as `action chan func()`, which receives 
functions to be executed against the `relay` state in a synchronised way. Because the state is only processed by the
go routine that consumes from the `actions` channel it isn't necessary to a mutex for guarding updates to
the session state.

This actor pattern was inspired by the talk "Ways To Do Things"
 ([slides](https://speakerdeck.com/peterbourgon/ways-to-do-things) and [video](https://www.youtube.com/watch?v=LHe1Cb_Ud_M)).

The `secrets` interface is for generating secrets. There are two secret generates: one that always generates the same
secret and was for testing purposes, and another that generates a six character pseudo-random secret.

## Shortcomings to be Addressed

If a receiver never connects to a waiting sender session, then the session lingers in the relay server forever.

If a receiver connects and doesn't consume data, then the session will linger in the relay server forever.

The relay server doesn't inform clients of errors, it just closes the client connection.
