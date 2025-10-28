package engine

type StepStatus string

const (
    StepPending StepStatus = "pending"
    StepRunning StepStatus = "running"
    StepSuccess StepStatus = "success"
    StepFailed  StepStatus = "failed"
	StepOK    StepStatus = StepSuccess
	StepError StepStatus = "error"
)

type StepResult struct {
    ID     string
    Status StepStatus
    Output string
    Error  error
}
