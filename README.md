# DirSearch

This software is a Go implementation of the original [dirsearch tool](https://github.com/maurosoria/dirsearch) written by `Mauro Soria`.
DirSearch is the very first tool I write in Go, mostly to play and experiment with Go's concurrency model, channels, and so forth :)

[![baby-gopher](https://raw.githubusercontent.com/drnic/babygopher-site/gh-pages/images/babygopher-badge.png)](http://www.babygopher.org)

## Purpose

DirSearch takes an input URL ( `-url` parameter ) and a wordlist ( `-wordlist` parameter ), it will then perform concurrent `HEAD` requests
using the lines of the wordlist as paths and files.

    Usage of dirsearch:
      -200only
            If enabled, will only display responses with 200 status code.
      -consumers int
            Number of concurrent consumers. (default 8)
      -ext string
            File extension. (default "php")
      -maxerrors int
            Maximum number of errors to get before killing the program. (default 20)
      -url string
            Base URL to start enumeration from.
      -wordlist string
            Wordlist file to use for enumeration. (default "dict.txt")

## License

This project is copyleft of [Simone Margaritelli](http://www.evilsocket.net/) and released under the GPL 3 license.

