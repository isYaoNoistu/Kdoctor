package lint

import "kdoctor/internal/snapshot"

func getCompose(snap *snapshot.Bundle) *snapshot.ComposeSnapshot {
	if snap == nil {
		return nil
	}
	return snap.Compose
}
