package common

import (
	"fmt"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"
)

type Conf struct {
	Node node `toml:"node"`
}

type node struct { //配置文件要通过tag来指定配置文件中的名称
	Dns  []string
	Port int
}

func init() {
	var cg Conf
	var cpath = "./apcc.toml"
	if _, err := toml.DecodeFile(cpath, &cg); err != nil {
		log.Panicln(err)
	}
	fmt.Println("Xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", cg)
	logrus.Debugf("读取配置文件%v\n", cg.Node)
	ServerNodeAddrs = cg.Node.Dns
	ListenPort = cg.Node.Port
}
