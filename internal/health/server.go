package health

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry = prometheus.NewRegistry()
	polls    = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_polls_total",
		Help: "Number of inverter polls performed",
	})
	errors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_poll_errors_total",
		Help: "Number of failed inverter polls",
	})
	lastPoll = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gateway_last_poll_timestamp",
		Help: "Unix timestamp of last poll",
	})
	brokerUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gateway_mqtt_connected",
		Help: "1 if MQTT broker is connected, 0 otherwise",
	})
)

func init() {
	registry.MustRegister(polls, errors, lastPoll, brokerUp)
}

func IncPolls()       { polls.Inc() }
func IncErrors()      { errors.Inc() }
func SetLastPoll(t time.Time) {
	lastPoll.Set(float64(t.Unix()))
}
func SetBrokerUp(up bool) {
	if up {
		brokerUp.Set(1)
	} else {
		brokerUp.Set(0)
	}
}

type Server struct {
	addr    string
	status  *Status
}

func NewServer(addr string, s *Status) *Server {
	return &Server{addr: addr, status: s}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !s.status.Healthy(2 * 30 * time.Second) {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(s.status.Snapshot())
	})
	return mux
}
