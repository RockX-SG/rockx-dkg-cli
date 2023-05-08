package workers

import (
	"context"

	"github.com/RockX-SG/frost-dkg-demo/internal/logger"
)

type Ctxlog string

type Job struct {
	ID string
	Fn func(*context.Context)
}

type Runner struct {
	incomingJobs chan *Job
	jobs         map[string]context.CancelFunc

	logger *logger.Logger
}

func NewRunner(logger *logger.Logger) *Runner {
	return &Runner{
		incomingJobs: make(chan *Job, 10),
		jobs:         make(map[string]context.CancelFunc),
		logger:       logger,
	}
}

func (r *Runner) AddJob(j *Job) {
	r.incomingJobs <- j
}

func (r *Runner) Run() {
	for job := range r.incomingJobs {
		ctxlog := context.WithValue(context.Background(), Ctxlog("logger"), r.logger)
		ctx, cancel := context.WithCancel(ctxlog)
		r.jobs[job.ID] = cancel
		go job.Fn(&ctx)
	}
}

func (r *Runner) Cancel(id string) {
	r.jobs[id]()
}
