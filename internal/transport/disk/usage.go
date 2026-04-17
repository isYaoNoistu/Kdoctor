package disk

type Usage struct {
	Path           string
	TotalBytes     int64
	AvailableBytes int64
	UsedBytes      int64
	UsedPercent    float64
}
