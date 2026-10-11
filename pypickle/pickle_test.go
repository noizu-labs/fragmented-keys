package pypickle

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"reflect"
	"testing"
)

// pickleCase pairs bytes captured from CPython with the Go value they should
// decode to.
//
// The pickle column was produced by:
//
//	python -c 'import pickle; print(pickle.dumps(VALUE, 4).hex())'
//
// on CPython 3.11 — the interpreter gnp-backend runs — using protocol 4, which
// is pickle.DEFAULT_PROTOCOL there and therefore what django-redis writes.
//
// byteIdentical marks cases where Marshal must reproduce CPython's bytes
// exactly. It is false where the two legitimately differ: multi-key dicts
// (CPython uses insertion order, Marshal sorts) and values containing a
// repeated string (CPython back-references it through the memo with BINGET,
// Marshal writes it out again). Both forms unpickle to the same object.
type pickleCase struct {
	name          string
	pickle        []byte
	want          any
	byteIdentical bool
}

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func cases() []pickleCase {
	return []pickleCase{
		{"none", mustHex("80044e2e"), nil, true},
		{"true", mustHex("8004882e"), true, true},
		{"false", mustHex("8004892e"), false, true},

		{"int_zero", mustHex("80044b002e"), int64(0), true},
		{"int_small", mustHex("80044b2a2e"), int64(42), true},
		{"int_255", mustHex("80044bff2e"), int64(255), true},
		{"int_256", mustHex("80049504000000000000004d00012e"), int64(256), true},
		{"int_65535", mustHex("80049504000000000000004dffff2e"), int64(65535), true},
		{"int_65536", mustHex("80049506000000000000004a000001002e"), int64(65536), true},
		{"int_max32", mustHex("80049506000000000000004affffff7f2e"), int64(2147483647), true},
		{"int_neg1", mustHex("80049506000000000000004affffffff2e"), int64(-1), true},
		{"int_min32", mustHex("80049506000000000000004a000000802e"), int64(-2147483648), true},
		{"int_big", mustHex("8004950c000000000000008a090000000000000000402e"), bigPow(70), true},
		{"int_bignegative", mustHex("8004950c000000000000008a090000000000000000c02e"), negBigPow(70), true},
		{"int_2p63", mustHex("8004950c000000000000008a090000000000000080002e"), bigPow(63), true},

		{"float_one", mustHex("8004950a00000000000000473ff00000000000002e"), 1.0, true},
		{"float_drift", mustHex("8004950a000000000000004742796fceb48003342e"), 1748000000000.2002, true},

		{"str_empty", mustHex("80049504000000000000008c00942e"), "", true},
		{"str_short", mustHex("8004950b000000000000008c0763617265657273942e"), "careers", true},
		{"str_utf8", mustHex("80049512000000000000008c0e636166c3a920e2809420f09f9880942e"), "café — 😀", true},

		{"bytes", mustHex("800495080000000000000043040001feff942e"), []byte{0x00, 0x01, 0xfe, 0xff}, true},

		{"list_empty", mustHex("80045d942e"), []any{}, true},
		{
			"list_scalars",
			mustHex("8004951a000000000000005d94284b014740040000000000008c057468726565944e88652e"),
			[]any{int64(1), 2.5, "three", nil, true},
			true,
		},
		{
			"list_nested",
			mustHex("80049514000000000000005d94285d94284b014b02655d948c01619461652e"),
			[]any{[]any{int64(1), int64(2)}, []any{"a"}},
			true,
		},

		{"tuple_empty", mustHex("8004292e"), Tuple{}, true},
		{"tuple1", mustHex("80049505000000000000004b0185942e"), Tuple{int64(1)}, true},
		{"tuple3", mustHex("8004950a000000000000004b018c0162944e87942e"), Tuple{int64(1), "b", nil}, true},
		{
			"tuple5",
			mustHex("8004950e00000000000000284b014b024b034b044b0574942e"),
			Tuple{int64(1), int64(2), int64(3), int64(4), int64(5)},
			true,
		},

		{"dict_empty", mustHex("80047d942e"), map[string]any{}, true},
		{"dict_one", mustHex("8004950a000000000000007d948c0161944b01732e"), map[string]any{"a": int64(1)}, true},
		{
			"dict_multi",
			mustHex("80049537000000000000007d94288c04736c7567948c0763617265657273948c016e944b038c026f6b94888c03737562947d948c0178945d94284b014b026573752e"),
			map[string]any{
				"slug": "careers",
				"n":    int64(3),
				"ok":   true,
				"sub":  map[string]any{"x": []any{int64(1), int64(2)}},
			},
			false, // insertion order vs sorted keys
		},
		{
			// Shaped like an ADR-015 section document, which is the payload
			// this codec exists to move between Django and Go.
			"doc_like",
			mustHex("800495be000000000000007d94288c04736c7567948c0763617265657273948c0474797065948c0e6d61726b6574696e675f70616765948c066c6f63616c65948c05656e5f5553948c0373656f947d948c057469746c65948c074361726565727394738c09757064617465644174948c14323032362d30382d30365430303a30303a30305a948c0873656374696f6e73945d947d94288c026964948c0273319468038c046865726f948c0570726f7073947d948c0768656164696e67948c044a6f696e94737561752e"),
			map[string]any{
				"slug":      "careers",
				"type":      "marketing_page",
				"locale":    "en_US",
				"seo":       map[string]any{"title": "Careers"},
				"updatedAt": "2026-08-06T00:00:00Z",
				"sections": []any{
					map[string]any{
						"id":    "s1",
						"type":  "hero",
						"props": map[string]any{"heading": "Join"},
					},
				},
			},
			false, // insertion order, and BINGET back-reference for "type"
		},
		{
			// Exercises the memo read path: CPython writes "dup" once and
			// refers back to it with BINGET twice.
			"repeated_str",
			mustHex("8004950f000000000000005d94288c036475709468016801652e"),
			[]any{"dup", "dup", "dup"},
			false,
		},
	}
}

