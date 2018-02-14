package cmds

import (
	"io"

	"github.com/kubepack/packserver/pkg/cmds/server"
	"github.com/spf13/cobra"
)

func NewCmdRun(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := server.NewKubepackServerOptions(out, errOut)

	cmd := &cobra.Command{
		Short: "Launch a Kubepack API server",
		Long:  "Launch a Kubepack API server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunWardleServer(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	o.RecommendedOptions.AddFlags(flags)
	o.Admission.AddFlags(flags)

	return cmd
}
