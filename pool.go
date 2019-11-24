package subprocess

import (
	"log"
	"time"
)

// Pool worker pool
type Pool struct {
	cmd          string
	workers      int
	workersQueue chan *Worker
}

// NewPool 创建一个新的worker pool
func NewPool(cmd string, workers int) (*Pool, error) {
	p := &Pool{
		cmd:          cmd,
		workers:      workers,
		workersQueue: make(chan *Worker, workers),
	}
	err := p.initWorkers()
	if err != nil {
		p.Close()
		return nil, err
	}
	return p, nil
}

func (p *Pool) initWorkers() error {
	for workerID := 0; workerID < p.workers; workerID++ {
		log.Printf("[info] Init New Worker(sub process) worker_id=[%d]\n", workerID)

		worker, err := NewWorker(p.cmd, workerID)
		if err != nil {
			log.Println("create new sub process worker failed,worker_id=", workerID)
			return err
		}
		p.workersQueue <- worker
	}
	return nil
}

// Talk 和子进程进行交互，若交互失败，会一直重试直到成功
func (p *Pool) Talk(txt string) string {
	return p.talk(0, txt)
}

func (p *Pool) talk(try uint64, txt string) string {
	worker := <-p.workersQueue

	defer func() {
		p.workersQueue <- worker
	}()

	resp, err := worker.Talk(txt)
	if err == nil {
		return resp
	}
	log.Printf("worker_id=%d,try=%d,worker_err=%s\n", worker.id, try, err.Error())
	time.Sleep(1 * time.Second)
	try = try + 1
	return p.talk(try, txt)
}

// Close 资源回收清理
func (p *Pool) Close() error {
	log.Println("pool closing ...")
	close(p.workersQueue)
	for worker := range p.workersQueue {
		worker.Close()
	}
	return nil
}
