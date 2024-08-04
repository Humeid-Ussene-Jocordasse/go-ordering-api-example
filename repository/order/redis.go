package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Humeid-Ussene-Jocordasse/orders-api/model"
	"github.com/redis/go-redis/v9"
)

// This connect to redis instance
type RedisRepo struct {
	Client *redis.Client
}

// creating the id key
func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

// Setting context as the first parameter is good practice in Golang
func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {

	// Marshaling and UnMarshaling, means encoding and decoding respectively, returns a []byte
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to decode order: %w", err)
	}
	key := orderIDKey(order.OrderId)

	// Adding a transaction Client
	txn := r.Client.TxPipeline()

	/*
		The both commands bellow are not going to be executed until there are committed

	*/

	// this will not override if a order with such id had already been created
	res := txn.SetNX(ctx, key, string(data), 0)

	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	// Adding the order to a new group
	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set %w", err)
	}

	//Committing the command, this will insure that the process is closed if all the commands are executed
	if _, err := txn.Exec(ctx); err != nil {
		fmt.Errorf("failed to exect %w", err)
	}
	return nil
}

var ErrNotExist = errors.New("order does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return model.Order{}, ErrNotExist
	}

	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		// this empty error is being returned because the other of doing this would be by creating a pointer
		// and is more expensive, variables are immutable in go, and always have a value...
		return model.Order{}, fmt.Errorf("failed to decode order json: %w", err)
	}

	return order, nil

}

func (r RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)

	// creating a transaction
	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()

	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("Get order: %w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove from orders set: %w", err)
	}

	// Executing the transaction
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}
	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {

	// Marshaling and UnMarshaling, means encoding and decoding respectively
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to decode order: %w", err)
	}
	key := orderIDKey(order.OrderId)
	res := r.Client.SetXX(ctx, key, string(data), 0)

	if err := res.Err(); err != nil {
		return fmt.Errorf("failed to set: %w", err)
	}
	return nil
}

type FindAllPage struct {
	Size   uint64
	Cursor uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", page.Cursor, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
	}

	// If there's no keys, stop the code right here
	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	// MultiGet, it gets multi objects from the db
	xs, err := r.Client.MGet(ctx, keys...).Result()

	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	// Creating a slice with the same length as the Resulting slice
	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		// casting to a string
		x := x.(string)
		var order model.Order

		err := json.Unmarshal([]byte(x), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode order json: %w", err)
		}
		orders[i] = order
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
