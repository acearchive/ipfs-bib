package archive

import "sort"

type sortAffinity string

const (
	sortAffinityGood          sortAffinity = "good"
	sortAffinityBad           sortAffinity = "bad"
	sortAffinityIndeterminate sortAffinity = "indeterminate"
)

func sortBibContents(contents []BibContents, affinity func(content BibContents) sortAffinity) {
	sort.SliceStable(contents, func(i, j int) bool {
		if affinity(contents[i]) == sortAffinityGood && affinity(contents[j]) == sortAffinityBad {
			return true
		} else {
			return false
		}
	})
}

func chooseBestBibContents(contents []BibContents) BibContents {
	if len(contents) == 1 {
		return contents[0]
	}

	// The least important criteria go first and the most important criteria go last.

	sortBibContents(contents, func(bibContent BibContents) sortAffinity {
		switch {
		case bibContent.Contents == nil:
			return sortAffinityIndeterminate
		case bibContent.Contents.FileName == "":
			return sortAffinityBad
		default:
			return sortAffinityGood
		}
	})

	sortBibContents(contents, func(bibContent BibContents) sortAffinity {
		if bibContent.Doi == nil {
			return sortAffinityBad
		} else {
			return sortAffinityGood
		}
	})

	sortBibContents(contents, func(bibContent BibContents) sortAffinity {
		switch {
		case bibContent.Contents == nil:
			return sortAffinityIndeterminate
		case IsPreferredMediaType(bibContent.Contents.MediaType):
			return sortAffinityGood
		default:
			return sortAffinityBad
		}
	})

	sortBibContents(contents, func(bibContent BibContents) sortAffinity {
		if bibContent.Contents == nil {
			return sortAffinityBad
		} else {
			return sortAffinityGood
		}
	})

	return contents[0]
}

func DeduplicateContents(contents []BibContents) []BibContents {
	byCiteName := make(map[BibCiteName][]BibContents)

	for _, bibContent := range contents {
		byCiteName[bibContent.Entry.CiteName] = append(byCiteName[bibContent.Entry.CiteName], bibContent)
	}

	deduplicated := make([]BibContents, 0, len(byCiteName))

	for _, contentsList := range byCiteName {
		deduplicated = append(deduplicated, chooseBestBibContents(contentsList))
	}

	return deduplicated
}
