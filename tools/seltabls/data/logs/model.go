// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package logs

import (
	"time"
)

type Log struct {
	ID             int64      `db:"id" json:"id"`
	Value          string     `db:"value" json:"value"`
	CreatedAt      *time.Time `db:"created_at" json:"created_at"`
	RequestID      *int64     `db:"request_id" json:"request_id"`
	ResponseID     *int64     `db:"response_id" json:"response_id"`
	NotificationID *int64     `db:"notification_id" json:"notification_id"`
}

type LogLevel struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type Notification struct {
	ID        int64     `db:"id" json:"id"`
	Method    string    `db:"method" json:"method"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Request struct {
	ID        int64  `db:"id" json:"id"`
	RpcMethod string `db:"rpc_method" json:"rpc_method"`
	RpcID     int64  `db:"rpc_id" json:"rpc_id"`
}

type Response struct {
	ID        int64     `db:"id" json:"id"`
	Rpc       string    `db:"rpc" json:"rpc"`
	RpcID     int64     `db:"rpc_id" json:"rpc_id"`
	Result    *string   `db:"result" json:"result"`
	Error     *string   `db:"error" json:"error"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
