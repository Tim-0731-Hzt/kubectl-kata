package cli

import (
	"fmt"
	"github.com/Tim-0731-Hzt/kubectl-kata/pkg/plugin"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

func init() {
	e := plugin.NewExecService()
	var execCmd = &cobra.Command{
		Use: "exec",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("exec called")
			err := e.Complete(cmd, args)
			if err != nil {
				return err
			}
			err = e.Validate()
			if err != nil {
				return err
			}
			err = e.Run()
			if err != nil {
				return err
			}
			return nil
		},
	}
	execCmd.Flags().StringVarP(&e.UserSpecifiedNamespace, "namespace", "n", "", "namespace (optional)")
	_ = viper.BindEnv("namespace", "KUBECTL_PLUGINS_CURRENT_NAMESPACE")
	_ = viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))

	execCmd.Flags().StringVarP(&e.UserSpecifiedPodName, "pod", "p", "", "pod (optional)")
	_ = viper.BindEnv("pod", "KUBECTL_PLUGINS_LOCAL_FLAG_POD")
	_ = viper.BindPFlag("pod", cmd.Flags().Lookup("pod"))

	cmd.AddCommand(execCmd)
}
