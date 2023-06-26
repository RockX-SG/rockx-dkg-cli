/*
 * ==================================================================
 *Copyright (C) 2022-2023 Altstake Technology Pte. Ltd. (RockX)
 *This file is part of rockx-dkg-cli <https://github.com/RockX-SG/rockx-dkg-cli>
 *CAUTION: THESE CODES HAVE NOT BEEN AUDITED
 *
 *rockx-dkg-cli is free software: you can redistribute it and/or modify
 *it under the terms of the GNU General Public License as published by
 *the Free Software Foundation, either version 3 of the License, or
 *(at your option) any later version.
 *
 *rockx-dkg-cli is distributed in the hope that it will be useful,
 *but WITHOUT ANY WARRANTY; without even the implied warranty of
 *MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *GNU General Public License for more details.
 *
 *You should have received a copy of the GNU General Public License
 *along with rockx-dkg-cli. If not, see <http://www.gnu.org/licenses/>.
 *==================================================================
 */

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
