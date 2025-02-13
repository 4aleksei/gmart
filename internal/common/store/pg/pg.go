package pg

import (
	"context"
	"database/sql"
	"errors"

	//"fmt"
	"time"

	"github.com/4aleksei/gmart/internal/common/store"
	"github.com/4aleksei/gmart/internal/common/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type (
	PgStore struct {
		pool        *pgxpool.Pool
		DatabaseURI string
	}
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrRowNotFound   = errors.New("not found")
)

func New() *PgStore {
	return &PgStore{}
}

func ProbePGConnection(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}
	return false
}

func ProbePGDublicate(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)
	}
	return false
}

func (s *PgStore) Start(ctx context.Context) error {

	ctxB, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	err := utils.RetryAction(ctxB, utils.RetryTimes(), func(ctx context.Context) error {
		var err error
		s.pool, err = pgxpool.New(ctx, s.DatabaseURI)
		return err
	})

	if err != nil {
		return err
	}

	ctxTimeOutPing, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	err = utils.RetryAction(ctxTimeOutPing, utils.RetryTimes(), func(ctx context.Context) error {
		ctxTime, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		return s.pool.Ping(ctxTime)
	}, ProbePGConnection)

	if err != nil {
		return err
	}

	return nil
}

func (s *PgStore) Stop(ctx context.Context) error {
	s.Close(context.Background())
	return nil
}

const (
	queryDefault = `INSERT INTO users (name, password , created_at) VALUES ($1,$2,now()) RETURNING name, password, user_id`

	queryROrderDefault = `INSERT INTO orders (order_id, user_id , status ,accrual , uploaded_at, changed_at)
	       VALUES ($1,$2, $3 , $4 ,now(),now()) RETURNING order_id, user_id , status ,accrual`

	selectDefault = `SELECT name, password, user_id  FROM users WHERE name = $1`

	selectOrdersDefault = `SELECT  order_id, user_id , status ,accrual , uploaded_at, changed_at  FROM orders WHERE user_id = $1`

	selectOneOrderDefault = `SELECT  order_id, user_id , status ,accrual , uploaded_at, changed_at  FROM orders WHERE order_id = $1`

	//onConflictStatementDelta = ` ON CONFLICT (name, kind)
	//	DO UPDATE SET delta=metrics.delta+excluded.delta,  updated_at = now() RETURNING name, kind, delta, value`
	//onConflictStatementValue = ` ON CONFLICT (name, kind)
	//	DO UPDATE SET  value=excluded.value , updated_at = now() RETURNING name, kind, delta, value`
)

func (s *PgStore) InsertOrder(ctx context.Context, o store.Order) error {
	row := s.pool.QueryRow(ctx, queryROrderDefault, o.OrderID, o.UserID, o.Status, o.Accrual)
	if row != nil {
		var u store.Order
		err := row.Scan(&u.OrderID, &u.UserID, &u.Status, &u.Accrual)
		if err != nil {
			if ProbePGDublicate(err) {
				return ErrAlreadyExists
			}
			return err
		}
	} else {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PgStore) GetOneOrder(ctx context.Context, id uint64) (store.Order, error) {
	row := s.pool.QueryRow(ctx, selectOneOrderDefault, id)
	var o store.Order
	if row != nil {
		err := row.Scan(&o.OrderID, &o.UserID, &o.Status, &o.Accrual, &o.TimeU, &o.TimeC)
		if err != nil {
			return o, err
		}
	} else {
		return o, ErrRowNotFound
	}
	return o, nil
}

func (s *PgStore) GetOrders(ctx context.Context, id uint64) ([]store.Order, error) {
	rows, err := s.pool.Query(ctx, selectOrdersDefault, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ores := make([]store.Order, 0, 10)
	for rows.Next() {
		var o store.Order
		err := rows.Scan(&o.OrderID, &o.UserID, &o.Status, &o.Accrual, &o.TimeU, &o.TimeC)
		if err != nil {
			return nil, err
		}
		ores = append(ores, o)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return ores, nil
}

func (s *PgStore) AddUser(ctx context.Context, u store.User) (store.User, error) {
	row := s.pool.QueryRow(ctx, queryDefault, u.Name, u.Password)
	if row != nil {
		//var m store.User
		err := row.Scan(&u.Name, &u.Password, &u.Id)
		if err != nil {
			if ProbePGDublicate(err) {
				return u, ErrAlreadyExists
			}
			return u, err
		}
	} else {
		return u, sql.ErrNoRows
	}
	return u, nil
}

func (s *PgStore) GetUser(ctx context.Context, u store.User) (store.User, error) {
	row := s.pool.QueryRow(ctx, selectDefault, u.Name)
	if row != nil {
		//var m store.User
		err := row.Scan(&u.Name, &u.Password, &u.Id)
		if err != nil {
			return u, err
		}
	} else {

		return u, ErrRowNotFound
	}
	return u, nil
}

func (s *PgStore) Close(ctx context.Context) {
	s.pool.Close()
}

func (s *PgStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
