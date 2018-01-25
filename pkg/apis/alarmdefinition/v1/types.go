/*
(C) Copyright 2018 Hewlett Packard Enterprise Development LP

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AlarmDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlarmDefinitionSpec   `json:"alarmDefinitionSpec"`
	Status AlarmDefinitionStatus `json:"status"`
}

type AlarmDefinitionSpec struct {
	ID                  string   `json:"id,omitempty"`
	Name                string   `json:"name,omitempty"`
	Expression          string   `json:"expression,omitempty"`
	Description         string   `json:"description,omitempty"`
	Severity            string   `json:"severity,omitempty"`
	MatchBy             []string `json:"match_by,omitempty"`
	AlarmActions        []string `json:"alarm_actions,omitempty"`
	OkActions           []string `json:"ok_actions,omitempty"`
	UndeterminedActions []string `json:"undetermined_actions,omitempty"`
}

type AlarmDefinitionStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AlarmDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AlarmDefinition `json:"items"`
}
