// package quartz provides the Quartz Protocol implementation
//
// Incoming data is parsed for valid commands and passed along to the handler,
// which decodes the command and calls the registered command handler.
// Clients and servers implement their own versions of command handlers to
// perform actions based on the commands.
package quartz
