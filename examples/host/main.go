// Package: main
// Description: Simple host application
// Author: John Doe
// Version: 1.0.0
package main

import (
	"log"
	"time"

	"github.com/conner80/ipc_plugin"
	"github.com/conner80/ipc_plugin/examples/shared"
	"github.com/google/uuid"
)

const socket = "/tmp/ipc_plugin.sock"

func main() {
	h, err := ipc_plugin.NewHost(socket, 30*time.Second, 5*time.Second, 5*time.Second, false)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	err = h.Ping()
	if err == nil {
		log.Println("ping success")
	} else {
		log.Printf("ping fault: %s\n", err.Error())
	}

	pid, err := h.GetPid()
	if err == nil {
		log.Printf("get PID success: %d", pid)
	} else {
		log.Printf("get PID fault: %s\n", err.Error())
	}

	inf, err := h.GetDescription()
	if err == nil {
		log.Printf("get info success: %s", inf)
	} else {
		log.Printf("get info fault: %s\n", err.Error())
	}

	lst, err := h.GetList()
	if err == nil {
		log.Printf("get methods list success: %v", lst)
	} else {
		log.Printf("get methods list fault: %s\n", err.Error())
	}

	var resp shared.User
	err = h.Call("get_user", uuid.Nil, &resp)
	if err == nil {
		log.Printf("call method 1 success: %v", resp)
	} else {
		log.Printf("call method 1 fault: %s\n", err.Error())
	}

	err = h.Call("get_user", uuid.MustParse("ebdb5d69-2f5c-4732-9daa-e1d2359122df"), &resp)
	if err == nil {
		log.Printf("call method 2 success: %v", resp)
	} else {
		log.Printf("call method 2 fault: %s\n", err.Error())
	}

	pid, err = h.Kill("key")
	if err == nil {
		log.Printf("kill success with PID %d", pid)
	} else {
		log.Printf("kill fault: %s\n", err.Error())
	}
}
