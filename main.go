// (c) 2022 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package main

import (
	"time"

	"github.com/elgopher/batch"
	"github.com/elgopher/batch-example/http"
	"github.com/elgopher/batch-example/logger"
	"github.com/elgopher/batch-example/postgres"
	"github.com/elgopher/batch-example/train"
)

func main() {
	db, err := postgres.Start()
	if err != nil {
		panic(err)
	}

	processor := batch.StartProcessor(
		batch.Options[*train.Train]{
			MinDuration:  100 * time.Millisecond,
			MaxDuration:  3 * time.Second,
			LoadResource: db.LoadTrain,
			SaveResource: db.SaveTrain,
		},
	)
	defer processor.Stop()

	go logger.LogMetrics(processor.SubscribeBatchMetrics())

	trainService := train.Service{
		BatchProcessor: processor,
	}

	if err = http.ListenAndServe(trainService); err != nil {
		panic(err)
	}
}
