package main

var FileLog Filelog
var DbName = "tender"
var Prefix = ""

func init() {
	CreateLogFile()
}

func main() {
	//fmt.Println(FileLog)
	Logging("Начало парсинга")
	Parser()
	/*b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	fmt.Println("hello")
	if err := xml.Unmarshal([]byte(b), &fl); err != nil {
		fmt.Println(err)
		return
	}*/
	/*fmt.Println(fl.Has_more_procedures)
	//fmt.Println(fl.Test)
	for _, r := range fl.Protocols {
		fmt.Println(r.RegistryNumber)
		fmt.Println(r.IdProtocol)
	}*/
	Logging("Конец парсинга")
}
