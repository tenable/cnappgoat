package cnappgoat

import (
	"fmt"
	"sort"
)

type Registry struct {
	scenarios map[string]*Scenario
	storage   *LocalStorage
}

type ListScenariosOptions struct {
	Module   Module
	Platform Platform
	State    State
}

type ListScenariosOption func(options *ListScenariosOptions)

func NewRegistry(s *LocalStorage) (*Registry, error) {
	registry := Registry{
		scenarios: make(map[string]*Scenario),
		storage:   s,
	}

	if s.WorkingDirectoryExists() {
		var err error
		registry.scenarios, err = s.LoadScenariosFromWorkingDir()
		if err != nil {
			return &registry, fmt.Errorf("error loading scenarios from working directory: %w", err)
		}
	}

	return &registry, nil
}

func (r *Registry) UpdateRegistryFromGit() error {
	scenariosFromGit, err := r.storage.UpdateScenariosFromGit()
	if err != nil {
		return fmt.Errorf("error updating scenarios from git: %w", err)
	}
	for _, scenario := range scenariosFromGit {
		r.scenarios[scenario.ScenarioParams.ID] = scenario
	}

	return nil
}

func (r *Registry) GetScenario(id string) (*Scenario, bool) {
	scenario, ok := r.scenarios[id]
	if !ok {
		return nil, false
	}

	return scenario, true
}

func (r *Registry) GetScenarios() []*Scenario {
	var scenarios []*Scenario
	for _, scenario := range r.scenarios {
		scenarios = append(scenarios, scenario)
	}

	sortScenarios(scenarios)
	return scenarios
}

func (r *Registry) ListScenarios() []*Scenario {
	var scenarios []*Scenario
	for name := range r.scenarios {
		scenarios = append(scenarios, r.scenarios[name])
	}
	sortScenarios(scenarios)
	return scenarios
}

func (r *Registry) ListScenariosWithOptions(opts ...ListScenariosOption) []*Scenario {
	options := ListScenariosOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	var scenarios []*Scenario
	for name := range r.scenarios {
		if (r.scenarios[name].ScenarioParams.Platform.Equals(options.Platform) || options.Platform.Equals("")) &&
			(r.scenarios[name].ScenarioParams.Module.Equals(options.Module) || options.Module.Equals("")) &&
			(r.scenarios[name].State.Equals(options.State) || options.State.Equals(State{})) {
			scenarios = append(scenarios, r.scenarios[name])
		}
	}

	sortScenarios(scenarios)
	return scenarios
}

func WithModule(module Module) ListScenariosOption {
	return func(opts *ListScenariosOptions) {
		opts.Module = module
	}
}

func WithPlatform(platform Platform) ListScenariosOption {
	return func(opts *ListScenariosOptions) {
		opts.Platform = platform
	}
}

func WithState(state State) ListScenariosOption {
	return func(opts *ListScenariosOptions) {
		opts.State = state
	}
}

func (r *Registry) SetState(scenario *Scenario, state State) error {
	r.scenarios[scenario.ScenarioParams.ID].State = state
	return r.storage.WriteStateToFile(r.scenarios[scenario.ScenarioParams.ID], state)
}

func (r *Registry) ImportScenarios(path string) (map[string]*Scenario, error) {
	return r.storage.updateScenariosFromFolder(path)
}

func sortScenarios(scenarios []*Scenario) {
	sort.Slice(scenarios, func(i, j int) bool {
		return scenarios[i].ScenarioParams.ID < scenarios[j].ScenarioParams.ID
	})
}
