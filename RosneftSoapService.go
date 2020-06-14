package main

import (
	"encoding/xml"
	"github.com/hooklift/gowsdl/soap"
	"time"
)

// against "unused imports"
var _ time.Time
var _ xml.Name

type ExportRequestType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap exportRequestType"`

	StartDate time.Time `xml:"startDate,omitempty"`

	EndDate time.Time `xml:"endDate,omitempty"`

	StartUpdateAt time.Time `xml:"startUpdateAt,omitempty"`

	EndUpdateAt time.Time `xml:"endUpdateAt,omitempty"`

	SectionCode string `xml:"sectionCode,omitempty"`

	RegistryNumber string `xml:"registryNumber,omitempty"`

	TypeId int32 `xml:"typeId,omitempty"`

	OrganizerINN string `xml:"organizerINN,omitempty"`

	CustomerINN string `xml:"customerINN,omitempty"`

	//Sort *SortRules `xml:"sort,omitempty"`

	Page int32 `xml:"page,omitempty"`
}

type Procedure struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap procedure"`

	RemoteId int32 `xml:"remoteId,omitempty"`

	RegistryNumber string `xml:"registryNumber,omitempty"`

	Title string `xml:"title,omitempty"`

	DatePublished time.Time `xml:"datePublished,omitempty"`

	DateUpdated time.Time `xml:"dateUpdated,omitempty"`

	DateEndRegistration time.Time `xml:"dateEndRegistration,omitempty"`

	DateEndSecondPartsReview time.Time `xml:"dateEndSecondPartsReview,omitempty"`

	DateEndPrequalification time.Time `xml:"dateEndPrequalification,omitempty"`

	ProcedureType *ProceduretypeType `xml:"procedureType,omitempty"`

	ContactEmail string `xml:"contactEmail,omitempty"`

	ContactPhone string `xml:"contactPhone,omitempty"`

	ContactPerson string `xml:"contactPerson,omitempty"`

	ReviewApplicsCity string `xml:"reviewApplicsCity,omitempty"`

	Currency string `xml:"currency,omitempty"`

	Organizer *OrganizationType `xml:"organizer,omitempty"`

	Documents *DocumentsType `xml:"documents,omitempty"`

	Lots struct {
		Lot []struct {
			RemoteId int32 `xml:"remoteId,omitempty"`

			Number int32 `xml:"number,omitempty"`

			Subject string `xml:"subject,omitempty"`

			StartPrice float32 `xml:"startPrice,omitempty"`

			Status string `xml:"status,omitempty"`

			Nds string `xml:"nds,omitempty"`

			StartPriceNoNds float32 `xml:"startPriceNoNds,omitempty"`

			StartPriceUndefined bool `xml:"startPriceUndefined,omitempty"`

			AlternativeApplics bool `xml:"alternativeApplics,omitempty"`

			Customers *CustomersType `xml:"customers,omitempty"`

			LotOkved *LotOkvedsType `xml:"lotOkved,omitempty"`

			LotOkved2 *LotOkvedsType `xml:"lotOkved2,omitempty"`

			DeliveryPlaces *DeliveryplacesType `xml:"deliveryPlaces,omitempty"`

			Nomenclature *NomenclatureType `xml:"nomenclature,omitempty"`

			Nomenclature2 *NomenclatureType `xml:"nomenclature2,omitempty"`

			LotUnits *LotUnitsType `xml:"lotUnits,omitempty"`

			Documents *DocumentsType `xml:"documents,omitempty"`

			Id int32 `xml:"id,attr,omitempty"`
		} `xml:"lot,omitempty"`
	} `xml:"lots,omitempty"`

	Id int32 `xml:"id,attr,omitempty"`
}

type Procedures struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap procedures"`

	Procedure []*Procedure `xml:"procedure,omitempty"`
}

