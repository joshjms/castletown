package job

import "sync"

var jp *JobPool

type JobPool struct {
	Jobs map[string]*Job

	mu sync.Mutex
}

func NewJobPool() {
	jp = &JobPool{
		Jobs: make(map[string]*Job),
	}
}

func GetJobPool() *JobPool {
	return jp
}

func (jp *JobPool) GetJob(id string) (*Job, bool) {
	jp.mu.Lock()
	defer jp.mu.Unlock()

	job, exists := jp.Jobs[id]
	return job, exists
}

func (jp *JobPool) AddOrAppendJob(job *Job) *Job {
	jp.mu.Lock()
	defer jp.mu.Unlock()

	if existingJob, exists := jp.Jobs[job.ID]; exists {
		existingJob.append(job)
	} else {
		jp.Jobs[job.ID] = job
	}

	return jp.Jobs[job.ID]
}

func (jp *JobPool) RemoveJob(id string) {
	jp.mu.Lock()
	defer jp.mu.Unlock()

	delete(jp.Jobs, id)
}

func (j *Job) append(other *Job) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.Files = append(j.Files, other.Files...)
	j.Procs = append(j.Procs, other.Procs...)
}
