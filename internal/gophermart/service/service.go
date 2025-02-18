package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/4aleksei/gmart/internal/common/models"
	"github.com/4aleksei/gmart/internal/common/store"
	"github.com/4aleksei/gmart/internal/common/store/pg"
	"github.com/4aleksei/gmart/internal/gophermart/config"

	"github.com/4aleksei/gmart/internal/common/httpclientpool"
	"github.com/4aleksei/gmart/internal/common/httpclientpool/job"
)

type ServiceStore interface {
	AddUser(context.Context, store.User) (store.User, error)
	GetUser(context.Context, store.User) (store.User, error)
	GetBalance(context.Context, uint64) (store.Balance, error)
	InsertOrder(context.Context, store.Order) error
	InsertWithdraw(context.Context, store.Withdraw) error
	GetOrders(context.Context, uint64) ([]store.Order, error)

	GetWithdrawals(context.Context, uint64) ([]store.Withdraw, error)

	GetOneOrder(context.Context, uint64) (store.Order, error)

	GetOrdersForProcessing(context.Context) ([]store.Order, error)
	UpdateOrdersBalancesBatch(context.Context, []store.Order) error
}

type HandleService struct {
	store ServiceStore
	key   string
	httpc *httpclientpool.PoolHandler
	jid   job.JobID
}

var (
	ErrAuthenticationFailed = errors.New("authentication_failed")

	ErrBadPass = errors.New("name or password empty")

	ErrBadTypeValue = errors.New("invalid typeValue")
	ErrBadValue     = errors.New("error value conversion")
	ErrBadKindType  = errors.New("error kind type")

	ErrBadValueUser = errors.New("parse user_id number error")

	ErrOrderAlreadyLoaded = errors.New("order Already Loaded")

	ErrOrderAlreadyLoadedOtherUser = errors.New("error order already loaded other")

	ErrBalanceNotEnough = errors.New("balance not enouth")
)

func NewService(s ServiceStore, cfg *config.Config, h *httpclientpool.PoolHandler) *HandleService {
	return &HandleService{
		key:   cfg.Key,
		store: s,
		httpc: h,
	}
}

func (s *HandleService) RegisterUser(ctx context.Context, user models.UserRegistration) (string, error) {

	//Check Name and Pass
	if user.Name == "" || user.Password == "" {
		return "", ErrBadPass
	}

	pass := user.Password // try hash password

	userAdded, err := s.store.AddUser(ctx, store.User{Name: user.Name, Password: pass})

	if err != nil {
		if errors.Is(err, pg.ErrAlreadyExists) {
			return "", ErrAuthenticationFailed
		}
		return "", err
	}

	id := strconv.FormatUint(userAdded.Id, 10)

	return id, nil
}

func (s *HandleService) LoginUser(ctx context.Context, user models.UserRegistration) (string, error) {
	//Check Name and Pass
	if user.Name == "" || user.Password == "" {
		return "", ErrBadPass
	}

	userGet, err := s.store.GetUser(ctx, store.User{Name: user.Name})
	if err != nil {
		if errors.Is(err, pg.ErrRowNotFound) {
			return "", ErrAuthenticationFailed
		}
		return "", err
	}
	if user.Password != userGet.Password {
		return "", ErrAuthenticationFailed
	}

	id := strconv.FormatUint(userGet.Id, 10)

	return id, nil
}

func (s *HandleService) PostWithdraw(ctx context.Context, userIdStr string, withdraw models.Withdraw) error {
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed %w : %w", ErrBadValueUser, err)
	}

	orderId, err := strconv.ParseUint(withdraw.OrderID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed %w : %w", ErrBadValue, err)
	}
	err = s.store.InsertWithdraw(ctx, store.Withdraw{OrderID: orderId, UserID: userId, Sum: withdraw.Sum})
	if err != nil {

		if errors.Is(err, pg.ErrBalanceNotEnough) {
			return ErrBalanceNotEnough
		}

		//variants!!!
		return err
	}
	return nil
}