type AddressType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap addressType"`

	Index string `xml:"index,omitempty"`

	Region string `xml:"region,omitempty"`

	Settlement string `xml:"settlement,omitempty"`

	City string `xml:"city,omitempty"`

	Street string `xml:"street,omitempty"`

	House string `xml:"house,omitempty"`

	CountryIsoNr string `xml:"countryIsoNr,omitempty"`
}

type OrganizationType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap organizationType"`

	Id int32 `xml:"id,omitempty"`

	FullName string `xml:"fullName,omitempty"`

	Inn string `xml:"inn,omitempty"`

	Legal *AddressType `xml:"legal,omitempty"`

	Postal *AddressType `xml:"postal,omitempty"`
}

type CustomerOrganizationType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap customerOrganizationType"`

	Id int32 `xml:"id,omitempty"`

	FullName string `xml:"fullName,omitempty"`

	Inn string `xml:"inn,omitempty"`

	Phone string `xml:"phone,omitempty"`

	Email string `xml:"email,omitempty"`

	Legal *AddressType `xml:"legal,omitempty"`

	Postal *AddressType `xml:"postal,omitempty"`
}

type CustomersType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap customersType"`

	Customer []*CustomerOrganizationType `xml:"customer,omitempty"`
}

type DocumentsType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap documentsType"`

	Document []*DocumentType `xml:"document,omitempty"`
}

type DocumentType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap documentType"`

	Id int32 `xml:"id,omitempty"`

	Filename string `xml:"filename,omitempty"`

	Removed bool `xml:"removed,omitempty"`

	File string `xml:"file,omitempty"`
}

type LotOkvedsType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap lotOkvedsType"`

	Okved_code []*LotOkvedType `xml:"okved_code,omitempty"`
}

type LotOkvedType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap lotOkvedType"`

	Code string `xml:"code,omitempty"`

	Name string `xml:"name,omitempty"`
}

type NomenclatureType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap nomenclatureType"`

	Item []struct {
		Code string `xml:"code,omitempty"`

		Name string `xml:"name,omitempty"`
	} `xml:"item,omitempty"`
}

type LotUnitsType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap lotUnitsType"`

	Unit []struct {
		Name string `xml:"name,omitempty"`

		OkeiCode string `xml:"okeiCode,omitempty"`

		Okpd2_code string `xml:"okpd2_code,omitempty"`

		Okved2_code string `xml:"okved2_code,omitempty"`

		Quantity string `xml:"quantity,omitempty"`
	} `xml:"unit,omitempty"`
}

type OkeicodesType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap okeicodesType"`

	Item []struct {
		Code int16 `xml:"code,omitempty"`

		Name string `xml:"name,omitempty"`
	} `xml:"item,omitempty"`
}

type DeliveryplacesType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap deliveryplacesType"`

	DeliveryPlace []*DeliveryplaceType `xml:"deliveryPlace,omitempty"`
}

type DeliveryplaceType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap deliveryplaceType"`

	Quantity string `xml:"quantity,omitempty"`

	Term string `xml:"term,omitempty"`

	OkeiCode string `xml:"okeiCode,omitempty"`

	Address string `xml:"address,omitempty"`
}

type ProceduretypeType struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap proceduretypeType"`

	Id int32 `xml:"id,omitempty"`

	Title string `xml:"title,omitempty"`
}

type SortRules struct {
	XMLName xml.Name `xml:"http://api.tektorg.ru/procedures/soap sortRules"`

	Field string `xml:"field,omitempty"`

	Operator string `xml:"operator,omitempty"`
}

type Int struct {}

type ExportProcedurePort interface {
	Procedures(request *ExportRequestType) (*Int, error)
}

type exportProcedurePort struct {
	client *soap.Client
}

func NewExportProcedurePort(client *soap.Client) ExportProcedurePort {
	return &exportProcedurePort{
		client: client,
	}
}

func (service *exportProcedurePort) Procedures(request *ExportRequestType) (*Int, error) {
	response := new(Int)
	err := service.client.Call("urn:procedures", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}