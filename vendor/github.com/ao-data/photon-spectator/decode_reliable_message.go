// Package photon_spectator decodes Photon transport and reliable payloads.
// Reliable message parameter values follow ExitGames Protocol18 rules: multi-byte
// scalars and collection length prefixes use little-endian (see protocol18.go).
package photon_spectator

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

const (
	NilType               = 42
	DictionaryType        = 68
	StringSliceType       = 97
	Int8Type              = 98
	Custom                = 99
	DoubleType            = 100
	EventDateType         = 101
	Float32Type           = 102
	Hashtable             = 104
	Int32Type             = 105
	Int16Type             = 107
	Int64Type             = 108
	Int32SliceType        = 110
	BooleanType           = 111
	OperationResponseType = 112
	OperationRequestType  = 113
	StringType            = 115
	Int8SliceType         = 120
	SliceType             = 121
	ObjectSliceType       = 122
)

const (
	gpTypeBoolean         = 2
	gpTypeByte            = 3
	gpTypeShort           = 4
	gpTypeFloat           = 5
	gpTypeDouble          = 6
	gpTypeString          = 7
	gpTypeNull            = 8
	gpTypeCompressedInt   = 9
	gpTypeCompressedLong  = 10
	gpTypeInt1            = 11
	gpTypeIntMinus1       = 12
	gpTypeInt2            = 13
	gpTypeIntMinus2       = 14
	gpTypeLong1           = 15
	gpTypeLongMinus1      = 16
	gpTypeLong2           = 17
	gpTypeLongMinus2      = 18
	gpTypeCustom          = 19
	gpTypeDictionary      = 20
	gpTypeHashtable       = 21
	gpTypeObjectArray     = 23
	gpTypeBoolFalse       = 27
	gpTypeBoolTrue        = 28
	gpTypeShortZero       = 29
	gpTypeIntZero         = 30
	gpTypeLongZero        = 31
	gpTypeFloatZero       = 32
	gpTypeDoubleZero      = 33
	gpTypeByteZero        = 34
	gpTypeArray           = 64
	gpTypeBooleanArray    = 65
	gpTypeByteArray       = 66
	gpTypeShortArray      = 67
	gpTypeDoubleArray     = 68
	gpTypeFloatArray      = 69
	gpTypeStringArray     = 70
	gpTypeHashtableArray  = 71
	gpTypeDictionaryArray = 72
	gpTypeCustomTypeArray = 73
	gpTypeIntArray        = 74
	gpTypeLongArray       = 75
)

type ReliableMessageParameters map[uint8]interface{}

// Deprecated: Use ReliableMessageParameters instead.
type ReliableMessageParamaters = ReliableMessageParameters

// Converts the parameters of a reliable message into a hash suitable for use in
// hashmap.
func DecodeReliableMessage(msg ReliableMessage) ReliableMessageParameters {
	params, _ := decodeReliableMessageParams(msg)
	return params
}

// decodeParamValue decodes one parameter value. Keys 252 (event) and 253 (operation)
// often use a compact unsigned prefix instead of a full Photon type tag; using
// decodeType for those bytes (e.g. gpTypeByte=3) would consume the next stream byte
// and desynchronize all following parameters.
func decodeParamValue(buf *bytes.Buffer, paramID uint8, marker uint8) interface{} {
	if paramID == 252 || paramID == 253 {
		if n, ok := decodeCompactUint16(buf, marker); ok {
			return n
		}
	}
	return decodeType(buf, marker)
}

func decodeReliableMessageParams(msg ReliableMessage) (ReliableMessageParameters, int) {
	buf := bytes.NewBuffer(msg.Data)
	params := make(map[uint8]interface{})
	errorCount := 0

	maxParams := int(msg.ParameterCount)
	if maxParams < 0 {
		maxParams = 0
	}
	if maxParams > len(msg.Data) {
		maxParams = len(msg.Data)
	}

	for i := 0; i < maxParams && buf.Len() > 0; i++ {
		paramID, err := buf.ReadByte()
		if err != nil {
			break
		}
		paramType, err := buf.ReadByte()
		if err != nil {
			break
		}

		value := decodeParamValue(buf, paramID, paramType)
		if decodeErrorString(value) {
			errorCount++
		}
		params[paramID] = value
	}

	return params, errorCount
}

