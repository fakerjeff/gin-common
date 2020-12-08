package jobs

import (
	"github.com/robfig/cron/v3"
)

type CronJobHook interface {
	BeforeJobRun(job *Job) error
	AfterJobRun(job *Job) error
}

type hook struct {
	hook CronJobHook
}

type JobResult struct {
	status int
}

func (r JobResult) Status() int {
	return r.status
}

type EtcdJobHook struct {
}

func (EtcdJobHook) BeforeJobRun(job *Job) error {
	// Todo： key:/cron/jobName/jobPlanTime/uuid
	// Todo: 向ETCD建立租约 并自动续约 通过revision判断是否获取锁
	// Todo：如果获取到锁执行定时任务操作
	// Todo：判断ETCD中任务的执行状态; key: /cron/jobName/result,
	// Todo：如果已执行则跳过操作，如果未执行则继续执行操作
	// Todo：如果未取道锁则等待至锁释放后执行上述操作
	panic("implement me")
}

func (EtcdJobHook) AfterJobRun(job *Job) error {
	// Todo：更新任务执行状态，记录任务执行结果，与ETCD解除租约
	panic("implement me")
}

func (h hook) process(job *Job) error {
	var err error
	err = h.hook.BeforeJobRun(job)
	if err != nil {
		return err
	}
	err = job.inner.Run()
	if err != nil {
		return err
	}
	err = h.hook.BeforeJobRun(job)
	if err != nil {
		return err
	}
	return err
}

type CronJob interface {
	Spec() string
	Run() error
}

type Job struct {
	Name      string
	jobID     int
	singleton bool
	cron      *cron.Cron
	inner     CronJob
	logger    cron.Logger
	hook      *hook
}

func (j *Job) Init(name string, c *cron.Cron, inner CronJob) {
	j.Name = name
	j.cron = c
	j.inner = inner
	j.jobID = j.register()
}

func (j *Job) JobID() int {
	return j.jobID
}

func (j *Job) Entry(jobId int) cron.Entry {
	return j.cron.Entry(cron.EntryID(jobId))
}

func (j *Job) Valid(jobId int) bool {
	entry := j.Entry(jobId)
	return entry.Valid()
}

func (j *Job) funcJob() {
	var err error
	if j.hook != nil {
		err = j.hook.process(j)
		if err != nil {
			panic(err)
		}
	} else {
		err = j.inner.Run()
		if err != nil {
			panic(err)
		}
	}
}

func (j *Job) register() int {
	var job cron.Job
	if j.singleton {
		job = cron.NewChain(cron.DelayIfStillRunning(j.logger)).Then(cron.FuncJob(j.funcJob))
	} else {
		job = cron.FuncJob(j.funcJob)
	}
	id, err := j.cron.AddJob(j.inner.Spec(), job)
	if err != nil {
		panic(err)
	}
	return int(id)
}

func (j *Job) SetSingleton(singleton bool) {
	j.singleton = singleton
}
