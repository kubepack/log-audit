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
package fake

import (
	apps "github.com/kubepack/packserver/apis/apps"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePacks implements PackInterface
type FakePacks struct {
	Fake *FakeApps
	ns   string
}

var packsResource = schema.GroupVersionResource{Group: "apps.kubepack.com", Version: "", Resource: "packs"}

var packsKind = schema.GroupVersionKind{Group: "apps.kubepack.com", Version: "", Kind: "Pack"}

// Get takes name of the pack, and returns the corresponding pack object, and an error if there is any.
func (c *FakePacks) Get(name string, options v1.GetOptions) (result *apps.Pack, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(packsResource, c.ns, name), &apps.Pack{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apps.Pack), err
}

// List takes label and field selectors, and returns the list of Packs that match those selectors.
func (c *FakePacks) List(opts v1.ListOptions) (result *apps.PackList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(packsResource, packsKind, c.ns, opts), &apps.PackList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &apps.PackList{}
	for _, item := range obj.(*apps.PackList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested packs.
func (c *FakePacks) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(packsResource, c.ns, opts))

}

// Create takes the representation of a pack and creates it.  Returns the server's representation of the pack, and an error, if there is any.
func (c *FakePacks) Create(pack *apps.Pack) (result *apps.Pack, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(packsResource, c.ns, pack), &apps.Pack{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apps.Pack), err
}

// Update takes the representation of a pack and updates it. Returns the server's representation of the pack, and an error, if there is any.
func (c *FakePacks) Update(pack *apps.Pack) (result *apps.Pack, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(packsResource, c.ns, pack), &apps.Pack{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apps.Pack), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePacks) UpdateStatus(pack *apps.Pack) (*apps.Pack, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(packsResource, "status", c.ns, pack), &apps.Pack{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apps.Pack), err
}

// Delete takes name of the pack and deletes it. Returns an error if one occurs.
func (c *FakePacks) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(packsResource, c.ns, name), &apps.Pack{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePacks) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(packsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &apps.PackList{})
	return err
}

// Patch applies the patch and returns the patched pack.
func (c *FakePacks) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *apps.Pack, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(packsResource, c.ns, name, data, subresources...), &apps.Pack{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apps.Pack), err
}
