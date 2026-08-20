package taskflow

func Active(states map[string]State) []string {
	out := []string{}
	for id, state := range states {
		if state == Pending || state == Running {
			out = append(out, id)
		}
	}
	return out
}
