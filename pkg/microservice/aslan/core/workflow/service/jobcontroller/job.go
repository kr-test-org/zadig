/*
Copyright 2022 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jobcontroller

import (
	"context"
	"sync"
	"time"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
	commonmodels "github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/models"
	"github.com/koderover/zadig/pkg/setting"
	"go.uber.org/zap"
)

type JobCtl interface {
	Run(ctx context.Context)
}

func runJob(ctx context.Context, job *commonmodels.JobTask, workflowCtx *commonmodels.WorkflowTaskCtx, logger *zap.SugaredLogger, ack func()) {
	job.Status = config.StatusRunning
	job.StartTime = time.Now().Unix()
	// set default timeout
	if job.Properties.Timeout <= 0 {
		job.Properties.Timeout = 600
	}
	// set default resource
	if job.Properties.ResourceRequest == setting.Request("") {
		job.Properties.ResourceRequest = setting.MinRequest
	}
	// set default resource
	if job.Properties.ClusterID == "" {
		job.Properties.ClusterID = setting.LocalClusterID
	}
	logger.Infof("start job: %s,status: %s", job.Name, job.Status)
	defer func() {
		job.EndTime = time.Now().Unix()
		logger.Infof("finish job: %s,status: %s", job.Name, job.Status)
		ack()
	}()
	var jobCtl JobCtl
	switch job.JobType {
	case "deploy":
		// TODO
	default:
		jobCtl = NewFreestyleJobCtl(job, workflowCtx, ack, logger)
	}
	jobCtl.Run(ctx)
}

func RunJobs(ctx context.Context, jobs []*commonmodels.JobTask, workflowCtx *commonmodels.WorkflowTaskCtx, concurrency int, logger *zap.SugaredLogger, ack func()) {
	jobPool := NewPool(ctx, jobs, workflowCtx, concurrency, logger, ack)
	jobPool.Run()
}

// Pool is a worker group that runs a number of tasks at a
// configured concurrency.
type Pool struct {
	Jobs        []*commonmodels.JobTask
	workflowCtx *commonmodels.WorkflowTaskCtx
	concurrency int
	jobsChan    chan *commonmodels.JobTask
	logger      *zap.SugaredLogger
	ack         func()
	ctx         context.Context
	wg          sync.WaitGroup
}

// NewPool initializes a new pool with the given tasks and
// at the given concurrency.
func NewPool(ctx context.Context, jobs []*commonmodels.JobTask, workflowCtx *commonmodels.WorkflowTaskCtx, concurrency int, logger *zap.SugaredLogger, ack func()) *Pool {
	return &Pool{
		Jobs:        jobs,
		concurrency: concurrency,
		workflowCtx: workflowCtx,
		jobsChan:    make(chan *commonmodels.JobTask),
		logger:      logger,
		ack:         ack,
		ctx:         ctx,
	}
}

// Run runs all job within the pool and blocks until it's
// finished.
func (p *Pool) Run() {
	for i := 0; i < p.concurrency; i++ {
		go p.work()
	}

	p.wg.Add(len(p.Jobs))
	for _, task := range p.Jobs {
		p.jobsChan <- task
	}

	// all workers return
	close(p.jobsChan)

	p.wg.Wait()
}

// The work loop for any single goroutine.
func (p *Pool) work() {
	for job := range p.jobsChan {
		runJob(p.ctx, job, p.workflowCtx, p.logger, p.ack)
		p.wg.Done()
	}
}