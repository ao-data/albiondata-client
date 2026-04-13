// Photon Protocol 18 (binary serialization used by ExitGames.Client.Photon.Protocol18)
//
// References (as of integration):
//   - Serialization overview: https://doc.photonengine.com/server/v4/reference/serialization-in-photon
//   - .NET API: https://doc-api.photonengine.com/en/dotnet/current/class_exit_games_1_1_client_1_1_photon_1_1_protocol18.html
//
// Wire rules used by this package for the *parameter blob* (ReliableMessage.Data):
//   - Multi-byte numeric fields and collection length prefixes follow little-endian order
//     (same as System.BitConverter / StreamBuffer on little-endian platforms).
//   - Primitives still carry an explicit type tag (GpType) before the payload, except where
//     the enclosing format omits it (e.g. hashtable/dictionary entries still carry per-key
//     and per-value type bytes).
//   - GpType values 0x80+ denote “custom type slim” (type id embedded in the high byte).
//
// The UDP/TCP Photon *transport headers* (peer id, command header, etc.) remain big-endian
// in photon_layer.go / photon_command.go — that layer is not Protocol18 payload.

package photon_spectator

// GpType matches ExitGames.Client.Photon.Protocol18.GpType (subset used in decoding).
// Legacy “TypeCode” values (e.g. StringType=115) are still accepted in decodeType for
// compatibility with mixed stacks.
const (
	GpTypeUnknown         uint8 = 0
	GpTypeBoolean         uint8 = 2
	GpTypeByte            uint8 = 3
	GpTypeShort           uint8 = 4
	GpTypeFloat           uint8 = 5
	GpTypeDouble          uint8 = 6
	GpTypeString          uint8 = 7
	GpTypeNull            uint8 = 8
	GpTypeCompressedInt   uint8 = 9
	GpTypeCompressedLong  uint8 = 10
	GpTypeInt1            uint8 = 11
	GpTypeInt1_           uint8 = 12
	GpTypeInt2            uint8 = 13
	GpTypeInt2_           uint8 = 14
	GpTypeL1              uint8 = 15
	GpTypeL1_             uint8 = 16
	GpTypeL2              uint8 = 17
	GpTypeL2_             uint8 = 18
	GpTypeCustom          uint8 = 19
	GpTypeDictionary      uint8 = 20
	GpTypeHashtable       uint8 = 21
	GpTypeObjectArray     uint8 = 23
	GpTypeBoolFalse       uint8 = 27
	GpTypeBoolTrue        uint8 = 28
	GpTypeShortZero       uint8 = 29
	GpTypeIntZero         uint8 = 30
	GpTypeLongZero        uint8 = 31
	GpTypeFloatZero       uint8 = 32
	GpTypeDoubleZero      uint8 = 33
	GpTypeByteZero        uint8 = 34
	GpTypeArray           uint8 = 64
	GpTypeBooleanArray    uint8 = 65
	GpTypeByteArray       uint8 = 66
	GpTypeShortArray      uint8 = 67
	GpTypeDoubleArray     uint8 = 68
	GpTypeFloatArray      uint8 = 69
	GpTypeStringArray     uint8 = 70
	GpTypeHashtableArray  uint8 = 71
	GpTypeDictionaryArray uint8 = 72
	GpTypeCustomTypeArray uint8 = 73
	GpTypeIntArray        uint8 = 74
	GpTypeLongArray       uint8 = 75
)

// CustomTypeSlimBase is the lowest GpType value for “custom slim” (0x80 | customCode).
const CustomTypeSlimBase uint8 = 0x80
