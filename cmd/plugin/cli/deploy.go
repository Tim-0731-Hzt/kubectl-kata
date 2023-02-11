package cli

import (
	"github.com/Tim-0731-Hzt/kubectl-kata/pkg/plugin"
	"github.com/spf13/cobra"
)

func init() {
	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "deploy kata-containers on each node",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := plugin.NewDeployService()
			if err := d.Complete(cmd, args); err != nil {
				return err
			}
			if err := d.Validate(); err != nil {
				return err
			}
			if err := d.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.AddCommand(deployCmd)
}
