package main

import (
	"fmt"
	"os"
	"time"
)

type Filelog string

func Logging(args ...interface{}) {
	file, err := os.OpenFile(string(FileLog), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		fmt.Println("Ошибка записи в файл лога", err)
		return
	}

	for _, v := range args {
		fmt.Fprintf(file, "%v  ", time.Now())
		fmt.Fprintf(file, "%v\n", v)
	}
}

func CreateLogFile() {
	dirlog := "./LogRosneft"
	if _, err := os.Stat(dirlog); os.IsNotExist(err) {
		err := os.MkdirAll(dirlog, 0711)

		if err != nil {
			fmt.Println("Не могу создать папку для лога")
			os.Exit(1)
		}
	}
	t := time.Now()
	ft := t.Format("2006-01-02T15:04:05")
	FileLog = Filelog(fmt.Sprintf("%s/log_rosneft_%v.log", dirlog, ft))
	/*file, err := os.Create(string(FileLog))
	defer file.Close()
	if err != nil {
		// handle the error here
		fmt.Println("Не могу создать файл лога")
	}*/

}
