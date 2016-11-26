package common

// A WorkGroup is wrapper to make managing multiple vital
//   concurrent functions cleaner
type WorkGroup struct {
	functions []func() error
	done      chan error
}

// NewWorkGroup returns a new work group.
func NewWorkGroup() *WorkGroup {
	jobs := WorkGroup{}
	jobs.done = make(chan error)
	return &jobs
}

// Add accepts a function to start work on concurrently.
func (jobs *WorkGroup) Add(work func() error) {
	jobs.functions = append(jobs.functions, work)
}

// Start goes through all added functions and starts them.
func (jobs *WorkGroup) Start() {
	for _, work := range jobs.functions {
		go func(job func() error) {
			err := job()
			jobs.done <- err
		}(work)
	}
}

// Wait watches the channel for any errors and returns them.
func (jobs *WorkGroup) Wait() error {
	return <-jobs.done
}
