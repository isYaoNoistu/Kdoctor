package disk

type Usage struct {
	Path            string
	TotalBytes      int64
	AvailableBytes  int64
	UsedBytes       int64
	UsedPercent     float64
	TotalInodes     int64
	AvailableInodes int64
	UsedInodes      int64
	UsedInodePct    float64
}
