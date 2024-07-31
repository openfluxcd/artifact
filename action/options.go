package action

import (
	"github.com/fluxcd/pkg/runtime/predicates"
	"github.com/openfluxcd/artifact/matchers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SourceMatcher = matchers.SourceMatcher

type RequestMapper func(list []runtime.Object) []reconcile.Request

type TriggerPredicate func(action ActionResource, art ArtifactSource) bool

type Options struct {
	ForOptions           []builder.ForOption
	NoCrossNamespaceRefs *bool
	AllowedSourceKinds   SourceMatcher
	TriggerPredicate     TriggerPredicate
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
	if o.TriggerPredicate != nil {
		opts.TriggerPredicate = o.TriggerPredicate
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

	if opts.RequestMapper == nil {
		opts.RequestMapper = DefaultRequestMapper
	}
	if opts.TriggerPredicate == nil {
		opts.TriggerPredicate = TriggerAlwaysPredicate
	}
	if opts.ForOptions == nil {
		opts.ForOptions = []builder.ForOption{builder.WithPredicates(
			predicate.Or(predicate.GenerationChangedPredicate{}, predicates.ReconcileRequestedPredicate{}),
		)}
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

type WithTriggerPredicate TriggerPredicate

func (o WithTriggerPredicate) Apply(opts *Options) {
	opts.TriggerPredicate = TriggerPredicate(o)
}

type foroptions []builder.ForOption

func WithForOptions(foroption ...builder.ForOption) Option {
	return foroptions(foroption)
}

func (o foroptions) Apply(opts *Options) {
	if len(o) == 0 && len(opts.ForOptions) == 0 {
		opts.ForOptions = []builder.ForOption{}
	} else {
		opts.ForOptions = append(opts.ForOptions, o...)
	}
}
