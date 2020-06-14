package main

type Protocol struct {
	RegistryNumber           string       `xml:"registryNumber"`
	Url                      string       `xml:"url_to_showcase"`
	IdProtocol               string       `xml:"id,attr"`
	DatePublished            string       `xml:"datePublished"`
	DateUpdated              string       `xml:"dateUpdated"`
	PurchaseObjectInfo       string       `xml:"title"`
	ProcedureTypeId          string       `xml:"procedureType>id"`
	ProcedureTypeName        string       `xml:"procedureType>title"`
	DateEndRegistration      string       `xml:"dateEndRegistration"`
	DateEndSecondPartsReview string       `xml:"dateEndSecondPartsReview"`
	Currency                 string       `xml:"currency"`
	Attachments              []Attachment `xml:"documents>document"`
	Lots                     []Lot        `xml:"lots>lot"`
	Organizer
}

type FileProtocols struct {
	TotalPage         int        `xml:"Body>proceduresResponse>totalPage"`
	CurrentPage       int        `xml:"Body>proceduresResponse>currentPage"`
	HasMoreProcedures int        `xml:"Body>proceduresResponse>has_more_procedures"`
	Protocols         []Protocol `xml:"Body>proceduresResponse>procedures>procedure"`
	Test              string     `xml:",innerxml"`
}

type Organizer struct {
	OrganizerfullNameU string `xml:"organizer>fullName"`
	OrganizerIndexU    string `xml:"organizer>legal>index"`
	OrganizerRegionU   string `xml:"organizer>legal>region"`
	OrganizerCityU     string `xml:"organizer>legal>city"`
	OrganizerStreetU   string `xml:"organizer>legal>street"`
	OrganizerHouseU    string `xml:"organizer>legal>house"`
	OrganizerIndexP    string `xml:"organizer>postal>index"`
	OrganizerRegionP   string `xml:"organizer>postal>region"`
	OrganizerCityP     string `xml:"organizer>postal>city"`
	OrganizerStreetP   string `xml:"organizer>postal>street"`
	OrganizerHouseP    string `xml:"organizer>postal>house"`
	ContactEmail       string `xml:"contactEmail"`
	ContactPhone       string `xml:"contactPhone"`
	ContactPerson      string `xml:"contactPerson"`
}

type Attachment struct {
	AttachName string `xml:"filename"`
	AttachUrl  string `xml:"file"`
}

type Lot struct {
	LotNumber      int             `xml:"number"`
	LotSubject     string          `xml:"subject"`
	StartPrice     float64         `xml:"startPrice"`
	Okpd2Code      string          `xml:"nomenclature2>item>code"`
	OkpdName       string          `xml:"nomenclature2>item>name"`
	Customers      []Customer      `xml:"customers>customer"`
	DeliveryPlaces []DeliveryPlace `xml:"deliveryPlaces>deliveryPlace"`
	AttachmentsLot []AttachmentLot `xml:"documents>document"`
}

type Customer struct {
	FullName string `xml:"fullName"`
}

type DeliveryPlace struct {
	Quantity string `xml:"quantity"`
	Term     string `xml:"term"`
	Address  string `xml:"address"`
}

type AttachmentLot struct {
	AttachName string `xml:"filename"`
	AttachUrl  string `xml:"file"`
}
