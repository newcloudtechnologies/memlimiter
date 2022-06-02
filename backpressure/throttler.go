package backpressure

import (
	"sync/atomic"

	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils"
	"github.com/villenny/fastrand64-go"

	"github.com/pkg/errors"
)

type throttler struct {
	// группа счётчиков запросов
	requestsTotal     utils.Counter
	requestsPassed    utils.Counter
	requestsThrottled utils.Counter

	// для нас важно, чтобы ГСЧ:
	// 1. выдавал равномерное распределение
	// 2. был потокобезопасным
	//
	// Чтобы не писать своё, была выбрана библиотека, в которой потокобезопасность есть из коробки,
	// но при этом есть небольшие сомнения в равномерности распределения.
	// По некоторым признакам, оно действительно равномерное, так как с помощью этого ГСЧ, выдающего только int'ы,
	// можно создать ГСЧ равномерного распределенных float'ов: https://prng.di.unimi.it/
	//
	// Здесь появились ответы, в которых говорится, что эмпирически равномерность распределения
	// наблюдается у всех ГСЧ семейства, кроме одного:
	// https://stackoverflow.com/questions/71050149/does-xoshiro-xoroshiro-prngs-provide-uniform-distribution
	// https://crypto.stackexchange.com/questions/98597
	rng *fastrand64.ThreadsafePoolRNG

	// число в диапазоне [0; 100], показывающее, какой процент запросов должен быть отбит
	threshold uint32
}

func (t *throttler) setThreshold(value uint32) error {
	if value > FullThrottling {
		return errors.New("implementation error: threshold value must belong to [0;100]")
	}

	atomic.StoreUint32(&t.threshold, value)

	return nil
}

func (t *throttler) getStats() *stats.ThrottlingStats {
	return &stats.ThrottlingStats{
		Total:     uint64(t.requestsTotal.Count()),
		Passed:    uint64(t.requestsPassed.Count()),
		Throttled: uint64(t.requestsThrottled.Count()),
	}
}

func (t *throttler) AllowRequest() bool {
	threshold := atomic.LoadUint32(&t.threshold)

	// если подавление отключено, разрешаем любые запросы
	if threshold == 0 {
		t.requestsPassed.Inc(1)

		return true
	}

	// подбрасываем монетку в диапазоне [0; 100], если выпавшее значение окажется меньше порогового значения,
	// запрос подавляется, если нет - разрешается
	value := t.rng.Uint32n(FullThrottling)

	allowed := value >= threshold

	if allowed {
		t.requestsPassed.Inc(1)
	} else {
		t.requestsThrottled.Inc(1)
	}

	return allowed
}

func newThrottler() *throttler {
	requestsTotal := utils.NewCounter(nil)

	return &throttler{
		rng:               fastrand64.NewSyncPoolXoshiro256ssRNG(),
		requestsTotal:     requestsTotal,
		requestsPassed:    utils.NewCounter(requestsTotal),
		requestsThrottled: utils.NewCounter(requestsTotal),
	}
}
