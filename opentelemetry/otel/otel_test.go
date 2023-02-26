package otel_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"

	"github.com/hitzhangjie/codemaster/otel/metrictest"
)

// meter can be a global/package variable.
var provider metric.MeterProvider
var exporter *metrictest.Exporter
var meter metric.Meter

func init() {
	boundaries := metrictest.WithExplicitBoundaries([]float64{10, 20, 30, 40, 50, 100})
	provider, exporter = metrictest.NewTestMeterProvider(boundaries)
	meter = provider.Meter("app_or_package_name")
}

func Test_OTEL_Counter(t *testing.T) {
	counter, _ := meter.SyncInt64().Counter(
		"some.prefix.counter",
		instrument.WithUnit("1"),
		instrument.WithDescription("TODO"),
	)

	counter.Add(context.Background(), 1)
	counter.Add(context.Background(), 1)
	counter.Add(context.Background(), 1)

	exporter.Collect(context.Background())
	record, err := exporter.GetByName("some.prefix.counter")
	assert.Nil(t, err)
	fmt.Printf("%+v\n", record)
}

// counterWithLabels demonstrates how to add different labels ("hits" and "misses")
// to measurements. Using this simple trick, you can get number of hits, misses,
// sum = hits + misses, and hit_rate = hits / (hits + misses).
func Test_OTEL_CounterWithLabels(t *testing.T) {
	ctx := context.Background()
	counter, _ := meter.SyncInt64().Counter(
		"some.prefix.cache",
		instrument.WithDescription("Cache hits and misses"),
	)
	for {
		if rand.Float64() < 0.3 {
			// increment hits
			counter.Add(ctx, 1, attribute.String("type", "hits"))
		} else {
			// increments misses
			counter.Add(ctx, 1, attribute.String("type", "misses"))
		}

		time.Sleep(time.Millisecond)
	}
}

// histogram demonstrates how to record a distribution of individual values, for example,
// request or query timings. With this instrument you get total number of records,
// avg/min/max values, and heatmaps/percentiles.
func Test_OTEL_Histogram(t *testing.T) {
	ctx := context.Background()
	hist, _ := meter.SyncInt64().Histogram(
		"some.prefix.histogram",
		instrument.WithUnit("microseconds"),
		instrument.WithDescription("TODO"),
	)

	var count int
	for {
		hist.Record(ctx, int64(rand.Int())%10)
		time.Sleep(time.Millisecond)
		count++
		if count > 10 {
			break
		}
	}

	exporter.Collect(context.Background())
	record, err := exporter.GetByName("some.prefix.histogram")
	assert.Nil(t, err)
	fmt.Printf("%+v\n", record)
}
