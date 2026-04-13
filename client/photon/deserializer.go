// Package photon implements a pure-Go port of the C# Protocol18Deserializer
// from https://github.com/JPCodeCraft/AlbionDataAvalonia.
package photon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

// Exported type code aliases for use in tests and external packages.
const (
	TypeNull    = typeNull
	TypeBoolean = typeBoolean
	TypeString  = typeString
	TypeIntZero = typeIntZero
)

// Protocol18 type codes (matches C# Protocol18Type enum)
const (
	typeUnknown          = byte(0)
	typeBoolean          = byte(2)
	typeByte             = byte(3)
	typeShort            = byte(4)
	typeFloat            = byte(5)
	typeDouble           = byte(6)
	typeString           = byte(7)
	typeNull             = byte(8)
	typeCompressedInt    = byte(9)
	typeCompressedLong   = byte(10)
	typeInt1             = byte(11) // 1-byte unsigned int
	typeInt1Neg          = byte(12) // 1-byte unsigned int, negated
	typeInt2             = byte(13) // 2-byte unsigned int
	typeInt2Neg          = byte(14) // 2-byte unsigned int, negated
	typeLong1            = byte(15) // 1-byte unsigned long
	typeLong1Neg         = byte(16) // 1-byte unsigned long, negated
	typeLong2            = byte(17) // 2-byte unsigned long
	typeLong2Neg         = byte(18) // 2-byte unsigned long, negated
	typeCustom           = byte(19)
	typeDictionary       = byte(20)
	typeHashtable        = byte(21)
	typeObjectArray      = byte(23)
	typeOperationRequest = byte(24)
	typeOperationResp    = byte(25)
	typeEventData        = byte(26)
	typeBoolFalse        = byte(27)
	typeBoolTrue         = byte(28)
	typeShortZero        = byte(29)
	typeIntZero          = byte(30)
	typeLongZero         = byte(31)
	typeFloatZero        = byte(32)
	typeDoubleZero       = byte(33)
	typeByteZero         = byte(34)
	typeArray            = byte(0x40) // bare array; 0x40|elemType = typed array
	customTypeSlimBase   = byte(0x80) // gpType >= 0x80: slim custom type, id in low 7 bits
)

// deserializeParameterTable parses a Protocol18 parameter table from raw bytes.
// Wire format: compressed-varint count | (key | typeCode | value)*
func deserializeParameterTable(data []byte) map[byte]interface{} {
	return readParameterTable(bytes.NewBuffer(data))
}

func readParameterTable(buf *bytes.Buffer) map[byte]interface{} {
	count := int(readCount(buf))
	params := make(map[byte]interface{}, count)
	for i := 0; i < count && buf.Len() > 0; i++ {
		key, err := buf.ReadByte()
		if err != nil {
			break
		}
		tc, err := buf.ReadByte()
		if err != nil {
			break
		}
		params[key] = deserialize(buf, tc)
	}
	return params
}

// deserialize decodes a single Protocol18 value given its type code.
func deserialize(buf *bytes.Buffer, tc byte) interface{} {
	if tc >= customTypeSlimBase {
		return deserializeCustom(buf, tc)
	}
	switch tc {
	case typeUnknown, typeNull:
		return nil
	case typeBoolean:
		b, _ := buf.ReadByte()
		return b != 0
	case typeByte:
		b, _ := buf.ReadByte()
		return b
	case typeShort:
		return readInt16(buf)
	case typeFloat:
		return readFloat32(buf)
	case typeDouble:
		return readFloat64(buf)
	case typeString:
		return readString(buf)
	case typeCompressedInt:
		return readCompressedInt32(buf)
	case typeCompressedLong:
		return readCompressedInt64(buf)
	case typeInt1:
		b, _ := buf.ReadByte()
		return int32(b)
	case typeInt1Neg:
		b, _ := buf.ReadByte()
		return -int32(b)
	case typeInt2:
		return int32(readUint16(buf))
	case typeInt2Neg:
		return -int32(readUint16(buf))
	case typeLong1:
		b, _ := buf.ReadByte()
		return int64(b)
	case typeLong1Neg:
		b, _ := buf.ReadByte()
		return -int64(b)
	case typeLong2:
		return int64(readUint16(buf))
	case typeLong2Neg:
		return -int64(readUint16(buf))
	case typeCustom:
		return deserializeCustom(buf, 0)
	case typeDictionary:
		return deserializeDictionary(buf)
	case typeHashtable:
		return deserializeHashtable(buf)
	case typeObjectArray:
		return deserializeObjectArray(buf)
	case typeOperationRequest:
		return deserializeOperationRequestInner(buf)
	case typeOperationResp:
		return deserializeOperationResponseInner(buf)
	case typeEventData:
		return deserializeEventDataInner(buf)
	case typeBoolFalse:
		return false
	case typeBoolTrue:
		return true
	case typeShortZero:
		return int16(0)
	case typeIntZero:
		return int32(0)
	case typeLongZero:
		return int64(0)
	case typeFloatZero:
		return float32(0)
	case typeDoubleZero:
		return float64(0)
	case typeByteZero:
		return byte(0)
	case typeArray:
		return deserializeNestedArray(buf)
	default:
		if tc&typeArray == typeArray {
			return deserializeTypedArray(buf, tc&^typeArray)
		}
		return fmt.Sprintf("ERROR - unknown type 0x%02X", tc)
	}
}

