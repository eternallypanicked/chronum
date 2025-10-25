package types

type StepDefinition struct {
    Name  string            `yaml:"name" json:"name"`
    Run   string            `yaml:"run" json:"run"`
    Needs []string          `yaml:"needs,omitempty" json:"needs,omitempty"`
    Env   map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Executor string            `yaml:"executor,omitempty" json:"executor,omitempty"`
}

type ChronumDefinition struct {
    Name     		string            `yaml:"name" json:"name"`
    Triggers 		[]string          `yaml:"triggers,omitempty" json:"triggers,omitempty"`
    Steps    		[]StepDefinition  `yaml:"steps" json:"steps"`
	MaxParallel 	int               `yaml:"maxParallel,omitempty" json:"maxParallel,omitempty"`
	StopOnFail   	bool              `yaml:"stopOnFail,omitempty" json:"stopOnFail,omitempty"`
    DefaultRetry 	int               `yaml:"retry,omitempty" json:"retry,omitempty"`
}
