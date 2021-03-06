// Copyright © 2017 The Kubicorn Authors
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
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg"
	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/initapi"

	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
)

type OutputData struct {
	Actual   *cluster.Cluster
	Expected *cluster.Cluster
}

var exo = &cli.ExplainOptions{}

// ExplainCmd represents the explain command
func ExplainCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "explain",
		Short: "Explain cluster",
		Long:  `Output expected and actual state of the given cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				exo.Name = cli.StrEnvDef("KUBICORN_NAME", "")
			} else if len(args) > 1 {
				logger.Critical("Too many arguments.")
				os.Exit(1)
			} else {
				exo.Name = args[0]
			}

			err := RunExplain(exo)
			if err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&exo.StateStore, "state-store", "s", cli.StrEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	cmd.Flags().StringVarP(&exo.StateStorePath, "state-store-path", "S", cli.StrEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	cmd.Flags().StringVarP(&exo.Output, "output", "o", cli.StrEnvDef("KUBICORN_OUTPUT", "json"), "Output format (currently only JSON supported)")

	// git flags
	cmd.Flags().StringVar(&exo.GitRemote, "git-config", cli.StrEnvDef("KUBICORN_GIT_CONFIG", "git"), "The git remote url to use")

	// s3 flags
	cmd.Flags().StringVar(&exo.S3AccessKey, "s3-access", cli.StrEnvDef("KUBICORN_S3_ACCESS_KEY", ""), "The s3 access key.")
	cmd.Flags().StringVar(&exo.S3SecretKey, "s3-secret", cli.StrEnvDef("KUBICORN_S3_SECRET_KEY", ""), "The s3 secret key.")
	cmd.Flags().StringVar(&exo.BucketEndpointURL, "s3-endpoint", cli.StrEnvDef("KUBICORN_S3_ENDPOINT", ""), "The s3 endpoint url.")
	cmd.Flags().BoolVar(&exo.BucketSSL, "s3-ssl", cli.BoolEnvDef("KUBICORN_S3_SSL", true), "The s3 bucket name to be used for saving the git state for the cluster.")
	cmd.Flags().StringVar(&exo.BucketName, "s3-bucket", cli.StrEnvDef("KUBICORN_S3_BUCKET", ""), "The s3 bucket name to be used for saving the s3 state for the cluster.")

	return cmd
}

func RunExplain(options *cli.ExplainOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to apply")
	}

	// Expand state store path
	options.StateStorePath = cli.ExpandPath(options.StateStorePath)

	// Register state store
	stateStore, err := options.NewStateStore()
	if err != nil {
		return err
	} else if !stateStore.Exists() {
		return fmt.Errorf("State store [%s] does not exists, can't edit", name)
	}

	cluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("Unable to get cluster [%s]: %v", name, err)
	}

	cluster, err = initapi.InitCluster(cluster)
	if err != nil {
		return err
	}

	runtimeParams := &pkg.RuntimeParameters{}

	if len(ao.AwsProfile) > 0 {
		runtimeParams.AwsProfile = ao.AwsProfile
	}

	reconciler, err := pkg.GetReconciler(cluster, runtimeParams)
	if err != nil {
		return fmt.Errorf("Unable to get reconciler: %v", err)
	}

	var d OutputData
	d.Actual, err = reconciler.Actual(cluster)
	if err != nil {
		return fmt.Errorf("Unable to get actual cluster: %v", err)
	}
	d.Expected, err = reconciler.Expected(cluster)
	if err != nil {
		return fmt.Errorf("Unable to get expected cluster: %v", err)
	}

	if exo.Output == "json" {
		o, err := json.MarshalIndent(d, "", "\t")
		if err != nil {
			return fmt.Errorf("Unable to parse cluster: %v", err)
		}
		fmt.Printf("%s\n", o)
	} else {
		return fmt.Errorf("Unsupported output format")
	}

	return nil
}
