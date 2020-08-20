package funcs

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"../basic"
	"../g"
)

var fpath string

func PingMetrics() {
	RemoteIpLock.RLock()
	defer RemoteIpLock.RUnlock()
	iplist := RemoteIpList
	// 这里使用的是falcon agent自带的fping工具，版本是Version 3.10，centos6的统不支持此版本
	// 所以在调用fping -v 错误的情况下，就使用ping做监控，默认使用fping做监控，省系统资源
	// centos6的系统fping工具不支持具体延迟的功能，所以还需要使用ping
	fpath = filepath.Join(g.Config().Plugin.Dir, g.FPING_PATH)
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
			fping(iplist)
			flag = false
		}
	}
	cmd.Wait()
	if flag {
		// centos6 系统
		for _, ip := range iplist {
			go ping(ip)
		}
	}
}

// 使用fping做ping监控
func fping(ips []string) {
	endpoint, err := g.Endpoint()
	if err != nil {
		log.Println(err)
		return
	}

	fpingCmd := []string{"-b", "0", "-c", "20", "-i", "1", "-p", "3000", "-q"}
	fpingCmd = append(fpingCmd, ips...)
	// fping的动作放到异步去执行
	// fmt.Println(fpingCmd)
	go func(fpingCmd []string, hostname string) {
		fpath = filepath.Join(g.Config().Plugin.Dir, g.FPING_PATH)
		cmd := exec.Command(fpath, fpingCmd...)
		// fping的最后输出是打印到了stderr里面的
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Println(err)
			return
		}
		cmd.Start()
		reader := bufio.NewReader(stderr)
		minDelay := "0"
		avgDelay := "0"
		maxDelay := "0"
		lossPk := "100"
		ip := ""
		mvs := []*MetricValue{}
		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}

			if strings.Contains(line, "xmt/rcv/%loss") {
				packge := strings.Fields(line)
				ip = packge[0]
				lossPkS := strings.Split(packge[4], "/")[2]
				lossPk = strings.Split(lossPkS, "%")[0]

				LossPk := GaugeValue("ping.loss", lossPk, fmt.Sprintf("ip=%s", ip))
				if len(packge) >= 7 {
					rrt := strings.Split(packge[7], "/")
					minDelay = rrt[0]
					avgDelay = rrt[1]
					maxDelay = rrt[2]
					// log.Debug(ip, minDelay, avgDelay, maxDelay, lossPk)
					MinDelay := GaugeValue("ping.min", minDelay, fmt.Sprintf("ip=%s", ip))
					MaxDelay := GaugeValue("ping.max", maxDelay, fmt.Sprintf("ip=%s", ip))
					AvgDelay := GaugeValue("ping.avg", avgDelay, fmt.Sprintf("ip=%s", ip))
					mvs = append(mvs, LossPk, MinDelay, AvgDelay, MaxDelay)
				} else {
					mvs = append(mvs, LossPk)
				}
			}
		}
		cmd.Wait()
		now := time.Now().Unix()
		for j := 0; j < len(mvs); j++ {
			mvs[j].Step = 60
			mvs[j].Endpoint = hostname
			mvs[j].Timestamp = now

		}
		fmt.Println("ping result:", mvs)
	}(fpingCmd, hostname)
}

// 使用ping做ping监控
func ping(ip string) {
	cmd := exec.Command("ping", "-w", "60", "-i", "3", "-c", "20", ip)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}
	cmd.Start()
	hostname, err := g.Hostname()
	if err != nil {
		return
	}
	reader := bufio.NewReader(stdout)
	minDelay := "0"
	avgDelay := "0"
	maxDelay := "0"
	lossPk := "100"
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		if strings.Contains(line, "packets transmitted") {
			packge := strings.Fields(line)
			lossPk = strings.Split(packge[5], "%")[0]
		}
		if strings.Contains(line, "rtt min/avg/max/mdev") {
			rrttmp := strings.Fields(line)
			rrt := strings.Split(rrttmp[3], "/")
			minDelay = rrt[0]
			avgDelay = rrt[1]
			maxDelay = rrt[2]
		}
	}
	cmd.Wait()
	LossPk := GaugeValue("ping.loss", lossPk, fmt.Sprintf("ip=%s", ip))
	MinDelay := GaugeValue("ping.min", minDelay, fmt.Sprintf("ip=%s", ip))
	MaxDelay := GaugeValue("ping.max", maxDelay, fmt.Sprintf("ip=%s", ip))
	AvgDelay := GaugeValue("ping.avg", avgDelay, fmt.Sprintf("ip=%s", ip))
	mvs := []*MetricValue{LossPk, MinDelay, MaxDelay, AvgDelay}
	now := time.Now().Unix()
	for j := 0; j < len(mvs); j++ {
		mvs[j].Step = 60
		mvs[j].Endpoint = hostname
		mvs[j].Timestamp = now

	}
	fmt.Println("ping result:", mvs)
}
