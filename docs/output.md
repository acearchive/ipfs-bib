# JSON Output Format

This document describes the format of the JSON output produced when the
`--json` flag is passed.

## Response Object

| Key | Type | Description |
| --- | --- | --- |
| `cid` | string | The CID of the root directory containing all the archived sources. |
| `totalEntries` | number | The total number of entries in the provided bibtex file or Zotero library. |
| `totalArchived` | number | The number of entries that the tool was able to find a source for and archive to IPFS, which may be less than `totalEntries`. |
| `archived` | array | An **Archived Entry Object** for each entry that was archived to IPFS. |
| `notArchived` | array | A **Not Archived Entry Object** for each entry that was not archived to IPFS. |

## Archived Entry Object

| Key | Type | Description |
| --- | --- | --- |
| `citeName` | string | The bibtex cite name for the entry. |
| `doi` | string \| null | The DOI of the entry, excluding the doi: or https://doi.org/ prefix (e.g. 10.1038/nphys1170). If no DOI was found, this is `null`. |
| `mediaType` | string | The media type (MIME type) of the archived source content (e.g. application/pdf). |
| `fileCid` | string | The CID of the archived source file. |
| `fileName` | string | The name of the archived source file. |
| `directoryCid` | string | The CID of the directory containing the archived source file. |
| `directoryName` | string | The name of the directory containing the archived source file. |
| `ipfsUrl` | string | The `ipfs://` URL of the archived source file, including a `?filename=` query parameter. |
| `gatewayUrl` | string | The gateway URL of the archived source file, including a `?filename=` query parameter. This uses the public subdomain gateway configured in the config file. |
| `contentOrigin` | string | A **Content Origin Enum** describing where the source content was archived from. |

## Not Archived Entry Object

| Key | Type | Description |
| --- | --- | --- |
| `citeName` | string | The bibtex cite name for the entry. |
| `doi` | string\|null | The DOI of the entry, excluding the doi: or https://doi.org/ prefix (e.g. 10.1038/nphys1170). If no DOI was found, this is `null`. |

## Content Origin Enum

| Value | Description |
| --- | --- |
| `url` | The source content was located by following the URL or DOI (via https://doi.org) in the bibtex/Zotero citation. |
| `local` | The source content was a local file referenced in the bibtex file via a `file` field. |
| `zotero` | The source content was a Zotero attachment pulled from the public Zotero library. |
| `unpaywall` | The source content was pulled from Unpaywall. |
| `resolver` | The source content was pulled from one of the link resolvers defined in the config file. |
