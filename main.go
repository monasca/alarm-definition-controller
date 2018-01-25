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

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/monasca/golang-monascaclient/monascaclient"
	"github.com/monasca/golang-monascaclient/monascaclient/models"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"

	"github.com/monasca/alarm-definition-controller/pkg/apis/alarmdefinition/v1"
	clientset "github.com/monasca/alarm-definition-controller/pkg/client/clientset/versioned"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	controllerAgentName             = "alarmDefinitionController"
	alarmDefinitionControllerSuffix = " - adc"
)

var (
	masterURL           = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig          = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	monascaUrl          = flag.String("monasca", getEnvDefault("MONASCA_API_URL", "http://monasca-api:8070/v2.0"), "The URI of the monasca api")
	namespace           = flag.String("namespace", getEnvDefault("NAMESPACE", "default"), "The namespace to watch for definitions")
	prometheusEndpoint  = flag.String("prometheus_endpoint", getEnvDefault("PROMETHEUS_ENDPOINT", "127.0.0.1"), "The endpoint to expose prometheus metrics")
	defaultNotification = flag.String("default-notification", getEnvDefault("DEFAULT_NOTIFICATION", ""), "A default notification method to apply to new definitions")
	pollInterval        = flag.Int("poll-interval", 15, "The interval in seconds to poll the resources")

	// prometheus metrics
	definitionErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "alarm_definition_errors",
			Help: "Number of errors encountered while creating and updating alarm definitions"})

	defaultNotificationID string

	alarmDefinitionCache = make(map[string]models.AlarmDefinitionElement)
)

func main() {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	definitionClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building example clientset: %s", err.Error())
	}

	go func() {
		// Start prometheus endpoint
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(prometheusEndpoint, nil))
	}()

	pollDefinitions(definitionClient, kubeClient)
}

func pollDefinitions(defClient clientset.Interface, kubeClient kubernetes.Interface) {
	log.Print("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	monascaclient.SetBaseURL(*monascaUrl)
	setKeystoneToken()
	updateCache()
	log.Printf("Found existing alarms %v", alarmDefinitionCache)

	if *defaultNotification != "" {
		go func() {
			log.Printf("Searching for default notification method named %s", defaultNotification)

			failureCount := 0
		pollLoop:
			for true {
				notifications, err := monascaclient.GetNotificationMethods(nil)
				if err != nil {
					log.Printf("Error fetching notification methods: %s", err.Error())
					failureCount++
					if failureCount >= 3 {
						log.Fatal("Could not retrieve notifications after three tries, quitting.")
					}
				} else {
					for _, notif := range notifications.Elements {
						if notif.Name == *defaultNotification {
							log.Printf("Found notification with ID %s", notif.ID)
							defaultNotificationID = notif.ID
							break pollLoop
						}
					}
					log.Printf("Could not find a notification named %s in the list", defaultNotification)
					failureCount = 0
				}
				time.Sleep(time.Duration(*pollInterval) * time.Second)
			}
		}()
	} else {
		log.Print("No default notification specified, skipping lookup")
	}

	first := true

	// Events and errors are not expected to be generated very often so
	// only allow the controller to buffer 100 of each.
	for {
		// Sleep for the poll interval if not the first time around.
		// Do this here so we sleep for the poll interval every time,
		// even after errors occur.
		if !first {
			time.Sleep(time.Duration(*pollInterval) * time.Second)
		}
		first = false

		err := setKeystoneToken()
		if err != nil {
			log.Printf("Failed to retrieve new keystone token: %s", err.Error())
			continue
		}

		opts := metav1.ListOptions{}
		l, err := defClient.Monasca().AlarmDefinitions(*namespace).List(opts)
		if err != nil {
			log.Printf("Error fetching resources: %s", err.Error())
			continue
		}

		// loop to remove alarms
		for id, cached := range alarmDefinitionCache {
			exists := false
			for _, discovered := range l.Items {
				// check for equality
				if cached.ID == discovered.Spec.ID {
					exists = true
				}
			}
			if !exists {
				// remove definitions from monasca
				err := removeAlarmDefinition(id)
				if err != nil {
					log.Printf("Error removing definition: %s", err.Error())
					definitionErrors.Inc()
					continue
				}
			}
		}

		// loop to add/update alarms
	discoveredLoop:
		for _, item := range l.Items {
			// if not marked with ID, add new
			if item.Spec.ID == "" {
				err := addAlarmDefinition(*namespace, &item, defClient)
				if err != nil {
					// If 409 is returned, we probably had a desync between cache
					// and monasca. This can happen if monasca returned an error, but
					// still created the definition.
					if strings.HasPrefix(err.Error(), "Error: 409") {
						log.Print("Mismatch between definitions and cache, updating cache")
						updateCache()
						definitionErrors.Inc()
						continue
					}
					log.Print(err)
					recorder.Event(&item, corev1.EventTypeWarning, err.Error(), "")
				}
				continue
			}

			for id, cached := range alarmDefinitionCache {
				// if exists, check if needs update
				if item.Spec.ID == id && !equal(&item, cached) {
					//update if possible
					err := updateAlarmDefinition(&item, defClient)
					if err != nil {
						log.Print(err)
						definitionErrors.Inc()
						recorder.Event(&item, corev1.EventTypeWarning, err.Error(), "")
					}
					continue discoveredLoop
				}
			}

		}
	}
}

func updateCache() error {
	existing, err := monascaclient.GetAlarmDefinitions(nil)
	if err != nil {
		log.Print(err)
		return err
	}
	//clear map
	for k := range alarmDefinitionCache {
		delete(alarmDefinitionCache, k)
	}
	for _, item := range existing.Elements {
		//ignore all alarm definitions that do not have the adc suffix
		if strings.HasSuffix(item.Name, alarmDefinitionControllerSuffix) {
			alarmDefinitionCache[item.ID] = item
		}
	}

	return nil
}

func setKeystoneToken() error {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		log.Printf("Error obtaining auth settings from env: %s", err.Error())
		return err
	}

	openstackProvider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		log.Printf("Error authenticating with keystone: %s", err.Error())
		return err
	}
	//log.Print(openstackProvider.TokenID)
	token := openstackProvider.TokenID
	headers := http.Header{}
	headers.Add("X-Auth-Token", token)
	monascaclient.SetHeaders(headers)
	return nil
}

