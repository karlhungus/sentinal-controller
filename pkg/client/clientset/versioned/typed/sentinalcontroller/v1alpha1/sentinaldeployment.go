/*
Copyright 2018 The Kubernetes Authors.

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

package v1alpha1

import (
	v1alpha1 "github.com/karlhungus/sentinal-controller/pkg/apis/sentinalcontroller/v1alpha1"
	scheme "github.com/karlhungus/sentinal-controller/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SentinalDeploymentsGetter has a method to return a SentinalDeploymentInterface.
// A group's client should implement this interface.
type SentinalDeploymentsGetter interface {
	SentinalDeployments(namespace string) SentinalDeploymentInterface
}

// SentinalDeploymentInterface has methods to work with SentinalDeployment resources.
type SentinalDeploymentInterface interface {
	Create(*v1alpha1.SentinalDeployment) (*v1alpha1.SentinalDeployment, error)
	Update(*v1alpha1.SentinalDeployment) (*v1alpha1.SentinalDeployment, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.SentinalDeployment, error)
	List(opts v1.ListOptions) (*v1alpha1.SentinalDeploymentList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SentinalDeployment, err error)
	SentinalDeploymentExpansion
}

// sentinalDeployments implements SentinalDeploymentInterface
type sentinalDeployments struct {
	client rest.Interface
	ns     string
}

// newSentinalDeployments returns a SentinalDeployments
func newSentinalDeployments(c *SentinalcontrollerV1alpha1Client, namespace string) *sentinalDeployments {
	return &sentinalDeployments{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the sentinalDeployment, and returns the corresponding sentinalDeployment object, and an error if there is any.
func (c *sentinalDeployments) Get(name string, options v1.GetOptions) (result *v1alpha1.SentinalDeployment, err error) {
	result = &v1alpha1.SentinalDeployment{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sentinaldeployments").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SentinalDeployments that match those selectors.
func (c *sentinalDeployments) List(opts v1.ListOptions) (result *v1alpha1.SentinalDeploymentList, err error) {
	result = &v1alpha1.SentinalDeploymentList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sentinaldeployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sentinalDeployments.
func (c *sentinalDeployments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("sentinaldeployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a sentinalDeployment and creates it.  Returns the server's representation of the sentinalDeployment, and an error, if there is any.
func (c *sentinalDeployments) Create(sentinalDeployment *v1alpha1.SentinalDeployment) (result *v1alpha1.SentinalDeployment, err error) {
	result = &v1alpha1.SentinalDeployment{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("sentinaldeployments").
		Body(sentinalDeployment).
		Do().
		Into(result)
	return
}

// Update takes the representation of a sentinalDeployment and updates it. Returns the server's representation of the sentinalDeployment, and an error, if there is any.
func (c *sentinalDeployments) Update(sentinalDeployment *v1alpha1.SentinalDeployment) (result *v1alpha1.SentinalDeployment, err error) {
	result = &v1alpha1.SentinalDeployment{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sentinaldeployments").
		Name(sentinalDeployment.Name).
		Body(sentinalDeployment).
		Do().
		Into(result)
	return
}

// Delete takes name of the sentinalDeployment and deletes it. Returns an error if one occurs.
func (c *sentinalDeployments) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sentinaldeployments").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sentinalDeployments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sentinaldeployments").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched sentinalDeployment.
func (c *sentinalDeployments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SentinalDeployment, err error) {
	result = &v1alpha1.SentinalDeployment{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("sentinaldeployments").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
