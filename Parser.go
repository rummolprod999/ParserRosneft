package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var FileProt string
var fl FileProtocols

func DownLoadFile() string {
	tNow := time.Now()
	tMinus3H := tNow.Add(time.Hour * -4)
	s := fmt.Sprintf("http://ws-rn.tektorg.ru/export/procedures?start_date=%v.000000%%2b03:00&end_date=%v.000000%%2b03:00", tMinus3H.Format("2006-01-02T15:04:05"), tNow.Format("2006-01-02T15:04:05"))
	//fmt.Println(s)
	count := 0
	for {
		if count > 20 {
			Logging(fmt.Sprintf("Не скачали файл за %d попыток", count))
			return ""
		}
		resp, err := http.Get(s)
		if err != nil {
			Logging("Ошибка скачивания", s, err)
			count++
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			count++
			continue
		}
		resp.Body.Close()
		return string(body)
	}
}

func Parser() {
	s := DownLoadFile()
	if s == "" {
		Logging("Получили пустую строку")
		return
	}
	if err := xml.Unmarshal([]byte(s), &fl); err != nil {
		Logging("Ошибка при парсинге строки", err)
		return
	}
	if fl.Has_more_procedures == 1 {
		Logging("Слишком много процедур в запросе")
	}
	for _, r := range fl.Protocols {
		ParserProtocol(r)
	}

}
func ParserProtocol(p Protocol) {
	fmt.Println(p.RegistryNumber)
}
