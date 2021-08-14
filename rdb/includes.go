package rdb

type IncludeOption func(cnf *includesConfig)

func Include(includes ...string) IncludeOption {
	return func(cnf *includesConfig) {
		cnf.includes = append(cnf.includes, includes...)
	}
}

func IncludeAllCounters() IncludeOption {
	return func(cnf *includesConfig) {
		cnf.allCounters = true
	}
}

type includesConfig struct {
	includes    []string
	allCounters bool
}

func applyModelIncludes(opts ...IncludeOption) []string {
	return resolveIncludes(opts...).includes
}

func resolveIncludes(opts ...IncludeOption) *includesConfig {
	cnf := new(includesConfig)
	for _, opt := range opts {
		opt(cnf)
	}
	return cnf
}
