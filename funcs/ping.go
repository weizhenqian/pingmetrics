package funcs

import (
        "bufio"
		"io"
		"os"
		"encoding/json"
        "log"
        "fmt"
		"sync"
        "time"
        "os/exec"
        "strings"
        "../g"
        "../basic"
)

var wg sync.WaitGroup
var mvs = []*basic.MetricValue{}

func Pings(flag bool) {
	if flag {
		// centos6 系统
		for _, ip := range Iplist {
			go ping(ip)
			wg.Add(1)
		}
		wg.Wait()
		b, err := json.Marshal(mvs)
		if err != nil {
			fmt.Println("error:", err)
		}
		os.Stdout.Write(b)
		close(EndPing)
	}
}

// 使用ping做ping监控
func ping(ip string) {
	defer func() {
		wg.Done()
	}()
	cmd := exec.Command("ping", "-w", "60", "-i", "3", "-c", "20", ip)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}
	cmd.Start()
	hostname, err := g.Endpoint()
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
	LossPk := basic.GaugeValue("ping.loss", lossPk, fmt.Sprintf("ip=%s", ip))
	MinDelay := basic.GaugeValue("ping.min", minDelay, fmt.Sprintf("ip=%s", ip))
	MaxDelay := basic.GaugeValue("ping.max", maxDelay, fmt.Sprintf("ip=%s", ip))
	AvgDelay := basic.GaugeValue("ping.avg", avgDelay, fmt.Sprintf("ip=%s", ip))
	mvs = append(mvs, LossPk, MinDelay, AvgDelay, MaxDelay)
	now := time.Now().Unix()
	for j := 0; j < len(mvs); j++ {
		mvs[j].Step = 60
		mvs[j].Endpoint = hostname
		mvs[j].Timestamp = now

	}
}
