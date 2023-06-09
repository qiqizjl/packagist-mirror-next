/*
Copyright © 2022 SeanWang
*/
package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"packagist-mirror-next/internal/http/router"
	"packagist-mirror-next/internal/svc"
)

// httpServerCmd represents the httpServer command
var httpServerCmd = &cobra.Command{
	Use:   "httpServer",
	Short: "Http Server",
	Run: func(cmd *cobra.Command, args []string) {
		svcCtx, err := svc.NewServiceContext("httpServer")
		if err != nil {
			panic(err)
		}
		//gin.SetMode(gin.ReleaseMode)
		g := gin.Default()
		g.Use(func(ctx *gin.Context) {
			ctx.Set("svcCtx", svcCtx)
		})
		g.GET("/ping", func(ctx *gin.Context) {
			ctx.String(200, "pong")
		})
		distRoute := router.NewDist(svcCtx)
		g.GET("/:owner/:repo/:version", distRoute.DistGet)
		statRoute := router.NewStat(svcCtx)
		g.GET("/stat", statRoute.APIStat)
		metadataRoute := router.NewMetadata(svcCtx)
		g.GET("/metadata/requeue-package", metadataRoute.RequeuePackage)
		g.Run(":8080")
	},
}

func init() {
	rootCmd.AddCommand(httpServerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// httpServerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// httpServerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
