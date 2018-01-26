/*
Copyright 2018 The Kubernetes sample-controller Authors.

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

package v1

import (
	v1 "github.com/monasca/alarm-definition-controller/pkg/apis/alarmdefinition/v1"
	scheme "github.com/monasca/alarm-definition-controller/pkg/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AlarmDefinitionsGetter has a method to return a AlarmDefinitionInterface.
// A group's client should implement this interface.
type AlarmDefinitionsGetter interface {
	AlarmDefinitions(namespace string) AlarmDefinitionInterface
}

// AlarmDefinitionInterface has methods to work with AlarmDefinition resources.
type AlarmDefinitionInterface interface {
	Create(*v1.AlarmDefinition) (*v1.AlarmDefinition, error)
	Update(*v1.AlarmDefinition) (*v1.AlarmDefinition, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.AlarmDefinition, error)
	List(opts meta_v1.ListOptions) (*v1.AlarmDefinitionList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.AlarmDefinition, err error)
	AlarmDefinitionExpansion
}

// alarmDefinitions implements AlarmDefinitionInterface
type alarmDefinitions struct {
	client rest.Interface
	ns     string
}

// newAlarmDefinitions returns a AlarmDefinitions
func newAlarmDefinitions(c *MonascaV1Client, namespace string) *alarmDefinitions {
	return &alarmDefinitions{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the alarmDefinition, and returns the corresponding alarmDefinition object, and an error if there is any.
func (c *alarmDefinitions) Get(name string, options meta_v1.GetOptions) (result *v1.AlarmDefinition, err error) {
	result = &v1.AlarmDefinition{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("alarmdefinitions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AlarmDefinitions that match those selectors.
func (c *alarmDefinitions) List(opts meta_v1.ListOptions) (result *v1.AlarmDefinitionList, err error) {
	result = &v1.AlarmDefinitionList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("alarmdefinitions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested alarmDefinitions.
func (c *alarmDefinitions) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("alarmdefinitions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a alarmDefinition and creates it.  Returns the server's representation of the alarmDefinition, and an error, if there is any.
func (c *alarmDefinitions) Create(alarmDefinition *v1.AlarmDefinition) (result *v1.AlarmDefinition, err error) {
	result = &v1.AlarmDefinition{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("alarmdefinitions").
		Body(alarmDefinition).
		Do().
		Into(result)
	return
}

// Update takes the representation of a alarmDefinition and updates it. Returns the server's representation of the alarmDefinition, and an error, if there is any.
func (c *alarmDefinitions) Update(alarmDefinition *v1.AlarmDefinition) (result *v1.AlarmDefinition, err error) {
	result = &v1.AlarmDefinition{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("alarmdefinitions").
		Name(alarmDefinition.Name).
		Body(alarmDefinition).
		Do().
		Into(result)
	return
}

// Delete takes name of the alarmDefinition and deletes it. Returns an error if one occurs.
func (c *alarmDefinitions) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("alarmdefinitions").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *alarmDefinitions) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("alarmdefinitions").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched alarmDefinition.
func (c *alarmDefinitions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.AlarmDefinition, err error) {
	result = &v1.AlarmDefinition{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("alarmdefinitions").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
