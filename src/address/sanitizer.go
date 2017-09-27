// Package sanitize :: For sanitizing text.
package address

import (
	"bytes"
	"regexp"
	"strings"
)

// We are very restrictive as this is intended for ascii url slugs
//var illegalPath = regexp.MustCompile(`[^[:alnum:]\~\-\./]`)

// Remove all other unrecognised characters apart from
var illegalChars = regexp.MustCompile(`[^[:alnum:]-.,/]`)
var illegalCharsForName = regexp.MustCompile(`[^[:alpha:]-.]`)

// A list of characters we consider separators in normal strings and replace with our canonical separator - rather than removing.
var (
	separators = regexp.MustCompile(`[ &_=+:]`)
	//dashes = regexp.MustCompile(`[\-]+`)
	whiteSpaces = regexp.MustCompile("\\s+")
)

// Name makes a string safe to use by replacing non-ascii characters.
func sanitize(s string, isValidateName bool) string {
	// Remove illegal characters for names, replacing some common separators with -
	if isValidateName {
		s = cleanString(s, illegalCharsForName)
	} else {
		s = cleanString(s, illegalChars)
	}
	// NB this may be of length 0, caller must check
	return s
}

// A very limited list of transliterations to catch common european names translated to urls.
// This set could be expanded with at least caps and many more characters.
var transliterations = map[rune]string{
	'À': "A",
	'Á': "A",
	'Â': "A",
	'Ã': "A",
	'Ä': "A",
	'Å': "AA",
	'Æ': "AE",
	'Ç': "C",
	'È': "E",
	'É': "E",
	'Ê': "E",
	'Ë': "E",
	'Ì': "I",
	'Í': "I",
	'Î': "I",
	'Ï': "I",
	'Ð': "D",
	'Ł': "L",
	'Ñ': "N",
	'Ò': "O",
	'Ó': "O",
	'Ô': "O",
	'Õ': "O",
	'Ö': "O",
	'Ø': "OE",
	'Ù': "U",
	'Ú': "U",
	'Ü': "U",
	'Û': "U",
	'Ý': "Y",
	'Þ': "Th",
	'ß': "ss",
	'à': "a",
	'á': "a",
	'â': "a",
	'ã': "a",
	'ä': "a",
	'å': "aa",
	'æ': "ae",
	'ç': "c",
	'è': "e",
	'é': "e",
	'ê': "e",
	'ë': "e",
	'ì': "i",
	'í': "i",
	'î': "i",
	'ï': "i",
	'ð': "d",
	'ł': "l",
	'ñ': "n",
	'ń': "n",
	'ò': "o",
	'ó': "o",
	'ô': "o",
	'õ': "o",
	'ō': "o",
	'ö': "o",
	'ø': "oe",
	'ś': "s",
	'ù': "u",
	'ú': "u",
	'û': "u",
	'ū': "u",
	'ü': "u",
	'ý': "y",
	'þ': "th",
	'ÿ': "y",
	'ż': "z",
	'Œ': "OE",
	'œ': "oe",
}

// Accents replaces a set of accented characters with ascii equivalents.
func Accents(s string) string {
	// Replace some common accent characters
	b := bytes.NewBufferString("")
	for _, c := range s {
		// Check transliterations first
		if val, ok := transliterations[c]; ok {
			b.WriteString(val)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// cleanString replaces separators with - and removes characters listed in the regexp provided from string.
// Accents, spaces, and all characters not in A-Za-z0-9 are replaced.
func cleanString(s string, r *regexp.Regexp) string {

	// Remove any trailing space to avoid ending on -
	s = strings.Trim(s, " ")

	// Flatten accents first so that if we remove non-ascii we still get a legible name
	s = Accents(s)

	// Replace certain joining characters with a dash
	s = separators.ReplaceAllString(s, " ")

	// Remove all other unrecognised characters - NB we do allow any printable characters
	s = r.ReplaceAllString(s, " ")

	// Remove any multiple dashes caused by replacements above
	s = whiteSpaces.ReplaceAllString(s, " ")

	s = strings.TrimSpace(s)
	return s
}
