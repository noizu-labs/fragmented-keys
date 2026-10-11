package pypickle

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sort"
)

// ErrUnsupportedType means a Go value has no representation in the JSON-shaped
// pickle subset — a struct, channel, function, or a map with non-string keys.
var ErrUnsupportedType = errors.New("pypickle: unsupported Go type")

// ErrDepthExceeded means the value nests deeper than maxDepth, which usually
// indicates a cycle. Pickle can express cycles via the memo; this package
// deliberately cannot, because the JSON model has no cycles either.
var ErrDepthExceeded = errors.New("pypickle: maximum nesting depth exceeded")

const (
	// batchSize matches CPython's Pickler._BATCHSIZE, so container framing
	// lines up with what Python would have produced.
	batchSize = 1000

	// frameMin matches CPython's _Framer._FRAME_SIZE_MIN. Frames shorter than
	// this are written without a FRAME header at all, so pickle.dumps(None, 4)
	// is four bytes and carries no frame. Emitting one anyway would still be
	// valid pickle, but would not be byte-identical to CPython.
	frameMin = 4

	// frameTarget matches CPython's _FRAME_SIZE_TARGET. Above this, CPython
	// splits output across multiple frames; Marshal always writes one. The
	// result is still valid, but byte-identity with CPython only holds below
	// this size.
	frameTarget = 64 * 1024

	// maxDepth bounds recursion on cyclic or hostile input.
	maxDepth = 256
)

// Marshal encodes v as a protocol-4 pickle that CPython can load.
//
// Supported: nil, bool, all signed and unsigned integer kinds, *big.Int,
// float32, float64, string, []byte, []any, Tuple, map[string]any, and — via
// reflection — any slice, array, or string-keyed map whose elements are
// themselves supported. Pointers are dereferenced; a nil pointer encodes as
// None.
//
// Output is byte-identical to pickle.dumps(v, 4) for scalars, strings, bytes,
// and tuples. Dicts differ: CPython emits keys in insertion order, which a Go
// map does not have, so Marshal sorts keys to stay deterministic. Both forms
// unpickle to the same object.
func Marshal(v any) ([]byte, error) {
	e := &encoder{}
	if err := e.encode(v, 0); err != nil {
		return nil, err
	}

	body := e.buf.Bytes()
	out := make([]byte, 0, len(body)+11)
	out = append(out, opProto, DefaultProtocol)

	// Protocol 4 wraps the remainder in a frame whose length covers everything
	// after the frame header, STOP included — but only once that payload
	// reaches frameMin bytes, matching CPython.
	frameLen := len(body) + 1
	if frameLen >= frameMin {
		out = append(out, opFrame)
		out = binary.LittleEndian.AppendUint64(out, uint64(frameLen))
	}
	out = append(out, body...)
	out = append(out, opStop)
	return out, nil
}

type encoder struct {
	buf bytes.Buffer
}

//nolint:gocyclo // Mirrors CPython's save() dispatch; flat is clearer here.
func (e *encoder) encode(v any, depth int) error {
	if depth > maxDepth {
		return ErrDepthExceeded
	}

	switch t := v.(type) {
	case nil:
		e.buf.WriteByte(opNone)
		return nil
	case bool:
		if t {
			e.buf.WriteByte(opNewTrue)
		} else {
			e.buf.WriteByte(opNewFalse)
		}
		return nil

	case int:
		return e.encodeInt(int64(t))
	case int8:
		return e.encodeInt(int64(t))
	case int16:
		return e.encodeInt(int64(t))
	case int32:
		return e.encodeInt(int64(t))
	case int64:
		return e.encodeInt(t)
	case uint8:
		return e.encodeInt(int64(t))
	case uint16:
		return e.encodeInt(int64(t))
	case uint32:
		return e.encodeInt(int64(t))
	case uint:
		return e.encodeUint(uint64(t))
	case uint64:
		return e.encodeUint(t)
	case *big.Int:
		if t == nil {
			e.buf.WriteByte(opNone)
			return nil
		}
		return e.encodeBig(t)

	case float32:
		return e.encodeFloat(float64(t))
	case float64:
		return e.encodeFloat(t)

	case string:
		e.encodeString(t)
		return nil
	case []byte:
		e.encodeBytes(t)
		return nil

	case Tuple:
		return e.encodeTuple(t, depth)
	case []any:
		return e.encodeList(t, depth)
	case map[string]any:
		return e.encodeDict(t, depth)
	}

	return e.encodeReflect(v, depth)
}

