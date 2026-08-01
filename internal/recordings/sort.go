package recordings

import "sort"

// sortSessionsDesc sorts sessions by CreatedAt descending (newest first).
func sortSessionsDesc(sessions []Session) {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})
}

// sortEvents sorts events by SequenceOrder ascending.
func sortEvents(events []Event) {
	sort.Slice(events, func(i, j int) bool {
		return events[i].SequenceOrder < events[j].SequenceOrder
	})
}
