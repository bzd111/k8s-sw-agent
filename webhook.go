package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	admissionWebhookAnnotationInjectKey = "sw-injector-webhook/inject"
)

// var pod *corev1.Pod = &corev1.Pod{}

type podMutate struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (p *podMutate) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}
	podMutateLog := log.WithName("podMutate")

	err := p.decoder.Decode(req, pod)
	if err != nil {
		podMutateLog.Error(err, "failed decoder pod")
		return admission.Errored(http.StatusBadRequest, err)
	}

	if !needAdd(pod) {
		podMutateLog.Info("don't inject concepts")
		return admission.Allowed("ok")
	}

	addConcepts(pod)
	podMutateLog.Info("will inject some concepts")
	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		podMutateLog.Error(err, "failed marshal pod")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (p *podMutate) InjectDecoder(d *admission.Decoder) error {
	p.decoder = d
	return nil
}

func needAdd(pod *corev1.Pod) bool {
	var required bool
	annotations := pod.ObjectMeta.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}
	switch strings.ToLower(annotations[admissionWebhookAnnotationInjectKey]) {
	case "y", "yes", "true", "on":
		required = true
	default:
		required = false
	}
	return required
}

func addConcepts(pod *corev1.Pod) {
	// InitContainer
	vm := corev1.VolumeMount{
		MountPath: "/skywalking/agent",
		Name:      "sw-agent",
	}
	needAddInitContainer := corev1.Container{
		Name:         "sw-agent-sidecar",
		Image:        "innerpeacez/sw-agent-sidecar:latest",
		Command:      []string{"sh"},
		Args:         []string{"-c", "mkdir -p /skywalking/agent && cp -r /app/sw-agent/* /skywalking/agent"},
		VolumeMounts: []corev1.VolumeMount{vm},
	}

	// Volume
	needAddVolumes := corev1.Volume{
		Name:         "sw-agent",
		VolumeSource: corev1.VolumeSource{EmptyDir: nil},
	}

	// VolumeMount
	needAddVolumeMount := corev1.VolumeMount{
		MountPath: "/usr/skywalking/agent",
		Name:      "sw-agent",
	}
	if pod.Spec.Volumes != nil {
		pod.Spec.Volumes = append(pod.Spec.Volumes, needAddVolumes)
	} else {
		pod.Spec.Volumes = []corev1.Volume{needAddVolumes}
	}

	if pod.Spec.InitContainers != nil {
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, needAddInitContainer)
	} else {
		pod.Spec.InitContainers = []corev1.Container{needAddInitContainer}
	}

	for i := 0; i < len(pod.Spec.Containers); i++ {
		pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts, needAddVolumeMount)
	}
	// for _, container := range pod.Spec.Containers {
	//     if container.VolumeMounts != nil {
	//         container.VolumeMounts = append(container.VolumeMounts, needAddVolumeMount)
	//     } else {
	//         container.VolumeMounts = []corev1.VolumeMount{needAddVolumeMount}
	//     }
	// }
}
