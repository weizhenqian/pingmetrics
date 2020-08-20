package funcs

import (
	"github.com/toolkits/nux"
	"../g"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

var STATE = map[string]string{
	"01": "ESTABLISHED",
	"02": "SYN_SENT",
	"03": "SYN_RECV",
	"04": "FIN_WAIT1",
	"05": "FIN_WAIT2",
	"06": "TIME_WAIT",
	"07": "CLOSE",
	"08": "CLOSE_WAIT",
	"09": "LAST_ACK",
	"0A": "LISTEN",
	"0B": "CLOSING",
}

var (
	RemoteIpList []string
	RemoteIpLock = new(sync.RWMutex)
)

func UpdateSSNetState() {
	portList, err := nux.TcpPorts()
	if err != nil {
		log.Println("ERROR: Get listen port faild: ", err)
	}
	//把本机监听的端口放到map里面，方便后面使用
	var portMap = make(map[int64]struct{})
	var emptys struct{}
	for _, port := range portList {
		portMap[port] = emptys
	}
	lines4 := loadData("/proc/net/tcp")
	lines6 := loadData("/proc/net/tcp6")
	lines := append(lines4, lines6...)
	ip := g.Endpoint()
	remoteIpSet := SetNew()
	for _, line := range lines {
		//fmt.Println(line)
		l := removeAllSpace(strings.Split(line, " "))
		LocalIp, LocalPort := convert2IpPort(l[1])
		port, _ := strconv.ParseInt(LocalPort, 10, 64)

		//如果连接的local port和本地监听的端口一直，说明是其他机器调用本机服务，
		//这种不需要往回ping，所以过滤掉
		if _, ok := portMap[port]; ok {
			continue
		}

		RemoteIp, _ := convert2IpPort(l[2])
		State := STATE[l[3]]

		if RemoteIp == ip {
			continue
		}
		if RemoteIp == "127.0.0.1" {
			continue
		}
		if RemoteIp == "0.0.0.0" {
			continue
		}
		if RemoteIp == "::1" {
			continue
		}
		if State == "LISTEN" {
			continue
		}
		//排除config配置中忽略的IP列表
		Ignorelist := g.GlobalConfig.Ingore
		if IsContain(Ignorelist,RemoteIp) {
			continue
		}

		//只过滤通过本地IP去调用服务的IP，有的机器多网卡，只会检测一块网卡，
		//因为有的公网机器ping公网地址没有意义
		if LocalIp == ip {
			remoteIpSet.Add(RemoteIp)
		}

	}

	if Config().Debug {
		log.Println("INFO: Ping ip list: ", remoteIpSet.List())
	}
	updateRemoteIpList(remoteIpSet.List())

}

func updateRemoteIpList(list []string) {
	RemoteIpLock.RLock()
	defer RemoteIpLock.RUnlock()
	RemoteIpList = list
}

func loadData(tcpFile string) []string {

	var str []string

	fin, err := os.Open(tcpFile)

	defer fin.Close()

	if err != nil {
		fmt.Println(tcpFile, err)
		return str
	}

	r := bufio.NewReader(fin)

	for {
		buf, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		str = append(str, buf)
	}

	if len(str) > 0 {
		return str[1:]
	}
	return str
}

func hex2Dec(hexstr string) string {
	i, _ := strconv.ParseInt(hexstr, 16, 0)
	return strconv.FormatInt(i, 10)
}

func hex2Ip(hexstr string) (string, string) {
	var ip string
	if len(hexstr) != 8 && len(hexstr) != 32 {
		err := "parse error"
		return ip, err
	}

	if len(hexstr) == 32 {
		hexstr = string(hexstr[24:32])
	}

	i1, _ := strconv.ParseInt(hexstr[6:8], 16, 0)
	i2, _ := strconv.ParseInt(hexstr[4:6], 16, 0)
	i3, _ := strconv.ParseInt(hexstr[2:4], 16, 0)
	i4, _ := strconv.ParseInt(hexstr[0:2], 16, 0)
	ip = fmt.Sprintf("%d.%d.%d.%d", i1, i2, i3, i4)

	return ip, ""
}

func convert2IpPort(str string) (string, string) {
	l := strings.Split(str, ":")
	if len(l) != 2 {
		return str, ""
	}

	ip, err := hex2Ip(l[0])
	if err != "" {
		return str, ""
	}

	return ip, hex2Dec(l[1])
}

func removeAllSpace(l []string) []string {
	var ll []string
	for _, v := range l {
		if v != "" {
			ll = append(ll, v)
		}
	}

	return ll
}
