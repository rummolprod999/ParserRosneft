package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tiaguinho/gosoap"
	"os/exec"
	"strings"
	"time"
)

type ParserRosneftSoap struct {
	maxPage int
}

func soap_request(section string, page int) FileProtocols {
	soap, err := gosoap.SoapClient("http://api.tektorg.ru/procedures/wsdl")
	if err != nil {
		Logging(err)
		panic("error get soap")
	}
	//var p = ExportRequestType{StartDate: time.Now().Add(time.Hour * -10), SectionCode: section, Page: int32(page)}
	params := gosoap.Params{
		"page": 5,
	}
	var r FileProtocols
	res, err := soap.Call("procedures", params)
	if err != nil {
		Logging(err)
		panic("error get soap")
	}
	myString := string(res.Body)
	myString = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" xmlns:ns1="http://tektorg.loc/api" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <SOAP-ENV:Body>%s  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, myString)
	res.Body = []byte(myString)
	res.Unmarshal(&r)
	return r
}

func (p *ParserRosneftSoap) Parser() {
	protocols := callSOAPClient("1", 1)
	if protocols.TotalPage == 0 {
		Logging("total page not found")
		return
	}
	p.maxPage = protocols.TotalPage
	p.ParserProtocols(protocols)
	for i := 2; i <= p.maxPage; i++ {
		protocols = callSOAPClient("1", i)
		p.ParserProtocols(protocols)
	}
}

