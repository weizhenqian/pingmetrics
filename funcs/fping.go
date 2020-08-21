package funcs

import (
        "bufio"
        "io"
        "log"
	"fmt"
	"time"
        "os/exec"
        "path/filepath"
        "strings"
        "../g"
	"../basic"
)

// 使用fping做ping监控
func fping(ips []string) {
	hostname, err := g.Endpoint()
	if err != nil {
		log.Println(err)
		return
	}

	fpingCmd := []string{"-b", "0", "-c", "20", "-i", "1", "-p", "3000", "-q"}
	fpingCmd = append(fpingCmd, ips...)
	// fping的动作放到异步去执行
	// fmt.Println(fpingCmd)
	go func(fpingCmd []string, hostname string) {
		fpath := filepath.Join(g.Config().Fpath)
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
		mvs := []*basic.MetricValue{}
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

				LossPk := basic.GaugeValue("ping.loss", lossPk, fmt.Sprintf("ip=%s", ip))
				if len(packge) >= 7 {
					rrt := strings.Split(packge[7], "/")
					minDelay = rrt[0]
					avgDelay = rrt[1]
					maxDelay = rrt[2]
					// log.Debug(ip, minDelay, avgDelay, maxDelay, lossPk)
					MinDelay := basic.GaugeValue("ping.min", minDelay, fmt.Sprintf("ip=%s", ip))
					MaxDelay := basic.GaugeValue("ping.max", maxDelay, fmt.Sprintf("ip=%s", ip))
					AvgDelay := basic.GaugeValue("ping.avg", avgDelay, fmt.Sprintf("ip=%s", ip))
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
		//关闭chan
		close(EndPing)
	}(fpingCmd, hostname)
}
