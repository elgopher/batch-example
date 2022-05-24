// (c) 2022 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/elgopher/batch-example/train"
)

type Store struct {
	db *sql.DB
}

func Start() (Store, error) {
	db, err := sql.Open("postgres", "user=postgres password=postgres host=postgres dbname=postgres sslmode=disable")
	if err != nil {
		return Store{}, err
	}

	db.SetMaxOpenConns(100)

	return Store{db: db}, nil
}

func (d Store) LoadTrain(ctx context.Context, key string) (*train.Train, error) {
	var state []byte
	var version int

	err := d.db.
		QueryRowContext(ctx, `SELECT state, ver FROM train WHERE key=$1`, key).
		Scan(&state, &version)

	if err == sql.ErrNoRows {
		t := train.New(30)
		t.Metadata = 0 // for new train, version is 0
		return t, nil
	}

	if err != nil {
		return nil, fmt.Errorf("executing query failed: %w", err)
	}

	t := &train.Train{}
	if err = json.Unmarshal(state, t); err != nil {
		return nil, fmt.Errorf("train unmarshalling failed: %w", err)
	}

	t.Metadata = version

	return t, nil
}

func (d Store) SaveTrain(ctx context.Context, key string, t *train.Train) error {
	previousVersion := t.Metadata.(int)
	t.Metadata = previousVersion + 1

	rollbackVersionChange := func() {
		t.Metadata = previousVersion
	}

	state, err := json.Marshal(t)
	if err != nil {
		rollbackVersionChange()
		return fmt.Errorf("train marshalling failed: %w", err)
	}

	upsert := `INSERT INTO train(key, state, ver)
				VALUES($1, $2, $3)
				ON CONFLICT ON CONSTRAINT train_pkey 
				DO UPDATE SET state=$2, ver=$3 WHERE train.key=$1 AND train.ver=$4`
	result, err := d.db.ExecContext(ctx, upsert, key, state, t.Metadata, previousVersion)
	if err != nil {
		rollbackVersionChange()
		return fmt.Errorf("ExecContext failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		rollbackVersionChange()
		return fmt.Errorf("RowsAffected failed: %w", err)
	}

	if rows == 0 {
		rollbackVersionChange()
		return fmt.Errorf("concurrent modification of key %s", key)
	}

	return nil
}
