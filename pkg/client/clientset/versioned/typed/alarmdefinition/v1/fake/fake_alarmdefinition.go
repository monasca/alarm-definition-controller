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

package fake

import (
	alarmdefinition_v1 "github.com/monasca/alarm-definition-controller/pkg/apis/alarmdefinition/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAlarmDefinitions implements AlarmDefinitionInterface
type FakeAlarmDefinitions struct {
	Fake *FakeMonascaV1
	ns   string
}

var alarmdefinitionsResource = schema.GroupVersionResource{Group: "monasca.io", Version: "v1", Resource: "alarmdefinitions"}

var alarmdefinitionsKind = schema.GroupVersionKind{Group: "monasca.io", Version: "v1", Kind: "AlarmDefinition"}

// Get takes name of the alarmDefinition, and returns the corresponding alarmDefinition object, and an error if there is any.
func (c *FakeAlarmDefinitions) Get(name string, options v1.GetOptions) (result *alarmdefinition_v1.AlarmDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(alarmdefinitionsResource, c.ns, name), &alarmdefinition_v1.AlarmDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*alarmdefinition_v1.AlarmDefinition), err
}

// List takes label and field selectors, and returns the list of AlarmDefinitions that match those selectors.
func (c *FakeAlarmDefinitions) List(opts v1.ListOptions) (result *alarmdefinition_v1.AlarmDefinitionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(alarmdefinitionsResource, alarmdefinitionsKind, c.ns, opts), &alarmdefinition_v1.AlarmDefinitionList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &alarmdefinition_v1.AlarmDefinitionList{}
	for _, item := range obj.(*alarmdefinition_v1.AlarmDefinitionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested alarmDefinitions.
func (c *FakeAlarmDefinitions) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(alarmdefinitionsResource, c.ns, opts))

}

// Create takes the representation of a alarmDefinition and creates it.  Returns the server's representation of the alarmDefinition, and an error, if there is any.
func (c *FakeAlarmDefinitions) Create(alarmDefinition *alarmdefinition_v1.AlarmDefinition) (result *alarmdefinition_v1.AlarmDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(alarmdefinitionsResource, c.ns, alarmDefinition), &alarmdefinition_v1.AlarmDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*alarmdefinition_v1.AlarmDefinition), err
}

// Update takes the representation of a alarmDefinition and updates it. Returns the server's representation of the alarmDefinition, and an error, if there is any.
func (c *FakeAlarmDefinitions) Update(alarmDefinition *alarmdefinition_v1.AlarmDefinition) (result *alarmdefinition_v1.AlarmDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(alarmdefinitionsResource, c.ns, alarmDefinition), &alarmdefinition_v1.AlarmDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*alarmdefinition_v1.AlarmDefinition), err
}

// Delete takes name of the alarmDefinition and deletes it. Returns an error if one occurs.
func (c *FakeAlarmDefinitions) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(alarmdefinitionsResource, c.ns, name), &alarmdefinition_v1.AlarmDefinition{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAlarmDefinitions) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(alarmdefinitionsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &alarmdefinition_v1.AlarmDefinitionList{})
	return err
}

// Patch applies the patch and returns the patched alarmDefinition.
func (c *FakeAlarmDefinitions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *alarmdefinition_v1.AlarmDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(alarmdefinitionsResource, c.ns, name, data, subresources...), &alarmdefinition_v1.AlarmDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*alarmdefinition_v1.AlarmDefinition), err
}
