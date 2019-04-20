package cmd

import (
	"fmt"
	"github.com/colindr/gosync/transfer"
	"github.com/spf13/viper"
)

func StartDaemon() {
	addr := fmt.Sprintf("%v:%v", viper.Get("host"), viper.Get("port"))

	fmt.Println("Listening at", addr)

	transfer.Daemon(addr)
}
