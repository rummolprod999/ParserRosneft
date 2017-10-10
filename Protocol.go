package main


type Protocol struct {
	RegistryNumber string `xml:"registryNumber"`
	IdProtocol string `xml:"id,attr"`
	DatePublished string `xml:"datePublished"`
	DateUpdated string `xml:"dateUpdated"`
	PurchaseObjectInfo string `xml:"title"`
	Organizer
	ProcedureTypeId string `xml:"procedureType>id"`
	ProcedureTypeName string `xml:"procedureType>title"`
	DateEndRegistration string `xml:"dateEndRegistration"`
	DateEndSecondPartsReview string `xml:"dateEndSecondPartsReview"`
}

type FileProtocols struct {
	HasMoreProcedures int        `xml:"Body>proceduresResponse>has_more_procedures"`
	Protocols         []Protocol `xml:"Body>proceduresResponse>procedures>procedure"`
	Test              string     `xml:",innerxml"`
}

type Organizer struct{
	OrganizerfullNameU string `xml:"organizer>fullName"`
	OrganizerIndexU string `xml:"organizer>legal>index"`
	OrganizerRegionU string `xml:"organizer>legal>region"`
	OrganizerCityU string `xml:"organizer>legal>city"`
	OrganizerStreetU string `xml:"organizer>legal>street"`
	OrganizerHouseU string `xml:"organizer>legal>house"`
	OrganizerIndexP string `xml:"organizer>postal>index"`
	OrganizerRegionP string `xml:"organizer>postal>region"`
	OrganizerCityP string `xml:"organizer>postal>city"`
	OrganizerStreetP string `xml:"organizer>postal>street"`
	OrganizerHouseP string `xml:"organizer>postal>house"`
	ContactEmail string `xml:"contactEmail"`
	ContactPhone string `xml:"contactPhone"`
	ContactPerson string `xml:"contactPerson"`
}