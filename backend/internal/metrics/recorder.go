package metrics

import "time"

// LLMRecorder records LLM call metrics.
type LLMRecorder interface {
	RecordLLMCall(model, label, result string, duration time.Duration)
	RecordLLMTokens(model, direction string, count int)
	RecordLLMFallback(model string)
}

// AgentRecorder records agent run metrics.
type AgentRecorder interface {
	RecordAgentRun(agentType, result string)
}

// SchedulerRecorder records scheduler job metrics.
type SchedulerRecorder interface {
	RecordSchedulerRun(jobType, result string)
	RecordAlertTriggered()
}

// HTTPRecorder records HTTP request metrics.
type HTTPRecorder interface {
	RecordHTTPRequest(method, route, status string, duration time.Duration)
}

// Recorder combines all metric sub-interfaces.
type Recorder interface {
	LLMRecorder
	AgentRecorder
	SchedulerRecorder
	HTTPRecorder
}
