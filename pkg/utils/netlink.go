package utils

import (
	"bytes"
	"strings"
	"syscall"
)

type UEventConn struct {
	sock int
}

type UEvent struct {
	Name string
	Env  map[string]string
}

func NewUEvent(buf []byte) *UEvent {
	event := &UEvent{
		Env: map[string]string{},
	}

	fields := bytes.Split(buf, []byte{0})

	if len(fields) > 0 {
		event.Name = string(fields[0])
		for _, field := range fields[1:] {
			field := string(field)
			if key, val, found := strings.Cut(field, "="); found {
				event.Env[key] = val
			}
		}
	}

	return event
}

func (c *UEventConn) Dial() error {
	sock, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_KOBJECT_UEVENT)
	if err != nil {
		return err
	}

	addr := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: syscall.NETLINK_KOBJECT_UEVENT,
	}

	if err := syscall.Bind(sock, addr); err != nil {
		syscall.Close(sock)
	}

	c.sock = sock

	return nil
}

// Close allow to close file descriptor and socket bound
func (c *UEventConn) Close() error {
	return syscall.Close(c.sock)
}

func (c *UEventConn) ReadEvent() (*UEvent, error) {
	buf := make([]byte, 0x10000)

	n, _, err := syscall.Recvfrom(c.sock, buf, 0)
	if err != nil {
		return nil, err
	}

	return NewUEvent(buf[:n]), nil
}
