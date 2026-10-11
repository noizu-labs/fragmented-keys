// Package pyfloat renders and parses float64 values exactly as CPython's
// repr()/str() does.
//
// This matters because fragmented keys interpolate tag versions into the raw
// key string with the host language's default float-to-string conversion, and
// the result is then MD5-hashed. A single digit of disagreement between the
// Python and Go renderings produces a completely different cache key, so the
// two runtimes would silently stop sharing cache entries.
//
// Go's own formatting is close but not equal: FormatFloat(v, 'f', -1, 64)
// drops the trailing ".0" that CPython keeps on integral values, and 'g'
// switches to exponent notation on a different threshold. Format below
// reproduces CPython's rule exactly.
//
// Verified byte-for-byte against CPython 3.11 over the +0.1 increment drift
// sequence, the 1e16 notation switch point, subnormals, and random bit
// patterns; see pyfloat_test.go and testdata/pyfloat_repr.txt.
package pyfloat

import (
	"math"
	"strconv"
	"strings"
)

// sciLowerExp and sciUpperExp bracket the decimal exponents CPython renders
// positionally. Outside [sciLowerExp, sciUpperExp) it switches to scientific
// notation. See CPython Python/pystrtod.c, format_float_short().
const (
	sciLowerExp = -4
	sciUpperExp = 16
)

// Format returns f rendered as CPython's repr(float) would render it.
//
// Non-finite values use CPython's spellings ("inf", "-inf", "nan") rather than
// Go's ("+Inf", "-Inf", "NaN"). A tag version should never be non-finite, but
// a corrupted cache entry must not silently produce a plausible-looking key.
func Format(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	case math.IsNaN(f):
		return "nan"
	}

	// The 'e' form with precision -1 gives the shortest round-tripping digit
	// string, which is the same digit string CPython's short repr produces.
	// Its exponent tells us which notation CPython would pick.
	sci := strconv.FormatFloat(f, 'e', -1, 64)
	mantissa, expDigits, _ := strings.Cut(sci, "e")
	exp, err := strconv.Atoi(expDigits)
	if err != nil {
		// Unreachable for finite input; fall back to the positional form
		// rather than emitting a malformed key fragment.
		return positional(f)
	}

	if exp < sciLowerExp || exp >= sciUpperExp {
		return mantissa + "e" + exponent(exp)
	}
	return positional(f)
}

// positional renders f without an exponent, forcing the trailing ".0" that
// CPython keeps on integral values and Go omits.
func positional(f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.ContainsRune(s, '.') {
		s += ".0"
	}
	return s
}

// exponent renders the exponent suffix with CPython's sign and its minimum
// width of two digits ("e+16", "e-05").
func exponent(exp int) string {
	sign := byte('+')
	if exp < 0 {
		sign = '-'
		exp = -exp
	}
	digits := strconv.Itoa(exp)
	if len(digits) < 2 {
		digits = "0" + digits
	}
	return string(sign) + digits
}

// Parse reads a version string produced by any conforming implementation.
//
// It accepts CPython's repr spellings, including the non-finite ones that
// strconv.ParseFloat does not recognise, so a value written by Python always
// round-trips back through Go.
func Parse(s string) (float64, error) {
	switch strings.TrimSpace(s) {
	case "inf", "Infinity":
		return math.Inf(1), nil
	case "-inf", "-Infinity":
		return math.Inf(-1), nil
	case "nan":
		return math.NaN(), nil
	}
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}
