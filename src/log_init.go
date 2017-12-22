package main

import (
	"fmt"
	"os"

	"github.com/cihub/seelog"
)

/* 创建一个新的 logger，可以传递给其他模块使用 */
func NewLogger() (seelog.LoggerInterface, error) {
	logConf := `
	<seelog type="asynctimer" asyncinterval="200" minlevel="debug"
		maxlevel="error">
		<outputs formatid="main">
			<buffered size="10000" flushperiod="200">
				<rollingfile type="date" filename="app.log"
					datepattern="20170113" maxrolls="30"/>
			</buffered>
		</outputs>
		<formats>
			<format id="main" format="%Time|%LEV|%File|%Line|%Msg%n"/>
		</formats>
	</seelog>`

	return seelog.LoggerFromConfigAsBytes([]byte(logConf))
}

/* 初始化 seelog，以后可以直接使用 seelog.Log... 打印信息 */
func LogInit() {
	logConf := `
	<seelog type="asynctimer" asyncinterval="200" minlevel="debug"
		maxlevel="error">
		<outputs formatid="main">
			<buffered size="10000" flushperiod="200">
				<rollingfile type="date" filename="app.log"
					datepattern="20170113" maxrolls="30"/>
			</buffered>
		</outputs>
		<formats>
			<format id="main" format="%Time|%LEV|%File|%Line|%Msg%n"/>
		</formats>
	</seelog>`

	logger, err := seelog.LoggerFromConfigAsBytes([]byte(logConf))
	if err != nil {
		fmt.Printf("LoggerFromConfigAsBytes Error:%s", err.Error())
		os.Exit(-1)
	}

	err = seelog.ReplaceLogger(logger)
	if err != nil {
		fmt.Printf("ReplaceLogger Error!%s\n", err.Error())
		os.Exit(-2)
	}
}


