package cnappgoat

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

type Scenario struct {
	Name           string         `yaml:"name"`
	Runtime        runtime        `yaml:"runtime"`
	Description    string         `yaml:"description,omitempty"`
	ScenarioParams scenarioParams `yaml:"cnappgoat-params"`
	State          State          `yaml:"-"`
	Hash           string         `yaml:"-"`
	SrcDir         string         `yaml:"-"`
}

type State struct {
	State string `yaml:"state"`
	Msg   string `yaml:"msg"`
}

type runtime struct {
	Name    string
	Options map[string]interface{}
}

type scenarioParams struct {
	Module       Module            `yaml:"module"`
	Platform     Platform          `yaml:"platform"`
	ID           string            `yaml:"id"`
	FriendlyName string            `yaml:"friendlyName"`
	Description  string            `yaml:"description"`
	ScenarioType string            `yaml:"scenarioType"`
	Config       map[string]string `yaml:"config"`
}

const (
	Deployed    = "deployed"
	Destroyed   = "destroyed"
	Error       = "error"
	NotDeployed = "not-deployed"
)

type Module string

const (
	CIEM Module = "CIEM"
	CSPM Module = "CSPM"
	CWPP Module = "CWPP"
	DSPM Module = "DSPM"
	IAC  Module = "IAC"
)

type Platform string

const (
	AWS   Platform = "AWS"
	Azure Platform = "AZURE"
	GCP   Platform = "GCP"
)

func (m Module) Equals(other Module) bool {
	return strings.EqualFold(string(m), string(other))
}

func (m Module) MarshalYAML() (interface{}, error) {
	return m.String(), nil
}

func (m Module) String() string {
	switch m {
	case CIEM:
		return "CIEM"
	case CSPM:
		return "CSPM"
	case CWPP:
		return "CWPP"
	case DSPM:
		return "DSPM"
	case IAC:
		return "IAC"
	default:
		panic("CNAPPgoat module name not formatted")
	}
}

func (m *Module) UnmarshalYAML(node *yaml.Node) error {
	name := node.Value
	module, err := ModuleFromString(name)
	if err != nil {
		return err
	}
	*m = module
	return nil
}

func (p Platform) Equals(other Platform) bool {
	return strings.EqualFold(string(p), string(other))
}

func (p Platform) String() string {
	switch p {
	case AWS:
		return "AWS"
	case Azure:
		return "AZURE"
	case GCP:
		return "GCP"
	default:
		panic("platform name not formatted")
	}
}

func (p Platform) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

func (p *Platform) UnmarshalYAML(node *yaml.Node) error {
	platform, err := PlatformFromString(node.Value)
	if err != nil {
		return err
	}
	*p = platform
	return nil
}

func (r *runtime) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var runtimeStr string
	err := unmarshal(&runtimeStr)
	if err == nil {
		r.Name = runtimeStr
		return nil
	}

	var runtimeObj map[string]interface{}
	if err = unmarshal(&runtimeObj); err != nil {
		return err
	}

	if name, ok := runtimeObj["name"].(string); ok {
		r.Name = name
	}
	r.Options = runtimeObj

	return nil
}

func (s State) String() string {
	return s.State
}

func (s State) Equals(state State) bool {
	return strings.EqualFold(s.State, state.State)
}

func ModuleFromString(name string) (Module, error) {
	switch strings.ToLower(name) {
	case strings.ToLower(CIEM.String()):
		return CIEM, nil
	case strings.ToLower(CSPM.String()):
		return CSPM, nil
	case strings.ToLower(CWPP.String()):
		return CWPP, nil
	case strings.ToLower(DSPM.String()):
		return DSPM, nil
	case strings.ToLower(IAC.String()):
		return IAC, nil
	default:
		return "", errors.New("unknown CNAPPgoat module name: " + name)
	}
}

func PlatformFromString(name string) (Platform, error) {
	switch strings.ToLower(name) {
	case strings.ToLower(AWS.String()):
		return AWS, nil
	case strings.ToLower(Azure.String()):
		return Azure, nil
	case strings.ToLower(GCP.String()):
		return GCP, nil
	default:
		return "", errors.New("unknown platform name: " + name)
	}
}

func StateFromString(state string) (State, error) {
	switch strings.ToLower(state) {
	case strings.ToLower(Deployed):
		return State{State: Deployed}, nil
	case strings.ToLower(Destroyed):
		return State{State: Destroyed}, nil
	case strings.ToLower(Error):
		return State{State: Error}, nil
	case strings.ToLower(NotDeployed):
		return State{State: NotDeployed}, nil
	default:
		return State{}, errors.New("unknown state: " + state)
	}
}
