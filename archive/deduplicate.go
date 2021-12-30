package archive

type bibMetadataPredicate = func(BibMetadata) bool

func bibMetadataHasContent(content BibMetadata) bool {
	return content.Contents != nil
}

func bibMetadataHasPreferredMediaType(content BibMetadata) bool {
	return content.Contents != nil && IsPreferredMediaType(content.Contents.MediaType)
}

func bibMetadataHasDoi(content BibMetadata) bool {
	return content.Doi != nil
}

func bibMetadataHasFileName(content BibMetadata) bool {
	return content.Contents != nil && content.Contents.FileName != ""
}

func (c BibMetadata) isBetterThanBy(other BibMetadata, predicates []bibMetadataPredicate) bool {
	for _, predicate := range predicates {
		if predicate(c) && !predicate(other) {
			return true
		}
	}

	return false
}

func (c BibMetadata) isBetterThan(other BibMetadata) bool {
	predicates := []bibMetadataPredicate{
		bibMetadataHasContent,
		bibMetadataHasPreferredMediaType,
		bibMetadataHasDoi,
		bibMetadataHasFileName,
	}

	return c.isBetterThanBy(other, predicates)
}

func DeduplicateContents(results chan DownloadResult) chan DownloadResult {
	deduplicated := make(chan DownloadResult, cap(results))
	bestByCiteName := make(map[BibCiteName]BibMetadata)

	go func() {
		for downloadResult := range results {
			if downloadResult.Error != nil {
				deduplicated <- downloadResult
				break
			}

			bibContents := downloadResult.Contents

			currentBest, currentBestExists := bestByCiteName[bibContents.Entry.CiteName]
			if !currentBestExists || bibContents.ToMetadata().isBetterThan(currentBest) {
				bestByCiteName[bibContents.Entry.CiteName] = bibContents.ToMetadata()
				deduplicated <- downloadResult
			}
		}

		close(deduplicated)
	}()

	return deduplicated
}
