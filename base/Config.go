package base

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

var Config = struct {
	Debug        bool
	RunLocal     bool
	ClusterId    string `yaml:"clusterId"`
	ServerId     string `yaml:"serverId"`
	ClusterTimer int    `yaml:"clusterTimer"`
	Http         struct {
		Local     string
		Proxy     string
		Static    int //Static Server listen port
		Cert, Key string
		LocalUrl  struct {
			TLS  bool
			Port int
		}
	}
	Fs struct {
		Root       string
		RatedSpace int    `yaml:"ratedSpace"`
		Meta       string `yaml:"meta"`
		Tasks      string
		Notify     string
	}
	Redis struct {
		Addr     string
		Password string
		Db       string
		// Key      string
	}
	Writer struct {
		Alg    string
		Params map[string]interface{}
	}
}{}

func init() {
	configFile := "conf.yml"
	var bs []byte
	var err error
	for i := 0; i < 5; i++ {
		bs, err = ioutil.ReadFile(configFile)
		if err == nil {
			break
		}
		configFile = "../" + configFile
	}
	fmt.Println("find conf.yml at:" + configFile)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(bs, &Config)
	if err != nil {
		panic(err)
	}
	err = checkServer()
	if nil != err {
		panic(err)
	}
}

func checkServer() error {
	localUrl, err := url.Parse(Config.Http.Local)
	if err != nil {
		return err
	}
	if "https" == localUrl.Scheme {
		Config.Http.LocalUrl.TLS = true
	}
	if "" == localUrl.Port() {
		if "http" == localUrl.Scheme {
			Config.Http.LocalUrl.Port = 80
		}
		if "https" == localUrl.Scheme {
			Config.Http.LocalUrl.Port = 443
		}
	} else {
		Config.Http.LocalUrl.Port, err = strconv.Atoi(localUrl.Port())
		if err != nil {
			return err
		}
	}
	return nil
}

func AbsRoot() string {
	r, _ := filepath.Abs(Config.Fs.Root)
	return r
}
func AbsPath(relativePath string) (string, error) {
	r, err := filepath.Abs(Config.Fs.Root)
	if err != nil {
		return "", err
	}
	if 0 == strings.Index(relativePath, r) {
		return relativePath, nil
	}
	return filepath.Join(r, relativePath), nil
}
func AbsPathMust(relativePath string) string {
	abs, _ := AbsPath(relativePath)
	return abs
}

func RelPath(absPath string) string {
	prefix, _ := filepath.Abs(Config.Fs.Root)
	i := strings.Index(absPath, prefix)
	if 0 != i {
		return absPath
	}
	return absPath[len(prefix):]
}
