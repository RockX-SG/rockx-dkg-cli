package workers

import (
	"context"
)

type Job struct {
	ID string
	Fn func(*context.Context)
}

type Runner struct {
	incomingJobs chan *Job
	jobs         map[string]context.CancelFunc
}

func NewRunner() *Runner {
	return &Runner{
		incomingJobs: make(chan *Job, 10),
		jobs:         make(map[string]context.CancelFunc),
	}
}

func (r *Runner) AddJob(j *Job) {
	r.incomingJobs <- j
}

func (r *Runner) Run() {
	for job := range r.incomingJobs {
		ctx, cancel := context.WithCancel(context.Background())
		r.jobs[job.ID] = cancel
		go job.Fn(&ctx)
	}
}

func (r *Runner) Cancel(id string) {
	r.jobs[id]()
}
