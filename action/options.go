package action

import (
	"github.com/openfluxcd/artifact/matchers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SourceMatcher = matchers.SourceMatcher

type RequestMapper func(list []runtime.Object) []reconcile.Request

type Options struct {
	NoCrossNamespaceRefs *bool
	AllowedSourceKinds   SourceMatcher
	RequestMapper        RequestMapper
}

func (o *Options) CrossNamespaceRefsForbidden() bool {
	return o.NoCrossNamespaceRefs != nil && *o.NoCrossNamespaceRefs
}

func (o *Options) Apply(opts *Options) {
	if o.AllowedSourceKinds != nil {
		opts.AllowedSourceKinds = o.AllowedSourceKinds
	}
	if o.NoCrossNamespaceRefs != nil {
		opts.NoCrossNamespaceRefs = o.NoCrossNamespaceRefs
	}
	if o.RequestMapper != nil {
		opts.RequestMapper = o.RequestMapper
	}
}

type Option interface {
	Apply(options *Options)
}

func EvalOptions(optList ...Option) *Options {
	opts := &Options{}
	for _, opt := range optList {
		opt.Apply(opts)
	}
	return opts
}

type nocrossnamespacerefs bool

func WithNoCrossNamespaceRefs(b ...bool) Option {
	if len(b) == 0 {
		return nocrossnamespacerefs(true)
	}
	return nocrossnamespacerefs(b[0])
}

func (o nocrossnamespacerefs) Apply(opts *Options) {
	b := bool(o)
	opts.NoCrossNamespaceRefs = &b
}

type allowedsourcekinds struct {
	SourceMatcher
}

func WithAllowedSourceKinds(m SourceMatcher) Option {
	return &allowedsourcekinds{m}
}

func (o *allowedsourcekinds) Apply(opts *Options) {
	opts.AllowedSourceKinds = o.SourceMatcher
}

type WithRequestMapper RequestMapper

func (o WithRequestMapper) Apply(opts *Options) {
	opts.RequestMapper = RequestMapper(o)
}
