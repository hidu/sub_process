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
			//主程序向子程序发送字符串 ping {ID}，
			//如 "ping 1",子程序对字符串解析后，返回内容为 "pong:ping 1"
			req := fmt.Sprintf("ping %d", id)
			resp := workersPool.Talk(req)
			log.Println(req, "\t---->\t", resp)
		}(i)
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	log.Println("finish")
}