// deserializeTypedArray decodes a typed array where all elements share the same type.
// Wire: compressed-count | element*
func deserializeTypedArray(buf *bytes.Buffer, elemType byte) interface{} {
	size := int(readCount(buf))
	switch elemType {
	case typeBoolean:
		result := make([]bool, size)
		packedBytes := (size + 7) / 8
		packed := make([]byte, packedBytes)
		buf.Read(packed)
		for i := 0; i < size; i++ {
			result[i] = (packed[i/8] & (1 << uint(i%8))) != 0
		}
		return result
	case typeByte:
		data := make([]byte, size)
		buf.Read(data)
		return data
	case typeShort:
		result := make([]int16, size)
		for i := range result {
			result[i] = readInt16(buf)
		}
		return result
	case typeFloat:
		result := make([]float32, size)
		for i := range result {
			result[i] = readFloat32(buf)
		}
		return result
	case typeDouble:
		result := make([]float64, size)
		for i := range result {
			result[i] = readFloat64(buf)
		}
		return result
	case typeString:
		result := make([]string, size)
		for i := range result {
			result[i] = readString(buf)
		}
		return result
	case typeCustom:
		// Shared custom type id for all elements
		customTypeID, _ := buf.ReadByte()
		result := make([]interface{}, size)
		for i := range result {
			result[i] = deserializeCustomPayload(buf, customTypeID, false)
		}
		return result
	case typeDictionary:
		result := make([]interface{}, size)
		for i := range result {
			result[i] = deserializeDictionary(buf)
		}
		return result
	case typeHashtable:
		result := make([]interface{}, size)
		for i := range result {
			result[i] = deserializeHashtable(buf)
		}
		return result
	case typeCompressedInt:
		result := make([]int32, size)
		for i := range result {
			result[i] = readCompressedInt32(buf)
		}
		return result
	case typeCompressedLong:
		result := make([]int64, size)
		for i := range result {
			result[i] = readCompressedInt64(buf)
		}
		return result
	default:
		result := make([]interface{}, size)
		for i := range result {
			result[i] = deserialize(buf, elemType)
		}
		return result
	}
}

// deserializeNestedArray decodes a bare array (0x40) whose elements share one type.
// Wire: compressed-count | typeCode | element*
func deserializeNestedArray(buf *bytes.Buffer) interface{} {
	size := int(readCount(buf))
	tc, err := buf.ReadByte()
	if err != nil {
		return nil
	}
	result := make([]interface{}, size)
	for i := range result {
		result[i] = deserialize(buf, tc)
	}
	return result
}

// deserializeObjectArray decodes an array where each element carries its own type byte.
// Wire: compressed-count | (typeCode | value)*
func deserializeObjectArray(buf *bytes.Buffer) interface{} {
	size := int(readCount(buf))
	result := make([]interface{}, size)
	for i := range result {
		tc, err := buf.ReadByte()
		if err != nil {
			break
		}
		result[i] = deserialize(buf, tc)
	}
	return result
}

// deserializeDictionary decodes a Protocol18 Dictionary.
// Wire: keyTypeCode | valueTypeCode | compressed-count | (key | value)*
// typeCode == 0 means each entry carries its own type byte.
func deserializeDictionary(buf *bytes.Buffer) map[interface{}]interface{} {
	keyTC, _ := buf.ReadByte()
	valTC, _ := buf.ReadByte()
	count := int(readCount(buf))
	out := make(map[interface{}]interface{}, count)
	for i := 0; i < count && buf.Len() > 0; i++ {
		var kt byte
		if keyTC == 0 {
			kt, _ = buf.ReadByte()
		} else {
			kt = keyTC
		}
		var vt byte
		if valTC == 0 {
			vt, _ = buf.ReadByte()
		} else {
			vt = valTC
		}
		key := deserialize(buf, kt)
		val := deserialize(buf, vt)
		if isComparable(key) {
			out[key] = val
		} else {
			out[fmt.Sprintf("UNHASHABLE_%d_%T", i, key)] = val
		}
	}
	return out
}

