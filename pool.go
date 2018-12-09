package subprocess

import (
	"log"
	"time"
)

type Pool struct {
	cmd          string
	workers      int
	workersQueue chan *Worker
}

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
	for worker_id := 0; worker_id < p.workers; worker_id++ {
		log.Printf("[info] Init New Worker(sub process) worker_id=[%d]\n", worker_id)

		worker, err := NewWorker(p.cmd, worker_id)
		if err != nil {
			log.Println("create new sub process worker failed,worker_id=", worker_id)
			return err
		}
		p.workersQueue <- worker
	}
	return nil
}

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

func (p *Pool) Close() error {
	close(p.workersQueue)
	for worker := range p.workersQueue {
		worker.Close()
	}
	return nil
}
