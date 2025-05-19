CCM in Go
=========

Package ccm implements a CCM, Counter with CBC-MAC as per
[RFC 3610](https://tools.ietf.org/html/rfc3610).

This package was orginally named
`bitbucket.org/dchapes/ripple/crypto/ccm`
as part of that Ripple Go repository
(CCM was needed for en/decrypting wallet blobs).

During the migration to Sourcehut I decided to extract just this package
into its own repository since it's likely to be the only thing of
use/value out of the old one
(which has been migrated to
[`hg.sr.ht/~dchapes/ripple`](https://hg.sr.ht/~dchapes/ripple)).

[![Go Reference](https://pkg.go.dev/badge/hg.sr.ht/~dchapes/ccm.svg)](https://pkg.go.dev/hg.sr.ht/~dchapes/ccm)
Online package documentation is available via
[pkg.go.dev](https://pkg.go.dev/hg.sr.ht/~dchapes/ccm).

All code contained within this repository is licensed under a simplified
BSD 2-clause license, see the LICENSE file for details.
