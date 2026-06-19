// package quartz implements the Evertz Quartz Router Control Protocol (RCP)
// for controlling routing crosspoints.
//
// Quartz is a serialized protocol, so each command must be responded to before
// processing subsequent commands.
//
// The protocol itself uses TCP/IP stream sockets so a [net.Conn] is used as
// the transport.
//
// For client implementations, commands are FIFO queued and written to the transport.
// Response messages are parsed and decoded, and the decoded
// data structure passed to the relevant command handler for application code
// consumption. Update messages may be received from the remote without a command being sent.
// In both cases [AddCommandHandler] may be used to add a handler for the command string.
//
// For server implementations, received commands are queued and responses are
// sent in the order received. Incomming commands are parsed and decoded, and
// the decoded data structure passed to the relevant command handler for application
// code consumption (see [AddCommandHandler]).
//
// Clients and servers implement their own versions of each handler to
// accept a decoded message from the protocol and perform business logic.
//
// Example:
//
//	client := net.Dial("tcp", "192.168.2.100:23")
//	p := quartz.NewProtocol(conn, slog.Default(), nil)
//	p.Start()
//	defer p.Stop()
package quartz
