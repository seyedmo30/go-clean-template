package worker

import (
	"context"
	"errors"
	"net/http"

	"__MODULE__/internal/config"
	"__MODULE__/internal/interfaces"
	"__MODULE__/pkg"

	"github.com/labstack/gommon/log"
	"github.com/robfig/cron/v3"
)

type worker struct {
	usecase interfaces.BackgroundJobUsecase
	conf    config.WorkerConfig
	cron    *cron.Cron
}

type jobFunc func(context.Context) error

func NewWorker(usecase interfaces.BackgroundJobUsecase, conf config.WorkerConfig) *worker {
	return &worker{
		usecase: usecase,
		conf:    conf,
		cron:    cron.New(),
	}
}

func (w *worker) Start() {
	log.Info("Starting worker...")

	// Register jobs with cron schedules
	// w.registerJobs(ctx, w.conf.ExpirePendingEndOfDay, w.usecase.ExpireEndOfDayPendingTransactions, true)

	// Start the cron scheduler
	w.cron.Start()
}

func (w *worker) registerJobs(ctx context.Context, schedule string, jobFunc jobFunc, dependent bool) {

	// Check if the job is dependent on the previous jobs being completed
	if dependent {
		var isRunning bool // Flag to track job execution state
		_, err := w.cron.AddFunc(schedule, func() {
			if isRunning {
				log.Warn("Skipping job execution as the previous job is still running")
				return
			}

			isRunning = true
			defer func() {
				isRunning = false
				if r := recover(); r != nil {
					log.Error("Recovered from panic in job", "error", r)
				}
			}()

			if err := jobFunc(ctx); err != nil {
				var errorWithCode *pkg.AppError
				if errors.As(err, &errorWithCode) && errorWithCode.InternalCode() != http.StatusNotFound {
					log.Error("Error executing task", "error", err.Error())
				}
			}
		})
		if err != nil {
			log.Error("Failed to register dependent job", "error", err.Error())
		}
	} else {
		_, err := w.cron.AddFunc(schedule, func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("Recovered from panic in job", "error", r)
				}
			}()
			if err := jobFunc(ctx); err != nil {
				var errorWithCode *pkg.AppError
				if errors.As(err, &errorWithCode) && errorWithCode.InternalCode() != http.StatusNotFound {
					log.Error("Error executing task", "error", err.Error())
				}
			}
		})
		if err != nil {
			log.Error("Failed to register independent job", "error", err.Error())
		}
	}
}
