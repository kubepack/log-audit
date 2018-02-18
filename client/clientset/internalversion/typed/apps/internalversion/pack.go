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
package internalversion

import (
	apps "github.com/kubepack/packserver/apis/apps"
	scheme "github.com/kubepack/packserver/client/clientset/internalversion/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PacksGetter has a method to return a PackInterface.
// A group's client should implement this interface.
type PacksGetter interface {
	Packs(namespace string) PackInterface
}

// PackInterface has methods to work with Pack resources.
type PackInterface interface {
	Create(*apps.Pack) (*apps.Pack, error)
	Update(*apps.Pack) (*apps.Pack, error)
	UpdateStatus(*apps.Pack) (*apps.Pack, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*apps.Pack, error)
	List(opts v1.ListOptions) (*apps.PackList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *apps.Pack, err error)
	PackExpansion
}

// packs implements PackInterface
type packs struct {
	client rest.Interface
	ns     string
}

// newPacks returns a Packs
func newPacks(c *AppsClient, namespace string) *packs {
	return &packs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the pack, and returns the corresponding pack object, and an error if there is any.
func (c *packs) Get(name string, options v1.GetOptions) (result *apps.Pack, err error) {
	result = &apps.Pack{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("packs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Packs that match those selectors.
func (c *packs) List(opts v1.ListOptions) (result *apps.PackList, err error) {
	result = &apps.PackList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("packs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested packs.
func (c *packs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("packs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a pack and creates it.  Returns the server's representation of the pack, and an error, if there is any.
func (c *packs) Create(pack *apps.Pack) (result *apps.Pack, err error) {
	result = &apps.Pack{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("packs").
		Body(pack).
		Do().
		Into(result)
	return
}

// Update takes the representation of a pack and updates it. Returns the server's representation of the pack, and an error, if there is any.
func (c *packs) Update(pack *apps.Pack) (result *apps.Pack, err error) {
	result = &apps.Pack{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("packs").
		Name(pack.Name).
		Body(pack).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *packs) UpdateStatus(pack *apps.Pack) (result *apps.Pack, err error) {
	result = &apps.Pack{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("packs").
		Name(pack.Name).
		SubResource("status").
		Body(pack).
		Do().
		Into(result)
	return
}

// Delete takes name of the pack and deletes it. Returns an error if one occurs.
func (c *packs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("packs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *packs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("packs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched pack.
func (c *packs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *apps.Pack, err error) {
	result = &apps.Pack{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("packs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
