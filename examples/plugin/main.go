// Package: main
// Description: Simple plugin application
// Author: John Doe
// Version: 1.0.0
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conner80/ipc_plugin"
	"github.com/conner80/ipc_plugin/examples/shared"
	"github.com/google/uuid"
)

// Socket filename for IPC via UNIX-sockets
const socket = "/tmp/ipc_plugin.sock"

// Plugin's engine
var p *ipc_plugin.Plugin

// Entry point
func main() {
	log.Print("plugin starting")
	sock := socket
	if len(os.Args) > 1 {
		sock = os.Args[1]
	}
	var err *ipc_plugin.Error
	p, err = ipc_plugin.NewPlugin(sock, "key", "Simple plugin")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	p.Use("get_user", handleFunc)
	log.Print("plugin started")
	GracefulShutdown()
}

// GracefulShutdown correct shutdown plugin application
func GracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	p.Done()
}

// handleFunc user's method's handler
func handleFunc(req *ipc_plugin.ProtocolMessage, resp *ipc_plugin.ProtocolMessage) error {
	var userID uuid.UUID
	err := req.GetData(&userID)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}

	if userID != uuid.MustParse("ebdb5d69-2f5c-4732-9daa-e1d2359122df") {
		return fmt.Errorf("user not found")
	}

	user := shared.User{
		ID:    uuid.MustParse("ebdb5d69-2f5c-4732-9daa-e1d2359122df"),
		Name:  "User",
		Mail:  "mail@mail.ru",
		Birth: time.Date(1980, time.September, 27, 15, 0, 0, 0, time.UTC),
	}

	if err := resp.SetData(user); err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	return nil
}
