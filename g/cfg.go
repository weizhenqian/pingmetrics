package g

import (
        "encoding/json"
        "sync"
        "os/exec"
	"strings"
        "github.com/golang/glog"
        "github.com/toolkits/file"
)

// global config file
type GlobalConfig struct {
        Debug      bool            `json:"debug"`
	Fpath      string          `json:"fpath"`
	Identity   *IdentityConfig `json:"identity"`
	Ingore     []string        `json:"ignore"`
}

// identity config file
type IdentityConfig struct {
	Specify		string		`json:"specify"`
	Shell		string		`json:"shell"`
}


var (
        ConfigFile string
        config     *GlobalConfig
        configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
        configLock.RLock()
        defer configLock.RUnlock()
        return config
}

// parse config file.
func ParseConfig(cfg string) {
        if cfg == "" {
                glog.Fatalln("use -c to specify configuration file")
        }

        if !file.IsExist(cfg) {
                glog.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
        }

        ConfigFile = cfg

        configContent, err := file.ToTrimString(cfg)
        if err != nil {
                glog.Fatalln("read config file:", cfg, "fail:", err)
        }

        var c GlobalConfig
        err = json.Unmarshal([]byte(configContent), &c)
        if err != nil {
                glog.Fatalln("parse config file:", cfg, "fail:", err)
        }

        configLock.Lock()
        defer configLock.Unlock()
        config = &c

        glog.Infoln("g:ParseConfig, ok, ", cfg)
}

// endpoint
func Endpoint() (string,error) {
        endpoint := Config().Identity.Specify
	if endpoint != "" {
		return endpoint, nil
	} else {
                cmd := Config().Identity.Shell
                bs, err := exec.Command("bash", "-c", cmd).CombinedOutput()
                if err != nil {
                        return endpoint, err
                } else {
                        endpoint = strings.TrimSpace(string(bs))
                        return endpoint, nil
                }
        }
}
