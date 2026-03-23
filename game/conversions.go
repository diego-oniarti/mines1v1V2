package game

import "encoding/binary"

func UpdatesToArray(ups []Update) []byte {
	ret := make([]byte, len(ups)*5)

	for i, up := range ups {
		binary.BigEndian.PutUint16(ret[i*5:],   uint16(up.x))
		binary.BigEndian.PutUint16(ret[i*5+2:], uint16(up.y))
		detail := byte(up.num)
		if up.flag { detail |= 0b10000 }
		ret[i*5+4] = detail
	}

	return ret
}
