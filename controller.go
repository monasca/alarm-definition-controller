// Copyright 2016 Google Inc.
// (C) Copyright 2017 Hewlett Packard Enterprise Development LP
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command simple-crontab-controller implements a crontab controller that
// watches for CronTab third party resources and runs a cron control
// loop for each. If the crontab is modified, then the cron loop is
// restarted with the new configuration. If it is deleted then the cron
// loop is stopped.
package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/openstack"
	"github.com/monasca/golang-monascaclient/monascaclient"
	"github.com/monasca/golang-monascaclient/monascaclient/models"
)

// TODO: support for multiple namespaces
// TODO: check into publishing events instead of patching the original resource
const (
	// A path to the endpoint for the AlarmDefinition custom resources.
	alarmDefinitionsEndpoint = "https://%s:%s/apis/monasca.io/%s/namespaces/%s/alarmdefinitions"
)

const alarmDefinitionControllerSuffix = " - adc"

var (
	pollInterval = flag.Int("poll-interval", 15, "The polling interval in seconds.")
	// The controller connects to the Kubernetes API via localhost. This is either
	// a locally running kubectl proxy or kubectl proxy running in a sidecar container.
	kubeServer          = flag.String("server", getEnvDefault("KUBERNETES_SERVICE_HOST", "127.0.0.1"), "The address of the Kubernetes API server.")
	kubePort            = flag.String("port", getEnvDefault("KUBERNETES_SERVICE_PORT_HTTPS", "443"), "The port of the Kubernetes API server")
	monServer           = flag.String("monasca", "http://monasca-api:8070/v2.0", "The URI of the monasca api")
	namespace           = flag.String("namespace", getEnvDefault("NAMESPACE", "default"), "The namespace to use.")
	version             = flag.String("version", getEnvDefault("VERSION", "v1"), "Version of alarm definition resource")
	defaultNotification = flag.String("default-notification", getEnvDefault("DEFAULT_NOTIFICATION", ""), "A default notification method to apply to new definitions")

	token      string
	httpClient *http.Client
	//cache to avoid repeated calls to monasca
	alarmDefinitionCache = map[string]models.AlarmDefinitionElement{}

	defaultNotificationID string
)

type kubeResponse struct {
	Kind       string
	Items      []Resource
	Metadata   MetaData
	APIVersion string
}

type MetaData struct {
	Name              string
	Namespace         string
	SelfLink          string
	UID               string
	ResourceVersion   string
	CreationTimestamp string
	Annotations       map[string]string
}

type Resource struct {
	Spec       alarmDefinitionResource `json:"alarmDefinitionSpec"`
	ApiVersion string
	Kind       string
	MetaData   MetaData
}

type alarmDefinitionResource struct {
	models.AlarmDefinitionElement
	Error string
}

func init() {
	token_byte, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		os.Exit(1)
	}
	token = string(token_byte)

	certs := x509.NewCertPool()

	pemData, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		// do error
	}
	certs.AppendCertsFromPEM(pemData)

	tlsConf := &tls.Config{
		RootCAs: certs,
	}

	transport := &http.Transport{TLSClientConfig: tlsConf}

	httpClient = &http.Client{
		Transport: transport,
	}
}

func getEnvDefault(name, def string) string {
	val := os.Getenv(name)
	if val == "" {
		val = def
	}
	return val
}

func equal(a1, a2 models.AlarmDefinitionElement) bool {
	if a1.Name != a2.Name {
		return false
	}
	if a1.Description != a2.Description {
		return false
	}
	if a1.Expression != a2.Expression {
		return false
	}
	if a1.Deterministic != a2.Deterministic {
		return false
	}
	if !equalStringList(a1.MatchBy, a2.MatchBy) {
		return false
	}
	if a1.Severity != a2.Severity {
		return false
	}
	if !equalStringList(a1.AlarmActions, a2.AlarmActions) {
		return false
	}
	if !equalStringList(a1.OkActions, a2.OkActions) {
		return false
	}
	if !equalStringList(a1.UndeterminedActions, a2.UndeterminedActions) {
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

func setKeystoneToken() error {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		log.Print(err)
		return err
	}

	openstackProvider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		log.Print(err)
		return err
	}
	//fmt.Println(openstackProvider.TokenID)
	token := openstackProvider.TokenID
	headers := http.Header{}
	headers.Add("X-Auth-Token", token)
	monascaclient.SetHeaders(headers)
	return nil
}

func updateCache() error {
	existing, err := monascaclient.GetAlarmDefinitions(nil)
	if err != nil {
		log.Print(err)
		return err
	}
	for _, item := range existing.Elements {
		//ignore all alarm definitions that do not have the adc suffix
		if strings.HasSuffix(item.Name, alarmDefinitionControllerSuffix) {
			alarmDefinitionCache[item.ID] = item
		}
	}

	return nil
}