func (p *ParserRosneftSoap) ParserProtocols(protocols *FileProtocols) {
	defer SaveStack()
	Dsn := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=true&readTimeout=60m&maxAllowedPacket=0&timeout=60m&writeTimeout=60m&autocommit=true", UserDb, PasswordDb, DbName)
	db, err := sql.Open("mysql", Dsn)
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
	}
	for _, pr := range protocols.Protocols{
		p.ParserProtocol(pr, db)
	}
}
func (pr *ParserRosneftSoap) ParserProtocol(p Protocol, db *sql.DB) error {
	defer SaveStack()
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

	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE id_xml = ? AND purchase_number = ? AND date_version = ? AND type_fz = ?", Prefix))
	res, err := stmt.Query(IdXml, RegistryNumber, DateUpdated, typeFz)
	stmt.Close()
	if err != nil {
		Logging("Ошибка выполения запроса", err)
		return err
	}
	if res.Next() {
		res.Close()
		return nil
	}
	res.Close()
	var cancelStatus = 0
	var updated = false
	if RegistryNumber != "" {
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ? AND cancel=0 AND type_fz = ?", Prefix))
		rows, err := stmt.Query(RegistryNumber, typeFz)
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return err
		}
		for rows.Next() {
			updated = true
			var idTender int
			var dateVersion time.Time
			err = rows.Scan(&idTender, &dateVersion)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return err
			}
			//fmt.Println(DateUpdated.Sub(dateVersion))
			if dateVersion.Sub(DateUpdated) <= 0 {
				stmtupd, _ := db.Prepare(fmt.Sprintf("UPDATE %stender SET cancel=1 WHERE id_tender = ?", Prefix))
				_, err = stmtupd.Exec(idTender)
				stmtupd.Close()

			} else {
				cancelStatus = 1
			}

		}
		rows.Close()
	}
	Href := fmt.Sprintf("http://rn.tektorg.ru/ru/procurement/procedures/%s", p.IdProtocol)
	PurchaseObjectInfo := p.PurchaseObjectInfo
	NoticeVersion := ""
	PrintForm := Href
	IdOrganizer := 0
	OrganizerfullName := p.OrganizerfullNameU
	if OrganizerfullName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name LIKE ? LIMIT 1", Prefix))
		rows, err := stmt.Query(OrganizerfullName)
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return err
		}
		if rows.Next() {
			err = rows.Scan(&IdOrganizer)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return err
			}
			rows.Close()
		} else {
			rows.Close()
			OrgPostAddress := strings.TrimSpace(fmt.Sprintf("%s %s %s %s %s", p.OrganizerIndexP, p.OrganizerRegionP, p.OrganizerCityP, p.OrganizerStreetP, p.OrganizerHouseP))
			OrgUrAddress := strings.TrimSpace(fmt.Sprintf("%s %s %s %s %s", p.OrganizerIndexU, p.OrganizerRegionU, p.OrganizerCityU, p.OrganizerStreetU, p.OrganizerHouseU))
			ContactEmail := p.ContactEmail
			ContactPhone := p.ContactPhone
			ContactPerson := p.ContactPerson
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?, contact_person = ?", Prefix))
			res, err := stmt.Exec(OrganizerfullName, OrgPostAddress, OrgUrAddress, ContactEmail, ContactPhone, ContactPerson)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки организатора", err)
				return err
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
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return err
		}
		if rows.Next() {
			err = rows.Scan(&IdPlacingWay)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return err
			}
			rows.Close()
		} else {
			rows.Close()
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %splacing_way SET code= ?, name= ?", Prefix))
			res, err := stmt.Exec(PwCode, PwName)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки placing way", err)
				return err
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
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return err
		}
		if rows.Next() {
			err = rows.Scan(&IdEtp)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return err
			}
			rows.Close()
		} else {
			rows.Close()
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %setp SET name = ?, url = ?, conf=0", Prefix))
			res, err := stmt.Exec(etpName, etpUrl)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки etp", err)
				return err
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
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_region = 0, id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, scoring_date = ?, bidding_date = ?, cancel = ?, date_version = ?, num_version = ?, notice_version = ?, xml = ?, print_form = ?", Prefix))
	rest, err := stmtt.Exec(IdXml, RegistryNumber, DatePublished, Href, PurchaseObjectInfo, typeFz, IdOrganizer, IdPlacingWay, IdEtp, EndDate, ScoringDate, BiddingDate, cancelStatus, DateUpdated, Version, NoticeVersion, UrlXml, PrintForm)
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return err
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		Updatetender++
	} else {
		Addtender++
	}
	for _, att := range p.Attachments {
		attachName := att.AttachName
		attachUrl := att.AttachUrl
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", Prefix))
		_, err := stmt.Exec(idTender, attachName, attachUrl)
		stmt.Close()
		if err != nil {
			Logging("Ошибка вставки attachment", err)
			return err
		}
	}

	for _, lot := range p.Lots {
		LotNumber := lot.LotNumber
		MaxPrice := lot.StartPrice
		Currency := p.Currency
		idLot := 0
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, max_price = ?, currency = ?", Prefix))
		res, err := stmt.Exec(idTender, LotNumber, MaxPrice, Currency)
		stmt.Close()
		if err != nil {
			Logging("Ошибка вставки lot", err)
			return err
		}
		id, _ := res.LastInsertId()
		idLot = int(id)
		for _, attL := range lot.AttachmentsLot {
			attachName := attL.AttachName
			attachUrl := attL.AttachUrl
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", Prefix))
			_, err := stmt.Exec(idTender, attachName, attachUrl)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки attachmentLot", err)
				return err
			}
		}
		idCustomer := 0
		if len(lot.Customers) > 0 {
			if lot.Customers[0].FullName != "" {
				stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name LIKE ? LIMIT 1", Prefix))
				rows, err := stmt.Query(lot.Customers[0].FullName)
				stmt.Close()
				if err != nil {
					Logging("Ошибка выполения запроса", err)
					return err
				}
				if rows.Next() {
					err = rows.Scan(&idCustomer)
					if err != nil {
						Logging("Ошибка чтения результата запроса", err)
						return err
					}
					rows.Close()
				} else {
					rows.Close()
					out, err := exec.Command("uuidgen").Output()
					if err != nil {
						Logging("Ошибка генерации UUID", err)
						return err
					}
					stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, is223=1, reg_num = ?", Prefix))
					res, err := stmt.Exec(lot.Customers[0].FullName, out)
					stmt.Close()
					if err != nil {
						Logging("Ошибка вставки организатора", err)
						return err
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
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки customer_requirement", err)
				return err

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
		stmtr.Close()
		if errr != nil {
			Logging("Ошибка вставки purchase_object", errr)
			return err
		}

	}
	e := TenderKwords(db, idTender)
	if e != nil {
		Logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, RegistryNumber)
	if e1 != nil {
		Logging("Ошибка обработки AddVerNumber", e1)
	}
	return nil
}

