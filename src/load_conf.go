package main

import (
	"io/ioutil"
	"encoding/json"
	"github.com/cihub/seelog"
)

/* 将配置文件 file 中内容读入配置结构体  */
func LoadConfig(file string, entity interface{}) (err error) {

	body, err := ioutil.ReadFile(file)
	if err != nil {
		seelog.Errorf("read file[%s] error[%s]", file, err.Error())
		return
	}

	err = json.Unmarshal(body, &entity)
	if err != nil {
		seelog.Errorf("Unmarshal error[%s]", err.Error())
		return
	}

	return
}
