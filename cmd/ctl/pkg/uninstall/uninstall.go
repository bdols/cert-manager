/*
Copyright 2022 The cert-manager Authors.

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

package uninstall

import (
	"context"
	"errors"
	"fmt"
	"helm.sh/helm/v3/pkg/storage/driver"
	"log"
	"os"
	"time"

	"github.com/cert-manager/cert-manager/cmd/ctl/pkg/build"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type options struct {
	settings *cli.EnvSettings
	client   *action.Uninstall
	cfg      *action.Configuration

	DisableHooks bool
	DryRun       bool
	Wait         bool

	genericclioptions.IOStreams
}

const (
	defaultCertManagerNamespace = "cert-manager"
	releaseName                 = "cert-manager"
)

func description() string {
	return build.WithTemplate(`This command uninstalls cert-manager. It uses the Helm libraries to do so.

It assumes cert-manager has been installed with this CLI.

Most of the features supported by 'helm uninstall' are also supported by this command.

Some example uses:
	$ {{.BuildName}} x uninstall
or
	$ {{.BuildName}} x uninstall --namespace my-cert-manager
or
	$ {{.BuildName}} x uninstall --dry-run
or
	$ {{.BuildName}} x uninstall --no-hooks
`)
}

func NewCmd(ctx context.Context, ioStreams genericclioptions.IOStreams) *cobra.Command {
	settings := cli.New()
	cfg := new(action.Configuration)

	options := options{
		settings: settings,
		cfg:      cfg,
		client:   action.NewUninstall(cfg),

		IOStreams: ioStreams,
	}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall cert-manager",
		Long:  description(),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := run(ctx, options)
			if err != nil {
				return fmt.Errorf("run: %v", err)
			}

			if options.DryRun {
				fmt.Fprintf(ioStreams.Out, "%s", res.Release.Manifest)
				return nil
			}

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	settings.AddFlags(cmd.Flags())

	// The Helm cli.New function does not provide an easy way to
	// override the default of the namespace flag.
	// See https://github.com/helm/helm/issues/9790
	//
	// set the default value shown in the usage message.
	cmd.Flag("namespace").DefValue = defaultCertManagerNamespace

	// The returned error is ignored because
	// pflag.stringValue.Set always returns a nil.
	cmd.Flag("namespace").Value.Set(defaultCertManagerNamespace)

	cmd.Flags().DurationVar(&options.client.Timeout, "timeout", 5*time.Minute, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	cmd.Flags().BoolVar(&options.Wait, "wait", true, "if set, will wait until all the resources are deleted before returning. It will wait for as long as --timeout")
	cmd.Flags().BoolVar(&options.DryRun, "dry-run", false, "simulate uninstall and output manifests to be deleted")
	cmd.Flags().BoolVar(&options.DisableHooks, "no-hooks", false, "prevent hooks from running during uninstallation (pre- and post-uninstall hooks)")

	return cmd
}

// run assumes cert-manager was installed as a Helm release named cert-manager.
// this is not configurable to avoid uninstalling non-cert-manager releases.
func run(ctx context.Context, o options) (*release.UninstallReleaseResponse, error) {
	log.SetFlags(0) // disable prefixing logs with timestamps.

	if err := o.cfg.Init(o.settings.RESTClientGetter(), o.settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, fmt.Errorf("o.cfg.Init: %v", err)
	}

	o.client.DisableHooks = o.DisableHooks
	o.client.DryRun = o.DryRun
	o.client.Wait = o.Wait

	res, err := o.client.Run(releaseName)

	if errors.Is(err, driver.ErrReleaseNotFound) {
		log.Fatalf("release %v not found in namespace %v, did you use the correct namespace?", releaseName, o.settings.Namespace())
	}

	return res, nil
}