func convertToADRequest(definition models.AlarmDefinitionElement) *models.AlarmDefinitionRequestBody {

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

func addAlarmDefinition(r Resource) error {
	if !strings.HasSuffix(r.Spec.Name, alarmDefinitionControllerSuffix) {
		r.Spec.Name = r.Spec.Name + alarmDefinitionControllerSuffix
	}
	if *defaultNotification != "" && len(r.Spec.AlarmActions) <= 0 {
		if defaultNotificationID == "" {
			return errors.New("Unable to apply default notification method: no ID found")
		}
		r.Spec.AlarmActions = []string{defaultNotificationID}
	}
	definitionRequest := convertToADRequest(r.Spec.AlarmDefinitionElement)
	result, err := monascaclient.CreateAlarmDefinition(definitionRequest)
	if err != nil {
		return err
	}
	applyDefinition(r, *result)

	alarmDefinitionCache[result.ID] = *result

	log.Printf("Added definition %v", r.Spec)
	return nil
}

func updateAlarmDefinition(id string, r Resource) error {
	definitionRequest := convertToADRequest(r.Spec.AlarmDefinitionElement)
	result, err := monascaclient.PatchAlarmDefinition(id, definitionRequest)
	if err != nil {
		return err
	}
	alarmDefinitionCache[id] = *result
	log.Printf("Updated definition %v", r.Spec)
	return nil
}

func removeAlarmDefinition(id string, definition models.AlarmDefinitionElement) error {
	err := monascaclient.DeleteAlarmDefinition(id)
	if err != nil {
		// if 404 is returned, assume definition is already gone
		if !strings.HasPrefix(err.Error(), "Error: 404") {
			return err
		}
	}
	delete(alarmDefinitionCache, id)
	log.Printf("Removed definition %v", definition)
	return nil
}

func applyDefinition(adr Resource, definition models.AlarmDefinitionElement) error {
	if adr.Spec.ID != "" {
		return errors.New("Cannot replace existing ID")
	}

	specPatch := map[string]models.AlarmDefinitionElement{}
	specPatch["alarmDefinitionSpec"] = definition

	err := patchResource(adr, specPatch)
	if err != nil {
		log.Print(err)
		return err
	}

	log.Printf("Applied ID to alarm definition %s", adr.Spec.ID)

	return nil
}

func applyError(adr Resource, alarmErr error) error {
	if adr.Spec.Error != "" {
		return errors.New("Not replacing existing error")
	}

	specPatch := map[string]string{"error": alarmErr.Error()}

	err := patchResource(adr, specPatch)
	if err != nil {
		log.Print(err)
		return err
	}

	log.Printf("Applied error on alarm definition %s", adr.Spec.Name)

	return nil
}

func patchResource(adr Resource, specPatch interface{}) error {
	url := fmt.Sprintf("https://%s:%s%s", *kubeServer, *kubePort, adr.MetaData.SelfLink)

	jsonStr, err := json.Marshal(specPatch)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Content-Type", "application/merge-patch+json")

	resp, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(resp.Status + string(data))
	}
	return nil
}

func pollDefinitions() {
	certs := x509.NewCertPool()

	pemData, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		// do error
	}
	certs.AppendCertsFromPEM(pemData)

	tlsConf := &tls.Config{
		RootCAs: certs,
	}

	transport := &http.Transport{TLSClientConfig: tlsConf}

	client := &http.Client{
		Transport: transport,
	}

	url := fmt.Sprintf(alarmDefinitionsEndpoint, *kubeServer, *kubePort, *version, *namespace)

	monascaclient.SetBaseURL(*monServer)
	setKeystoneToken()
	err = updateCache()
	if err != nil {
		log.Fatalf("Unable to update cache from monasca: %v", err)
	}
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
		// even after errors occurred.
		if !first {
			time.Sleep(time.Duration(*pollInterval) * time.Second)
		}
		first = false

		err := setKeystoneToken()
		if err != nil {
			log.Printf("Failed to retrieve new keystone token: %s", err.Error())
			continue
		}

		request, err := http.NewRequest("GET", url, nil)
		request.Header.Add("Authorization", "Bearer "+token)

		resp, err := client.Do(request)
		if err != nil {
			log.Printf("Could not connect to Kubernetes API: %v", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("Unexpected status from Kubernetes: %s", resp.Status)
			continue
		}

		respBytes, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			log.Print(err2)
		}

		var decodedResp kubeResponse
		err = json.Unmarshal(respBytes, &decodedResp)
		if err != nil {
			log.Printf("Could not decode JSON event object: %v", err)
			continue
		}

		l := decodedResp.Items

		// loop to remove alarms
		for id, cached := range alarmDefinitionCache {
			exists := false
			for _, discovered := range l {
				// check for equality
				if cached.ID == discovered.Spec.ID {
					exists = true
				}
			}
			if !exists {
				// remove definitions from monasca
				err := removeAlarmDefinition(id, cached)
				if err != nil {
					log.Print(err)
					continue
				}
			}
		}

	discoveredLoop:
		for _, item := range l { // loop to add/update alarms
			discovered := item.Spec.AlarmDefinitionElement

			// if not marked with ID, add new
			if item.Spec.ID == "" {
				err := addAlarmDefinition(item)
				if err != nil {
					// If 409 is returned, we probably had a desync between cache
					// and monasca. This can happen if monasca returned an error, but
					// still created the definition.
					if strings.HasPrefix(err.Error(), "Error: 409") {
						updateCache()
					}
					log.Print(err)
					applyError(item, err)
				}
				continue
			}

			for id, cached := range alarmDefinitionCache {
				// if exists, check if needs update
				if discovered.ID == id && !equal(discovered, cached) {
					//update if possible
					err := updateAlarmDefinition(id, item)
					if err != nil {
						log.Print(err)
						applyError(item, err)
					}
					continue discoveredLoop
				}
			}

		}
	}
}

func main() {
	flag.Parse()

	log.Print("Watching for definition objects...")

	pollDefinitions()
}
