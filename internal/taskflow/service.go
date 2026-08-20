package taskflow

import "fmt"

type Service struct{ states map[string]State }

func New() *Service                           { return &Service{states: map[string]State{}} }
func (s *Service) Set(id string, state State) { s.states[id] = state }
func (s *Service) State(id string) State      { return s.states[id] }

func (s *Service) Retry(id string) error {
	current := s.states[id]
	if !Transition(current, Retrying) {
		return fmt.Errorf("cannot retry %s from %s", id, current)
	}
	s.states[id] = Failed
	return nil
}

func (s *Service) Complete(id string) error {
	current := s.states[id]
	if !Transition(current, Succeeded) {
		return fmt.Errorf("cannot complete %s from %s", id, current)
	}
	s.states[id] = Succeeded
	return nil
}
