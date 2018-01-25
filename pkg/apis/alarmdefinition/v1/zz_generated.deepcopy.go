// +build !ignore_autogenerated

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

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlarmDefinition) DeepCopyInto(out *AlarmDefinition) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlarmDefinition.
func (in *AlarmDefinition) DeepCopy() *AlarmDefinition {
	if in == nil {
		return nil
	}
	out := new(AlarmDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AlarmDefinition) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlarmDefinitionList) DeepCopyInto(out *AlarmDefinitionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AlarmDefinition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlarmDefinitionList.
func (in *AlarmDefinitionList) DeepCopy() *AlarmDefinitionList {
	if in == nil {
		return nil
	}
	out := new(AlarmDefinitionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AlarmDefinitionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlarmDefinitionSpec) DeepCopyInto(out *AlarmDefinitionSpec) {
	*out = *in
	if in.MatchBy != nil {
		in, out := &in.MatchBy, &out.MatchBy
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AlarmActions != nil {
		in, out := &in.AlarmActions, &out.AlarmActions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.OkActions != nil {
		in, out := &in.OkActions, &out.OkActions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.UndeterminedActions != nil {
		in, out := &in.UndeterminedActions, &out.UndeterminedActions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlarmDefinitionSpec.
func (in *AlarmDefinitionSpec) DeepCopy() *AlarmDefinitionSpec {
	if in == nil {
		return nil
	}
	out := new(AlarmDefinitionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlarmDefinitionStatus) DeepCopyInto(out *AlarmDefinitionStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlarmDefinitionStatus.
func (in *AlarmDefinitionStatus) DeepCopy() *AlarmDefinitionStatus {
	if in == nil {
		return nil
	}
	out := new(AlarmDefinitionStatus)
	in.DeepCopyInto(out)
	return out
}