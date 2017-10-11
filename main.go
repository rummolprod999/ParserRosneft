package main

import (
	"fmt"
	"time"
)

var FileLog Filelog

const (
	DbName = "tenders_test"
	Prefix = ""
	UserDb = "root"
	PasswordDb = "Dft56Point"
)

func init() {
	CreateLogFile()
	tNow := time.Now()
	tMinus25H := tNow.Add(time.Hour * -25)
	UrlXml = fmt.Sprintf("http://ws-rn.tektorg.ru/export/procedures?start_date=%s.000000", tMinus25H.Format("2006-01-02T15:04:05"))
}

func main() {
	Logging("Начало парсинга")
	count := 0
	for {
		if HasMoreProcedures == 0 || count > 20{
			break
		}
		count++
		Logging("")
		Parser()
	}
	//Parser()
	Logging("Конец парсинга")
	Logging(fmt.Sprintf("Добавили тенедеров %d", Addtender))
}
