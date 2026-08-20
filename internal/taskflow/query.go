package taskflow

func Active(states map[string]State) []string {
	out := []string{}
	for id, state := range states {
		if state == Pending || state == Running || state == Retrying {
			out = append(out, id)
		}
	}
	return out
}
