package kuchniaviking

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"sort"
	"time"
)

func (kv *kuchniaViking) GetNearestDeliveries(deliveries []Delivery, limit int) ([]Delivery, error) {
	type deliveryWithDiff struct {
		delivery *Delivery
		diff     time.Duration
	}

	var futureDeliveries []deliveryWithDiff

	for i, delivery := range deliveries {
		deliveryTime, err := time.Parse("2006-01-02", delivery.Date)
		if err != nil {
			log.Error().Err(err).Msg("can't parse delivery date")
			continue
		}

		diff := deliveryTime.Sub(time.Now())
		if diff < 0 {
			continue
		}

		futureDeliveries = append(futureDeliveries, deliveryWithDiff{
			delivery: &deliveries[i],
			diff:     diff,
		})
	}

	if len(futureDeliveries) == 0 {
		return nil, fmt.Errorf("no future deliveries found")
	}

	sort.Slice(futureDeliveries, func(i, j int) bool {
		return futureDeliveries[i].diff < futureDeliveries[j].diff
	})

	resultCount := min(limit, len(futureDeliveries))
	result := make([]Delivery, resultCount)
	for i := 0; i < resultCount; i++ {
		result[i] = *futureDeliveries[i].delivery
	}

	return result, nil
}
