package utils

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectPointerType[T any] interface {
	*T
	client.Object
}

func CreateListForType[T any, P ObjectPointerType[T]](scheme *runtime.Scheme) client.ObjectList {
	var o T
	typ := reflect.TypeOf(o)
	pkg := typ.PkgPath()
	list := typ.Name() + "List"
	for _, t := range scheme.AllKnownTypes() {
		if t.Name() == list && t.PkgPath() == pkg {
			return reflect.New(t).Interface().(client.ObjectList)
		}
	}
	return nil
}

func GetGroupKindForType[T any, P ObjectPointerType[T]](scheme *runtime.Scheme) *schema.GroupKind {
	var o T
	typ := reflect.TypeOf(o)
	for k, t := range scheme.AllKnownTypes() {
		if t == typ {
			gk := k.GroupKind()
			return &gk
		}
	}
	return nil
}

func GetGroupKindForObject(scheme *runtime.Scheme, obj client.Object) *schema.GroupKind {
	typ := reflect.TypeOf(obj)
	if typ.Kind() != reflect.Ptr {
		return nil
	}
	typ = typ.Elem()

	for k, t := range scheme.AllKnownTypes() {
		if t == typ {
			gk := k.GroupKind()
			return &gk
		}
	}
	return nil
}
