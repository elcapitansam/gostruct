// Package gostruct provides a verb-driven Pack/Unpack interface for (de)serializing
// data with a []byte buffer.  It is modeled after Python's struct module.
package gostruct

import (
	"bytes"
	"errors"
	"unsafe"
)

var nativeIsLE bool

func init() {
	i := uint32(1)
	nativeIsLE = *(*byte)(unsafe.Pointer(&i)) == 1
}

func putBytes(buf []byte, val uint64, isLE bool, nbytes int) {
	for i := 0; i < nbytes; i++ {
		if isLE {
			buf[i] = byte(val)
		} else {
			buf[nbytes-i-1] = byte(val)
		}
		val >>= 8
	}
}

func getBytes(buf []byte, isLE bool, nbytes int) uint64 {
	v := uint64(0)
	for i := 0; i < nbytes; i++ {
		v <<= 8
		if isLE {
			v |= uint64(buf[nbytes-i-1])
		} else {
			v |= uint64(buf[i])
		}
	}
	return v
}

func atoi(s string) (int, int) {
	n := len(s)
	rv := 0
	var i int
	for i = 0; i < n; i++ {
		if s[i] < '0' || s[i] > '9' {
			break
		}
		rv *= 10
		rv += int(s[i] - '0')
	}
	return rv, i
}

func packarg(buf []byte, vfmt byte, arg interface{}, encLE bool) (int, error) {
	rv := 0
	switch vfmt {
	case 'b':
		buf[0] = byte(arg.(int8))
		rv = 1
	case 'B':
		buf[0] = byte(arg.(byte))
		rv = 1
	case '?':
		if x := arg.(bool); x == true {
			buf[0] = byte(1)
		} else {
			buf[0] = byte(0)
		}
		rv = 1
	case 'h':
		putBytes(buf, uint64(arg.(int16)), encLE, 2)
		rv = 2
	case 'H':
		putBytes(buf, uint64(arg.(uint16)), encLE, 2)
		rv = 2
	case 'i', 'l':
		putBytes(buf, uint64(arg.(int32)), encLE, 4)
		rv = 4
	case 'I', 'L':
		putBytes(buf, uint64(arg.(uint32)), encLE, 4)
		rv = 4
	case 'q':
		putBytes(buf, uint64(arg.(int64)), encLE, 8)
		rv = 8
	case 'Q':
		putBytes(buf, uint64(arg.(uint64)), encLE, 8)
		rv = 8
	case 'f':
		f := arg.(float32)
		xf := *(*uint32)(unsafe.Pointer(&f))
		putBytes(buf, uint64(xf), encLE, 4)
		rv = 4
	case 'd':
		f := arg.(float64)
		xf := *(*uint64)(unsafe.Pointer(&f))
		putBytes(buf, uint64(xf), encLE, 8)
		rv = 8
	default:
		return 0, errors.New("unknown format type " + string(vfmt))
	}
	return rv, nil
}

func argbytes(fmtbyte byte, count int) int {
	n := 0
	switch fmtbyte {
	case 'x', 'c', 'b', 'B', '?':
		n = 1
	case 'h', 'H':
		n = 2
	case 'i', 'l', 'I', 'L', 'f':
		n = 4
	case 'q', 'Q', 'd':
		n = 8
	case 's', 'p':
		return count
	}
	return n * count
}

