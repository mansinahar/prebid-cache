package backends

import (
	"context"
	"errors"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prebid/prebid-cache/config"
	"github.com/prebid/prebid-cache/metrics"
	log "github.com/sirupsen/logrus"
)

const setName = "uuid"
const binValue = "value"

type Aerospike struct {
	cfg     config.Aerospike
	client  *as.Client
	metrics *metrics.Metrics
}

func NewAerospikeBackend(cfg config.Aerospike, metrics *metrics.Metrics) *Aerospike {
	client, err := as.NewClient(cfg.Host, cfg.Port)
	if err != nil {
		log.Fatalf("Error creating Aerospike backend: %v", err)
		panic("Aerospike failure. This shouldn't happen.")
	}
	log.Infof("Connected to Aerospike at %s:%d", cfg.Host, cfg.Port)

	return &Aerospike{
		cfg:     cfg,
		client:  client,
		metrics: metrics,
	}
}

func (a *Aerospike) Get(ctx context.Context, key string) (string, error) {
	asKey, err := as.NewKey(a.cfg.Namespace, setName, key)
	if err != nil {
		return "", err
	}
	rec, err := a.client.Get(nil, asKey, "value")
	if err != nil {
		return "", err
	}
	if rec == nil {
		return "", errors.New("client.Get returned a nil record. Is aerospike configured properly?")
	}
	a.metrics.RecordExtraTTLSeconds(float64(rec.Expiration))
	return rec.Bins[binValue].(string), nil
}

func (a *Aerospike) Put(ctx context.Context, key string, value string, ttlSeconds int) error {
	asKey, err := as.NewKey(a.cfg.Namespace, setName, key)
	if err != nil {
		return err
	}
	if ttlSeconds == 0 {
		ttlSeconds = a.cfg.DefaultTTL
	}
	bins := as.BinMap{
		binValue: value,
	}
	err = a.client.Put(&as.WritePolicy{
		Expiration: uint32(ttlSeconds),
	}, asKey, bins)
	if err != nil {
		return err
	}
	return nil
}
