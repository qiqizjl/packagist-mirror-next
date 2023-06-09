package cmd

import (
	"github.com/spf13/cobra"
	"packagist-mirror-next/internal/clear"
	"packagist-mirror-next/internal/svc"
)

// httpServerCmd represents the httpServer command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear",
	Run: func(cmd *cobra.Command, args []string) {
		svcCtx, err := svc.NewServiceContext("clear")
		if err != nil {
			panic(err)
		}
		//gin.SetMode(gin.ReleaseMode)
		ctx := cmd.Context()
		clearProvider := clear.NewClearProvider(svcCtx, &ctx)
		if err := clearProvider.Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
