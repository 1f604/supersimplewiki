package pagelib

type Page struct { // TODO: make sure PageID is properly initialized. Same goes for other structs.
	PageID string
	Title  string
	Body   []byte
}

type ViewPageStruct struct {
	PageID    string
	PageTitle string
	HTML      []byte
}

type EditPageStruct struct {
	PageID    string
	PageTitle string
	Source    []byte
	Checksum  string
}
