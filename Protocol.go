package main

type Protocol struct {
	RegistryNumber string `xml:"registryNumber"`
	IdProtocol string `xml:"id,attr"`
}

type FileProtocols struct {
	Has_more_procedures int        `xml:"Body>proceduresResponse>has_more_procedures"`
	Protocols           []Protocol `xml:"Body>proceduresResponse>procedures>procedure"`
	Test                string     `xml:",innerxml"`
}
