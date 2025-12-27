package keystore

const HardenedOffset = 0x80000000

func Hardened(i uint32) uint32 {
	return i + HardenedOffset
}
