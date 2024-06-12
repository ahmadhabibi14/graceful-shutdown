package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testgrcsh/configs"
	"time"
)

// "operation()" is a clean up function on shutting down
type operation func(ctx context.Context) error

// "gracefulShutdown()" waits for termination syscalls and doing
// clean up operations after received it
func gracefulShutdown(ctx context.Context, timeout time.Duration, ops map[string]operation) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)

		// add any other syscalls that you want to be notified with
		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-s

		log.Println("shutting down")

		// set timeout for the ops to be done to prevent system hang
		timeoutFunc := time.AfterFunc(timeout, func ()  {
			log.Printf("timeout %d ms has been elapsed, force exit", timeout.Milliseconds())
			os.Exit(0)
		})

		defer timeoutFunc.Stop()

		var wg sync.WaitGroup

		// do the operations asynchronously to save time
		for key, op := range ops {
			wg.Add(1)

			innerOp := op
			innerKey := key

			go func() {
				defer wg.Done()

				log.Printf("cleaning up: %s", innerKey)
				if err := innerOp(ctx); err != nil {
					log.Printf("%s: clean up failed: %s", innerKey, err.Error())
					return
				}

				log.Printf("%s was shutdown gracefully", innerKey)
			}()
		}

		wg.Wait()

		close(wait)
	}()

	return wait
}

func main() {
	// initialize some resources
	configs.LoadEnv()
	db := configs.ConnectPostgresSQL()

	for i := 1; i < 6; i++ {
		log.Printf("[%d] program running", i)
		time.Sleep(1 * time.Second)
	}

	// wait for termination signal and register database & http server clean-up operations
	wait := gracefulShutdown(context.Background(), 2 * time.Second, map[string]operation{
		`database`: func(ctx context.Context) error {
			return db.Close()
		},
	})

	<-wait
}