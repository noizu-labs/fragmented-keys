package pypickle

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Tuple is a Python tuple. It is distinct from []any (a Python list) so that
// a value round-trips back to Python as the same type it arrived as.
type Tuple []any

var (
	// ErrUnsupportedOpcode means the pickle contains a construct outside the
	// JSON-shaped subset — typically a class instance, datetime, Decimal, or
	// set. Callers should treat this as a cache miss, not a failure.
	ErrUnsupportedOpcode = errors.New("pypickle: unsupported opcode")

	// ErrNonStringKey means a dict used a non-string key. Go's map[string]any
	// cannot represent it and JSON has no such concept.
	ErrNonStringKey = errors.New("pypickle: dict key is not a string")

	// ErrTruncated means the input ended mid-opcode.
	ErrTruncated = errors.New("pypickle: truncated input")

	// ErrProtocol means the pickle declares a protocol this package cannot read.
	ErrProtocol = errors.New("pypickle: unsupported protocol version")
)

// mark is the sentinel the pickle machine pushes for MARK, delimiting the
// start of a variable-length item group on the stack.
type mark struct{}

// Unmarshal decodes a Python pickle into the corresponding Go value.
//
//	None      -> nil
//	bool      -> bool
//	int       -> int64, or *big.Int when it exceeds 64 bits
//	float     -> float64
//	str       -> string
//	bytes     -> []byte
//	list      -> []any
//	tuple     -> Tuple
//	dict      -> map[string]any
//
// Protocols 0 through 5 are accepted. Anything outside the type list above
// returns ErrUnsupportedOpcode.
func Unmarshal(data []byte) (any, error) {
	d := &decoder{buf: data, memo: map[int]any{}}
	return d.run()
}

type decoder struct {
	buf   []byte
	pos   int
	stack []any
	memo  map[int]any
}

