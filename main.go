package main

import (
	"fmt"
	"time"
	"runtime"
	"os"
)

var FileLog Filelog

const (
	DbName = "tender"
	Prefix = ""
	UserDb = "root"
	PasswordDb = "1234"
)

func init() {
	CreateLogFile()
	tNow := time.Now()
	tMinus25H := tNow.Add(time.Hour * -25)
	UrlXml = fmt.Sprintf("http://ws-rn.tektorg.ru/export/procedures?start_date=%s.000000", tMinus25H.Format("2006-01-02T15:04:05"))
}

func SaveStack(){
	if p:= recover(); p != nil{
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		file, err := os.OpenFile(string(FileLog), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		defer file.Close()
		if err != nil {
			fmt.Println("Ошибка записи stack log", err)
			return
		}
		fmt.Fprintln(file, fmt.Sprintf("Fatal Error %v", p))
		fmt.Fprintf(file, "%v  ", string(buf[:n]))
	}

}

func main() {
	defer SaveStack()
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
