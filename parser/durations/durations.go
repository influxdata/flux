package durations

import (
	"fmt"
	"strconv"

	"github.com/influxdata/flux/ast"
)

var _cond_offsets []byte = []byte{
	0, 0, 0, 9, 9, 9, 9, 9,
	9, 10, 10, 11, 12, 12, 13, 14,
	15, 16, 16, 18, 20, 22, 23, 25,
	27, 29, 30, 32, 34, 36, 37, 39,
	41, 43, 44, 46, 48, 50, 51, 53,
	55, 57, 58, 58, 59, 60, 61, 62,
	62, 63, 64, 65, 66, 67, 67, 67,
	67, 67, 67, 67, 67, 67, 67, 67,
	67, 67, 67, 68, 69, 70, 70, 71,
	72, 73, 74, 76, 78, 80, 81, 83,
	85, 87, 87, 88, 90, 92, 94, 95,
	97, 99, 101, 101, 102, 104, 106, 108,
	109, 111, 113, 115, 115, 115, 115, 115,
	115,
}

var _cond_lengths []byte = []byte{
	0, 0, 9, 0, 0, 0, 0, 0,
	1, 0, 1, 1, 0, 1, 1, 1,
	1, 0, 2, 2, 2, 1, 2, 2,
	2, 1, 2, 2, 2, 1, 2, 2,
	2, 1, 2, 2, 2, 1, 2, 2,
	2, 1, 0, 1, 1, 1, 1, 0,
	1, 1, 1, 1, 1, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 1, 1, 1, 0, 1, 1,
	1, 1, 2, 2, 2, 1, 2, 2,
	2, 0, 1, 2, 2, 2, 1, 2,
	2, 2, 0, 1, 2, 2, 2, 1,
	2, 2, 2, 0, 0, 0, 0, 0,
	0,
}

var _cond_keys []uint16 = []uint16{
	100, 100, 104, 104, 109, 109, 110, 110,
	115, 115, 117, 117, 119, 119, 121, 121,
	206, 206, 111, 111, 111, 111, 111, 111,
	111, 111, 111, 111, 111, 111, 115, 115,
	111, 111, 115, 115, 111, 111, 115, 115,
	111, 111, 115, 115, 115, 115, 111, 111,
	115, 115, 111, 111, 115, 115, 111, 111,
	115, 115, 115, 115, 111, 111, 115, 115,
	111, 111, 115, 115, 111, 111, 115, 115,
	115, 115, 111, 111, 115, 115, 111, 111,
	115, 115, 111, 111, 115, 115, 115, 115,
	111, 111, 115, 115, 111, 111, 115, 115,
	111, 111, 115, 115, 115, 115, 111, 111,
	115, 115, 111, 111, 115, 115, 111, 111,
	115, 115, 115, 115, 115, 115, 115, 115,
	188, 188, 115, 115, 188, 188, 115, 115,
	188, 188, 115, 115, 115, 115, 111, 111,
	111, 111, 111, 111, 111, 111, 111, 111,
	111, 111, 115, 115, 111, 111, 115, 115,
	111, 111, 115, 115, 111, 111, 115, 115,
	115, 115, 111, 111, 115, 115, 111, 111,
	115, 115, 111, 111, 115, 115, 115, 115,
	111, 111, 115, 115, 111, 111, 115, 115,
	111, 111, 115, 115, 115, 115, 111, 111,
	115, 115, 111, 111, 115, 115, 111, 111,
	115, 115, 115, 115, 111, 111, 115, 115,
	111, 111, 115, 115, 111, 111, 115, 115,
	115, 115, 111, 111, 115, 115, 111, 111,
	115, 115, 111, 111, 115, 115,
}