func (d *decoder) run() (any, error) {
	for {
		op, err := d.byteAt()
		if err != nil {
			// A pickle without STOP is malformed, not merely exhausted.
			return nil, ErrTruncated
		}
		if op == opStop {
			v, err := d.pop()
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		if err := d.step(op); err != nil {
			return nil, err
		}
	}
}

// step executes one opcode.
//
// It is a single flat switch on purpose: the case order and naming track
// CPython's Lib/pickle.py load_* dispatch, so the two can be read side by side.
//
//nolint:gocyclo // Inherent to an opcode interpreter.
func (d *decoder) step(op byte) error {
	switch op {

	// ---- framing and protocol -------------------------------------------
	case opProto:
		v, err := d.byteAt()
		if err != nil {
			return ErrTruncated
		}
		if v > maxProtocol {
			return fmt.Errorf("%w: %d", ErrProtocol, v)
		}
		return nil
	case opFrame:
		// Frame lengths are advisory; the opcode stream is self-delimiting.
		_, err := d.uint64LE()
		return err

	// ---- singletons ------------------------------------------------------
	case opNone:
		return d.push(nil)
	case opNewTrue:
		return d.push(true)
	case opNewFalse:
		return d.push(false)

	// ---- integers --------------------------------------------------------
	case opBinInt1:
		b, err := d.byteAt()
		if err != nil {
			return ErrTruncated
		}
		return d.push(int64(b))
	case opBinInt2:
		v, err := d.uint16LE()
		if err != nil {
			return err
		}
		return d.push(int64(v))
	case opBinInt:
		v, err := d.uint32LE()
		if err != nil {
			return err
		}
		return d.push(int64(int32(v)))
	case opInt:
		line, err := d.line()
		if err != nil {
			return err
		}
		// Protocol 0 encodes bools as "I01\n" / "I00\n".
		switch line {
		case "01":
			return d.push(true)
		case "00":
			return d.push(false)
		}
		n, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return fmt.Errorf("pypickle: bad INT %q: %w", line, err)
		}
		return d.push(n)
	case opLong:
		line, err := d.line()
		if err != nil {
			return err
		}
		return d.pushBigString(strings.TrimSuffix(line, "L"))
	case opLong1:
		n, err := d.byteAt()
		if err != nil {
			return ErrTruncated
		}
		raw, err := d.take(int(n))
		if err != nil {
			return err
		}
		return d.push(normalizeBig(decodeLong(raw)))
	case opLong4:
		n, err := d.uint32LE()
		if err != nil {
			return err
		}
		raw, err := d.take(int(int32(n)))
		if err != nil {
			return err
		}
		return d.push(normalizeBig(decodeLong(raw)))

	// ---- floats ----------------------------------------------------------
	case opBinFloat:
		raw, err := d.take(8)
		if err != nil {
			return err
		}
		return d.push(math.Float64frombits(binary.BigEndian.Uint64(raw)))
	case opFloat:
		line, err := d.line()
		if err != nil {
			return err
		}
		f, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return fmt.Errorf("pypickle: bad FLOAT %q: %w", line, err)
		}
		return d.push(f)

	// ---- strings ---------------------------------------------------------
	case opShortBinUnicode:
		n, err := d.byteAt()
		if err != nil {
			return ErrTruncated
		}
		raw, err := d.take(int(n))
		if err != nil {
			return err
		}
		return d.push(string(raw))
	case opBinUnicode:
		n, err := d.uint32LE()
		if err != nil {
			return err
		}
		raw, err := d.take(int(n))
		if err != nil {
			return err
		}
		return d.push(string(raw))
	case opBinUnicode8:
		n, err := d.uint64LE()
		if err != nil {
			return err
		}
		raw, err := d.take(int(n))
		if err != nil {
			return err
		}
		return d.push(string(raw))
	case opUnicode:
		line, err := d.line()
		if err != nil {
			return err
		}
		return d.push(decodeRawUnicodeEscape(line))

	// ---- bytes -----------------------------------------------------------
	case opShortBinBytes:
		n, err := d.byteAt()
		if err != nil {
			return ErrTruncated
		}
		raw, err := d.take(int(n))
		if err != nil {
			return err
		}
		return d.push(copyBytes(raw))
	case opBinBytes:
		n, err := d.uint32LE()
		if err != nil {
			return err
		}
		raw, err := d.take(int(n))
		if err != nil {
			return err
		}
		return d.push(copyBytes(raw))
	case opBinBytes8, opByteArray8:
		n, err := d.uint64LE()
		if err != nil {
			return err
		}
		raw, err := d.take(int(n))
		if err != nil {
			return err
		}
		return d.push(copyBytes(raw))

	// Protocol 0/1 "string" opcodes. On Python 3 these decode to bytes when
	// the unpickler runs with encoding="bytes"; django-redis uses the default
	// ("ASCII" -> str), so we produce a string to match.
	case opShortBinString:
		n, err := d.byteAt()
		if err != nil {
			return ErrTruncated
		}
		raw, err := d.take(int(n))
		if err != nil {
			return err
		}
		return d.push(string(raw))
	case opBinString:
		n, err := d.uint32LE()
		if err != nil {
			return err
		}
		raw, err := d.take(int(int32(n)))
		if err != nil {
			return err
		}
		return d.push(string(raw))
	case opString:
		line, err := d.line()
		if err != nil {
			return err
		}
		s, err := unquotePyString(line)
		if err != nil {
			return err
		}
		return d.push(s)

	// ---- containers ------------------------------------------------------
	case opEmptyList:
		return d.push([]any{})
	case opEmptyDict:
		return d.push(map[string]any{})
	case opEmptyTuple:
		return d.push(Tuple{})
	case opMark:
		return d.push(mark{})

	case opList:
		items, err := d.popMark()
		if err != nil {
			return err
		}
		return d.push(items)
	case opTuple:
		items, err := d.popMark()
		if err != nil {
			return err
		}
		return d.push(Tuple(items))
	case opTuple1, opTuple2, opTuple3:
		n := int(op-opTuple1) + 1
		items, err := d.popN(n)
		if err != nil {
			return err
		}
		return d.push(Tuple(items))
	case opDict:
		items, err := d.popMark()
		if err != nil {
			return err
		}
		m := make(map[string]any, len(items)/2)
		if err := setItems(m, items); err != nil {
			return err
		}
		return d.push(m)

	case opAppend:
		v, err := d.pop()
		if err != nil {
			return err
		}
		return d.appendToList([]any{v})
	case opAppends:
		items, err := d.popMark()
		if err != nil {
			return err
		}
		return d.appendToList(items)

	case opSetItem:
		kv, err := d.popN(2)
		if err != nil {
			return err
		}
		return d.setOnDict(kv)
	case opSetItems:
		items, err := d.popMark()
		if err != nil {
			return err
		}
		return d.setOnDict(items)

	// ---- memo ------------------------------------------------------------
	case opMemoize:
		v, err := d.peek()
		if err != nil {
			return err
		}
		d.memo[len(d.memo)] = v
		return nil
	case opPut:
		line, err := d.line()
		if err != nil {
			return err
		}
		i, err := strconv.Atoi(line)
		if err != nil {
			return fmt.Errorf("pypickle: bad PUT index %q: %w", line, err)
		}
		v, err := d.peek()
		if err != nil {
			return err
		}
		d.memo[i] = v
		return nil
	case opBinPut:
		i, err := d.byteAt()
		if err != nil {
			return ErrTruncated
		}
		v, err := d.peek()
		if err != nil {
			return err
		}
		d.memo[int(i)] = v
		return nil
	case opLongBinPut:
		i, err := d.uint32LE()
		if err != nil {
			return err
		}
		v, err := d.peek()
		if err != nil {
			return err
		}
		d.memo[int(i)] = v
		return nil
	case opGet:
		line, err := d.line()
		if err != nil {
			return err
		}
		i, err := strconv.Atoi(line)
		if err != nil {
			return fmt.Errorf("pypickle: bad GET index %q: %w", line, err)
		}
		return d.pushMemo(i)
	case opBinGet:
		i, err := d.byteAt()
		if err != nil {
			return ErrTruncated
		}
		return d.pushMemo(int(i))
	case opLongBinGet:
		i, err := d.uint32LE()
		if err != nil {
			return err
		}
		return d.pushMemo(int(i))

	// ---- stack shuffling -------------------------------------------------
	case opPop:
		_, err := d.pop()
		return err
	case opPopMark:
		_, err := d.popMark()
		return err
	case opDup:
		v, err := d.peek()
		if err != nil {
			return err
		}
		return d.push(v)

	default:
		return fmt.Errorf("%w: 0x%02x (%q)", ErrUnsupportedOpcode, op, string(rune(op)))
	}
}