func (s *HandleService) RegisterOrder(ctx context.Context, userIdStr, orderIdStr string) error {

	orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed %w : %w", ErrBadValue, err)
	}
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed %w : %w", ErrBadValueUser, err)
	}

	err = s.store.InsertOrder(ctx, store.Order{OrderID: orderId, UserID: userId, Status: "NEW", Accrual: 0})
	if err != nil {
		if errors.Is(err, pg.ErrAlreadyExists) {
			one, err := s.store.GetOneOrder(ctx, orderId)
			if err != nil {
				return err
			}
			if one.UserID != userId {
				return ErrOrderAlreadyLoadedOtherUser
			}
			return ErrOrderAlreadyLoaded
		}
		return err
	}
	return nil
}

func (s *HandleService) GetOrders(ctx context.Context, userIdStr string) ([]models.Order, error) {
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
	}
	vals, err := s.store.GetOrders(ctx, userId)
	if err != nil {
		return nil, err
	}
	valsret := make([]models.Order, len(vals))
	for i, v := range vals {
		valsret[i] = models.Order{OrderID: strconv.FormatUint(v.OrderID, 10), Status: v.Status, Accrual: v.Accrual, Time: v.TimeU}
	}
	return valsret, nil
}

func (s *HandleService) GetWithdrawals(ctx context.Context, userIdStr string) ([]models.Withdraw, error) {
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
	}
	vals, err := s.store.GetWithdrawals(ctx, userId)
	if err != nil {
		return nil, err
	}
	valsret := make([]models.Withdraw, len(vals))
	for i, v := range vals {
		valsret[i] = models.Withdraw{OrderID: strconv.FormatUint(v.OrderID, 10), Sum: v.Sum, TimeC: v.TimeC}
	}
	return valsret, nil
}

func (s *HandleService) GetBalance(ctx context.Context, userIdStr string) (models.Balance, error) {
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	var valRet models.Balance
	if err != nil {
		return valRet, fmt.Errorf("failed %w : %w", ErrBadValue, err)
	}
	val, err := s.store.GetBalance(ctx, userId)
	if err != nil {
		if errors.Is(err, pg.ErrRowNotFound) {
			return valRet, nil
		}
		return valRet, err
	}
	valRet.Accrual = val.Accrual
	valRet.Withdrawn = val.Withdrawn
	return valRet, err
}

// Accrual Services

func (s *HandleService) GetOrdersForProcess(ctx context.Context) ([]store.Order, error) {
	vals, err := s.store.GetOrdersForProcessing(ctx)
	if err != nil {
		return nil, err
	}
	return vals, nil
}

func (s *HandleService) UpdateOrdersAndBalances(ctx context.Context, updOrders []store.Order) error {
	err := s.store.UpdateOrdersBalancesBatch(ctx, updOrders)
	if err != nil {
		return err
	}
	return nil
}

func (s *HandleService) newJid() job.JobID {
	s.jid++
	return s.jid
}

func (s *HandleService) sendRun(ctx context.Context, jobs chan job.Job, orders []store.Order) {
	defer close(jobs)

	for _, val := range orders {
		select {
		case <-ctx.Done():
			return
		default:
			id := s.newJid()
			jobs <- job.Job{ID: id, Value: val}
		}
	}

}

func (s *HandleService) SendOrdersToAccrual(ctx context.Context, orders []store.Order) (map[uint64]store.Order, int, error) {
	wg := &sync.WaitGroup{}
	jobs := make(chan job.Job, s.httpc.WorkerCount*2)
	results := make(chan job.Result, s.httpc.WorkerCount*2)

	go s.sendRun(ctx, jobs, orders)

	s.httpc.StartPool(ctx, jobs, results, wg)

	go func() {
		wg.Wait()
		close(results)
	}()
	var waitSec int
	resOrders := make(map[uint64]store.Order)
	for res := range results {
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		default:
			if res.Err == nil {
				if res.Result == 200 {
					resOrders[res.Value.OrderID] = res.Value
				} else if res.Result == 429 {
					if res.WaitSec > waitSec {
						waitSec = res.WaitSec
					}
				}
			}
		}
	}
	return resOrders, waitSec, nil
}
