package alert

import (
	"bytes"
	_ "embed" // nolint
	"fmt"
	"text/template"

	meta1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	k8Yaml "k8s.io/apimachinery/pkg/util/yaml"

	operatorapiv1 "github.com/rancher/compliance-operator/pkg/apis/compliance.cattle.io/v1"
	"github.com/rancher/wrangler/v3/pkg/name"
)

//go:embed templates/prometheusrule.template
var prometheusRuleTemplate string

const templateName = "prometheusrule.template"

func NewPrometheusRule(clusterscan *operatorapiv1.ClusterScan, clusterscanprofile *operatorapiv1.ClusterScanProfile, imageConfig *operatorapiv1.ScanImageConfig) (*monitoringv1.PrometheusRule, error) {
	configdata := map[string]interface{}{
		"namespace":       operatorapiv1.ClusterScanNS,
		"name":            name.SafeConcatName("rancher-compliance-alerts", clusterscan.Name),
		"severity":        imageConfig.AlertSeverity,
		"scanName":        clusterscan.Name,
		"scanProfileName": clusterscanprofile.Name,
		"alertOnFailure":  clusterscan.Spec.ScheduledScanConfig.ScanAlertRule.AlertOnFailure,
		"alertOnComplete": clusterscan.Spec.ScheduledScanConfig.ScanAlertRule.AlertOnComplete,
		"failOnWarn":      clusterscan.Spec.ScoreWarning == operatorapiv1.ClusterScanFailOnWarning,
	}
	scanAlertRule, err := generatePrometheusRule(clusterscan, configdata)
	if err != nil {
		return scanAlertRule, err
	}

	return scanAlertRule, nil
}

func generatePrometheusRule(clusterscan *operatorapiv1.ClusterScan, data map[string]interface{}) (*monitoringv1.PrometheusRule, error) {
	scanAlertRule := &monitoringv1.PrometheusRule{}
	obj, err := parseTemplate(clusterscan, data)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the template %w", err)
	}

	if err := obj.Decode(&scanAlertRule); err != nil {
		return nil, fmt.Errorf("Error decoding to template %w", err)
	}

	ownerRef := meta1.OwnerReference{
		APIVersion: "compliance.cattle.io/v1",
		Kind:       "ClusterScan",
		Name:       clusterscan.Name,
		UID:        clusterscan.GetUID(),
	}
	scanAlertRule.ObjectMeta.OwnerReferences = append(scanAlertRule.ObjectMeta.OwnerReferences, ownerRef)

	return scanAlertRule, nil
}

func parseTemplate(_ *operatorapiv1.ClusterScan, data map[string]interface{}) (*k8Yaml.YAMLOrJSONDecoder, error) {
	cmTemplate, err := template.New(templateName).Parse(prometheusRuleTemplate)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	err = cmTemplate.Execute(&b, data)
	if err != nil {
		return nil, err
	}

	return k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(b.String())), 1000), nil
}