// appendToList extends the list currently below the popped items.
func (d *decoder) appendToList(items []any) error {
	target, err := d.peek()
	if err != nil {
		return err
	}
	lst, ok := target.([]any)
	if !ok {
		return fmt.Errorf("pypickle: APPEND onto %T, want list", target)
	}
	d.stack[len(d.stack)-1] = append(lst, items...)
	return nil
}

// setOnDict applies alternating key/value pairs to the dict below them.
func (d *decoder) setOnDict(items []any) error {
	target, err := d.peek()
	if err != nil {
		return err
	}
	m, ok := target.(map[string]any)
	if !ok {
		return fmt.Errorf("pypickle: SETITEM onto %T, want dict", target)
	}
	return setItems(m, items)
}

func setItems(m map[string]any, items []any) error {
	if len(items)%2 != 0 {
		return fmt.Errorf("pypickle: odd number of dict items (%d)", len(items))
	}
	for i := 0; i < len(items); i += 2 {
		k, ok := items[i].(string)
		if !ok {
			return fmt.Errorf("%w: %T", ErrNonStringKey, items[i])
		}
		m[k] = items[i+1]
	}
	return nil
}

func (d *decoder) pushMemo(i int) error {
	v, ok := d.memo[i]
	if !ok {
		return fmt.Errorf("pypickle: memo index %d not set", i)
	}
	return d.push(v)
}

func (d *decoder) pushBigString(s string) error {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return fmt.Errorf("pypickle: bad LONG %q", s)
	}
	return d.push(normalizeBig(n))
}

// ---- stack helpers --------------------------------------------------------

// maxStack bounds memory use on a hostile or corrupt payload.
const maxStack = 1 << 20

func (d *decoder) push(v any) error {
	if len(d.stack) >= maxStack {
		return errors.New("pypickle: stack overflow")
	}
	d.stack = append(d.stack, v)
	return nil
}

func (d *decoder) pop() (any, error) {
	if len(d.stack) == 0 {
		return nil, errors.New("pypickle: stack underflow")
	}
	v := d.stack[len(d.stack)-1]
	d.stack = d.stack[:len(d.stack)-1]
	return v, nil
}

func (d *decoder) peek() (any, error) {
	if len(d.stack) == 0 {
		return nil, errors.New("pypickle: stack underflow")
	}
	return d.stack[len(d.stack)-1], nil
}

func (d *decoder) popN(n int) ([]any, error) {
	if len(d.stack) < n {
		return nil, errors.New("pypickle: stack underflow")
	}
	out := make([]any, n)
	copy(out, d.stack[len(d.stack)-n:])
	d.stack = d.stack[:len(d.stack)-n]
	return out, nil
}

