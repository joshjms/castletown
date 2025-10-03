package job

type JobPool struct {
	Jobs map[string]Job
}

func NewJobPool() *JobPool {
	return &JobPool{
		Jobs: make(map[string]Job),
	}
}

func (jp *JobPool) AddOrAppendJob(job Job) {
	if existingJob, exists := jp.Jobs[job.ID]; exists {
		existingJob.Files = append(existingJob.Files, job.Files...)
		existingJob.Procs = append(existingJob.Procs, job.Procs...)
		jp.Jobs[job.ID] = existingJob
	} else {
		jp.Jobs[job.ID] = job
	}
}

func (jp *JobPool) RemoveJob(id string) {
	delete(jp.Jobs, id)
}