// Pack serializes one or more args values into buf according to the fmt string.
// The buf buffer must be large enough to store the desired data, and each arg
// must match in type to the requested type in fmt.
func Pack(fmt string, buf []byte, args ...interface{}) error {
	boff := 0
	narg := 0
	encLE := nativeIsLE
	fmtlen := len(fmt)
	for i := 0; i < fmtlen; i++ {
		// read (and skip) count if it exists
		count, ncountb := atoi(fmt[i:])
		if ncountb == 0 {
			if fmt[i] == 'p' {
				count = len(args[narg].(string)) + 1
			} else {
				count = 1
			}
		} else {
			i += ncountb
		}
		fmtbyte := fmt[i]

		n := argbytes(fmtbyte, count)
		if (len(buf) - boff) < n {
			return errors.New("insufficient buf space for fmt")
		}

		// handle those for which count has special or no meaning
		switch fmtbyte {
		case ' ', '\t', '\r', '\n':
			continue
		case '<':
			encLE = true
			continue
		case '>', '!':
			encLE = false
			continue
		case '=':
			encLE = nativeIsLE
			continue
		case 's':
			n := copy(buf[boff:boff+count], args[narg].(string))
			for ; n < count; n++ {
				buf[boff+n] = byte(0)
			}
			boff += count
			narg++
			continue
		case 'p':
			if count > 256 {
				return errors.New("Pascal string output space cannot be larger than 256 bytes")
			}
			buf[boff] = byte(count - 1)
			n := copy(buf[boff+1:boff+count], args[narg].(string))
			for ; n < count-1; n++ {
				buf[boff+1+n] = byte(0)
			}
			boff += count
			narg++
			continue
		}

		// iterate through count items of remaining types
		for ; count > 0; count-- {
			switch fmtbyte {
			case 'x':
				buf[boff] = 0
				boff++
			default:
				argbytes, err := packarg(buf[boff:], fmtbyte, args[narg], encLE)
				if err != nil {
					return err
				}
				boff += argbytes
				narg++
			}
		}
	}
	return nil
}

// Unpack deserializes values from buf according to the fmt string, returning values
// via the []interface{} array.
func Unpack(fmt string, buf []byte) ([]interface{}, error) {
	boff := 0
	decLE := nativeIsLE
	fmtlen := len(fmt)
	args := make([]interface{}, 0)
	for i := 0; i < fmtlen; i++ {
		// read (and skip) count if it exists
		count, ncountb := atoi(fmt[i:])
		if ncountb == 0 {
			if fmt[i] == 'p' {
				if len(buf)-boff >= 1 {
					count = int(buf[boff]) + 1
				}
			} else {
				count = 1
			}
		} else {
			i += ncountb
		}
		fmtbyte := fmt[i]

		n := argbytes(fmtbyte, count)
		if (len(buf) - boff) < n {
			return nil, errors.New("insufficient buf space for fmt")
		}

		// handle those for which count has special or no meaning
		switch fmtbyte {
		case ' ', '\t', '\r', '\n':
			continue
		case '<':
			decLE = true
			continue
		case '>', '!':
			decLE = false
			continue
		case '=':
			decLE = nativeIsLE
			continue
		case 's':
			args = append(args, string(bytes.Trim(buf[boff:boff+count], "\x00")))
			boff += count
			continue
		case 'p':
			nb := int(buf[boff])
			boff++
			// use count to trim returned string?
			args = append(args, string(bytes.Trim(buf[boff:boff+nb], "\x00")))
			boff += nb
			continue
		}

		// iterate through count items of remaining types
		for ; count > 0; count-- {
			switch fmtbyte {
			case 'x':
				// skip pad bytes
				boff++
			case 'b':
				args = append(args, int8(buf[boff]))
				boff++
			case 'B':
				args = append(args, buf[boff])
				boff++
			case '?':
				args = append(args, buf[boff] == byte(1))
				boff++
			case 'h':
				args = append(args, int16(getBytes(buf[boff:], decLE, 2)))
				boff += 2
			case 'H':
				args = append(args, uint16(getBytes(buf[boff:], decLE, 2)))
				boff += 2
			case 'i', 'l':
				args = append(args, int32(getBytes(buf[boff:], decLE, 4)))
				boff += 4
			case 'I', 'L':
				args = append(args, uint32(getBytes(buf[boff:], decLE, 4)))
				boff += 4
			case 'q':
				args = append(args, int64(getBytes(buf[boff:], decLE, 8)))
				boff += 8
			case 'Q':
				args = append(args, uint64(getBytes(buf[boff:], decLE, 8)))
				boff += 8
			case 'f':
				xf := uint32(getBytes(buf[boff:], decLE, 4))
				args = append(args, *(*float32)(unsafe.Pointer(&xf)))
				boff += 4
			case 'd':
				xf := getBytes(buf[boff:], decLE, 8)
				args = append(args, *(*float64)(unsafe.Pointer(&xf)))
				boff += 8
			default:
				return args, errors.New("unknown format type " + string(fmtbyte))
			}
		}
	}
	return args, nil
}
