package client

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

const (
	debugFormatMaxDepth   = 5
	debugFormatMinZeroRun = 10
	debugFormatMaxString  = 240
	debugFormatMaxHexLen  = 128
)

var debugSpacesZeros = regexp.MustCompile(fmt.Sprintf(`(?: 0){%d,}`, debugFormatMinZeroRun))

// formatDebugPhotonParams formats decoded Photon parameters for terminal logging:
// long runs of numeric zeros and long runs of 0x00 in hex are collapsed.
func formatDebugPhotonParams(m map[uint8]interface{}) string {
	if m == nil {
		return "nil"
	}
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	var b strings.Builder
	b.Grow(len(keys) * 16)
	for i, ik := range keys {
		if i > 0 {
			b.WriteString(" ")
		}
		k := uint8(ik)
		fmt.Fprintf(&b, "%d:%s", k, formatDebugValue(m[k], 0))
	}
	return b.String()
}

func formatDebugValue(v interface{}, depth int) string {
	if depth > debugFormatMaxDepth {
		return "…"
	}
	if v == nil {
		return "nil"
	}

	switch t := v.(type) {
	case string:
		return formatDebugString(t)
	case []byte:
		return formatDebugByteSlice(t)
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		return fmt.Sprint(t)
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return formatDebugSlice(rv, depth)
	case reflect.Map:
		return formatDebugMap(rv, depth)
	case reflect.Ptr:
		if rv.IsNil() {
			return "nil"
		}
		return formatDebugValue(rv.Elem().Interface(), depth+1)
	default:
		s := fmt.Sprint(v)
		return collapseSpacesZeroRuns(s)
	}
}

func formatDebugString(s string) string {
	if len(s) <= debugFormatMaxString {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%q…(len=%d)", s[:debugFormatMaxString], len(s))
}

func formatDebugByteSlice(p []byte) string {
	if len(p) == 0 {
		return "[]byte{}"
	}
	h := hexCompactZeros(p)
	if len(h) > debugFormatMaxHexLen {
		h = h[:debugFormatMaxHexLen] + fmt.Sprintf("…(hexlen>%d, rawlen=%d)", debugFormatMaxHexLen, len(p))
	}
	return fmt.Sprintf("[]byte(len=%d) %s", len(p), h)
}

func hexCompactZeros(p []byte) string {
	var b strings.Builder
	b.Grow(len(p) * 2)
	z := 0
	flush := func() {
		if z == 0 {
			return
		}
		if z >= 4 {
			fmt.Fprintf(&b, "00×%d", z)
		} else {
			for range z {
				b.WriteString("00")
			}
		}
		z = 0
	}
	for _, by := range p {
		if by == 0 {
			z++
			continue
		}
		flush()
		fmt.Fprintf(&b, "%02x", by)
	}
	flush()
	return b.String()
}

func formatDebugSlice(rv reflect.Value, depth int) string {
	n := rv.Len()
	if n == 0 {
		return "[]"
	}
	if rv.Type().Elem().Kind() == reflect.Uint8 {
		return formatDebugByteSlice(rv.Bytes())
	}

	var b strings.Builder
	b.WriteByte('[')
	zeroRun := 0
	wrote := false
	flushZeros := func() {
		if zeroRun == 0 {
			return
		}
		if zeroRun < debugFormatMinZeroRun {
			for j := 0; j < zeroRun; j++ {
				if wrote {
					b.WriteByte(' ')
				}
				b.WriteByte('0')
				wrote = true
			}
		} else {
			if wrote {
				b.WriteByte(' ')
			}
			fmt.Fprintf(&b, "…(%d×0)", zeroRun)
			wrote = true
		}
		zeroRun = 0
	}

	for i := 0; i < n; i++ {
		el := rv.Index(i)
		if el.Kind() == reflect.Interface && !el.IsNil() {
			el = el.Elem()
		}
		if isNumericZero(el) {
			zeroRun++
			continue
		}
		flushZeros()
		if wrote {
			b.WriteByte(' ')
		}
		b.WriteString(formatDebugValue(el.Interface(), depth+1))
		wrote = true
	}
	flushZeros()
	b.WriteByte(']')
	return b.String()
}

func isNumericZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	default:
		return false
	}
}

func formatDebugMap(rv reflect.Value, depth int) string {
	if rv.Len() == 0 {
		return "{}"
	}
	keys := rv.MapKeys()
	var b strings.Builder
	b.WriteByte('{')
	for i, mk := range keys {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(formatDebugValue(mk.Interface(), depth+1))
		b.WriteByte(':')
		mv := rv.MapIndex(mk)
		if !mv.IsValid() {
			b.WriteString("<nil>")
		} else {
			b.WriteString(formatDebugValue(mv.Interface(), depth+1))
		}
	}
	b.WriteByte('}')
	return b.String()
}

func collapseSpacesZeroRuns(s string) string {
	return debugSpacesZeros.ReplaceAllStringFunc(s, func(m string) string {
		n := strings.Count(strings.TrimSpace(m), "0")
		if n < debugFormatMinZeroRun {
			return m
		}
		return fmt.Sprintf(" …(%d×0) ", n)
	})
}
