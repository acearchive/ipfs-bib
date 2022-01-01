# ipfs-bib

ipfs-bib is a tool for hosting bibliographic references on
[IPFS](https://ipfs.io). The tool pulls citations from a bibtex/biblatex file
or a public [Zotero](https://zotero.org) library, finds the full text of the
research on the legacy web, and rehosts it on a local IPFS node or exports it
to a [CAR archive](https://ipld.io/specs/transport/car/).

## Features

- Pull citations from a bibtex/biblatex file or a public Zotero library.
- Host content on a local IPFS node or export it to a CAR archive. You can pin
  content on your local node or add it to
  [MFS](https://docs.ipfs.io/concepts/file-systems/#mutable-file-system-mfs).
- Pin content with IPFS [pinning
  services](https://docs.ipfs.io/how-to/work-with-pinning-services/) that
  support the [pinning service
  API](https://github.com/ipfs/pinning-services-api-spec).
- Generate a new biblatex file containing the new URLs of the content on IPFS.
  Both `ipfs://` and gateway URLs are supported.
- Pulls open-access full-text articles from [Unpaywall](https://unpaywall.org/).
- Configure custom link resolvers for accessing full-text articles through your
  educational institution or any service that provides open access to research.
- Can access local full-text articles downloaded by your reference manager.
- Can take snapshots of web pages using
  [monolith](https://github.com/Y2Z/monolith) when a PDF isn't available. This
  requires monolith to be installed separately.
- Pulls embedded documents from sites that don't serve PDFs directly.
- Can produce JSON output for hacking and scripting.

## Configuration

The behavior of the program can be configured via a config file. See [the
default config file](./config/config.toml) for documentation and examples.

The tool will produce a single
[UnixFS](https://docs.ipfs.io/concepts/file-systems/#unix-file-system-unixfs)
directory in IPFS with a subdirectory for each source, and each of those
subdirectories will contain the full-text article. Both the name of the
directory and the name of the file can be configured in the config file.

Some fields in the config file accept a [Golang `text/template`
template](https://pkg.go.dev/text/template). In these templates, functions
provided by the [Sprig](https://masterminds.github.io/sprig/) library are
available.

## Usage

```
A tool for hosting bibliographic references on IPFS.

This command accepts the path of a bibtex/biblatex file, or `-` to read from stdin.
If --zotero is passed, this accepts a Zotero group ID instead.

Usage:
  ipfs-bib [options] <bibtex_file>

Flags:
      --car string          Rather than add the sources to an IPFS node, export them as a CAR archive at this path.
  -c, --config string       The path of the config file to use. Otherwise, use the default config.
      --dry-run             Download sources, but don't add them to IPFS or export them as a CAR.
  -h, --help                help for ipfs-bib
      --json                Produce machine-readable JSON output.
      --mfs string          Add the sources to MFS at this path.
  -o, --output string       Generate a new bibtex file at this path with the IPFS URLs added to each entry.
      --pin                 Pin the source files when adding them to the IPFS node.
      --pin-remote string   Pin the source files using each of the configured IPFS pinning services. Pass a name for the pin.
  -v, --verbose             Print verbose output.
      --version             version for ipfs-bib
      --zotero              Pull references from a public Zotero library. Pass a Zotero group ID.
```
