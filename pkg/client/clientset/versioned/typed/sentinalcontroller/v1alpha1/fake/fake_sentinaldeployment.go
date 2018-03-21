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

package fake

import (
	v1alpha1 "github.com/karlhungus/sentinal-controller/pkg/apis/sentinalcontroller/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSentinalDeployments implements SentinalDeploymentInterface
type FakeSentinalDeployments struct {
	Fake *FakeSentinalcontrollerV1alpha1
	ns   string
}

var sentinaldeploymentsResource = schema.GroupVersionResource{Group: "sentinalcontroller.k8s.io", Version: "v1alpha1", Resource: "sentinaldeployments"}

var sentinaldeploymentsKind = schema.GroupVersionKind{Group: "sentinalcontroller.k8s.io", Version: "v1alpha1", Kind: "SentinalDeployment"}

// Get takes name of the sentinalDeployment, and returns the corresponding sentinalDeployment object, and an error if there is any.
func (c *FakeSentinalDeployments) Get(name string, options v1.GetOptions) (result *v1alpha1.SentinalDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(sentinaldeploymentsResource, c.ns, name), &v1alpha1.SentinalDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SentinalDeployment), err
}

// List takes label and field selectors, and returns the list of SentinalDeployments that match those selectors.
func (c *FakeSentinalDeployments) List(opts v1.ListOptions) (result *v1alpha1.SentinalDeploymentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(sentinaldeploymentsResource, sentinaldeploymentsKind, c.ns, opts), &v1alpha1.SentinalDeploymentList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SentinalDeploymentList{}
	for _, item := range obj.(*v1alpha1.SentinalDeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested sentinalDeployments.
func (c *FakeSentinalDeployments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(sentinaldeploymentsResource, c.ns, opts))

}

// Create takes the representation of a sentinalDeployment and creates it.  Returns the server's representation of the sentinalDeployment, and an error, if there is any.
func (c *FakeSentinalDeployments) Create(sentinalDeployment *v1alpha1.SentinalDeployment) (result *v1alpha1.SentinalDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(sentinaldeploymentsResource, c.ns, sentinalDeployment), &v1alpha1.SentinalDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SentinalDeployment), err
}

// Update takes the representation of a sentinalDeployment and updates it. Returns the server's representation of the sentinalDeployment, and an error, if there is any.
func (c *FakeSentinalDeployments) Update(sentinalDeployment *v1alpha1.SentinalDeployment) (result *v1alpha1.SentinalDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(sentinaldeploymentsResource, c.ns, sentinalDeployment), &v1alpha1.SentinalDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SentinalDeployment), err
}

// Delete takes name of the sentinalDeployment and deletes it. Returns an error if one occurs.
func (c *FakeSentinalDeployments) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(sentinaldeploymentsResource, c.ns, name), &v1alpha1.SentinalDeployment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSentinalDeployments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(sentinaldeploymentsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.SentinalDeploymentList{})
	return err
}

// Patch applies the patch and returns the patched sentinalDeployment.
func (c *FakeSentinalDeployments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SentinalDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(sentinaldeploymentsResource, c.ns, name, data, subresources...), &v1alpha1.SentinalDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SentinalDeployment), err
}
