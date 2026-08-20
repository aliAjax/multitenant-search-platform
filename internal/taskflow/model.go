package taskflow

type State string

const (
	Pending   State = "pending"
	Running   State = "running"
	Retrying  State = "retrying"
	Succeeded State = "succeeded"
	Failed    State = "failed"
)

var transitions = map[State]map[State]bool{
	Pending:  {Running: true},
	Running:  {Succeeded: true, Failed: true},
	Failed:   {Retrying: true},
	Retrying: {},
}

func Transition(from, to State) bool { return transitions[from][to] }
