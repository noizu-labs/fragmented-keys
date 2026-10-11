package pypickle

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// TestDifferentialAgainstCPython round-trips randomly generated values through
// a real CPython interpreter and asserts nothing changes.
//
// Each value makes the full journey:
//
//	Go value -> Marshal -> CPython pickle.loads -> pickle.dumps(proto 4)
//	         -> Unmarshal -> compare with the original Go value
//
// A failure anywhere on that path means Go and Django would disagree about a
// cache payload. Hand-picked fixtures cover the shapes I thought of; this
// covers the ones I did not.
//
// Enable with:
//
//	FK_PYTHON=python3 go test ./pypickle/
//
// Skipped by default so the suite has no interpreter dependency.
func TestDifferentialAgainstCPython(t *testing.T) {
	py := os.Getenv("FK_PYTHON")
	if py == "" {
		t.Skip("set FK_PYTHON to a CPython 3 interpreter to run the differential test")
	}
	if _, err := exec.LookPath(py); err != nil {
		t.Skipf("FK_PYTHON=%q is not executable: %v", py, err)
	}

	// Widen the sweep with FK_FUZZ_N / FK_FUZZ_SEED when chasing a suspected
	// codec bug; the defaults keep a normal test run quick.
	n := envInt("FK_FUZZ_N", 600)
	seed := int64(envInt("FK_FUZZ_SEED", 20260806))
	rng := rand.New(rand.NewSource(seed))

	values := make([]any, 0, n)
	encoded := make([]string, 0, n)
	for i := 0; i < n; i++ {
		v := randomValue(rng, 0)
		b, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal(%#v): %v", v, err)
		}
		values = append(values, v)
		encoded = append(encoded, hex.EncodeToString(b))
	}

	out, err := runPythonRoundTrip(py, encoded)
	if err != nil {
		t.Fatalf("python round trip: %v", err)
	}
	if len(out) != len(values) {
		t.Fatalf("python returned %d lines, sent %d", len(out), len(values))
	}

	var failures int
	for i, line := range out {
		if strings.HasPrefix(line, "ERR ") {
			t.Errorf("case %d: CPython could not load Go's pickle of %#v: %s",
				i, values[i], strings.TrimPrefix(line, "ERR "))
			failures++
			continue
		}
		raw, err := hex.DecodeString(line)
		if err != nil {
			t.Fatalf("case %d: bad hex from python: %v", i, err)
		}
		got, err := Unmarshal(raw)
		if err != nil {
			t.Errorf("case %d: Unmarshal of CPython's re-dump failed: %v", i, err)
			failures++
			continue
		}
		if !deepEqualPy(got, values[i]) {
			t.Errorf("case %d: round trip changed the value\n got: %#v\nwant: %#v", i, got, values[i])
			failures++
		}
		if failures > 10 {
			t.Fatal("too many failures; stopping")
		}
	}
	t.Logf("round-tripped %d randomly generated values through CPython (seed %d)", len(values), seed)
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

// runPythonRoundTrip feeds hex-encoded pickles to CPython, which loads each and
// re-dumps it at protocol 4, returning the result as hex.
func runPythonRoundTrip(py string, hexLines []string) ([]string, error) {
	const script = `
import sys, pickle
out = []
for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    try:
        obj = pickle.loads(bytes.fromhex(line))
        out.append(pickle.dumps(obj, 4).hex())
    except Exception as e:
        out.append("ERR " + repr(e))
sys.stdout.write("\n".join(out))
`
	cmd := exec.Command(py, "-c", script)
	cmd.Stdin = strings.NewReader(strings.Join(hexLines, "\n"))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, stderr.String())
	}

	var lines []string
	sc := bufio.NewScanner(&stdout)
	sc.Buffer(make([]byte, 1<<22), 1<<22)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// randomValue produces a value from the supported type space, nesting up to a
