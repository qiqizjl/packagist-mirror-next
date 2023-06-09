package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"packagist-mirror-next/internal/process"
	"packagist-mirror-next/internal/svc"
	"syscall"
)

// processCmd represents the httpServer command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process For Mirroring",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		go processPprof()
		svcCtx, err := svc.NewServiceContext("process")
		if err != nil {
			panic(err)
		}
		ctx := cmd.Context()
		if err := process.InitProcess(svcCtx, &ctx); err != nil {
			panic(err)
		}
		if err := process.StartProcess(); err != nil {
			panic(err)
		}

		//  等待进程结束
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		s := <-c
		fmt.Printf("Got signal: %s\r\n", s.String())
		if err := process.StopProcess(); err != nil {
			panic(err)
		}
		svcCtx.NSQ.Stop()
		//time.Sleep(30 * time.Second)
	},
}

func init() {
	rootCmd.AddCommand(processCmd)
}

func processPprof() {
	http.ListenAndServe(":18080", nil)
}
