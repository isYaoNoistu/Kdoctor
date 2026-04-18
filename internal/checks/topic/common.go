package topic

import "kdoctor/internal/snapshot"

func topicSnap(bundle *snapshot.Bundle) *snapshot.TopicSnapshot {
	if bundle == nil {
		return nil
	}
	return bundle.Topic
}

func metricsSnap(bundle *snapshot.Bundle) *snapshot.MetricsSnapshot {
	if bundle == nil {
		return nil
	}
	return bundle.Metrics
}