func decodeCompactUint16(buf *bytes.Buffer, firstByte uint8) (uint16, bool) {
	value := uint64(firstByte & 0x7F)
	if firstByte&0x80 == 0 {
		return uint16(value), true
	}
	shift := uint(7)
	for i := 0; i < 3; i++ {
		next, err := buf.ReadByte()
		if err != nil {
			return 0, false
		}
		value |= uint64(next&0x7F) << shift
		if next&0x80 == 0 {
			return uint16(value), true
		}
		shift += 7
	}
	return 0, false
}

func decodeErrorString(v interface{}) bool {
	s, ok := v.(string)
	return ok && len(s) >= 7 && s[:7] == "ERROR -"
}

func decodeType(buf *bytes.Buffer, paramType uint8) interface{} {
	switch paramType {
	case NilType, 0:
		// Do nothing
		return nil
	case gpTypeNull:
		return nil
	case Int8Type:
		return decodeInt8Type(buf)
	case gpTypeByte:
		return int16(decodeInt8Type(buf))
	case Float32Type:
		return decodeFloat32Type(buf)
	case gpTypeFloat:
		return decodeFloat32Type(buf)
	case DoubleType:
		return decodeFloat64Type(buf)
	case gpTypeDouble:
		return decodeFloat64Type(buf)
	case Int32Type:
		return decodeInt32Type(buf)
	case gpTypeCompressedInt:
		return decodeCompressedInt32Type(buf)
	case Int16Type:
		return decodeInt16Type(buf)
	case gpTypeShort:
		return decodeInt16Type(buf)
	case Int64Type:
		return decodeInt64Type(buf)
	case gpTypeCompressedLong:
		return decodeCompressedInt64Type(buf)
	case StringType:
		return decodeStringType(buf)
	case gpTypeString:
		return decodeStringType(buf)
	case BooleanType:
		result, err := decodeBooleanType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Boolean - %v", err.Error())
		}
		return result
	case gpTypeBoolean:
		result, err := decodeBooleanType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Boolean - %v", err.Error())
		}
		return result
	case gpTypeBoolFalse:
		return false
	case gpTypeBoolTrue:
		return true
	case gpTypeShortZero:
		return int16(0)
	case gpTypeIntZero:
		return int16(0)
	case gpTypeLongZero:
		return int64(0)
	case gpTypeFloatZero:
		return float32(0)
	case gpTypeDoubleZero:
		return float64(0)
	case gpTypeByteZero:
		return int16(0)
	case gpTypeInt1:
		b, _ := buf.ReadByte()
		return int16(int8(b))
	case gpTypeIntMinus1:
		return int16(-1)
	case gpTypeInt2:
		var v int16
		binary.Read(buf, binary.LittleEndian, &v)
		return v
	case gpTypeIntMinus2:
		return int16(-2)
	case gpTypeLong1:
		b, _ := buf.ReadByte()
		return int64(int8(b))
	case gpTypeLongMinus1:
		return int64(-1)
	case gpTypeLong2:
		var v int16
		binary.Read(buf, binary.LittleEndian, &v)
		return int64(v)
	case gpTypeLongMinus2:
		return int64(-2)
	case Int8SliceType:
		result, err := decodeSliceInt8Type(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Int8 - %v", err.Error())
		}
		return result
	case gpTypeByteArray:
		result, err := decodeSliceInt8Type(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Int8 - %v", err.Error())
		}
		return result
	case Int32SliceType:
		result, err := decodeSliceInt32Type(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Int32 - %v", err.Error())
		}
		return result
	case gpTypeIntArray:
		result, err := decodeCompressedInt32SliceType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Int32 - %v", err.Error())
		}
		return result
	case StringSliceType:
		result, err := decodeSliceStringType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice String - %v", err.Error())
		}
		return result
	case gpTypeStringArray:
		result, err := decodeSliceStringType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice String - %v", err.Error())
		}
		return result
	case gpTypeFloatArray:
		result, err := decodeSliceFloat32Type(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Float32 - %v", err.Error())
		}
		return result
	case gpTypeShortArray:
		result, err := decodeSliceInt16Type(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Int16 - %v", err.Error())
		}
		return result
	case gpTypeLongArray:
		result, err := decodeCompressedInt64SliceType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Int64 - %v", err.Error())
		}
		return result
	case gpTypeBooleanArray: // 65
		result, err := decodeSliceBoolType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Bool - %v", err.Error())
		}
		return result
	case gpTypeHashtableArray: // 71
		result, err := decodeSliceHashtableType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Hashtable - %v", err.Error())
		}
		return result
	case gpTypeDictionaryArray: // 72
		result, err := decodeSliceDictionaryType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Dictionary - %v", err.Error())
		}
		return result
	case gpTypeCustomTypeArray: // 73
		result, err := decodeSliceCustomType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice Custom - %v", err.Error())
		}
		return result
	case SliceType:
		array, err := decodeSlice(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice - %v", err.Error())
		}
		return array
	case gpTypeArray:
		array, err := decodeSlice(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Slice - %v", err.Error())
		}
		return array
	case ObjectSliceType:
		array, err := decodeObjectSlice(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Object Slice - %v", err.Error())
		}
		return array
	case gpTypeObjectArray:
		array, err := decodeObjectSlice(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Object Slice - %v", err.Error())
		}
		return array
	case DictionaryType:
		dict, err := decodeDictionaryType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Dictionary - %v", err.Error())
		}
		return dict
	case gpTypeDictionary:
		dict, err := decodeDictionaryType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Dictionary - %v", err.Error())
		}
		return dict
	case Hashtable:
		ht, err := decodeHashtableType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Hashtable - %v", err.Error())
		}
		return ht
	case gpTypeHashtable:
		ht, err := decodeHashtableType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Hashtable - %v", err.Error())
		}
		return ht
	case Custom:
		result, err := decodeCustomType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Custom - %v", err.Error())
		}
		return result
	case gpTypeCustom:
		result, err := decodeCustomType(buf)
		if err != nil {
			return fmt.Sprintf("ERROR - Custom - %v", err.Error())
		}
		return result
	}

	// Protocol18 typed-array shorthand: 0x40 + elementGpType.
	// This appears in recent Albion packets with markers like 0x4F, 0x52, etc.
	if paramType >= 0x40 && paramType < 0x80 {
		elementType := paramType - 0x40
		result, err := decodeTypedArrayByElementType(buf, elementType)
		if err == nil {
			return result
		}
		return fmt.Sprintf("ERROR - Typed Array (elemType=%d) - %v", elementType, err.Error())
	}

	// Protocol18 custom type slim: GpType >= 0x80 embeds custom type id in low 7 bits.
	if paramType >= CustomTypeSlimBase {
		result, err := decodeCustomTypeSlim(buf, paramType)
		if err != nil {
			return fmt.Sprintf("ERROR - Custom Slim - %v", err.Error())
		}
		return result
	}
	if paramType < 0x40 {
		return int16(paramType)
	}

	return fmt.Sprintf("ERROR - Invalid type of %v (remaining=%d bytes)", paramType, buf.Len())
}

