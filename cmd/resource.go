// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/walk1ng/k8swatch/pkg/config"
)

// resourceCmd represents the resource command
var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "manage resources to be watched",
	Long: `
manage resources to be watched`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			logrus.Warn("Too few arguments to Command \"resource\".\nMinimum 2 arguments required: subcommand, resource flags")
		}
		cmd.Help()
	},
}

var resourceAddCmd = &cobra.Command{
	Use:   "add",
	Short: "adds the specific resources to be watched",
	Long: `
adds the specific resources to be watched`,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.New()
		if err != nil {
			logrus.Fatal(err)
		}
		manageResource("add", cmd, conf)
	},
}

var resourceRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "removes the specific resources to be watched",
	Long: `
removes the specific resources to be watched`,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.New()
		if err != nil {
			logrus.Fatal(err)
		}
		manageResource("remove", cmd, conf)
	},
}

// add or remove the resource to be watched
func manageResource(operation string, cmd *cobra.Command, config *config.Config) {

	var flags = []struct {
		ResourceStr string
		EnableWatch *bool
	}{
		{
			"po",
			&config.Resource.Pod,
		},
		{
			"deploy",
			&config.Resource.Service,
		},
		{
			"rc",
			&config.Resource.ReplicationController,
		},
		{
			"rs",
			&config.Resource.ReplicaSet,
		},
		{
			"ds",
			&config.Resource.DaemonSet,
		},
		{
			"svc",
			&config.Resource.Service,
		},
		{
			"job",
			&config.Resource.Job,
		},
		{
			"pv",
			&config.Resource.PersistentVolume,
		},
		{
			"ns",
			&config.Resource.Namespace,
		},
		{
			"secret",
			&config.Resource.Secret,
		},
		{
			"cm",
			&config.Resource.ConfigMap,
		},
		{
			"ing",
			&config.Resource.Ingress,
		},
	}

	for _, flag := range flags {
		b, err := cmd.Flags().GetBool(flag.ResourceStr)
		if err == nil {
			if b {
				switch operation {
				case "add":
					*flag.EnableWatch = true
					logrus.Infof("resource %s added", flag.ResourceStr)
				case "remove":
					*flag.EnableWatch = false
					logrus.Infof("resource %s removed", flag.ResourceStr)
				}
			}

		} else {
			logrus.Fatal(flag.ResourceStr, err)
		}

		if err := config.Write(); err != nil {
			logrus.Fatal(err)
		}
	}
}

func init() {
	rootCmd.AddCommand(resourceCmd)

	resourceCmd.AddCommand(
		resourceAddCmd,
		resourceRemoveCmd,
	)

	// resource flags as PersistentFlags to resourceCmd
	resourceCmd.PersistentFlags().Bool("po", false, "watch for Pods")
	resourceCmd.PersistentFlags().Bool("deploy", false, "watch for Deployments")
	resourceCmd.PersistentFlags().Bool("rc", false, "watch for ReplicationControllers")
	resourceCmd.PersistentFlags().Bool("rs", false, "watch for ReplicaSets")
	resourceCmd.PersistentFlags().Bool("ds", false, "watch for DaemonSets")
	resourceCmd.PersistentFlags().Bool("svc", false, "watch for Services")
	resourceCmd.PersistentFlags().Bool("job", false, "watch for Jobs")
	resourceCmd.PersistentFlags().Bool("pv", false, "watch for PersistentVolumes")
	resourceCmd.PersistentFlags().Bool("ns", false, "watch for Namespaces")
	resourceCmd.PersistentFlags().Bool("secret", false, "watch for Secrets")
	resourceCmd.PersistentFlags().Bool("cm", false, "watch for ConfigMaps")
	resourceCmd.PersistentFlags().Bool("ing", false, "watch for Ingresses")

}
