package main

import (
	"flag"
	"fmt"
	"github.com/hidu/subprocess"
	"log"
	"sync"
	"time"
)

var cmd = flag.String("cmd", "php pong.php", "sum process")

func main() {
	flag.Parse()
	workersPool, err := subprocess.NewPool(*cmd, 10)
	if err != nil {
		log.Fatalln("create NewPool failed:", err)
	}
	defer workersPool.Close()

	var wg sync.WaitGroup

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := fmt.Sprintf("ping %d", id)
			resp := workersPool.Talk(req)
			log.Println(req, "\t---->\t", resp)
		}(i)
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	log.Println("finish")
}