var _cond_spaces []byte = []byte{
	24, 25, 33, 20, 26, 28, 23, 21,
	29, 2, 3, 22, 2, 3, 22, 14,
	2, 14, 3, 14, 22, 14, 14, 2,
	14, 3, 14, 22, 14, 15, 2, 15,
	3, 15, 22, 15, 15, 2, 15, 3,
	15, 22, 15, 27, 2, 27, 3, 27,
	22, 27, 27, 2, 27, 3, 27, 22,
	27, 16, 17, 28, 18, 18, 19, 19,
	29, 29, 20, 2, 3, 22, 2, 3,
	22, 14, 2, 14, 3, 14, 22, 14,
	14, 2, 14, 3, 14, 22, 14, 15,
	2, 15, 3, 15, 22, 15, 15, 2,
	15, 3, 15, 22, 15, 27, 2, 27,
	3, 27, 22, 27, 27, 2, 27, 3,
	27, 22, 27,
}

var _key_offsets []int16 = []int16{
	0, 0, 2, 89, 91, 93, 95, 97,
	99, 100, 102, 103, 106, 108, 111, 114,
	119, 120, 122, 124, 126, 130, 133, 137,
	141, 147, 148, 150, 152, 156, 159, 163,
	167, 173, 176, 180, 184, 190, 195, 201,
	207, 215, 216, 218, 219, 222, 223, 224,
	226, 227, 228, 231, 234, 235, 235, 237,
	237, 239, 239, 241, 241, 243, 243, 245,
	245, 247, 247, 248, 249, 252, 254, 257,
	260, 265, 266, 268, 270, 274, 277, 281,
	285, 291, 291, 292, 294, 296, 300, 303,
	307, 311, 317, 319, 322, 326, 330, 336,
	341, 347, 353, 361, 361, 363, 363, 365,
	365,
}

var _trans_keys []uint16 = []uint16{
	49, 57, 1657, 1913, 2169, 5751, 6007, 6263,
	7780, 8036, 8292, 9832, 10088, 10344, 19059, 19315,
	19571, 29293, 29549, 29805, 30061, 30317, 30573, 30829,
	31085, 31341, 31597, 31853, 32109, 32365, 32621, 32877,
	33133, 33389, 33645, 33901, 34157, 34413, 34669, 34925,
	35181, 35437, 35693, 35949, 36205, 36461, 36717, 36973,
	37229, 37485, 37741, 37997, 38253, 38509, 38765, 39021,
	39277, 39533, 39789, 40045, 40301, 40557, 40813, 41069,
	41325, 41581, 41837, 42093, 42349, 42605, 42861, 43117,
	43373, 43629, 43885, 44141, 44397, 44653, 44909, 45165,
	47733, 47989, 48245, 49870, 50126, 50382, 50798, 48,
	57, 49, 57, 49, 57, 49, 57, 49,
	57, 49, 57, 2671, 49, 57, 3183, 3695,
	3951, 4207, 49, 57, 2671, 49, 57, 3183,
	49, 57, 3695, 3951, 4207, 49, 57, 20083,
	49, 57, 2671, 20083, 3183, 20083, 3695, 3951,
	4207, 20083, 20083, 49, 57, 2671, 20083, 49,
	57, 3183, 20083, 49, 57, 3695, 3951, 4207,
	20083, 49, 57, 28787, 2671, 28787, 3183, 28787,
	3695, 3951, 4207, 28787, 28787, 49, 57, 2671,
	28787, 49, 57, 3183, 28787, 49, 57, 3695,
	3951, 4207, 28787, 49, 57, 45683, 45939, 46195,
	2671, 45683, 45939, 46195, 3183, 45683, 45939, 46195,
	3695, 3951, 4207, 45683, 45939, 46195, 45683, 45939,
	46195, 49, 57, 2671, 45683, 45939, 46195, 49,
	57, 3183, 45683, 45939, 46195, 49, 57, 3695,
	3951, 4207, 45683, 45939, 46195, 49, 57, 46707,
	49, 57, 47219, 47731, 47987, 48243, 48828, 48755,
	49, 57, 49340, 49267, 49852, 50108, 50364, 49779,
	50035, 50291, 50803, 49, 57, 49, 57, 49,
	57, 49, 57, 49, 57, 49, 57, 2671,
	3183, 3695, 3951, 4207, 49, 57, 2671, 49,
	57, 3183, 49, 57, 3695, 3951, 4207, 49,
	57, 20083, 2671, 20083, 3183, 20083, 3695, 3951,
	4207, 20083, 20083, 49, 57, 2671, 20083, 49,
	57, 3183, 20083, 49, 57, 3695, 3951, 4207,
	20083, 49, 57, 28787, 2671, 28787, 3183, 28787,
	3695, 3951, 4207, 28787, 28787, 49, 57, 2671,
	28787, 49, 57, 3183, 28787, 49, 57, 3695,
	3951, 4207, 28787, 49, 57, 49, 57, 45683,
	45939, 46195, 2671, 45683, 45939, 46195, 3183, 45683,
	45939, 46195, 3695, 3951, 4207, 45683, 45939, 46195,
	45683, 45939, 46195, 49, 57, 2671, 45683, 45939,
	46195, 49, 57, 3183, 45683, 45939, 46195, 49,
	57, 3695, 3951, 4207, 45683, 45939, 46195, 49,
	57, 49, 57, 49, 57, 32, 9, 13,
}