func decodeSlice(buf *bytes.Buffer) (interface{}, error) {
	var length uint16
	var sliceType uint8

	binary.Read(buf, binary.LittleEndian, &length)
	binary.Read(buf, binary.LittleEndian, &sliceType)

	switch sliceType {
	case Float32Type, gpTypeFloat: // 102 or 5
		array := make([]float32, length)
		for j := range array {
			array[j] = decodeFloat32Type(buf)
		}
		return array, nil
	case Int32Type: // 105 — raw int32 (legacy)
		array := make([]int32, length)
		for j := range array {
			array[j] = decodeInt32Type(buf)
		}
		return array, nil
	case gpTypeCompressedInt: // 9 — varint int32 (Protocol18)
		array := make([]int32, length)
		for j := range array {
			array[j] = decodeCompressedInt32Type(buf)
		}
		return array, nil
	case Int16Type, gpTypeShort: // 107 or 4
		array := make([]int16, length)
		for j := range array {
			binary.Read(buf, binary.LittleEndian, &array[j])
		}
		return array, nil
	case Int64Type: // 108 — raw int64 (legacy)
		array := make([]int64, length)
		for j := range array {
			array[j] = decodeInt64Type(buf)
		}
		return array, nil
	case gpTypeCompressedLong: // 10 — varint int64 (Protocol18)
		array := make([]int64, length)
		for j := range array {
			array[j] = decodeCompressedInt64Type(buf)
		}
		return array, nil
	case StringType, gpTypeString: // 115 or 7
		array := make([]string, length)
		for j := range array {
			array[j] = decodeStringType(buf)
		}
		return array, nil
	case BooleanType, gpTypeBoolean: // 111 or 2
		array := make([]bool, length)
		for j := range array {
			result, err := decodeBooleanType(buf)
			if err != nil {
				return array, err
			}
			array[j] = result
		}
		return array, nil
	case Int8SliceType, gpTypeByteArray: // 120 or 66
		array := make([][]int8, length)
		for j := range array {
			result, err := decodeSliceInt8Type(buf)
			if err != nil {
				return nil, err
			}
			array[j] = result
		}
		return array, nil
	case SliceType, gpTypeArray: // 121 or 64
		array := make([]interface{}, length)
		for j := range array {
			subArray, err := decodeSlice(buf)
			if err != nil {
				return nil, err
			}
			array[j] = subArray
		}
		return array, nil
	default:
		// Unknown/Protocol18 element type: decode each element generically.
		array := make([]interface{}, length)
		for j := range array {
			array[j] = decodeType(buf, sliceType)
		}
		return array, nil
	}
}