// encodeReflect handles the container shapes not named explicitly above —
// []string, map[string]string, []int, and so on — by walking them generically.
func (e *encoder) encodeReflect(v any, depth int) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			e.buf.WriteByte(opNone)
			return nil
		}
		return e.encode(rv.Elem().Interface(), depth)

	case reflect.Slice, reflect.Array:
		// []byte is handled above; any other byte slice type still counts.
		if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() == reflect.Uint8 {
			e.encodeBytes(rv.Bytes())
			return nil
		}
		items := make([]any, rv.Len())
		for i := range items {
			items[i] = rv.Index(i).Interface()
		}
		return e.encodeList(items, depth)

	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("%w: map with %s keys", ErrNonStringKey, rv.Type().Key())
		}
		m := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			m[iter.Key().String()] = iter.Value().Interface()
		}
		return e.encodeDict(m, depth)
	}

	return fmt.Errorf("%w: %T", ErrUnsupportedType, v)
}

// ---- scalars --------------------------------------------------------------

// encodeInt mirrors CPython's save_long opcode selection so small integers
// come out byte-identical.
func (e *encoder) encodeInt(n int64) error {
	switch {
	case n >= 0 && n <= 0xff:
		e.buf.WriteByte(opBinInt1)
		e.buf.WriteByte(byte(n))
	case n >= 0 && n <= 0xffff:
		e.buf.WriteByte(opBinInt2)
		e.buf.Write(binary.LittleEndian.AppendUint16(nil, uint16(n)))
	case n >= math.MinInt32 && n <= math.MaxInt32:
		e.buf.WriteByte(opBinInt)
		e.buf.Write(binary.LittleEndian.AppendUint32(nil, uint32(int32(n))))
	default:
		return e.encodeBig(big.NewInt(n))
	}
	return nil
}

func (e *encoder) encodeUint(n uint64) error {
	if n <= math.MaxInt64 {
		return e.encodeInt(int64(n))
	}
	return e.encodeBig(new(big.Int).SetUint64(n))
}

// encodeBig writes LONG1/LONG4 using CPython's encode_long representation:
// little-endian two's complement in the shortest form that keeps the sign bit
// unambiguous, with zero encoded as no bytes at all.
func (e *encoder) encodeBig(n *big.Int) error {
	raw := encodeLong(n)
	if len(raw) < 256 {
		e.buf.WriteByte(opLong1)
		e.buf.WriteByte(byte(len(raw)))
	} else {
		e.buf.WriteByte(opLong4)
		e.buf.Write(binary.LittleEndian.AppendUint32(nil, uint32(len(raw))))
	}
	e.buf.Write(raw)
	return nil
}

func (e *encoder) encodeFloat(f float64) error {
	e.buf.WriteByte(opBinFloat)
	e.buf.Write(binary.BigEndian.AppendUint64(nil, math.Float64bits(f)))
	return nil
}

func (e *encoder) encodeString(s string) {
	b := []byte(s)
	switch {
	case len(b) < 256:
		e.buf.WriteByte(opShortBinUnicode)
		e.buf.WriteByte(byte(len(b)))
	case uint64(len(b)) > math.MaxUint32:
		e.buf.WriteByte(opBinUnicode8)
		e.buf.Write(binary.LittleEndian.AppendUint64(nil, uint64(len(b))))
	default:
		e.buf.WriteByte(opBinUnicode)
		e.buf.Write(binary.LittleEndian.AppendUint32(nil, uint32(len(b))))
	}
	e.buf.Write(b)
	e.memoize()
}

func (e *encoder) encodeBytes(b []byte) {
	switch {
	case len(b) < 256:
		e.buf.WriteByte(opShortBinBytes)
		e.buf.WriteByte(byte(len(b)))
	case uint64(len(b)) > math.MaxUint32:
		e.buf.WriteByte(opBinBytes8)
		e.buf.Write(binary.LittleEndian.AppendUint64(nil, uint64(len(b))))
	default:
		e.buf.WriteByte(opBinBytes)
		e.buf.Write(binary.LittleEndian.AppendUint32(nil, uint32(len(b))))
	}
	e.buf.Write(b)
	e.memoize()
}