var _single_lengths []byte = []byte{
	0, 0, 85, 0, 0, 0, 0, 0,
	1, 0, 1, 3, 0, 1, 1, 3,
	1, 0, 2, 2, 4, 1, 2, 2,
	4, 1, 2, 2, 4, 1, 2, 2,
	4, 3, 4, 4, 6, 3, 4, 4,
	6, 1, 0, 1, 3, 1, 1, 0,
	1, 1, 3, 3, 1, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 1, 1, 3, 0, 1, 1,
	3, 1, 2, 2, 4, 1, 2, 2,
	4, 0, 1, 2, 2, 4, 1, 2,
	2, 4, 0, 3, 4, 4, 6, 3,
	4, 4, 6, 0, 0, 0, 0, 0,
	1,
}

var _range_lengths []byte = []byte{
	0, 1, 1, 1, 1, 1, 1, 1,
	0, 1, 0, 0, 1, 1, 1, 1,
	0, 1, 0, 0, 0, 1, 1, 1,
	1, 0, 0, 0, 0, 1, 1, 1,
	1, 0, 0, 0, 0, 1, 1, 1,
	1, 0, 1, 0, 0, 0, 0, 1,
	0, 0, 0, 0, 0, 0, 1, 0,
	1, 0, 1, 0, 1, 0, 1, 0,
	1, 0, 0, 0, 0, 1, 1, 1,
	1, 0, 0, 0, 0, 1, 1, 1,
	1, 0, 0, 0, 0, 0, 1, 1,
	1, 1, 1, 0, 0, 0, 0, 1,
	1, 1, 1, 0, 1, 0, 1, 0,
	1,
}

var _index_offsets []int16 = []int16{
	0, 0, 2, 89, 91, 93, 95, 97,
	99, 101, 103, 105, 109, 111, 114, 117,
	122, 124, 126, 129, 132, 137, 140, 144,
	148, 154, 156, 159, 162, 167, 170, 174,
	178, 184, 188, 193, 198, 205, 210, 216,
	222, 230, 232, 234, 236, 240, 242, 244,
	246, 248, 250, 254, 258, 260, 261, 263,
	264, 266, 267, 269, 270, 272, 273, 275,
	276, 278, 279, 281, 283, 287, 289, 292,
	295, 300, 302, 305, 308, 313, 316, 320,
	324, 330, 331, 333, 336, 339, 344, 347,
	351, 355, 361, 363, 367, 372, 377, 384,
	389, 395, 401, 409, 410, 412, 413, 415,
	416,
}

