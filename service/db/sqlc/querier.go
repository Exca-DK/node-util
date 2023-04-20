// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0

package db

import (
	"context"
)

type Querier interface {
	AddQueries(ctx context.Context, arg AddQueriesParams) (Node, error)
	CreateEnrRecord(ctx context.Context, arg CreateEnrRecordParams) (EnrRecord, error)
	CreateNode(ctx context.Context, id string) (Node, error)
	CreateRlpRecord(ctx context.Context, arg CreateRlpRecordParams) (RlpRecord, error)
	CreateScanRecord(ctx context.Context, arg CreateScanRecordParams) (ScanRecord, error)
	DeleteEnrRecord(ctx context.Context, id string) error
	DeleteNode(ctx context.Context, id string) error
	DeleteNodeScanRecords(ctx context.Context, node string) error
	DeleteRlpRecord(ctx context.Context, id string) error
	DeleteScanRecord(ctx context.Context, id int64) error
	GetEnrRecord(ctx context.Context, id string) (EnrRecord, error)
	GetNode(ctx context.Context, id string) (Node, error)
	GetNodeScanRecords(ctx context.Context, node string) (ScanRecord, error)
	GetRlpRecord(ctx context.Context, id string) (RlpRecord, error)
	IncrementQuery(ctx context.Context, id string) (Node, error)
	ListNodesByActivity(ctx context.Context, arg ListNodesByActivityParams) ([]Node, error)
	ListNodesByCreation(ctx context.Context, arg ListNodesByCreationParams) ([]Node, error)
	MarkEnrSeen(ctx context.Context, arg MarkEnrSeenParams) error
	UpdateEnrFields(ctx context.Context, arg UpdateEnrFieldsParams) error
	UpdateRlpRecord(ctx context.Context, arg UpdateRlpRecordParams) error
}

var _ Querier = (*Queries)(nil)