// popMark pops everything above the most recent mark, and the mark itself.
func (d *decoder) popMark() ([]any, error) {
	for i := len(d.stack) - 1; i >= 0; i-- {
		if _, isMark := d.stack[i].(mark); isMark {
			out := make([]any, len(d.stack)-i-1)
			copy(out, d.stack[i+1:])
			d.stack = d.stack[:i]
			return out, nil
		}
	}
	return nil, errors.New("pypickle: no MARK on stack")
}

// ---- input helpers --------------------------------------------------------

func (d *decoder) byteAt() (byte, error) {
	if d.pos >= len(d.buf) {
		return 0, ErrTruncated
	}
	b := d.buf[d.pos]
	d.pos++
	return b, nil
}

func (d *decoder) take(n int) ([]byte, error) {
	if n < 0 || d.pos+n > len(d.buf) {
		return nil, ErrTruncated
	}
	b := d.buf[d.pos : d.pos+n]
	d.pos += n
	return b, nil
}

func (d *decoder) uint16LE() (uint16, error) {
	b, err := d.take(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

func (d *decoder) uint32LE() (uint32, error) {
	b, err := d.take(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (d *decoder) uint64LE() (uint64, error) {
	b, err := d.take(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

// line reads through the next newline, which protocol 0 uses as a terminator.
func (d *decoder) line() (string, error) {
	i := strings.IndexByte(string(d.buf[d.pos:]), '\n')
	if i < 0 {
		return "", ErrTruncated
	}
	s := string(d.buf[d.pos : d.pos+i])
	d.pos += i + 1
	return strings.TrimSuffix(s, "\r"), nil
}

// ---- long integer encoding ------------------------------------------------

// decodeLong reverses CPython's encode_long: little-endian two's complement,
// shortest form, empty meaning zero.
func decodeLong(raw []byte) *big.Int {
	if len(raw) == 0 {
		return big.NewInt(0)
	}
	// Reverse into big-endian.
	be := make([]byte, len(raw))
	for i, b := range raw {
		be[len(raw)-1-i] = b
	}
	n := new(big.Int).SetBytes(be)
	// Negative when the most significant bit of the top byte is set.
	if be[0]&0x80 != 0 {
		shift := new(big.Int).Lsh(big.NewInt(1), uint(len(be)*8))
		n.Sub(n, shift)
	}
	return n
}

// copyBytes detaches a slice from the input buffer. It always allocates, so
// an empty bytes value decodes to []byte{} rather than a nil slice — b"" and
// None are different values in Python and must stay different here.
func copyBytes(raw []byte) []byte {
	out := make([]byte, len(raw))
	copy(out, raw)
	return out
}

// normalizeBig narrows to int64 when the value fits, so ordinary integers
// arrive as int64 rather than *big.Int.
func normalizeBig(n *big.Int) any {
	if n.IsInt64() {
		return n.Int64()
	}
	return n
}

// ---- protocol 0 string helpers --------------------------------------------

// unquotePyString parses the repr-quoted form used by the STRING opcode.
func unquotePyString(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return "", fmt.Errorf("pypickle: bad STRING %q", s)
	}
	q := s[0]
	if (q != '\'' && q != '"') || s[len(s)-1] != q {
		return "", fmt.Errorf("pypickle: bad STRING quoting %q", s)
	}
	return decodeEscapes(s[1 : len(s)-1]), nil
}

// decodeRawUnicodeEscape handles the UNICODE opcode's raw-unicode-escape form,
// where only \uXXXX sequences are special.
func decodeRawUnicodeEscape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '\\' && i+5 < len(s)+1 && i+1 < len(s) && s[i+1] == 'u' && i+6 <= len(s) {
			if n, err := strconv.ParseUint(s[i+2:i+6], 16, 32); err == nil {
				b.WriteRune(rune(n))
				i += 6
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// decodeEscapes handles the backslash escapes CPython's repr emits.
func decodeEscapes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '\\':
			b.WriteByte('\\')
		case '\'':
			b.WriteByte('\'')
		case '"':
			b.WriteByte('"')
		case 'x':
			if i+2 < len(s) {
				if n, err := strconv.ParseUint(s[i+1:i+3], 16, 8); err == nil {
					b.WriteByte(byte(n))
					i += 2
					continue
				}
			}
			b.WriteByte('x')
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