var _indicies []byte = []byte{
	1, 0, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16,
	17, 18, 19, 20, 21, 22, 23, 24,
	25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 37, 38, 39, 40,
	41, 42, 43, 44, 45, 46, 47, 48,
	49, 50, 51, 52, 53, 54, 55, 56,
	57, 58, 59, 60, 61, 62, 63, 64,
	65, 66, 67, 68, 69, 70, 71, 72,
	73, 74, 75, 76, 77, 78, 79, 80,
	81, 82, 83, 84, 85, 86, 87, 2,
	0, 88, 0, 89, 0, 90, 0, 91,
	0, 92, 0, 93, 0, 94, 0, 95,
	0, 93, 95, 96, 0, 97, 0, 93,
	97, 0, 95, 97, 0, 93, 95, 96,
	97, 0, 98, 0, 99, 0, 93, 98,
	0, 95, 98, 0, 93, 95, 96, 98,
	0, 98, 97, 0, 93, 98, 97, 0,
	95, 98, 97, 0, 93, 95, 96, 98,
	97, 0, 100, 0, 93, 100, 0, 95,
	100, 0, 93, 95, 96, 100, 0, 100,
	97, 0, 93, 100, 97, 0, 95, 100,
	97, 0, 93, 95, 96, 100, 97, 0,
	98, 100, 101, 0, 93, 98, 100, 101,
	0, 95, 98, 100, 101, 0, 93, 95,
	96, 98, 100, 101, 0, 98, 100, 101,
	97, 0, 93, 98, 100, 101, 97, 0,
	95, 98, 100, 101, 97, 0, 93, 95,
	96, 98, 100, 101, 97, 0, 102, 0,
	103, 0, 104, 0, 102, 104, 105, 0,
	106, 0, 107, 0, 108, 0, 109, 0,
	110, 0, 106, 109, 111, 0, 107, 110,
	112, 0, 113, 0, 0, 88, 0, 0,
	89, 0, 0, 90, 0, 0, 91, 0,
	0, 92, 0, 0, 94, 0, 0, 93,
	0, 95, 0, 93, 95, 96, 0, 97,
	0, 93, 97, 0, 95, 97, 0, 93,
	95, 96, 97, 0, 98, 0, 93, 98,
	0, 95, 98, 0, 93, 95, 96, 98,
	0, 98, 97, 0, 93, 98, 97, 0,
	95, 98, 97, 0, 93, 95, 96, 98,
	97, 0, 0, 100, 0, 93, 100, 0,
	95, 100, 0, 93, 95, 96, 100, 0,
	100, 97, 0, 93, 100, 97, 0, 95,
	100, 97, 0, 93, 95, 96, 100, 97,
	0, 99, 0, 98, 100, 101, 0, 93,
	98, 100, 101, 0, 95, 98, 100, 101,
	0, 93, 95, 96, 98, 100, 101, 0,
	98, 100, 101, 97, 0, 93, 98, 100,
	101, 97, 0, 95, 98, 100, 101, 97,
	0, 93, 95, 96, 98, 100, 101, 97,
	0, 0, 103, 0, 0, 108, 0, 0,
	115, 115, 114,
}

var _trans_targs []byte = []byte{
	0, 2, 2, 3, 53, 54, 4, 55,
	56, 5, 57, 58, 6, 59, 60, 7,
	61, 62, 8, 10, 11, 12, 13, 14,
	15, 65, 66, 67, 68, 69, 70, 71,
	72, 16, 18, 19, 20, 21, 22, 23,
	24, 73, 74, 75, 76, 77, 78, 79,
	80, 25, 26, 27, 28, 29, 30, 31,
	32, 82, 83, 84, 85, 86, 87, 88,
	89, 33, 34, 35, 36, 37, 38, 39,
	40, 91, 92, 93, 94, 95, 96, 97,
	98, 41, 43, 44, 45, 48, 50, 52,
	2, 2, 2, 2, 2, 9, 2, 63,
	64, 2, 17, 2, 81, 90, 42, 2,
	99, 100, 46, 47, 2, 49, 101, 51,
	102, 103, 104, 0,
}