func decodeInt8Type(buf *bytes.Buffer) (temp int8) {
	binary.Read(buf, binary.LittleEndian, &temp)
	return
}

func decodeFloat32Type(buf *bytes.Buffer) (temp float32) {
	binary.Read(buf, binary.LittleEndian, &temp)
	return
}

func decodeInt16Type(buf *bytes.Buffer) (temp int16) {
	binary.Read(buf, binary.LittleEndian, &temp)
	return
}

func decodeInt32Type(buf *bytes.Buffer) (temp int32) {
	binary.Read(buf, binary.LittleEndian, &temp)
	return
}

func decodeInt64Type(buf *bytes.Buffer) (temp int64) {
	binary.Read(buf, binary.LittleEndian, &temp)
	return
}

func decodeStringType(buf *bytes.Buffer) string {
	var length uint8

	binary.Read(buf, binary.LittleEndian, &length)

	strBytes := make([]byte, length)
	buf.Read(strBytes)

	return string(strBytes[:])
}

func decodeBooleanType(buf *bytes.Buffer) (bool, error) {
	var value uint8

	binary.Read(buf, binary.LittleEndian, &value)

	if value == 0 {
		return false, nil
	} else if value == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("Invalid value for boolean of %d", value)
	}

}

func decodeSliceInt8Type(buf *bytes.Buffer) ([]int8, error) {
	var length uint32

	err := binary.Read(buf, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}

	array := make([]int8, length)

	for j := 0; j < int(length); j++ {
		var temp int8
		err := binary.Read(buf, binary.LittleEndian, &temp)
		if err != nil {
			return nil, err
		}
		array[j] = temp
	}

	return array, nil
}

