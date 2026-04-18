package transaction

import "kdoctor/internal/snapshot"

func topicExists(bundle *snapshot.Bundle, topicName string) bool {
	if bundle == nil || bundle.Topic == nil {
		return false
	}
	for _, topic := range bundle.Topic.Topics {
		if topic.Name == topicName {
			return true
		}
	}
	return false
}
