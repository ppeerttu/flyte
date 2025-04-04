// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package config

import (
	"encoding/json"
	"reflect"

	"fmt"

	"github.com/spf13/pflag"
)

// If v is a pointer, it will get its element value or the zero value of the element type.
// If v is not a pointer, it will return it as is.
func (K8sPluginConfig) elemValueOrNil(v interface{}) interface{} {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		if reflect.ValueOf(v).IsNil() {
			return reflect.Zero(t.Elem()).Interface()
		} else {
			return reflect.ValueOf(v).Interface()
		}
	} else if v == nil {
		return reflect.Zero(t).Interface()
	}

	return v
}

func (K8sPluginConfig) mustJsonMarshal(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (K8sPluginConfig) mustMarshalJSON(v json.Marshaler) string {
	raw, err := v.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(raw)
}

// GetPFlagSet will return strongly types pflags for all fields in K8sPluginConfig and its nested types. The format of the
// flags is json-name.json-sub-name... etc.
func (cfg K8sPluginConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("K8sPluginConfig", pflag.ExitOnError)
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "inject-finalizer"), defaultK8sConfig.InjectFinalizer, "Instructs the plugin to inject a finalizer on startTask and remove it on task termination.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "default-cpus"), defaultK8sConfig.DefaultCPURequest.String(), "Defines a default value for cpu for containers if not specified.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "default-memory"), defaultK8sConfig.DefaultMemoryRequest.String(), "Defines a default value for memory for containers if not specified.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "scheduler-name"), defaultK8sConfig.SchedulerName, "Defines scheduler name.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "co-pilot.name"), defaultK8sConfig.CoPilot.NamePrefix, "Flyte co-pilot sidecar container name prefix. (additional bits will be added after this)")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "co-pilot.image"), defaultK8sConfig.CoPilot.Image, "Flyte co-pilot Docker Image FQN")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "co-pilot.default-input-path"), defaultK8sConfig.CoPilot.DefaultInputDataPath, "Default path where the volume should be mounted")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "co-pilot.default-output-path"), defaultK8sConfig.CoPilot.DefaultOutputPath, "Default path where the volume should be mounted")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "co-pilot.input-vol-name"), defaultK8sConfig.CoPilot.InputVolumeName, "Name of the data volume that is created for storing inputs")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "co-pilot.output-vol-name"), defaultK8sConfig.CoPilot.OutputVolumeName, "Name of the data volume that is created for storing outputs")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "co-pilot.cpu"), defaultK8sConfig.CoPilot.CPU, "Used to set cpu for co-pilot containers")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "co-pilot.memory"), defaultK8sConfig.CoPilot.Memory, "Used to set memory for co-pilot containers")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "co-pilot.storage"), defaultK8sConfig.CoPilot.Storage, "Default storage limit for individual inputs / outputs")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "delete-resource-on-finalize"), defaultK8sConfig.DeleteResourceOnFinalize, "Instructs the system to delete the resource upon successful execution of a k8s pod rather than have the k8s garbage collector clean it up. This ensures that no resources are kept around (potentially consuming cluster resources). This,  however,  will cause k8s log links to expire as soon as the resource is finalized.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "default-pod-template-name"), defaultK8sConfig.DefaultPodTemplateName, "Name of the PodTemplate to use as the base for all k8s pods created by FlytePropeller.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "default-pod-template-resync"), defaultK8sConfig.DefaultPodTemplateResync.String(), "Frequency of resyncing default pod templates")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "send-object-events"), defaultK8sConfig.SendObjectEvents, "If true,  will send k8s object events in TaskExecutionEvent updates.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "update-base-backoff-duration"), defaultK8sConfig.UpdateBaseBackoffDuration, "Initial delay in exponential backoff when updating a resource in milliseconds.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "update-backoff-retries"), defaultK8sConfig.UpdateBackoffRetries, "Number of retries for exponential backoff when updating a resource.")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "add-tolerations-for-extended-resources"), defaultK8sConfig.AddTolerationsForExtendedResources, "Name of the extended resources for which tolerations should be added.")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "enable-distributed-error-aggregation"), defaultK8sConfig.EnableDistributedErrorAggregation, "If true,  will aggregate errors of different worker pods for distributed tasks.")
	return cmdFlags
}
