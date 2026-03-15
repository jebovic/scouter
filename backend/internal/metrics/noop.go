package metrics

import "time"

// NoopRecorder is a no-op implementation for tests and when metrics are disabled.
type NoopRecorder struct{}

func (NoopRecorder) RecordLLMCall(_, _, _ string, _ time.Duration)        {}
func (NoopRecorder) RecordLLMTokens(_, _ string, _ int)                   {}
func (NoopRecorder) RecordLLMFallback(_ string)                            {}
func (NoopRecorder) RecordAgentRun(_, _ string)                            {}
func (NoopRecorder) RecordSchedulerRun(_, _ string)                        {}
func (NoopRecorder) RecordAlertTriggered()                                 {}
func (NoopRecorder) RecordHTTPRequest(_, _, _ string, _ time.Duration)     {}
