package ecs

import (
	"bufio"
	"io"

	"golang.org/x/crypto/ssh"
)

var (
	endSymbol = []byte("$$_end_$$")
)

type Ssh struct {
	c         *ssh.Client
	s         *ssh.Session
	stdinPipe io.WriteCloser
	outCh     chan []byte
	signalCh  chan bool
}

func NewSsh(host, rootPwd string) (*Ssh, error) {
	cfg := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.Password(rootPwd)},
	}
	cfg.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	c, err := ssh.Dial("tcp", host+":22", cfg)
	if err != nil {
		return nil, err
	}

	s, err := c.NewSession()
	if err != nil {
		c.Close()
		return nil, err
	}

	stdout, err := s.StdoutPipe()
	if err != nil {
		c.Close()
		return nil, err
	}

	stderr, err := s.StderrPipe()
	if err != nil {
		c.Close()
		return nil, err
	}

	stdin, err := s.StdinPipe()
	if err != nil {
		c.Close()
		return nil, err
	}

	if err := s.Shell(); err != nil {
		c.Close()
		return nil, err
	}

	ssh := &Ssh{
		c:         c,
		s:         s,
		stdinPipe: stdin,
		outCh:     make(chan []byte, 8192),
		signalCh:  make(chan bool),
	}

	go ssh.logLoop(stdout)
	go ssh.logLoop(stderr)

	return ssh, nil
}

func (s *Ssh) logLoop(r io.Reader) {
	br := bufio.NewReaderSize(r, 8192)

	for {
		bytes, err := br.ReadBytes('\n')

		if err != nil {
			if err != io.EOF {
				Error("error reading pipe: %v", err)
			}
			s.signalCh <- true
			return
		}

		bytes = bytes[:len(bytes)-1]
		if len(bytes) == len(endSymbol) && string(bytes) == string(endSymbol) {
			s.signalCh <- true
			return
		} else {
			s.outCh <- bytes
		}
	}
}

func (s *Ssh) Next() []byte {
	return <-s.outCh
}

func (s *Ssh) Run(cmd string) error {
	if _, err := s.stdinPipe.Write([]byte(cmd + "\n")); err != nil {
		return err
	}
	if _, err := s.stdinPipe.Write([]byte("echo -e '\n';echo -e '" + string(endSymbol) + "';echo -e '\n'\n")); err != nil {
		return err
	}
	<-s.signalCh
	return nil
}

func (s *Ssh) Close() error {
	return s.c.Close()
}
