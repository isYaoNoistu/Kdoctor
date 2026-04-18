package snapshot

type GroupSnapshot struct {
	Collected bool               `json:"collected"`
	Available bool               `json:"available"`
	Targets   []GroupLagSnapshot `json:"targets,omitempty"`
	Errors    []string           `json:"errors,omitempty"`
}

type GroupLagSnapshot struct {
	Name            string                  `json:"name,omitempty"`
	GroupID         string                  `json:"group_id"`
	Topic           string                  `json:"topic"`
	State           string                  `json:"state,omitempty"`
	Coordinator     string                  `json:"coordinator,omitempty"`
	MemberCount     int                     `json:"member_count,omitempty"`
	TotalLag        int64                   `json:"total_lag,omitempty"`
	MaxPartitionLag int64                   `json:"max_partition_lag,omitempty"`
	MaxLagPartition int32                   `json:"max_lag_partition,omitempty"`
	MissingOffsets  int                     `json:"missing_offsets,omitempty"`
	Error           string                  `json:"error,omitempty"`
	Partitions      []GroupPartitionLagInfo `json:"partitions,omitempty"`
}

type GroupPartitionLagInfo struct {
	Partition          int32 `json:"partition"`
	CommittedOffset    int64 `json:"committed_offset"`
	EndOffset          int64 `json:"end_offset"`
	Lag                int64 `json:"lag"`
	HasCommittedOffset bool  `json:"has_committed_offset"`
}
