package pattern

type EntryNameValues struct {
	Doi     string
	Title   string
	Year    string
	Authors string
	EntryId string
}

func (e *EntryNameValues) Parser() Parser {
	return NewParser(map[Var]string{
		'd': e.Doi,
		't': e.Title,
		'y': e.Year,
		'a': e.Authors,
		'i': e.EntryId,
	})
}

type ProxySchemaValues struct {
	Doi      string
	Hostname string
	Path     string
}

func (e *ProxySchemaValues) Parser() Parser {
	return NewParser(map[Var]string{
		'd': e.Doi,
		'h': e.Hostname,
		'p': e.Path,
	})
}