func decodeDictionaryType(buf *bytes.Buffer) (map[interface{}]interface{}, error) {
	var keyTypeCode uint8
	var valueTypeCode uint8
	var dictionarySize uint16

	err := binary.Read(buf, binary.LittleEndian, &keyTypeCode)
	if err != nil {
		return nil, err
	}
	err = binary.Read(buf, binary.LittleEndian, &valueTypeCode)
	if err != nil {
		return nil, err
	}
	err = binary.Read(buf, binary.LittleEndian, &dictionarySize)
	if err != nil {
		return nil, err
	}

	// TODO: The map[interface{}]interface{} may not actually work in real use-cases
	dictionary := make(map[interface{}]interface{})
	for i := uint16(0); i < dictionarySize; i++ {
		// TODO: We may need to read another byte for either key or value if they equal 0 or 42 in order to determine actual type
		key := decodeType(buf, keyTypeCode)
		value := decodeType(buf, valueTypeCode)
		if isMapKeyComparable(key) {
			dictionary[key] = value
			continue
		}
		// Avoid panics when Protocol18 sends collection/object keys.
		dictionary[fmt.Sprintf("UNHASHABLE_KEY_%d_%T", i, key)] = value
	}

	return dictionary, nil
}

func decodeFloat64Type(buf *bytes.Buffer) (temp float64) {
	binary.Read(buf, binary.LittleEndian, &temp)
	return
}

func decodeSliceInt32Type(buf *bytes.Buffer) ([]int32, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]int32, length)
	for j := range array {
		if err := binary.Read(buf, binary.LittleEndian, &array[j]); err != nil {
			return nil, err
		}
	}
	return array, nil
}

func decodeSliceInt16Type(buf *bytes.Buffer) ([]int16, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]int16, length)
	for j := range array {
		if err := binary.Read(buf, binary.LittleEndian, &array[j]); err != nil {
			return nil, err
		}
	}
	return array, nil
}

func decodeSliceFloat32Type(buf *bytes.Buffer) ([]float32, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]float32, length)
	for j := range array {
		if err := binary.Read(buf, binary.LittleEndian, &array[j]); err != nil {
			return nil, err
		}
	}
	return array, nil
}

func decodeSliceStringType(buf *bytes.Buffer) ([]string, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]string, length)
	for j := range array {
		array[j] = decodeStringType(buf)
	}
	return array, nil
}

func decodeCompressedInt32Type(buf *bytes.Buffer) int32 {
	value, err := decodeUVarint(buf, 5)
	if err != nil {
		return 0
	}
	return int32(decodeZigZag(value))
}

func decodeCompressedInt64Type(buf *bytes.Buffer) int64 {
	value, err := decodeUVarint(buf, 10)
	if err != nil {
		return 0
	}
	return decodeZigZag(value)
}

func decodeCompressedInt32SliceType(buf *bytes.Buffer) ([]int32, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]int32, length)
	for i := range array {
		array[i] = decodeCompressedInt32Type(buf)
	}
	return array, nil
}

func decodeCompressedInt64SliceType(buf *bytes.Buffer) ([]int64, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]int64, length)
	for i := range array {
		array[i] = decodeCompressedInt64Type(buf)
	}
	return array, nil
}

func decodeUVarint(buf *bytes.Buffer, maxBytes int) (uint64, error) {
	var value uint64
	var shift uint
	for i := 0; i < maxBytes; i++ {
		if buf.Len() == 0 {
			return 0, fmt.Errorf("unexpected EOF while reading varint")
		}
		b, err := buf.ReadByte()
		if err != nil {
			return 0, err
		}
		value |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return value, nil
		}
		shift += 7
	}
	return 0, fmt.Errorf("varint too large")
}

func decodeZigZag(v uint64) int64 {
	return int64((v >> 1) ^ uint64(-(int64(v & 1))))
}

