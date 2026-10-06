commonmark-0.31.2.json is the complete, unmodified official CommonMark 0.31.2
example suite (652 examples), downloaded from:
https://spec.commonmark.org/0.31.2/spec.json

Specification: https://spec.commonmark.org/0.31.2/
Upstream source: https://github.com/commonmark/commonmark-spec

The examples are derived from the CommonMark specification (spec.txt),
Copyright (C) 2014-16 John MacFarlane, released under Creative Commons
Attribution-ShareAlike 4.0:
https://creativecommons.org/licenses/by-sa/4.0/

These terms apply to the imported specification examples. The library's
source code continues to use the license in LICENSE.txt.

Run the conformance suite offline with: go test . -run '^TestCommonMark$'
Each case compares exact HTML, including escaping and whitespace, with
CommonMark enabled and either NoExtensions or CommonExtensions selected.
