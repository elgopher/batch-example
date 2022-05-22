// (c) 2022 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

	return Store{db: db}, nil
}

func (d Store) LoadTrain(ctx context.Context, key string) (*train.Train, error) {
	row := d.db.QueryRowContext(ctx, "SELECT state, ver FROM train WHERE key=$1", key)
	if row.Err() != nil {
		return nil, fmt.Errorf("QueryRowContext failed: %w", row.Err())
	}

	var state []byte
	var version int

	err := row.Scan(&state, &version)
	if err == sql.ErrNoRows {
		t := train.New(30)
		t.Metadata = 0
		return t, nil
	}
	if err != nil {
		return nil, fmt.Errorf("row Scan failed: %w", err)
	}

	t := &train.Train{}
	if err = json.Unmarshal(state, t); err != nil {
		return nil, fmt.Errorf("train unmarshalling failed: %w", err)
	}

	t.Metadata = version

	return t, nil
}

func (d Store) SaveTrain(ctx context.Context, key string, t *train.Train) error {
	state, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("train marshalling failed: %w", err)
	}

	version := t.Metadata

	upsert := `INSERT INTO train(key, state, ver)
				VALUES($1, $2, $3)
				ON CONFLICT ON CONSTRAINT train_pkey 
				DO UPDATE SET state=$2, ver=train.ver+1 WHERE train.key=$1 AND train.ver=$3`
	result, err := d.db.ExecContext(ctx, upsert, key, state, version)
	if err != nil {
		return fmt.Errorf("ExecContext failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("RowsAffected failed: %w", err)
	}

	if rows == 0 {
		return errors.New("concurrent modification error")
	}

	return nil
}