var _trans_actions []byte = []byte{
	1, 2, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	3, 4, 5, 6, 7, 0, 8, 0,
	0, 9, 0, 10, 0, 0, 0, 11,
	0, 0, 0, 0, 12, 0, 0, 0,
	0, 0, 0, 0,
}

var _eof_actions []byte = []byte{
	0, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 13, 13, 14,
	14, 15, 15, 16, 16, 17, 17, 18,
	18, 19, 19, 19, 19, 19, 19, 19,
	19, 19, 19, 19, 19, 19, 19, 19,
	19, 20, 19, 19, 19, 19, 19, 19,
	19, 19, 20, 19, 19, 19, 19, 19,
	19, 19, 19, 21, 21, 22, 22, 23,
	0,
}

const start int = 1
const first_final int = 53

const en_main int = 1
const en_fail int = 104

type machine struct {
	data       []byte
	cs         int
	p, pe, eof int
	pb         int
	curline    int
	err        error

	root       *ast.Program
	expression ast.Expression
	durations  []ast.Duration

	durationrank int
}

func NewMachine() *machine {
	m := &machine{}

	return m
}

// Err returns the error that occurred on the last call to Parse.
//
// If the result is nil, then the line was parsed successfully.
func (m *machine) Err() error {
	return m.err
}

func (m *machine) text() []byte {
	return m.data[m.pb:m.p]
}

func (m *machine) initDurationsRank() {
	m.durationrank = 11
}

func getDuration(bytes []byte, unit string) ast.Duration {
	v1 := bytes[:len(bytes)-len(unit)]
	v2, _ := strconv.Atoi(string(v1))
	if unit == "μs" {
		unit = "us"
	}
	return ast.Duration{
		Magnitude: int64(v2),
		Unit:      unit,
	}
}

