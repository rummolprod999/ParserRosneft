package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
)

var fl FileProtocols
var UrlXml string
func DownLoadFile() string {
	tNow := time.Now()
	tMinus3H := tNow.Add(time.Hour * -4)
	UrlXml = fmt.Sprintf("http://ws-rn.tektorg.ru/export/procedures?start_date=%v.000000%%2b03:00&end_date=%v.000000%%2b03:00", tMinus3H.Format("2006-01-02T15:04:05"), tNow.Format("2006-01-02T15:04:05"))
	//fmt.Println(s)
	count := 0
	for {
		if count > 20 {
			Logging(fmt.Sprintf("Не скачали файл за %d попыток", count))
			return ""
		}
		resp, err := http.Get(UrlXml)
		if err != nil {
			Logging("Ошибка скачивания", UrlXml, err)
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
		break
	}

}
func ParserProtocol(p Protocol) {
	RegistryNumber := p.RegistryNumber
	DatePublished := p.DatePublished
	DateUpdated :=p.DateUpdated
	if DateUpdated == (time.Time{}){
		DateUpdated = DatePublished
	}
	IdXml := p.IdProtocol
	//Version := 0

	fmt.Println(IdXml)
	fmt.Println(RegistryNumber)
	fmt.Println(DateUpdated.Format("2006-01-02T15:04:05"))
	Dsn := fmt.Sprintf("root:1234@/%s?charset=utf8&parseTime=true&readTimeout=60m", DbName)
	db, err := sql.Open("mysql", Dsn)
	defer db.Close()
	if err != nil{
		Logging("Ошибка подключения к БД", err)
	}
	stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE id_xml = ? AND purchase_number = ? AND date_version = ?", Prefix))
	res, err := stmt.Query(IdXml, RegistryNumber, DateUpdated.Format("2006-01-02T15:04:05"))
	if err != nil{
		Logging("Ошибка подключения к БД", err)
	}
	if res.Next(){
		Logging("Такой тендер уже есть", RegistryNumber)
	}
	
}
