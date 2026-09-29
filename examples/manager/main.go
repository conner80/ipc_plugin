// Package: main
// Description: Simple host application with manager
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

func main() {
	log.Println("manager example starting...")
	mgr := ipc_plugin.NewPluginManager("/tmp", []string{"get_user"})
	if err := mgr.LoadPlugin("112233", "key", "bin/plugin", 30*time.Second, 5*time.Second, 5*time.Second); err != nil {
		mgr.Done()
		log.Fatalf("plugin load error: %s", err.Error())
		return
	}
	plg := mgr.GetPlugin("112233")
	if plg != nil {
		var resp shared.User
		err := plg.Call("get_user", uuid.MustParse("ebdb5d69-2f5c-4732-9daa-e1d2359122df"), &resp)
		if err != nil {
			log.Printf("Call method error: %s", err.Error())
		} else {
			log.Printf("call method success: %v", resp)
		}
	}

	mgr.Done()
}
