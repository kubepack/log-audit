/*
Copyright 2018 The Kubepack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pack

import (
	"fmt"

	"github.com/kubepack/packserver/apis/apps"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
)

func NewStrategy(typer runtime.ObjectTyper) flunderStrategy {
	return flunderStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*apps.Pack)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a Pack.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), PackToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchPack is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchPack(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// PackToSelectableFields returns a field set that represents the object.
func PackToSelectableFields(obj *apps.Pack) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type flunderStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (flunderStrategy) NamespaceScoped() bool {
	return true
}

func (flunderStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
}

func (flunderStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
}

func (flunderStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (flunderStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (flunderStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (flunderStrategy) Canonicalize(obj runtime.Object) {
}

func (flunderStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
