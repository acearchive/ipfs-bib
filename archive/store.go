package archive

import "context"

type SourceStore interface {
	AddSource(ctx context.Context, source *BibSource) (id *BibSourceId, err error)
}
