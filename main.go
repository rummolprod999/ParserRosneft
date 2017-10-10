package main

import "fmt"

var FileLog Filelog

const (
	DbName = "tender"
	Prefix = ""
)

func init() {
	CreateLogFile()
}

func main() {
	Logging("Начало парсинга")
	Parser()
	Logging("Конец парсинга")
	Logging(fmt.Sprintf("Добавили тенедеров %d", Addtender))
}
