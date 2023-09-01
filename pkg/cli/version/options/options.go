package options

var (
	Domain  = "externalname-resolver-controller"
	Current = NewOptions()
)

func NewOptions() *Options {
	options := new(Options)

	return options
}

type Options struct {
	Semver bool
	Commit bool
}