func (m *machine) Parse(input []byte) *ast.Program {
	m.data = input
	m.p = 0
	m.pb = 0
	m.pe = len(input)
	m.eof = len(input)
	m.err = nil
	m.root = nil
	m.initDurationsRank()

	{
		m.cs = start
	}

	{
		var _klen int
		var _keys int
		var _trans int
		var _widec uint16

		if (m.p) == (m.pe) {
			goto _test_eof
		}
		if m.cs == 0 {
			goto _out
		}
	_resume:
		_widec = uint16((m.data)[(m.p)])
		_klen = int(_cond_lengths[m.cs])
		_keys = int(_cond_offsets[m.cs] * 2)
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + (_klen << 1) - 2)
		COND_LOOP:
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + (((_upper - _lower) >> 1) & ^1)
				switch {
				case _widec < uint16(_cond_keys[_mid]):
					_upper = _mid - 2
				case _widec > uint16(_cond_keys[_mid+1]):
					_lower = _mid + 2
				default:
					switch _cond_spaces[int(_cond_offsets[m.cs])+((_mid-_keys)>>1)] {
					case 0:
						_widec = 256 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 10 {
							_widec += 256
						}
					case 1:
						_widec = 768 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 10 {
							_widec += 256
						}
					case 2:
						_widec = 2304 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 9 {
							_widec += 256
						}
					case 3:
						_widec = 2816 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 9 {
							_widec += 256
						}
					case 4:
						_widec = 4352 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 8 {
							_widec += 256
						}
					case 5:
						_widec = 4864 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 8 {
							_widec += 256
						}
					case 6:
						_widec = 6400 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 7 {
							_widec += 256
						}
					case 7:
						_widec = 6912 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 7 {
							_widec += 256
						}
					case 8:
						_widec = 8448 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 6 {
							_widec += 256
						}
					case 9:
						_widec = 8960 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 6 {
							_widec += 256
						}
					case 10:
						_widec = 10496 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 5 {
							_widec += 256
						}
					case 11:
						_widec = 13056 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 5 {
							_widec += 256
						}
					case 12:
						_widec = 17664 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 4 {
							_widec += 256
						}
					case 13:
						_widec = 18176 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 4 {
							_widec += 256
						}
					case 14:
						_widec = 19712 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 3 {
							_widec += 256
						}
					case 15:
						_widec = 28416 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 3 {
							_widec += 256
						}
					case 16:
						_widec = 46336 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 2 {
							_widec += 256
						}
					case 17:
						_widec = 46848 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 2 {
							_widec += 256
						}
					case 18:
						_widec = 48384 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 2 {
							_widec += 256
						}
					case 19:
						_widec = 48896 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 2 {
							_widec += 256
						}
					case 20:
						_widec = 50432 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 1 {
							_widec += 256
						}
					case 21:
						_widec = 1280 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 10 {
							_widec += 256
						}
						if m.durationrank > 10 {
							_widec += 512
						}
					case 22:
						_widec = 3328 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 9 {
							_widec += 256
						}
						if m.durationrank > 9 {
							_widec += 512
						}
					case 23:
						_widec = 5376 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 8 {
							_widec += 256
						}
						if m.durationrank > 8 {
							_widec += 512
						}
					case 24:
						_widec = 7424 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 7 {
							_widec += 256
						}
						if m.durationrank > 7 {
							_widec += 512
						}
					case 25:
						_widec = 9472 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 6 {
							_widec += 256
						}
						if m.durationrank > 6 {
							_widec += 512
						}
					case 26:
						_widec = 18688 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 4 {
							_widec += 256
						}
						if m.durationrank > 4 {
							_widec += 512
						}
					case 27:
						_widec = 45312 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 3 {
							_widec += 256
						}
						if m.durationrank > 3 {
							_widec += 512
						}
					case 28:
						_widec = 47360 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 2 {
							_widec += 256
						}
						if m.durationrank > 2 {
							_widec += 512
						}
					case 29:
						_widec = 49408 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 2 {
							_widec += 256
						}
						if m.durationrank > 2 {
							_widec += 512
						}
					case 30:
						_widec = 11008 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 9 {
							_widec += 256
						}
						if m.durationrank > 9 {
							_widec += 512
						}
						if m.durationrank > 5 {
							_widec += 1024
						}
					case 31:
						_widec = 13568 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 9 {
							_widec += 256
						}
						if m.durationrank > 9 {
							_widec += 512
						}
						if m.durationrank > 5 {
							_widec += 1024
						}
						if m.durationrank > 5 {
							_widec += 2048
						}
					case 32:
						_widec = 20224 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 9 {
							_widec += 256
						}
						if m.durationrank > 9 {
							_widec += 512
						}
						if m.durationrank > 5 {
							_widec += 1024
						}
						if m.durationrank > 5 {
							_widec += 2048
						}
						if m.durationrank > 3 {
							_widec += 4096
						}
					case 33:
						_widec = 28928 + (uint16((m.data)[(m.p)]) - 0)
						if m.durationrank > 9 {
							_widec += 256
						}
						if m.durationrank > 9 {
							_widec += 512
						}
						if m.durationrank > 5 {
							_widec += 1024
						}
						if m.durationrank > 5 {
							_widec += 2048
						}
						if m.durationrank > 3 {
							_widec += 4096
						}
						if m.durationrank > 3 {
							_widec += 8192
						}
					}
					break COND_LOOP
				}
			}
		}

		_keys = int(_key_offsets[m.cs])
		_trans = int(_index_offsets[m.cs])

		_klen = int(_single_lengths[m.cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + _klen - 1)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + ((_upper - _lower) >> 1)
				switch {
				case _widec < _trans_keys[_mid]:
					_upper = _mid - 1
				case _widec > _trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_range_lengths[m.cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + (_klen << 1) - 2)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + (((_upper - _lower) >> 1) & ^1)
				switch {
				case _widec < _trans_keys[_mid]:
					_upper = _mid - 2
				case _widec > _trans_keys[_mid+1]:
					_lower = _mid + 2
				default:
					_trans += int((_mid - int(_keys)) >> 1)
					goto _match
				}
			}
			_trans += _klen
		}

	_match:
		_trans = int(_indicies[_trans])
		m.cs = int(_trans_targs[_trans])

		if _trans_actions[_trans] == 0 {
			goto _again
		}

		switch _trans_actions[_trans] {
		case 2:

			m.pb = m.p

		case 1:

			m.err = fmt.Errorf("unable to match [col %d]", m.p)
			(m.p)--

			m.cs = 104
			goto _again

		case 3:

			m.durations = append(m.durations, getDuration(m.text(), "y"))
			m.durationrank = 10

			m.pb = m.p

		case 8:

			m.durations = append(m.durations, getDuration(m.text(), "mo"))
			m.durationrank = 9

			m.pb = m.p

		case 4:

			m.durations = append(m.durations, getDuration(m.text(), "w"))
			m.durationrank = 8

			m.pb = m.p

		case 5:

			m.durations = append(m.durations, getDuration(m.text(), "d"))
			m.durationrank = 7

			m.pb = m.p

		case 6:

			m.durations = append(m.durations, getDuration(m.text(), "h"))
			m.durationrank = 6

			m.pb = m.p

		case 9:

			m.durations = append(m.durations, getDuration(m.text(), "m"))
			m.durationrank = 5

			m.pb = m.p

		case 7:

			m.durations = append(m.durations, getDuration(m.text(), "s"))
			m.durationrank = 4

			m.pb = m.p

		case 10:

			m.durations = append(m.durations, getDuration(m.text(), "ms"))
			m.durationrank = 3

			m.pb = m.p

		case 11:

			m.durations = append(m.durations, getDuration(m.text(), "us"))
			m.durationrank = 2

			m.pb = m.p

		case 12:

			m.durations = append(m.durations, getDuration(m.text(), "μs"))
			m.durationrank = 2

			m.pb = m.p

		}

	_again:
		if m.cs == 0 {
			goto _out
		}
		if (m.p)++; (m.p) != (m.pe) {
			goto _resume
		}
	_test_eof:
		{
		}
		if (m.p) == (m.eof) {
			switch _eof_actions[m.cs] {
			case 1:

				m.err = fmt.Errorf("unable to match [col %d]", m.p)
				(m.p)--

				m.cs = 104
				goto _again

			case 13:

				m.durations = append(m.durations, getDuration(m.text(), "y"))
				m.durationrank = 10

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 18:

				m.durations = append(m.durations, getDuration(m.text(), "mo"))
				m.durationrank = 9

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 14:

				m.durations = append(m.durations, getDuration(m.text(), "w"))
				m.durationrank = 8

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 15:

				m.durations = append(m.durations, getDuration(m.text(), "d"))
				m.durationrank = 7

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 16:

				m.durations = append(m.durations, getDuration(m.text(), "h"))
				m.durationrank = 6

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 19:

				m.durations = append(m.durations, getDuration(m.text(), "m"))
				m.durationrank = 5

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 17:

				m.durations = append(m.durations, getDuration(m.text(), "s"))
				m.durationrank = 4

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 20:

				m.durations = append(m.durations, getDuration(m.text(), "ms"))
				m.durationrank = 3

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 21:

				m.durations = append(m.durations, getDuration(m.text(), "us"))
				m.durationrank = 2

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 22:

				m.durations = append(m.durations, getDuration(m.text(), "μs"))
				m.durationrank = 2

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			case 23:

				m.durations = append(m.durations, getDuration(m.text(), "ns"))
				m.durationrank = 1

				// re-init (e.g, reset) durations rank
				m.initDurationsRank()
				// populate expression
				m.expression = &ast.DurationLiteral{
					Values: m.durations,
				}
				// empty durations slice
				m.durations = nil

				m.root = &ast.Program{
					Body: append([]ast.Statement{}, &ast.ExpressionStatement{
						Expression: m.expression,
					}),
					// todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
				}

			}
		}

	_out:
		{
		}
	}

	if m.cs < first_final {
		return nil
	}

	return m.root
}
