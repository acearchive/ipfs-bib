# ipfs-bib

ipfs-bib is a tool for hosting bibliographic references on
[IPFS](https://ipfs.io). The tool pulls citations from a bibtex/biblatex file
or a public [Zotero](https://zotero.org) library, finds the full text of the
research on the legacy web, and rehosts it on a local IPFS node or exports it
to a [CAR archive](ipns://ipld.io/specs/transport/car/).

## Features

- Pull citations from a bibtex/biblatex file or a public Zotero library.
- Host content on a local IPFS node or export it to a CAR archive.
- Generate a new biblatex file containing the new URLs of the content on IPFS.
  Both `ipfs://` and gateway URLs are supported.
- Pulls open-access full-text articles from [Unpaywall](https://unpaywall.org/).
- Configure custom link resolvers for accessing full-text articles through your
  educational institution or any service that provides open access to research.
- Can access local full-text articles downloaded by your reference manager.
- Can take snapshots of web pages using
  [monolith](https://github.com/Y2Z/monolith).
- Pulls embedded documents from sites that don't serve PDFs directly.
- Can produce machine-readable JSON output for hacking and scripting.
