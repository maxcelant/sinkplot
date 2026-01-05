package config

type v1_ConfigSchema struct {
	App App `json:"app" yaml:"app"`
}

type App struct {
	Name      string  `json:"name" yaml:"name"`
	Listeners []int   `json:"listeners" yaml:"listeners"`
	Routes    []Route `json:"routes" yaml:"routes"`
	Sinks     []Sink  `json:"sinks" yaml:"sinks"`
}

type Route struct {
	Name     string    `json:"name" yaml:"name"`
	Matchers []Matcher `json:"matchers" yaml:"matchers"`
}

type Matcher struct {
	Endpoint string   `json:"endpoint" yaml:"endpoint"`
	Methods  []string `json:"methods" yaml:"methods"`
	Match    string   `json:"match" yaml:"match"` // exact | prefix
	Sink     string   `json:"sink" yaml:"sink"`
}

type Sink struct {
	Name      string     `json:"name" yaml:"name"`
	Upstreams []Upstream `json:"upstreams" yaml:"upstreams"`
}

type Upstream struct {
	Address string `json:"address" yaml:"address"`
	Port    int    `json:"port" yaml:"port"`
	Weight  int    `json:"weight" yaml:"weight"`
}
