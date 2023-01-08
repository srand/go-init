package utils

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Process struct {
	Command []string
	CGroup  string
	cmd     *exec.Cmd
}

func NewProcess(command []string) *Process {
	return &Process{
		Command: command,
	}
}

func (p *Process) Pid() int {
	if p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return 0
}

func (p *Process) Start() error {
	cmd := []string{"/sbin/init-exec"}
	if p.CGroup != "" {
		cmd = append(cmd, "--cgroup", p.CGroup)
	}
	cmd = append(cmd, "--")
	cmd = append(cmd, p.Command...)

	p.cmd = exec.Command(cmd[0], cmd[1:]...)
	p.cmd.Stderr = os.Stderr
	err := p.cmd.Start()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (p *Process) Terminate() {
	syscall.Kill(p.cmd.Process.Pid, syscall.SIGTERM)
}

func (p *Process) Kill() {
	syscall.Kill(p.cmd.Process.Pid, syscall.SIGKILL)
}