// small depth so containers get exercised inside containers.
func randomValue(rng *rand.Rand, depth int) any {
	kinds := 9
	if depth >= 3 {
		kinds = 6 // scalars only, to terminate
	}
	switch rng.Intn(kinds) {
	case 0:
		return nil
	case 1:
		return rng.Intn(2) == 0
	case 2:
		// Span the BININT1/BININT2/BININT/LONG1 opcode boundaries.
		switch rng.Intn(5) {
		case 0:
			return int64(rng.Intn(256))
		case 1:
			return int64(rng.Intn(65536))
		case 2:
			return int64(rng.Int31()) - (1 << 30)
		case 3:
			return -int64(rng.Intn(1 << 20))
		default:
			return int64(rng.Uint64() >> 1)
		}
	case 3:
		// Include the drifted-version shapes tag storage actually produces.
		switch rng.Intn(4) {
		case 0:
			return rng.NormFloat64()
		case 1:
			return 1748000000000.0 + float64(rng.Intn(50))*0.1
		case 2:
			return math.Float64frombits(rng.Uint64() &^ (0x7FF << 52))
		default:
			return float64(rng.Intn(1000))
		}
	case 4:
		return randomString(rng)
	case 5:
		b := make([]byte, rng.Intn(40))
		rng.Read(b)
		return b
	case 6:
		n := rng.Intn(6)
		out := make([]any, n)
		for i := range out {
			out[i] = randomValue(rng, depth+1)
		}
		return out
	case 7:
		n := rng.Intn(4)
		out := make(Tuple, n)
		for i := range out {
			out[i] = randomValue(rng, depth+1)
		}
		return out
	default:
		n := rng.Intn(6)
		out := make(map[string]any, n)
		for i := 0; i < n; i++ {
			out[randomString(rng)] = randomValue(rng, depth+1)
		}
		return out
	}
}

func randomString(rng *rand.Rand) string {
	// Mix ASCII, multi-byte UTF-8, and lengths straddling the
	// SHORT_BINUNICODE/BINUNICODE boundary at 256 bytes.
	alphabets := []string{
		"abcdefghijklmnopqrstuvwxyz_-0123456789",
		"áéíóúñüçßœ—…",
		"日本語テスト漢字",
		"😀🎉🔥🌍",
	}
	a := []rune(alphabets[rng.Intn(len(alphabets))])
	n := rng.Intn(12)
	if rng.Intn(20) == 0 {
		n = 250 + rng.Intn(20) // straddle the 256-byte opcode switch
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteRune(a[rng.Intn(len(a))])
	}
	return b.String()
}

// deepEqualPy compares values, treating *big.Int numerically and tolerating
// the nil/empty slice distinction Go draws but Python does not.
func deepEqualPy(a, b any) bool {
	ab, aok := a.(*big.Int)
	bb, bok := b.(*big.Int)
	if aok || bok {
		switch {
		case aok && bok:
			return ab.Cmp(bb) == 0
		case aok:
			if n, ok := b.(int64); ok {
				return ab.Cmp(big.NewInt(n)) == 0
			}
		case bok:
			if n, ok := a.(int64); ok {
				return bb.Cmp(big.NewInt(n)) == 0
			}
		}
		return false
	}

	// NaN never equals itself; treat two NaNs as matching.
	if af, ok := a.(float64); ok {
		if bf, ok2 := b.(float64); ok2 && math.IsNaN(af) && math.IsNaN(bf) {
			return true
		}
	}

	switch at := a.(type) {
	case []any:
		bt, ok := b.([]any)
		if !ok || len(at) != len(bt) {
			return false
		}
		for i := range at {
			if !deepEqualPy(at[i], bt[i]) {
				return false
			}
		}
		return true
	case Tuple:
		bt, ok := b.(Tuple)
		if !ok || len(at) != len(bt) {
			return false
		}
		for i := range at {
			if !deepEqualPy(at[i], bt[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		bt, ok := b.(map[string]any)
		if !ok || len(at) != len(bt) {
			return false
		}
		for k, v := range at {
			bv, present := bt[k]
			if !present || !deepEqualPy(v, bv) {
				return false
			}
		}
		return true
	}
	return reflect.DeepEqual(a, b)
}
