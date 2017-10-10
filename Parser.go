package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var fl FileProtocols
var UrlXml string
var Addtender = 0

func DownLoadFile() string {
	tNow := time.Now()
	tMinus3H := tNow.Add(time.Hour * -4)
	UrlXml = fmt.Sprintf("http://ws-rn.tektorg.ru/export/procedures?start_date=%v.000000%%2b03:00&end_date=%v.000000%%2b03:00", tMinus3H.Format("2006-01-02T15:04:05"), tNow.Format("2006-01-02T15:04:05"))
	//fmt.Println(s)
	count := 0
	for {
		if count > 50 {
			Logging(fmt.Sprintf("Не скачали файл за %d попыток", count))
			return ""
		}
		resp, err := http.Get(UrlXml)
		if err != nil {
			Logging("Ошибка скачивания", UrlXml, err)
			count++
			time.Sleep(time.Second * 5)
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

func DownLoadFileTest() string {
	b, err := ioutil.ReadFile("./procedures.xml")
	if err != nil {
		return ""
	}
	return string(b)
}

func Parser() {
	//s := DownLoadFile()
	s := DownLoadFileTest()
	if s == "" {
		Logging("Получили пустую строку")
		return
	}
	if err := xml.Unmarshal([]byte(s), &fl); err != nil {
		Logging("Ошибка при парсинге строки", err)
		return
	}
	if fl.HasMoreProcedures == 1 {
		Logging("Слишком много процедур в запросе")
	}
	for _, r := range fl.Protocols {
		ParserProtocol(r)
		break
	}

}
func ParserProtocol(p Protocol) {
	layout := "2006-01-02T15:04:05"
	RegistryNumber := p.RegistryNumber
	DatePublishedS := p.DatePublished[:19]
	DateUpdatedS := p.DateUpdated
	if DateUpdatedS == "" {
		DateUpdatedS = DatePublishedS
	}
	DateUpdatedS = DateUpdatedS[:19]
	DatePublished, _ := time.Parse(layout, DatePublishedS)
	DateUpdated, _ := time.Parse(layout, DateUpdatedS)

	IdXml := p.IdProtocol
	Version := 0
	Dsn := fmt.Sprintf("root:1234@/%s?charset=utf8&parseTime=true&readTimeout=60m", DbName)
	db, err := sql.Open("mysql", Dsn)
	defer db.Close()
	if err != nil {
		Logging("Ошибка подключения к БД", err)
	}
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE id_xml = ? AND purchase_number = ? AND date_version = ?", Prefix))
	res, err := stmt.Query(IdXml, RegistryNumber, DateUpdated)
	if err != nil {
		Logging("Ошибка выполения запроса", err)
	}
	if res.Next() {
		Logging("Такой тендер уже есть", RegistryNumber)
		return
	}
	var cancelStatus = 0
	if RegistryNumber != "" {
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ?", Prefix))
		rows, err := stmt.Query(RegistryNumber)
		if err != nil {
			Logging("Ошибка выполения запроса", err)
		}
		for rows.Next() {
			var idTender int
			var dateVersion time.Time
			err = rows.Scan(&idTender, &dateVersion)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
			}
			//fmt.Println(DateUpdated.Sub(dateVersion))
			if DateUpdated.Sub(dateVersion) <= 0 {
				cancelStatus = 1
			} else {
				stmtupd, _ := db.Prepare(fmt.Sprintf("UPDATE %stender SET cancel=1 WHERE id_tender = ?", Prefix))
				_, err = stmtupd.Exec(idTender)
			}

		}
		//fmt.Println(cancelStatus)
	}
	Href := fmt.Sprintf("http://rn.tektorg.ru/ru/procurement/procedures/%s", p.IdProtocol)
	PurchaseObjectInfo := p.PurchaseObjectInfo
	NoticeVersion := ""
	PrintForm := Href
	IdOrganizer := 0
	OrganizerfullName := p.OrganizerfullNameU
	if OrganizerfullName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name LIKE '%%'|| ? ||'%%' LIMIT 1", Prefix))
		rows, err := stmt.Query(OrganizerfullName)
		if err != nil {
			Logging("Ошибка выполения запроса", err)
		}
		if rows.Next() {
			err = rows.Scan(&IdOrganizer)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
			}
		} else {
			OrgPostAddress := strings.TrimSpace(fmt.Sprintf("%s %s %s %s %s", p.OrganizerIndexP, p.OrganizerRegionP, p.OrganizerCityP, p.OrganizerStreetP, p.OrganizerHouseP))
			OrgUrAddress := strings.TrimSpace(fmt.Sprintf("%s %s %s %s %s", p.OrganizerIndexU, p.OrganizerRegionU, p.OrganizerCityU, p.OrganizerStreetU, p.OrganizerHouseU))
			ContactEmail := p.ContactEmail
			ContactPhone := p.ContactPhone
			ContactPerson := p.ContactPerson
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?, contact_person = ?", Prefix))
			res, err := stmt.Exec(OrganizerfullName, OrgPostAddress, OrgUrAddress, ContactEmail, ContactPhone, ContactPerson)
			if err != nil {
				Logging("Ошибка чтения вставки организатора", err)
			}
			id, err := res.LastInsertId()
			IdOrganizer = int(id)
		}
	}
	IdPlacingWay := 0
	PwCode := p.ProcedureTypeId
	PwName := p.ProcedureTypeName
	if PwCode != "" && PwName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_placing_way FROM %splacing_way WHERE code = ? AND name = ? LIMIT 1", Prefix))
		rows, err := stmt.Query(PwCode, PwName)
		if err != nil {
			Logging("Ошибка выполения запроса", err)
		}
		if rows.Next() {
			err = rows.Scan(&IdPlacingWay)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
			}
		} else {
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %splacing_way SET code= ?, name= ?", Prefix))
			res, err := stmt.Exec(PwCode, PwName)
			if err != nil {
				Logging("Ошибка чтения вставки placing way", err)
			}
			id, err := res.LastInsertId()
			IdPlacingWay = int(id)
		}
	}

	IdEtp := 0
	etpName := "ЭТП ТЭК-Торг секция ОАО «НК «Роснефть»"
	etpUrl := "https://rn.tektorg.ru/"
	if etpName != "" && etpUrl != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_etp FROM %setp WHERE name = ? AND url = ? LIMIT 1", Prefix))
		rows, err := stmt.Query(etpName, etpUrl)
		if err != nil {
			Logging("Ошибка выполения запроса", err)
		}
		if rows.Next() {
			err = rows.Scan(&IdEtp)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
			}
		} else {
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %setp SET name = ?, url = ?, conf=0", Prefix))
			res, err := stmt.Exec(etpName, etpUrl)
			if err != nil {
				Logging("Ошибка чтения вставки etp", err)
			}
			id, err := res.LastInsertId()
			IdEtp = int(id)
		}
	}

	var EndDate, BiddingDate, ScoringDate = time.Time{}, time.Time{}, time.Time{}
	EndDateS := p.DateEndRegistration
	if EndDateS != "" {
		EndDate, _ = time.Parse(layout, EndDateS[:19])
	}
	ScoringDateS := p.DateEndSecondPartsReview
	if ScoringDateS != "" {
		ScoringDate, _ = time.Parse(layout, ScoringDateS[:19])
	}
	typeFz := 1
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_region = 0, id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, scoring_date = ?, bidding_date = ?, cancel = ?, date_version = ?, num_version = ?, notice_version = ?, xml = ?, print_form = ?", Prefix))
	rest, err := stmtt.Exec(IdXml, RegistryNumber, DatePublished, Href, PurchaseObjectInfo, typeFz, IdOrganizer, IdPlacingWay, IdEtp, EndDate, ScoringDate, BiddingDate, cancelStatus, DateUpdated, Version, NoticeVersion, UrlXml, PrintForm)
	if err != nil {
		Logging("Ошибка чтения вставки tender", err)
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	fmt.Println(idTender)
	fmt.Println(cancelStatus)
	Addtender++

}
