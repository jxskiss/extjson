package extjson

import "os"

// IncludeRoot specifies the root directory to use with the extended file
// including feature.
func IncludeRoot(dir string) ExtOption {
	return ExtOption{apply: func(options *extOptions) {
		options.IncludeRoot = dir
	}}
}

// ExtOption represents an option to customize the extended features.
type ExtOption struct {
	apply func(options *extOptions)
}

type extOptions struct {
	IncludeRoot string
}

func (o *extOptions) apply(opts ...ExtOption) *extOptions {
	for _, opt := range opts {
		opt.apply(o)
	}
	return o
}

func (o *extOptions) getIncludeRoot() (string, error) {
	if o.IncludeRoot != "" {
		return o.IncludeRoot, nil
	}
	return os.Getwd()
}
