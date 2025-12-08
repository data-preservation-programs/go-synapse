package constants

const (
	KiB = 1024
	MiB = 1 << 20
	GiB = 1 << 30
	TiB = 1 << 40

	MaxUploadSize = GiB * 127 / 128

	MinUploadSize = 127
)
