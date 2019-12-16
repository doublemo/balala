// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package process

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRuntimeContainer(t *testing.T) {
	n := NewRuntimeContainer()
	n.Add(&RuntimeActor{
		Exec: func() error {
			return nil
		},

		Interrupt: func(err error) {
			fmt.Println("TEST1:", err)
		},
		Close: func() {
			fmt.Println("Close:", "TEST1")
		},
	})

	n.Add(&RuntimeActor{
		Exec: func() error {
			return errors.New("TEST2")
		},

		Interrupt: func(err error) {
			fmt.Println("TEST2:", err)
		},

		Close: func() {
			fmt.Println("Close:", "TEST2")
		},
	})

	go func() {
		x := make(chan struct{})
		n.Add(&RuntimeActor{
			Exec: func() error {
				for {
					select {
					case <-x:
						return errors.New("TEST3")
					}
				}
			},

			Interrupt: func(err error) {
				fmt.Println("TEST3:", err)
			},

			Close: func() {
				fmt.Println("Close:", "TEST3")
				close(x)
			},
		})
	}()
	time.Sleep(time.Millisecond * 100)
	go func() {
		time.Sleep(time.Second * 4)
		n.Stop()
		fmt.Println("stop:", "ok3")
	}()

	if err := n.Run(); err != nil {
		fmt.Printf("Run: %v\n", err)
	}

	time.Sleep(time.Second * 5)
}