// deserializeHashtable decodes a Protocol18 Hashtable (same wire format as Dictionary).
func deserializeHashtable(buf *bytes.Buffer) map[interface{}]interface{} {
	return deserializeDictionary(buf)
}

// deserializeCustom handles both regular (typeCustom=19) and slim (>=0x80) custom types.
func deserializeCustom(buf *bytes.Buffer, gpType byte) interface{} {
	var customID byte
	isSlim := gpType >= customTypeSlimBase
	if isSlim {
		customID = gpType & 0x7F
	} else {
		customID, _ = buf.ReadByte()
	}
	return deserializeCustomPayload(buf, customID, isSlim)
}

func deserializeCustomPayload(buf *bytes.Buffer, customID byte, isSlim bool) interface{} {
	size := int(readCount(buf))
	if size < 0 || size > buf.Len() {
		if isSlim {
			data := make([]byte, buf.Len())
			buf.Read(data)
			return map[string]interface{}{"type": customID, "data": data}
		}
		return nil
	}
	data := make([]byte, size)
	buf.Read(data)
	return map[string]interface{}{"type": customID, "data": data}
}

// Nested operation types (24, 25, 26) — return as generic maps.
func deserializeOperationRequestInner(buf *bytes.Buffer) interface{} {
	opCode, _ := buf.ReadByte()
	params := readParameterTable(buf)
	return map[string]interface{}{"operationCode": opCode, "parameters": params}
}

func deserializeOperationResponseInner(buf *bytes.Buffer) interface{} {
	if buf.Len() < 3 {
		return nil
	}
	opCode, _ := buf.ReadByte()
	returnCode := readInt16(buf)
	debugMsg := ""
	if buf.Len() > 0 {
		tc, _ := buf.ReadByte()
		if v, ok := deserialize(buf, tc).(string); ok {
			debugMsg = v
		}
	}
	params := readParameterTable(buf)
	return map[string]interface{}{
		"operationCode": opCode,
		"returnCode":    returnCode,
		"debugMessage":  debugMsg,
		"parameters":    params,
	}
}

func deserializeEventDataInner(buf *bytes.Buffer) interface{} {
	code, _ := buf.ReadByte()
	params := readParameterTable(buf)
	return map[string]interface{}{"code": code, "parameters": params}
}

// ── low-level readers ────────────────────────────────────────────────────────

func readInt16(buf *bytes.Buffer) int16 {
	var v int16
	binary.Read(buf, binary.LittleEndian, &v)
	return v
}

func readUint16(buf *bytes.Buffer) uint16 {
	var v uint16
	binary.Read(buf, binary.LittleEndian, &v)
	return v
}

func readFloat32(buf *bytes.Buffer) float32 {
	var bits uint32
	binary.Read(buf, binary.LittleEndian, &bits)
	return math.Float32frombits(bits)
}

func readFloat64(buf *bytes.Buffer) float64 {
	var bits uint64
	binary.Read(buf, binary.LittleEndian, &bits)
	return math.Float64frombits(bits)
}

// readString reads a Protocol18 string: compressed-varint length | UTF-8 bytes.
func readString(buf *bytes.Buffer) string {
	length := int(readCompressedUint32(buf))
	if length <= 0 || length > buf.Len() {
		return ""
	}
	b := make([]byte, length)
	buf.Read(b)
	return string(b)
}

// readCount reads the collection/array length prefix (compressed varint).
func readCount(buf *bytes.Buffer) uint32 {
	return readCompressedUint32(buf)
}

func readCompressedUint32(buf *bytes.Buffer) uint32 {
	var value uint32
	shift := uint(0)
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return 0
		}
		value |= uint32(b&0x7F) << shift
		if b&0x80 == 0 {
			return value
		}
		shift += 7
		if shift >= 35 {
			return 0
		}
	}
}

func readCompressedUint64(buf *bytes.Buffer) uint64 {
	var value uint64
	shift := uint(0)
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return 0
		}
		value |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return value
		}
		shift += 7
		if shift >= 70 {
			return 0
		}
	}
}

func readCompressedInt32(buf *bytes.Buffer) int32 {
	v := readCompressedUint32(buf)
	return int32((v >> 1) ^ uint32(-(int32(v & 1))))
}

func readCompressedInt64(buf *bytes.Buffer) int64 {
	v := readCompressedUint64(buf)
	return int64((v >> 1) ^ uint64(-(int64(v & 1))))
}

func isComparable(v interface{}) bool {
	if v == nil {
		return true
	}
	return reflect.TypeOf(v).Comparable()
}