func addAlarmDefinition(namespace string, r *v1.AlarmDefinition, sampleclientset clientset.Interface) error {
	if *defaultNotification != "" && len(r.Spec.AlarmActions) <= 0 {
		if defaultNotificationID == "" {
			return fmt.Errorf("Unable to apply default notification method: no ID found")
		}
		r.Spec.AlarmActions = []string{defaultNotificationID}
	}
	definitionRequest := convertToADRequest(r)
	glog.Warning(definitionRequest)
	result, err := monascaclient.CreateAlarmDefinition(definitionRequest)
	if err != nil {
		return err
	}
	alarmDefinitionCache[result.ID] = *result
	newDef := convertToADResource(r, result)
	_, err = sampleclientset.Monasca().AlarmDefinitions(namespace).Update(newDef)
	if err != nil {
		log.Printf("Failed to update resource %s", newDef.Name)
		return err
	}

	log.Printf("Added definition %v", r.Spec)
	return nil
}

func updateAlarmDefinition(r *v1.AlarmDefinition, sampleclientset clientset.Interface) error {
	definitionRequest := convertToADRequest(r)
	result, err := monascaclient.PatchAlarmDefinition(r.Spec.ID, definitionRequest)
	if err != nil {
		return err
	}
	newDef := convertToADResource(r, result)
	_, err = sampleclientset.Monasca().AlarmDefinitions(*namespace).Update(newDef)
	if err != nil {
		log.Printf("Failed to update resource %s", newDef.Name)
		return err
	}
	alarmDefinitionCache[result.ID] = *result
	log.Printf("Updated definition %v", r.Spec)
	return nil
}

func removeAlarmDefinition(id string) error {
	err := monascaclient.DeleteAlarmDefinition(id)
	if err != nil {
		// if 404 is returned, assume definition is already gone
		if !strings.HasPrefix(err.Error(), "Error: 404") {
			return err
		}
	}
	delete(alarmDefinitionCache, id)
	log.Printf("Removed definition %v", id)
	return nil
}

func convertToADRequest(r *v1.AlarmDefinition) *models.AlarmDefinitionRequestBody {
	if !strings.HasSuffix(r.Spec.Name, alarmDefinitionControllerSuffix) {
		r.Spec.Name = r.Spec.Name + alarmDefinitionControllerSuffix
	}
	definition := r.Spec

	request := &models.AlarmDefinitionRequestBody{
		Name:        &definition.Name,
		Description: &definition.Description,
		Expression:  &definition.Expression,
	}
	if len(definition.MatchBy) > 0 {
		request.MatchBy = &definition.MatchBy
	}
	if definition.Severity != "" {
		request.Severity = &definition.Severity
	}
	if len(definition.AlarmActions) > 0 {
		request.AlarmActions = &definition.AlarmActions
	}
	if len(definition.OkActions) > 0 {
		request.OkActions = &definition.OkActions
	}
	if len(definition.UndeterminedActions) > 0 {
		request.UndeterminedActions = &definition.UndeterminedActions
	}
	return request
}

func convertToADResource(r *v1.AlarmDefinition, d *models.AlarmDefinitionElement) *v1.AlarmDefinition {
	newDef := r.DeepCopy()
	newDef.Spec = v1.AlarmDefinitionSpec{
		ID:                  d.ID,
		Name:                d.Name,
		Description:         d.Description,
		Expression:          d.Expression,
		Deterministic:       d.Deterministic,
		Severity:            d.Severity,
		MatchBy:             d.MatchBy,
		AlarmActions:        d.AlarmActions,
		OkActions:           d.OkActions,
		UndeterminedActions: d.UndeterminedActions,
	}
	return newDef
}

func equal(r *v1.AlarmDefinition, existing models.AlarmDefinitionElement) bool {
	if r.Spec.ID == "" {
		return false
	}
	if r.Spec.ID != existing.ID {
		return false
	}
	if r.Spec.Name != existing.Name {
		return false
	}
	if r.Spec.Description != existing.Description {
		return false
	}
	if r.Spec.Expression != existing.Expression {
		return false
	}
	if r.Spec.Deterministic != existing.Deterministic {
		return false
	}
	if !equalStringList(r.Spec.MatchBy, existing.MatchBy) {
		return false
	}
	if r.Spec.Severity != existing.Severity {
		return false
	}
	if !equalStringList(r.Spec.AlarmActions, existing.AlarmActions) {
		return false
	}
	if !equalStringList(r.Spec.OkActions, existing.OkActions) {
		return false
	}
	if !equalStringList(r.Spec.UndeterminedActions, existing.UndeterminedActions) {
		return false
	}
	return true
}

func equalStringList(listA []string, listB []string) bool {
	if len(listA) != len(listB) {
		return false
	}
outer:
	for _, itemA := range listA {
		for _, itemB := range listB {
			if itemA == itemB {
				continue outer
			}
		}
		return false
	}
	return true
}

func getEnvDefault(name, def string) string {
	val := os.Getenv(name)
	if val == "" {
		val = def
	}
	return val
}
