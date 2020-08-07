# squid
Squid convert to HTML a single markdown file or an entire project offline.
It saves the rendered project or file into a destination directory that can be specified in input.
Additionally it copies eventual non-markdown files or assets found in the project tree, preserving their relative position to the markdown files.
The html produced is standalone and contains its own embedded css.
A different css template can be used with the `-css` flag.

Run `squid --help` for additional information.

### Usage
```
$ squid [OPTIONS] SOURCE [DESTINATION]
```