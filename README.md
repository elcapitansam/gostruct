# gostruct
Golang data (de)serialization package in the style of Python's struct module.  

## Functions

### Pack(fmt string, buf []byte, args ...interface{}) error

Pack serializes data into the provided []byte buffer according to the fmt string.  Errors occur when the buffer supplied is too small for the requested serialization, a request is made to embed a Pascal string larger than 255 bytes, and when unknown format types are requested.

### Unpack(fmt string, buf []byte) ([]interface{}, error)

Unpack deserializes data from the provided []byte buffer according to the fmt string, returning values via an []interface{} array.  The returned values must be manually type decoded for further use.

## Errors

Errors occur when the buffer is too small for the requested (de)serialization, and when a Pascal string larger than 255 or unknown format characters are attempted.

## Format Strings

Format strings are composed of format characters specifying the size and type of the datum being (de)serialized.  Additionally format characters control byte order and numerical counts may be used to repeat formats for concise format representation.

### Byte Order

Byte order defaults to native-endian format on each call to Pack/Unpack.  It can be changed with byte order format characters as follows.  The byte order persists in processing the format string; it is only necessary to set it once for subsequent format characters.  It is acceptable to change byte order as often as necessary in a single format string.

| Character | Byte Order |
| - | - |
| = | native |
| < | little-endian |
| > | big-endian |
| ! | network (= big-endian) |

### Format Characters

Format characters are as follows.  Size is the count in bytes as consumed by the (de)serialization buffer.

| Character | Go Type | Size | Notes |
| - | -:|:-:| - |
| x | no value | 1 | Zero pad byte, used for spacing/alignment.  No arg is supplied or returned. |
| b | int8 | 1 | |
| B | uint8 | 1 | |
| ? | bool | 1 | True is serialized as 1, False as 0. |
| h | int16 | 2 | |
| H | uint16 | 2 | |
| i, l | int32 | 4 | |
| I, L | uint32 | 4 | |
| q | int64 | 8 | |
| Q | uint64 | 8 | |
| f | float32 | 4 | |
| d | float64 | 8 | |
| s | string | | If size is not specified in count, it defaults to 1. |
| p | string | | Pascal string, max size 255. |

An optional count may be specified before each format character.  For non-string types this is a repeat count; e.g., a format of "6x" requests 6 zero pad bytes.  For string types the count specifies the length of the string space in the buffer.  On Pack(), strings which are too long will be truncated and strings which are too short will be zero byte padded.  As a convenience, when Unpack()'ing strings trailing zero bytes will be removed from the returned string.

When using Pascal strings if count is not specified the string will be variably (de)serialized according to string length, up to 255 bytes.  If a count is specified, it cannot be larger than 256.

### Ignored Characters

Whitespace in the format string is ignored for convenience.

## Not implemented

Specific features of the Python struct module are not implemented.  The c and P formats are not available.  The @ character to select native alignment, and the entirety of logic around native word alignment is not available.
