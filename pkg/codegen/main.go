package main

import (
	"os"

	controllergen "github.com/rancher/wrangler/v3/pkg/controller-gen"
	"github.com/rancher/wrangler/v3/pkg/controller-gen/args"

	v1 "github.com/rancher/compliance-operator/pkg/apis/compliance.cattle.io/v1"
	"github.com/rancher/compliance-operator/pkg/crds"
)

func main() {
	os.Unsetenv("GOPATH")
	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher/compliance-operator/pkg/generated",
		Boilerplate:   "hack/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"compliance.cattle.io": {
				Types: []interface{}{
					v1.ClusterScan{},
					v1.ClusterScanProfile{},
					v1.ClusterScanReport{},
					v1.ClusterScanBenchmark{},
				},
				GenerateTypes: true,
			},
		},
	})

	err := crds.WriteCRD()
	if err != nil {
		panic(err)
	}
}
