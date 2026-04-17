package snapshot

type KafkaSnapshot struct {
	ClusterID           string           `json:"cluster_id,omitempty"`
	Brokers             []BrokerSnapshot `json:"brokers,omitempty"`
	ControllerID        *int32           `json:"controller_id,omitempty"`
	ControllerAddress   string           `json:"controller_address,omitempty"`
	ExpectedBrokerCount int              `json:"expected_broker_count,omitempty"`
}

type BrokerSnapshot struct {
	ID      int32  `json:"id"`
	Address string `json:"address"`
}
