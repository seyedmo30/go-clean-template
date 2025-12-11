/*
Package cmd provides the CLI interface for the open-banking service.
*/
package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"__MODULE__/internal/adapter/http"
	"__MODULE__/internal/client/integration"
	"__MODULE__/internal/repository"
	"__MODULE__/internal/usecase"
	"__MODULE__/internal/worker"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// serveCmd is the command that starts the kafka service and
// dependencies like mysql, providers, etc.
// Graceful shutdown on SIGTERM and SIGINT is also handled.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start the server",
	Run: func(_ *cobra.Command, _ []string) {
		log.Info("starting the server")
		rp := repository.NewServiceRepository(conf)
		// dbConn, err := repository.GetDB()
		// if err != nil {
		// 	log.Error("failed to connect to the database: " + err.Error())
		// 	os.Exit(1)
		// }
		// settingService, err := config.NewSettingsService(dbConn, conf.SettingsTTL)
		// if err != nil {
		// 	log.Error("failed to create settings service: " + err.Error())
		// 	os.Exit(1)
		// }
		// fmt.Println(settingService)
		pr := integration.NewUserProviderService(conf)

		err := pr.RegisterNewProvider(integration.JsonPlaceholderProvider, integration.JsonPlaceholderProvider, "")
		if err != nil {
			log.Error("failed to get RegisterNewProvider: " + err.Error())
			os.Exit(1)
		}
		userSvc, err := pr.GetUserService(integration.JsonPlaceholderProvider)
		if err != nil {
			log.Error("failed to get user provider instance: " + err.Error())
			os.Exit(1)
		}

		userUsecase := usecase.NewUserUsecase(rp, userSvc, 50)

		worker.NewWorker(userUsecase, conf.WorkerConfig).Start()

		// echo server
		e := echo.New()
		// setup validator and routes
		http.SetupValidator(e)
		http.RegisterUserRoutes(e, &userUsecase)
		http.RegisterSwagger(e)

		// run echo in a goroutine so we can block on signals
		serverErrCh := make(chan error, 1)
		go func() {
			addr := ":" + "8009"
			log.Info("http server starting", "addr", addr)
			if err := e.Start(addr); err != nil && err != echo.ErrInternalServerError {
				serverErrCh <- err
			}
			serverErrCh <- nil
		}()

		log.Info("shutdown complete")

		var gracefulStop = make(chan os.Signal, 1)
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		sig := <-gracefulStop
		log.Info("shutdown commencing...", "signal", sig)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
