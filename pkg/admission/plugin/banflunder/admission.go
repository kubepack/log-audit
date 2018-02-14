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

package banflunder

import (
	"fmt"
	"io"

	"github.com/kubepack/packserver/apis/apps"
	informers "github.com/kubepack/packserver/client/informers/internalversion"
	listers "github.com/kubepack/packserver/client/listers/apps/internalversion"
	"github.com/kubepack/packserver/pkg/admission/wardleinitializer"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/admission"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register("BanPack", func(config io.Reader) (admission.Interface, error) {
		return New()
	})
}

type DisallowPack struct {
	*admission.Handler
	lister listers.UserLister
}

var _ = wardleinitializer.WantsInternalWardleInformerFactory(&DisallowPack{})

// Admit ensures that the object in-flight is of kind Pack.
// In addition checks that the Name is not on the banned list.
// The list is stored in Users API objects.
func (d *DisallowPack) Admit(a admission.Attributes) error {
	// we are only interested in flunders
	if a.GetKind().GroupKind() != apps.Kind("Pack") {
		return nil
	}

	metaAccessor, err := meta.Accessor(a.GetObject())
	if err != nil {
		return err
	}
	flunderName := metaAccessor.GetName()

	fischers, err := d.lister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, fischer := range fischers {
		for _, disallowedPack := range fischer.DisallowedPacks {
			if flunderName == disallowedPack {
				return errors.NewForbidden(
					a.GetResource().GroupResource(),
					a.GetName(),
					fmt.Errorf("this name may not be used, please change the resource name"),
				)
			}
		}
	}
	return nil
}

// SetInternalWardleInformerFactory gets Lister from SharedInformerFactory.
// The lister knows how to lists Users.
func (d *DisallowPack) SetInternalWardleInformerFactory(f informers.SharedInformerFactory) {
	d.lister = f.Apps().InternalVersion().Users().Lister()
}

// ValidaValidateInitializationte checks whether the plugin was correctly initialized.
func (d *DisallowPack) ValidateInitialization() error {
	if d.lister == nil {
		return fmt.Errorf("missing user lister")
	}
	return nil
}

// New creates a new ban pack admission plugin
func New() (*DisallowPack, error) {
	return &DisallowPack{
		Handler: admission.NewHandler(admission.Create),
	}, nil
}
