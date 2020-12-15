package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestDiskWrite(t *testing.T) {
	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(number int) {
			f, err := os.OpenFile("/tmp/test.file.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				t.Fatal(err)
			}

			for j := 0; j < 1000; j++ {
				r := strings.NewReader(fmt.Sprintf("goroutine: %d, loop: %d\n", number, j))
				_, err = io.Copy(f, r)
				if err != nil {
					t.Fatal(err)
				}
			}
			err = f.Close()
			if err != nil {
				t.Fatal(f)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
