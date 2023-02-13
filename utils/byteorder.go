package utils

import (
	"encoding/binary"
	"unsafe"
)

var HostByteOrder binary.ByteOrder
var NetworkByteOrder = binary.BigEndian

func init() {
	// dynamically detect the host's byte order
	// https://stackoverflow.com/a/53286786

	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		HostByteOrder = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		HostByteOrder = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}
