// Package pypickle reads and writes the JSON-shaped subset of Python's pickle
// format, so a Go process can share cache values with a Django process.
//
// django-redis serializes every non-int cache value with
// pickle.dumps(value, pickle.DEFAULT_PROTOCOL) and then optionally
// zlib-compresses it. To read or write those entries, Go has to speak the same
// format. This package implements the subset of pickle that corresponds to
// JSON's data model — None, bool, int, float, str, bytes, list, tuple, dict —
// which covers cache payloads and the stringified tag versions that
// fragmented-keys stores.
//
// Anything outside that subset (class instances, datetimes, Decimals, sets,
// object references) decodes to ErrUnsupportedOpcode. Callers are expected to
// treat that as a cache miss rather than an error worth surfacing: a Go
// service should not fail a request because Django cached something only
// Python can understand.
//
// # Reading
//
// Unmarshal accepts protocols 0 through 5, since a cache may hold entries
// written by older Python versions or by a differently configured
// PICKLE_VERSION.
//
// # Writing
//
// Marshal emits protocol 4, matching pickle.DEFAULT_PROTOCOL on CPython 3.8
// through 3.13 — which is what django-redis uses unless PICKLE_VERSION is set.
// Output is byte-identical to CPython for scalars, strings, bytes, and tuples.
// It is not byte-identical for dicts, because Python preserves insertion order
// while a Go map has none; Marshal sorts keys instead, trading exact byte
// equality for deterministic output. Django reads either form identically.
package pypickle

// Pickle opcodes. Names and values follow CPython's Lib/pickle.py.
const (
	opMark            = '('  // push special markobject on stack
	opStop            = '.'  // every pickle ends with STOP
	opPop             = '0'  // discard topmost stack item
	opPopMark         = '1'  // discard stack top through topmost markobject
	opDup             = '2'  // duplicate top stack item
	opFloat           = 'F'  // push float object; decimal string argument
	opInt             = 'I'  // push integer or bool; decimal string argument
	opBinInt          = 'J'  // push four-byte signed int
	opBinInt1         = 'K'  // push 1-byte unsigned int
	opLong            = 'L'  // push long; decimal string argument
	opBinInt2         = 'M'  // push 2-byte unsigned int
	opNone            = 'N'  // push None
	opPersID          = 'P'  // push persistent object; unsupported
	opBinPersID       = 'Q'  // same, but via stack; unsupported
	opReduce          = 'R'  // apply callable to argtuple; unsupported
	opString          = 'S'  // push string; NL-terminated string argument
	opBinString       = 'T'  // push string; counted binary string argument
	opShortBinString  = 'U'  // push string; counted binary string < 256 bytes
	opUnicode         = 'V'  // push Unicode string; raw-unicode-escaped
	opBinUnicode      = 'X'  // push Unicode string; counted UTF-8 string
	opAppend          = 'a'  // append stack top to list below it
	opBuild           = 'b'  // call __setstate__ or update __dict__; unsupported
	opGlobal          = 'c'  // push self.find_class(modname, name); unsupported
	opDict            = 'd'  // build a dict from stack items
	opEmptyDict       = '}'  // push empty dict
	opAppends         = 'e'  // extend list on stack by topmost stack slice
	opGet             = 'g'  // push item from memo on stack; index is string arg
	opBinGet          = 'h'  // push item from memo on stack; index is 1-byte arg
	opInst            = 'i'  // build & push class instance; unsupported
	opLongBinGet      = 'j'  // push item from memo on stack; index is 4-byte arg
	opList            = 'l'  // build list from topmost stack items
	opEmptyList       = ']'  // push empty list
	opObj             = 'o'  // build & push class instance; unsupported
	opPut             = 'p'  // store stack top in memo; index is string arg
	opBinPut          = 'q'  // store stack top in memo; index is 1-byte arg
	opLongBinPut      = 'r'  // store stack top in memo; index is 4-byte arg
	opSetItem         = 's'  // add key+value pair to dict
	opTuple           = 't'  // build tuple from topmost stack items
	opEmptyTuple      = ')'  // push empty tuple
	opSetItems        = 'u'  // modify dict by adding topmost key+value pairs
	opBinFloat        = 'G'  // push float; arg is 8-byte big-endian float
	opProto           = 0x80 // identify pickle protocol
	opNewObj          = 0x81 // build object by applying cls.__new__; unsupported
	opExt1            = 0x82 // push object from extension registry; unsupported
	opExt2            = 0x83 // ditto, 2-byte index; unsupported
	opExt4            = 0x84 // ditto, 4-byte index; unsupported
	opTuple1          = 0x85 // build 1-tuple from stack top
	opTuple2          = 0x86 // build 2-tuple from two topmost stack items
	opTuple3          = 0x87 // build 3-tuple from three topmost stack items
	opNewTrue         = 0x88 // push True
	opNewFalse        = 0x89 // push False
	opLong1           = 0x8a // push long from < 256 bytes
	opLong4           = 0x8b // push really big long
	opShortBinUnicode = 0x8c // push short string; UTF-8 length < 256 bytes
	opBinUnicode8     = 0x8d // push very long string
	opBinBytes8       = 0x8e // push very long bytes string
	opEmptySet        = 0x8f // push empty set on the stack; unsupported
	opFrozenSet       = 0x91 // build frozenset from topmost stack items; unsupported
	opAddItems        = 0x90 // modify set by adding topmost stack items; unsupported
	opNewObjEx        = 0x92 // like NEWOBJ but work with keyword only args; unsupported
	opStackGlobal     = 0x93 // same as GLOBAL but using names on the stacks; unsupported
	opMemoize         = 0x94 // store top of the stack in memo
	opFrame           = 0x95 // indicate the beginning of a new frame
	opBinBytes        = 'B'  // push bytes; counted binary string argument
	opShortBinBytes   = 'C'  // push bytes; counted binary string < 256 bytes
	opByteArray8      = 0x96 // push bytearray
	opNextBuffer      = 0x97 // push next out-of-band buffer; unsupported
	opReadonlyBuffer  = 0x98 // make top of stack readonly; unsupported
)

// DefaultProtocol is the protocol Marshal emits. It matches
// pickle.DEFAULT_PROTOCOL on CPython 3.8+, which is what django-redis uses
// when PICKLE_VERSION is left unset.
const DefaultProtocol = 4

// maxProtocol is the highest protocol Unmarshal will accept.
const maxProtocol = 5