func decodeTypedArrayByElementType(buf *bytes.Buffer, elementType uint8) ([]interface{}, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]interface{}, length)
	for i := range array {
		array[i] = decodeType(buf, elementType)
	}
	return array, nil
}

// decodeObjectSlice decodes an array where each element carries its own type byte.
func decodeObjectSlice(buf *bytes.Buffer) ([]interface{}, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]interface{}, length)
	for j := range array {
		var elemType uint8
		if err := binary.Read(buf, binary.LittleEndian, &elemType); err != nil {
			return nil, err
		}
		array[j] = decodeType(buf, elemType)
	}
	return array, nil
}

// decodeHashtableType decodes a Photon Hashtable where each key and value carry their own type byte.
func decodeHashtableType(buf *bytes.Buffer) (map[interface{}]interface{}, error) {
	var size uint16
	if err := binary.Read(buf, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	ht := make(map[interface{}]interface{}, size)
	for i := uint16(0); i < size; i++ {
		var keyType uint8
		if err := binary.Read(buf, binary.LittleEndian, &keyType); err != nil {
			return nil, err
		}
		key := decodeType(buf, keyType)

		var valType uint8
		if err := binary.Read(buf, binary.LittleEndian, &valType); err != nil {
			return nil, err
		}
		val := decodeType(buf, valType)
		if isMapKeyComparable(key) {
			ht[key] = val
		} else {
			ht[fmt.Sprintf("UNHASHABLE_KEY_%d_%T", i, key)] = val
		}
	}
	return ht, nil
}

func isMapKeyComparable(v interface{}) bool {
	if v == nil {
		return true
	}
	t := reflect.TypeOf(v)
	return t.Comparable()
}

// decodeCustomType decodes a Photon Custom type: customTypeID (1 byte) + length (2 bytes) + raw bytes.
func decodeCustomType(buf *bytes.Buffer) ([]byte, error) {
	var customTypeID uint8
	if err := binary.Read(buf, binary.LittleEndian, &customTypeID); err != nil {
		return nil, err
	}
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	data := make([]byte, length)
	if _, err := buf.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}

func decodeSliceBoolType(buf *bytes.Buffer) ([]bool, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]bool, length)
	for j := range array {
		result, err := decodeBooleanType(buf)
		if err != nil {
			return nil, err
		}
		array[j] = result
	}
	return array, nil
}

func decodeSliceHashtableType(buf *bytes.Buffer) ([]map[interface{}]interface{}, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]map[interface{}]interface{}, length)
	for j := range array {
		ht, err := decodeHashtableType(buf)
		if err != nil {
			return nil, err
		}
		array[j] = ht
	}
	return array, nil
}

func decodeSliceDictionaryType(buf *bytes.Buffer) ([]map[interface{}]interface{}, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([]map[interface{}]interface{}, length)
	for j := range array {
		dict, err := decodeDictionaryType(buf)
		if err != nil {
			return nil, err
		}
		array[j] = dict
	}
	return array, nil
}

func decodeSliceCustomType(buf *bytes.Buffer) ([][]byte, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	array := make([][]byte, length)
	for j := range array {
		data, err := decodeCustomType(buf)
		if err != nil {
			return nil, err
		}
		array[j] = data
	}
	return array, nil
}

// decodeCustomTypeSlim decodes Protocol18 custom type slim encoding where the type id is in the gpType itself.
// Albion Online registers custom slim types as fixed 21-byte blobs (no length prefix).
func decodeCustomTypeSlim(buf *bytes.Buffer, gpType uint8) (map[string]interface{}, error) {
	customTypeID := gpType & 0x7F
	const customTypeSlimSize = 21
	data := make([]byte, customTypeSlimSize)
	n, err := buf.Read(data)
	if err != nil && n == 0 {
		return nil, err
	}
	return map[string]interface{}{
		"type": customTypeID,
		"data": data[:n],
	}, nil
}