func bigPow(n int) *big.Int    { return new(big.Int).Lsh(big.NewInt(1), uint(n)) }
func negBigPow(n int) *big.Int { return new(big.Int).Neg(bigPow(n)) }

// TestUnmarshalCPythonFixtures decodes bytes CPython actually produced.
func TestUnmarshalCPythonFixtures(t *testing.T) {
	for _, c := range cases() {
		t.Run(c.name, func(t *testing.T) {
			got, err := Unmarshal(c.pickle)
			if err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if !equalPy(got, c.want) {
				t.Errorf("Unmarshal = %#v (%T), want %#v (%T)", got, got, c.want, c.want)
			}
		})
	}
}

// TestMarshalMatchesCPython asserts byte-identity where it is expected, and
// semantic equivalence everywhere else.
func TestMarshalMatchesCPython(t *testing.T) {
	for _, c := range cases() {
		t.Run(c.name, func(t *testing.T) {
			got, err := Marshal(c.want)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if c.byteIdentical {
				if !bytes.Equal(got, c.pickle) {
					t.Errorf("Marshal bytes differ from CPython\n got: %x\nwant: %x", got, c.pickle)
				}
				return
			}
			// Not byte-identical by design — require that our own bytes still
			// decode back to the same value.
			back, err := Unmarshal(got)
			if err != nil {
				t.Fatalf("Unmarshal(Marshal(v)): %v", err)
			}
			if !equalPy(back, c.want) {
				t.Errorf("round trip = %#v, want %#v", back, c.want)
			}
		})
	}
}

// TestRoundTrip covers values with no CPython fixture attached.
func TestRoundTrip(t *testing.T) {
	values := []any{
		nil, true, false,
		int64(0), int64(-7), int64(1 << 40), bigPow(200), negBigPow(200),
		0.0, -0.5, 1748000000000.901,
		"", "careers", "café — 😀",
		[]byte{}, []byte("bytes"),
		[]any{}, []any{int64(1), []any{"deep", map[string]any{"k": nil}}},
		Tuple{}, Tuple{int64(1), Tuple{int64(2)}},
		map[string]any{}, map[string]any{"a": int64(1), "b": []any{true, false}},
	}
	for _, v := range values {
		b, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal(%#v): %v", v, err)
		}
		got, err := Unmarshal(b)
		if err != nil {
			t.Fatalf("Unmarshal(%#v): %v", v, err)
		}
		if !equalPy(got, v) {
			t.Errorf("round trip = %#v, want %#v", got, v)
		}
	}
}

// TestLargeContainersBatch pushes past CPython's 1000-item batch boundary,
// where the encoder switches from one APPENDS group to several.
func TestLargeContainersBatch(t *testing.T) {
	list := make([]any, 2500)
	for i := range list {
		list[i] = int64(i % 256)
	}
	dict := make(map[string]any, 2500)
	for i := 0; i < 2500; i++ {
		dict[string(rune('a'+i%26))+string(rune('a'+i/26))] = int64(i)
	}
	for _, v := range []any{list, dict} {
		b, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := Unmarshal(b)
		if err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if !equalPy(got, v) {
			t.Errorf("large container did not round trip")
		}
	}
}

// TestUnsupportedConstructsAreRejected confirms the failure mode is a clean
// error rather than a wrong value — callers treat this as a cache miss.
func TestUnsupportedConstructsAreRejected(t *testing.T) {
	// Real CPython output for constructs Django caches but Go cannot model.
	// Each leans on STACK_GLOBAL/REDUCE or the set opcodes.
	unsupported := map[string]string{
		"datetime.date":   "80049520000000000000008c086461746574696d65948c0464617465949394430407ea080694859452942e",
		"set":             "80049509000000000000008f94284b014b02902e",
		"decimal.Decimal": "80049521000000000000008c07646563696d616c948c07446563696d616c9493948c03312e3594859452942e",
	}
	for name, h := range unsupported {
		if v, err := Unmarshal(mustHex(h)); err == nil {
			t.Errorf("decoding a %s pickle returned %#v; want an error so callers treat it as a miss", name, v)
		}
	}

	// A Go struct has no place in the JSON-shaped subset.
	type payload struct{ A int }
	if _, err := Marshal(payload{A: 1}); err == nil {
		t.Error("expected ErrUnsupportedType marshalling a struct, got nil")
	}

	// Non-string map keys likewise.
	if _, err := Marshal(map[int]string{1: "a"}); err == nil {
		t.Error("expected ErrNonStringKey marshalling map[int]string, got nil")
	}
}

func TestTruncatedInput(t *testing.T) {
	full := mustHex("8004950b000000000000008c0763617265657273942e")
	for n := 1; n < len(full); n++ {
		if _, err := Unmarshal(full[:n]); err == nil {
			t.Errorf("Unmarshal of %d-byte prefix succeeded; want error", n)
		}
	}
}

// equalPy compares decoded values, treating []byte by content and *big.Int by
// numeric value. reflect.DeepEqual alone mishandles neither, but big.Int
// comparison by pointer identity would.
func equalPy(a, b any) bool {
	ab, aok := a.(*big.Int)
	bb, bok := b.(*big.Int)
	if aok || bok {
		return aok && bok && ab.Cmp(bb) == 0
	}
	return reflect.DeepEqual(a, b)
}
