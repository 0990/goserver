package util

import "hash/crc32"

// 字符串转为16位整形哈希
func StringHash(s string) (hash uint16) {
	for _, c := range s {

		ch := uint16(c)

		hash = hash + ((hash) << 5) + ch + (ch << 7)
	}

	return
}

func CRC32Hash(s string) uint32 {
	v := crc32.ChecksumIEEE([]byte(s))
	if v < 0 {
		return -v
	}
	return v
}
