package funcs

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"../g"
)

var fpath string
var EndPing = make(chan bool)
var flag bool
var Iplist []string

func PingMetrics() {
	RemoteIpLock.RLock()
	defer RemoteIpLock.RUnlock()
	Iplist = RemoteIpList
	// 这里使用的是falcon agent自带的fping工具，版本是Version 3.10，centos6的统不支持此版本
	// 所以在调用fping -v 错误的情况下，就使用ping做监控，默认使用fping做监控，省系统资源
	// centos6的系统fping工具不支持具体延迟的功能，所以还需要使用ping
	fpath = filepath.Join(g.Config().Fpath)
	cmd := exec.Command(fpath, "-v")
	// fping的最后输出是打印到了stderr里面的
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	flag := true
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		if strings.Contains(line, g.FPING_VERSION) {
			// centos7 系统
			fping(Iplist)
			flag = false
		}
	}
	cmd.Wait()
	Pings(flag)
	for _ = range EndPing {}
}
