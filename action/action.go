package action

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/fluxcd/pkg/runtime/acl"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	artifactv1 "github.com/openfluxcd/artifact/api/v1alpha1"
	"github.com/openfluxcd/artifact/matchers"
	"github.com/openfluxcd/artifact/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SourceRefIndexKey     = ".metadata.sourceref"
	ArtifactOwnerIndexKey = ".metadata.artifactowner"
)

type ArtifactSource interface {
	GetArtifact() *sourcev1.Artifact
}

type ActionResourcePointerType[T any] interface {
	*T
	ActionResource
}

type ActionResource interface {
	ctrlclient.Object
	GetSourceRef() (utils.SourceRefProvider, error)
}

var _ ArtifactSource = sourcev1.Source(nil)

func requestsForRevisionChangeOf[T any, P ActionResourcePointerType[T]](client ctrlclient.Client, scheme *runtime.Scheme, opts *Options) handler.MapFunc {
	// Queues requests for all kustomization resources that either reference the artifact resource directly in their
	// source ref or that reference another resource that is referenced by the owner reference of the artifact in their
	// source ref.
	return func(ctx context.Context, obj ctrlclient.Object) []reconcile.Request {
		log := ctrl.LoggerFrom(ctx)
		src, ok := obj.(ArtifactSource)
		if !ok {
			log.Error(fmt.Errorf("expected an object conformed with GetArtifact() method, but got a %T", obj),
				"failed to get reconcile requests for revision change")
			return nil
		}
		// If we do not have an artifact, we have no requests to make
		if src.GetArtifact() == nil {
			return nil
		}

		actions := lookupBySourceObj[T, P](ctx, client, scheme, obj, src.GetArtifact())

		if art, ok := src.(*artifactv1.Artifact); ok {
			list := utils.CreateListForType[T, P](scheme)
			gk := utils.GetGroupKindForObject(scheme, obj)
			if err := client.List(ctx, list, ctrlclient.MatchingFields{
				SourceRefIndexKey: fmt.Sprintf("%s/%s/%s/%s", gk.Group, gk.Kind, obj.GetNamespace(), obj.GetName()),
			}); err != nil {
				log.Error(err, "failed to list objects for revision change")
				return nil
			}

			for _, ref := range art.OwnerReferences {
				actions = append(actions, lookupByCoordinates[T, P](ctx, client, scheme, utils.ExtractGroupName(ref.APIVersion), ref.Kind, art.Namespace, ref.Name, art.GetArtifact())...)
			}
		}
		for i := 0; i < len(actions); i++ {
			if !opts.TriggerPredicate(actions[i].(ActionResource), src) {
				actions = append(actions[:i], actions[i+1:]...)
				i--
			}
		}

		return opts.RequestMapper(actions)
	}
}

func lookupBySourceObj[T any, P ActionResourcePointerType[T]](ctx context.Context, client ctrlclient.Client, scheme *runtime.Scheme, srcobj ctrlclient.Object, art *sourcev1.Artifact) []runtime.Object {
	gk := utils.GetGroupKindForObject(scheme, srcobj)
	return lookupByCoordinates[T, P](ctx, client, scheme, gk.Group, gk.Kind, srcobj.GetNamespace(), srcobj.GetName(), art)
}

func lookupByCoordinates[T any, P ActionResourcePointerType[T]](ctx context.Context, client ctrlclient.Client, scheme *runtime.Scheme, group, kind, ns, name string, art *sourcev1.Artifact) []runtime.Object {
	log := ctrl.LoggerFrom(ctx)
	list := utils.CreateListForType[T, P](scheme)
	if err := client.List(ctx, list, ctrlclient.MatchingFields{
		SourceRefIndexKey: fmt.Sprintf("%s/%s/%s/%s", group, kind, ns, name),
	}); err != nil {
		log.Error(err, "failed to list objects for revision change")
		return nil
	}

	actions, _ := meta.ExtractList(list)
	// filter action by comparing status and art revision

	return actions
}

