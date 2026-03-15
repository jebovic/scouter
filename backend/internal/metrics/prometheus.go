package metrics

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusRecorder implements Recorder using a private Prometheus registry.
type PrometheusRecorder struct {
	registry *prometheus.Registry

	llmCalls     *prometheus.CounterVec
	llmDuration  *prometheus.HistogramVec
	llmTokens    *prometheus.CounterVec
	llmFallbacks *prometheus.CounterVec

	agentRuns *prometheus.CounterVec

	schedulerRuns   *prometheus.CounterVec
	schedulerAlerts prometheus.Counter

	httpRequests *prometheus.CounterVec
	httpDuration *prometheus.HistogramVec
}

// NewPrometheusRecorder creates a recorder with a fresh private registry.
// pool is used to register gauge funcs for active missions and unread notifications.
func NewPrometheusRecorder(pool *pgxpool.Pool) (*PrometheusRecorder, *prometheus.Registry) {
	reg := prometheus.NewRegistry()

	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	llmBuckets := []float64{0.5, 1, 2, 5, 10, 30, 60, 120, 180}

	r := &PrometheusRecorder{
		registry: reg,

		llmCalls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "scouter_llm_calls_total",
			Help: "Total LLM calls by model, label, and result.",
		}, []string{"model", "label", "result"}),

		llmDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "scouter_llm_call_duration_seconds",
			Help:    "LLM call latency in seconds.",
			Buckets: llmBuckets,
		}, []string{"model"}),

		llmTokens: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "scouter_llm_tokens_total",
			Help: "Total LLM tokens by model and direction (input/output).",
		}, []string{"model", "direction"}),

		llmFallbacks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "scouter_llm_fallbacks_total",
			Help: "Total LLM fallback events by model.",
		}, []string{"model"}),

		agentRuns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "scouter_agent_runs_total",
			Help: "Total agent runs by type and result.",
		}, []string{"agent_type", "result"}),

		schedulerRuns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "scouter_scheduler_runs_total",
			Help: "Total scheduler job runs by type and result.",
		}, []string{"job_type", "result"}),

		schedulerAlerts: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "scouter_scheduler_alerts_triggered_total",
			Help: "Total price alerts triggered by the scheduler.",
		}),

		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "scouter_http_requests_total",
			Help: "Total HTTP requests by method, route, and status.",
		}, []string{"method", "route", "status"}),

		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "scouter_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
	}

	reg.MustRegister(
		r.llmCalls, r.llmDuration, r.llmTokens, r.llmFallbacks,
		r.agentRuns,
		r.schedulerRuns, r.schedulerAlerts,
		r.httpRequests, r.httpDuration,
	)

	// Gauge funcs queried at scrape time.
	if pool != nil {
		reg.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "scouter_active_missions",
			Help: "Number of active (non-archived) missions.",
		}, func() float64 {
			var count float64
			_ = pool.QueryRow(context.Background(),
				"SELECT COUNT(*) FROM missions WHERE archived_at IS NULL").Scan(&count)
			return count
		}))

		reg.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "scouter_unread_notifications",
			Help: "Number of unread notifications.",
		}, func() float64 {
			var count float64
			_ = pool.QueryRow(context.Background(),
				"SELECT COUNT(*) FROM notifications WHERE read_at IS NULL").Scan(&count)
			return count
		}))
	}

	return r, reg
}

func (r *PrometheusRecorder) RecordLLMCall(model, label, result string, duration time.Duration) {
	r.llmCalls.WithLabelValues(model, label, result).Inc()
	r.llmDuration.WithLabelValues(model).Observe(duration.Seconds())
}

func (r *PrometheusRecorder) RecordLLMTokens(model, direction string, count int) {
	r.llmTokens.WithLabelValues(model, direction).Add(float64(count))
}

func (r *PrometheusRecorder) RecordLLMFallback(model string) {
	r.llmFallbacks.WithLabelValues(model).Inc()
}

func (r *PrometheusRecorder) RecordAgentRun(agentType, result string) {
	r.agentRuns.WithLabelValues(agentType, result).Inc()
}

func (r *PrometheusRecorder) RecordSchedulerRun(jobType, result string) {
	r.schedulerRuns.WithLabelValues(jobType, result).Inc()
}

func (r *PrometheusRecorder) RecordAlertTriggered() {
	r.schedulerAlerts.Inc()
}

func (r *PrometheusRecorder) RecordHTTPRequest(method, route, status string, duration time.Duration) {
	r.httpRequests.WithLabelValues(method, route, status).Inc()
	r.httpDuration.WithLabelValues(method, route).Observe(duration.Seconds())
}
