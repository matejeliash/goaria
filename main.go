package main

import (
	"context"
	_ "embed"
	"goaria/internal/ariarpc"
	"goaria/internal/ariarunner"
	"goaria/internal/server"
	"os"
	"os/signal"
	"syscall"
)

const aria2URL = "http://localhost:6800/jsonrpc"
const rpcToken = "mysecret"

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rpcSecret := "mysecret"

	done := make(chan error, 1)

	ariaClient := ariarpc.NewAriaClient(rpcSecret)

	ariarunner.RunAriaProcess(ctx, rpcSecret, done, ariaClient)

	server := server.NewServer("44444", ariaClient)
	server.Run(done)

}