func Setup[T any, P ActionResourcePointerType[T]](ctx context.Context, mgr ctrl.Manager, client ctrlclient.Client, options ...Option) (*builder.Builder, error) {
	var _obj T
	obj := P(&_obj)

	if err := mgr.GetCache().IndexField(ctx, obj, SourceRefIndexKey,
		SourceReferenceIndex[P]()); err != nil {
		return nil, fmt.Errorf("failed setting index fields: %w", err)
	}

	if err := mgr.GetCache().IndexField(ctx, &artifactv1.Artifact{}, ArtifactOwnerIndexKey,
		utils.OwnerReferenceIndex()); err != nil {
		return nil, fmt.Errorf("failed setting index fields: %w", err)
	}

	opts := EvalOptions(options...)
	bldr := ctrl.NewControllerManagedBy(mgr)

	bldr.For(obj, opts.ForOptions...)
	for gk, o := range matchers.BuiltinFluxSourceKinds {
		if opts.AllowedSourceKinds == nil || opts.AllowedSourceKinds.Match(gk) {
			bldr = bldr.Watches(
				o.DeepCopyObject().(ctrlclient.Object),
				handler.EnqueueRequestsFromMapFunc(requestsForRevisionChangeOf[T, P](client, mgr.GetScheme(), opts)),
				builder.WithPredicates(SourceRevisionChangePredicate{}),
			)
		}
	}
	return bldr, nil
}

func SourceReferenceIndex[T ActionResource]() func(o ctrlclient.Object) []string {
	return func(o ctrlclient.Object) []string {
		k, ok := o.(T)
		if !ok {
			var _nil T
			panic(fmt.Sprintf("Expected a resource of type %T, got %T", _nil, o))
		}
		sourceref, err := k.GetSourceRef()
		if err != nil {
			return nil
		}
		key := utils.KeyForReference(k, sourceref)
		if key != "" {
			return []string{key}
		}

		return nil
	}
}

func GetSource(ctx context.Context, client ctrlclient.Client, action ActionResource, options ...Option) (ArtifactSource, error) {
	raw, err := action.GetSourceRef()
	if err != nil {
		return nil, err
	}
	ref := utils.NormalizedSourceRef(raw, action.GetNamespace())

	opts := EvalOptions(options...)
	if opts.CrossNamespaceRefsForbidden() && ref.GetNamespace() != action.GetNamespace() {
		return nil, acl.AccessDeniedError(
			fmt.Sprintf("can't access '%s/%s', cross-namespace references have been blocked",
				ref.GetGroupKind().Kind, ref.GetNamespace()))
	}

	gk := ref.GetGroupKind()

	if opts.AllowedSourceKinds != nil && !opts.AllowedSourceKinds.Match(gk) {
		return nil, fmt.Errorf("source objects of kind %s are not allowed", gk)
	}

	if obj := matchers.BuiltinFluxSourceKinds.Create(gk); obj != nil {
		src, ok := obj.(ArtifactSource)
		if !ok {
			return nil, fmt.Errorf("source object %s is not an ArtifactSource", gk)
		}
		err := client.Get(ctx, ref.GetObjectKey(), obj)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil, err
			}
			return nil, fmt.Errorf("unable to get source '%s': %w", ref.GetObjectKey(), err)
		}
		return src, nil
	} else {
		namespacedName := types.NamespacedName{
			Namespace: ref.GetNamespace(),
		}

		key := utils.KeyForReference(action, ref)
		artList := &artifactv1.ArtifactList{}
		if key != "" {
			err := client.List(ctx, artList, ctrlclient.MatchingFields{
				ArtifactOwnerIndexKey: key,
			})
			if err != nil {
				return nil, err
			}
			switch len(artList.Items) {
			case 0:
				return nil, fmt.Errorf("no artifact resource found for %s", key)
			case 1:
				namespacedName.Name = artList.Items[0].Name
			default:
				return nil, fmt.Errorf("multiple artifacts found for %s", key)
			}

			var art artifactv1.Artifact
			err = client.Get(ctx, namespacedName, &art)
			if err != nil {
				if apierrors.IsNotFound(err) {
					return nil, err
				}
				return nil, fmt.Errorf("unable to get source '%s': %w", namespacedName, err)
			}
			return &art, nil
		} else {
			return nil, fmt.Errorf("no source ref specified")
		}
	}
}
