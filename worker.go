package subprocess

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"
)

// Worker 工作子进程
type Worker struct {
	cmdStr string
	id     int
	cmd    *exec.Cmd
	reader *bufio.Reader
	writer io.WriteCloser
}

// NewWorker 创建一个新的子进程
func NewWorker(cmdStr string, id int) (*Worker, error) {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return nil, fmt.Errorf("call NewWorker with empty command line,id=%d", id)
	}
	worker := &Worker{
		cmdStr: cmdStr,
		id:     id,
	}
	var err error
	for {
		err = worker.start()
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	return worker, err
}

func (worker *Worker) log(msg ...interface{}) {
	var pid int
	if worker.cmd != nil {
		pid = worker.cmd.Process.Pid
	}
	s := fmt.Sprintf("Worker worker_id=%d cmd=[%s] pid=%d", worker.id, worker.cmdStr, pid)
	sl := fmt.Sprint(msg...)
	log.Println(s, sl)
}

func (worker *Worker) start() (err error) {
	worker.log("Starting ...")

	if worker.cmd != nil && worker.cmd.Process != nil {
		worker.cmd.Process.Kill()
	}

	defer func() {
		if err == nil {
			worker.log("Started Success")
		} else {
			worker.log("Started Failed,error:", err)
		}
	}()

	cmd := exec.Command("sh", "-c", worker.cmdStr)
	//设置进程组
	//	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var stdin io.WriteCloser
	stdin, err = cmd.StdinPipe()
	if err != nil {
		worker.log("cmd.StdinPipe with error:", err)
		return err
	}
	var stdout io.ReadCloser
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		worker.log("cmd.StdoutPipe with error:", err)
		return err
	}

	var erout io.ReadCloser

	erout, err = cmd.StderrPipe()
	if err != nil {
		worker.log("cmd.StderrPipe with error:", err)
		return err
	}

	go func() {
		defer erout.Close()
		berr := bufio.NewReader(erout)
		for {
			l, e := berr.ReadString('\n')
			worker.log("StdErr:", strings.TrimSpace(l), e)
			if e != nil {
				worker.log("StderrPipe Read With error ", e, ",stop read stderr")
				break
			}
		}
	}()

	err = cmd.Start()
	worker.cmd = cmd
	worker.reader = bufio.NewReader(stdout)
	worker.writer = stdin
	return err
}

func (worker *Worker) processExixts() bool {
	return worker.cmd.ProcessState != nil && !worker.cmd.ProcessState.Exited()
}

// Talk 使用子进程对数据进行处理
func (worker *Worker) Talk(request string) (resp string, err error) {
	if strings.Contains(request, "\n") {
		request = strings.Replace(request, "\n", "\\n", -1)
	}
write:
	// 交互协议：输出内容为一行，以\n结束
	writeStr := fmt.Sprintf("%s\n", request)
	var n int
	n, err = io.WriteString(worker.writer, writeStr)
	if err == nil && n != len(writeStr) {
		err = fmt.Errorf("WriteString wrong,expect wrote=%d,now wrote=%d", len(writeStr), n)
	}
	if err != nil {
		worker.log("WriteString error:", err)
		if !worker.processExixts() {
			worker.start()
			goto write
		}
		// 其他未知异常
		return "WriteString (request) error", err
	}
	// 读取处理好的数据
	resp, err = worker.reader.ReadString('\n')
	if err != nil {
		worker.log("read error:", err)
		if !worker.processExixts() {
			worker.start()
			goto write
		}
		return "ReadString (response) error", err
	}

	resp = strings.TrimRight(resp, "\n")
	resp = strings.Replace(resp, "\\n", "\n", -1)
	return
}

// Close 资源回收，包括关闭子进程等逻辑
func (worker *Worker) Close() (err error) {
	worker.log("Closing ......")
	defer func() {
		worker.log("Closed with error:", err)
	}()
	buf := new(bytes.Buffer)
	if worker.processExixts() {
		if err := worker.cmd.Process.Kill(); err != nil {
			worker.log("onClose kill process with error:", err.Error())
			buf.WriteString(err.Error())
		}
	}
	if err := worker.writer.Close(); err != nil {
		worker.log("onClose writer close  with error:", err.Error())
		buf.WriteString(";")
		buf.WriteString(err.Error())
	}

	if buf.Len() == 0 {
		return nil
	}
	err = fmt.Errorf(buf.String())
	return
}
