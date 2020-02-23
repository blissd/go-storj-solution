# golang-storj-solution

Solution for the storj file send interview problem defined [here](https://gist.githubusercontent.com/jtolds/0cde4aa3e07b20d6a42686ad3bc9cb53).

## Package Structure

The project uses the package structure recommended by [Go Best Practices 2016](https://peter.bourgon.org/go-best-practices-2016/#repository-structure).
Command line tools are in their own directories under `cmd/`, and libraries are under `pkg/`.

## Protocol
Each client must talk to two entities. A client first talks to a relay server to establish a transfer session, 
and then talks to its peer client to transfer the files.

Messages are represented as variable size data frames prefixed by a single byte indicating the frame length.
As such the payload to a message frame can be at most 254 bytes, which is long enough to accommodate the session
secret and the file name.

The initial message from a client to the relay server specifies the client type (sender or receiver), but subsequent
messages don't specify any type. Instead, the type is inferred from the message ordering.

The messages are used by clients to establish their connection with the relay server, and then by client to exchange
information about the file (the name and size). After the session and file facts are established, no further messages
are exchanged. The file data is simply streamed from sender to receiver.

### The 'wire' Package
The `wire` package defines functions for encoding and decoding data types into frames. The package defines
a `FrameEncoder` and a `FrameDecoder` which are intended to wrap standard Golang `io.Reader`s and `io.Writer`s.
The use of encoders for framing is inspired by the JSON and XML encoders already present in Golang.

### Sender Message Exchange


## Clients
The sender and receiver clients use the `session` package to communicate with the relay server. The `session` package
is a higher-level wrapper around the `wire` package to provide a more client friendly API. 

## Relay
The relay server defines a `Relay` struct type that is used to handle session establishment and transfers. The
`Relay` is in effect an actor because it has an `actions` channel defined as `action chan func()` which receives 
functions to be executes against the `Relay` state in a synchronised way. Because the state is only processed by the
go routine that consumes from the `actions` channel it isn't necessary to a mutex for guarding updates to
the session state.

This actor pattern was inspired by the talk "Ways To Do Things"
 ([slides](https://speakerdeck.com/peterbourgon/ways-to-do-things) and [video](https://www.youtube.com/watch?v=LHe1Cb_Ud_M)).