// ---- containers -----------------------------------------------------------

func (e *encoder) encodeTuple(t Tuple, depth int) error {
	if len(t) == 0 {
		// CPython emits the singleton EMPTY_TUPLE and does not memoize it.
		e.buf.WriteByte(opEmptyTuple)
		return nil
	}
	if len(t) <= 3 {
		for _, item := range t {
			if err := e.encode(item, depth+1); err != nil {
				return err
			}
		}
		e.buf.WriteByte(byte(opTuple1 + len(t) - 1))
		e.memoize()
		return nil
	}
	e.buf.WriteByte(opMark)
	for _, item := range t {
		if err := e.encode(item, depth+1); err != nil {
			return err
		}
	}
	e.buf.WriteByte(opTuple)
	e.memoize()
	return nil
}

func (e *encoder) encodeList(items []any, depth int) error {
	e.buf.WriteByte(opEmptyList)
	e.memoize()

	// Batched exactly as CPython's _batch_appends does.
	for start := 0; start < len(items); start += batchSize {
		end := min(start+batchSize, len(items))
		batch := items[start:end]
		if len(batch) == 1 {
			if err := e.encode(batch[0], depth+1); err != nil {
				return err
			}
			e.buf.WriteByte(opAppend)
			continue
		}
		e.buf.WriteByte(opMark)
		for _, item := range batch {
			if err := e.encode(item, depth+1); err != nil {
				return err
			}
		}
		e.buf.WriteByte(opAppends)
	}
	return nil
}

func (e *encoder) encodeDict(m map[string]any, depth int) error {
	e.buf.WriteByte(opEmptyDict)
	e.memoize()

	// Go maps have no order, so sort for reproducible output. CPython would
	// have used insertion order; the unpickled dict is equal either way.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for start := 0; start < len(keys); start += batchSize {
		end := min(start+batchSize, len(keys))
		batch := keys[start:end]
		if len(batch) == 1 {
			e.encodeString(batch[0])
			if err := e.encode(m[batch[0]], depth+1); err != nil {
				return err
			}
			e.buf.WriteByte(opSetItem)
			continue
		}
		e.buf.WriteByte(opMark)
		for _, k := range batch {
			e.encodeString(k)
			if err := e.encode(m[k], depth+1); err != nil {
				return err
			}
		}
		e.buf.WriteByte(opSetItems)
	}
	return nil
}

// memoize emits MEMOIZE for the object just written. This package never emits
// GET, so the memo is write-only; it exists so the byte stream matches what
// CPython produces and so any future reader's indices line up.
func (e *encoder) memoize() {
	e.buf.WriteByte(opMemoize)
}

// encodeLong is the inverse of decodeLong: CPython's encode_long.
func encodeLong(n *big.Int) []byte {
	if n.Sign() == 0 {
		return nil
	}

	// Work in two's complement, big-endian, then reverse.
	var be []byte
	if n.Sign() > 0 {
		be = n.Bytes()
		// Keep the sign bit clear so the value reads back as positive.
		if be[0]&0x80 != 0 {
			be = append([]byte{0x00}, be...)
		}
	} else {
		// Two's complement of |n| in the smallest byte width that fits.
		bits := n.BitLen()
		width := bits/8 + 1
		for {
			shift := new(big.Int).Lsh(big.NewInt(1), uint(width*8))
			tc := new(big.Int).Add(shift, n) // shift - |n|
			candidate := tc.Bytes()
			// Left-pad to the full width.
			if len(candidate) < width {
				pad := make([]byte, width-len(candidate))
				candidate = append(pad, candidate...)
			}
			// The sign bit must be set for the value to read back negative.
			if candidate[0]&0x80 != 0 {
				be = candidate
				break
			}
			width++
		}
	}

	le := make([]byte, len(be))
	for i, b := range be {
		le[len(be)-1-i] = b
	}
	return le
}
