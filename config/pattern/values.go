package pattern

import (
	"errors"
	"fmt"
	"github.com/nickng/bibtex"
	"net/url"
	"strings"
)

var ErrInvalidReferenceUrl = errors.New("invalid URL from bibtex file")

type EntryNameValues struct {
	Doi     *string
	Title   *string
	Year    *string
	Authors *string
	EntryId *string
}

func entryField(entry bibtex.BibEntry, field string) *string {
	if value, ok := entry.Fields[field]; ok {
		stringValue := value.String()
		return &stringValue
	}

	return nil
}

func NewEntryNameValues(entry bibtex.BibEntry) *EntryNameValues {
	values := &EntryNameValues{}

	values.Doi = entryField(entry, "doi")
	values.Title = entryField(entry, "title")
	if date := entryField(entry, "date"); date != nil {
		// The date field should be in YYYY-MM-DD format.
		values.Year = &strings.SplitN(*date, "-", 2)[0]
	}
	values.Authors = entryField(entry, "author")
	values.EntryId = &entry.CiteName

	return values
}

func (e EntryNameValues) Parser() Parser {
	return NewParser(map[Var]*string{
		'd': e.Doi,
		't': e.Title,
		'y': e.Year,
		'a': e.Authors,
		'i': e.EntryId,
	})
}

type ProxySchemaValues struct {
	Doi      *string
	Hostname *string
	Path     *string
}

func (e ProxySchemaValues) Parser() Parser {
	return NewParser(map[Var]*string{
		'd': e.Doi,
		'h': e.Hostname,
		'p': e.Path,
	})
}

func NewProxySchemaValues(entry bibtex.BibEntry) (*ProxySchemaValues, error) {
	values := &ProxySchemaValues{}

	values.Doi = entryField(entry, "doi")

	if rawUrl := entryField(entry, "url"); rawUrl != nil {
		entryUrl, err := url.Parse(*rawUrl)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidReferenceUrl, err)
		}

		urlPath := strings.TrimPrefix(entryUrl.Path, "/")
		values.Hostname = &entryUrl.Host
		values.Path = &urlPath
	}

	return values, nil
}
