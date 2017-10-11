package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	//"golang.org/x/tools/go/gcimporter15/testdata"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var fl FileProtocols
var UrlXml string
var Addtender = 0
var HasMoreProcedures = 1

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
func DownLoadFileDay() string {
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
	s := DownLoadFileDay()
	//s := DownLoadFileTest()
	if s == "" {
		Logging("Получили пустую строку")
		return
	}
	if err := xml.Unmarshal([]byte(s), &fl); err != nil {
		Logging("Ошибка при парсинге строки", err)
		return
	}
	HasMoreProcedures = fl.HasMoreProcedures
	if fl.HasMoreProcedures == 1 {
		Logging("Слишком много процедур в запросе")
		if len(fl.Protocols) > 0 {
			DatePublishedS := fl.Protocols[len(fl.Protocols)-1].DatePublished[:19]
			DateUpdatedS := fl.Protocols[len(fl.Protocols)-1].DateUpdated
			if DateUpdatedS == "" {
				DateUpdatedS = DatePublishedS
			}
			DateUpdatedS = DateUpdatedS[:19]
			UrlXml = fmt.Sprintf("http://ws-rn.tektorg.ru/export/procedures?start_date=%s.000000", DateUpdatedS)
			Logging("Новый URL", UrlXml)
		}
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
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ? AND cancel=0", Prefix))
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
			if dateVersion.Sub(DateUpdated) <= 0 {
				stmtupd, _ := db.Prepare(fmt.Sprintf("UPDATE %stender SET cancel=1 WHERE id_tender = ?", Prefix))
				_, err = stmtupd.Exec(idTender)

			} else {
				cancelStatus = 1
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
				Logging("Ошибка вставки организатора", err)
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
				Logging("Ошибка вставки placing way", err)
			}
			id, err := res.LastInsertId()
			IdPlacingWay = int(id)
		}
	}

	IdEtp := 0
	etpName := "ЭТП ТЭК-Торг секция ОАО «НК «Роснефть»"
	etpUrl := "https://rn.tektorg.ru/"
	if true {
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
				Logging("Ошибка вставки etp", err)
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
		Logging("Ошибка вставки tender", err)
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	Addtender++
	for _, att := range p.Attachments {
		attachName := att.AttachName
		attachUrl := att.AttachUrl
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", Prefix))
		_, err := stmt.Exec(idTender, attachName, attachUrl)
		if err != nil {
			Logging("Ошибка вставки attachment", err)
		}
	}
	for _, lot := range p.Lots {
		LotNumber := lot.LotNumber
		MaxPrice := lot.StartPrice
		//Subject := lot.LotSubject
		Currency := p.Currency
		idLot := 0
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, max_price = ?, currency = ?", Prefix))
		res, err := stmt.Exec(idTender, LotNumber, MaxPrice, Currency)
		if err != nil {
			Logging("Ошибка вставки lot", err)
		}
		id, _ := res.LastInsertId()
		idLot = int(id)
		//fmt.Println(idLot)
		idCustomer := 0
		if len(lot.Customers) > 0 {
			if lot.Customers[0].FullName != "" {
				stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name LIKE '%%'|| ? ||'%%' LIMIT 1", Prefix))
				rows, err := stmt.Query(lot.Customers[0].FullName)
				if err != nil {
					Logging("Ошибка выполения запроса", err)
				}
				if rows.Next() {
					err = rows.Scan(&idCustomer)
					if err != nil {
						Logging("Ошибка чтения результата запроса", err)
					}
				} else {
					stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, is223=1, reg_num = ?", Prefix))
					res, err := stmt.Exec(lot.Customers[0].FullName, "00000223000000000")
					if err != nil {
						Logging("Ошибка вставки организатора", err)
					}
					id, err := res.LastInsertId()
					idCustomer = int(id)
				}
			}
		}

		for _, cusR := range lot.DeliveryPlaces {
			deliveryPlace := cusR.Address
			deliveryTerm := cusR.Term
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_place = ?, delivery_term = ?", Prefix))
			_, err := stmt.Exec(idLot, idCustomer, deliveryPlace, deliveryTerm)
			if err != nil {
				Logging("Ошибка вставки customer_requirement", err)
			}
		}
		QuantityValue := ""
		if len(lot.DeliveryPlaces) == 1 {
			QuantityValue = lot.DeliveryPlaces[0].Quantity
		}

		okpd2Code := lot.Okpd2Code
		okpdName := lot.OkpdName
		okpd2GroupCode, okpd2GroupLevel1Code := GetOkpd(okpd2Code)
		stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, okpd2_code = ?, okpd2_group_code = ?, okpd2_group_level1_code = ?, okpd_name = ?, name = ?, quantity_value = ?, customer_quantity_value = ?", Prefix))
		_, errr := stmtr.Exec(idLot, idCustomer, okpd2Code, okpd2GroupCode, okpd2GroupLevel1Code, okpdName, lot.LotSubject, QuantityValue, QuantityValue)
		if errr != nil {
			Logging("Ошибка вставки purchase_object", errr)
		}

	}
	TenderKwords(db, idTender)

}
func TenderKwords(db *sql.DB, idTender int) {
	resString := ""
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT DISTINCT po.name, po.okpd_name FROM %spurchase_object AS po LEFT JOIN %slot AS l ON l.id_lot = po.id_lot WHERE l.id_tender = ?", Prefix, Prefix))
	rows, err := stmt.Query(idTender)
	if err != nil {
		Logging("Ошибка выполения запроса", err)
	}
	for rows.Next() {
		var name sql.NullString
		var okpdName sql.NullString
		err = rows.Scan(&name, &okpdName)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
		}
		if name.Valid {
			resString = fmt.Sprintf("%s %s ", resString, name.String)
		}
		if okpdName.Valid {
			resString = fmt.Sprintf("%s %s ", resString, okpdName.String)
		}
	}

	stmt1, _ := db.Prepare(fmt.Sprintf("SELECT DISTINCT file_name FROM %sattachment WHERE id_tender = ?", Prefix))
	rows1, err := stmt1.Query(idTender)
	if err != nil {
		Logging("Ошибка выполения запроса", err)
	}
	for rows1.Next() {
		var attName sql.NullString
		err = rows1.Scan(&attName)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
		}
		if attName.Valid {
			resString = fmt.Sprintf("%s %s ", resString, attName.String)
		}
	}

	idOrg := 0
	stmt2, _ := db.Prepare(fmt.Sprintf("SELECT purchase_object_info, id_organizer FROM %stender WHERE id_tender = ?", Prefix))
	rows2, err := stmt2.Query(idTender)
	if err != nil {
		Logging("Ошибка выполения запроса", err)
	}
	for rows2.Next() {
		var idOrgNull sql.NullInt64
		var purOb sql.NullString
		err = rows2.Scan(&purOb, &idOrgNull)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
		}
		if idOrgNull.Valid {
			idOrg = int(idOrgNull.Int64)
		}
		if purOb.Valid {
			resString = fmt.Sprintf("%s %s ", resString, purOb.String)
		}

	}

	if idOrg != 0 {
		stmt3, _ := db.Prepare(fmt.Sprintf("SELECT full_name, inn FROM %sorganizer WHERE id_organizer = ?", Prefix))
		rows3, err := stmt3.Query(idOrg)
		if err != nil {
			Logging("Ошибка выполения запроса", err)
		}
		for rows3.Next() {
			var innOrg sql.NullString
			var nameOrg sql.NullString
			err = rows3.Scan(&nameOrg, &innOrg)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
			}
			if innOrg.Valid {

				resString = fmt.Sprintf("%s %s ", resString, innOrg.String)
			}
			if nameOrg.Valid {
				resString = fmt.Sprintf("%s %s ", resString, nameOrg.String)
			}

		}
	}
	stmt4, _ := db.Prepare(fmt.Sprintf("SELECT DISTINCT cus.inn, cus.full_name FROM %scustomer AS cus LEFT JOIN %spurchase_object AS po ON cus.id_customer = po.id_customer LEFT JOIN %slot AS l ON l.id_lot = po.id_lot WHERE l.id_tender = ?", Prefix, Prefix, Prefix))
	rows4, err := stmt4.Query(idTender)
	if err != nil {
		Logging("Ошибка выполения запроса", err)
	}
	for rows4.Next() {
		var innC sql.NullString
		var fullNameC sql.NullString
		err = rows4.Scan(&innC, &fullNameC)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
		}
		if innC.Valid {

			resString = fmt.Sprintf("%s %s ", resString, innC.String)
		}
		if fullNameC.Valid {
			resString = fmt.Sprintf("%s %s ", resString, fullNameC.String)
		}
	}
	re := regexp.MustCompile(`\s+`)
	resString = re.ReplaceAllString(resString, " ")
	stmtr, _ := db.Prepare(fmt.Sprintf("UPDATE %stender SET tender_kwords = ? WHERE id_tender = ?", Prefix))
	_, errr := stmtr.Exec(resString, idTender)
	if errr != nil {
		Logging("Ошибка вставки TenderKwords", errr)
	}
}
func GetOkpd(s string) (int, string) {
	okpd2GroupCode := 0
	okpd2GroupLevel1Code := ""
	if len(s) > 1 {
		if strings.Index(s, ".") != -1 {
			okpd2GroupCode, _ = strconv.Atoi(s[:2])
		}
	}
	if len(s) > 3 {
		if strings.Index(s, ".") != -1 {
			okpd2GroupLevel1Code = s[3:4]
		}
	}
	return okpd2GroupCode, okpd2GroupLevel1Code
}
