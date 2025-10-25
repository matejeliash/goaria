package ariarunner

import (
	"context"
	"errors"
	"fmt"
	"goaria/internal/ariarpc"
	"io"
	"log"
	"os/exec"
	"time"
)

// run aria process, termiante when signal is reveived via ctx,
// aria2c process is restarted when aria2c crashed,
// only way to gracefully terminate aria is to sent signal
func RunAriaProcess(ctx context.Context, RpcSecret string, done chan error, ariaClient *ariarpc.AriaClient) {
	go func() {
		for {

			cmd := exec.Command(
				"aria2c",
				"--enable-rpc",
				"--rpc-secret="+RpcSecret,
				"--file-allocation=none",
			)
			// change to this for showing aria2c output to screen
			// cmd.Stdout = os.Stdout
			// cmd.Stderr = os.Stderr

			cmd.Stdout = io.Discard
			cmd.Stderr = io.Discard

			// if app doesn't start wait second and try to start again
			if err := cmd.Start(); err != nil {
				log.Printf("[aria2c] failed to start error: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			ariaDone := make(chan error, 1)
			go func() { ariaDone <- cmd.Wait() }()

			select {
			case <-ctx.Done():
				fmt.Println("[aria2c] ctrl+c detected")
				// []TODO here put shutting down aria RPC command
				ariaClient.ShutdownAriaProcess()

				<-ariaDone // waiting till aria2c process finished gracefully
				done <- errors.New("complete shutdown")

				return
			case err := <-ariaDone:
				fmt.Printf("aria2c crashed: %v\n restarting aria", err)
				time.Sleep(1 * time.Second)
			}

		}

	}()

}
