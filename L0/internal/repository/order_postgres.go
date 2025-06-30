package repository

import (
	"L0/internal/config"
	"L0/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"math/rand"
)

var (
	ErrShardNotFound    = errors.New("shard not found")
	ErrOrderNotFound    = errors.New("order not found")
	ErrDeliveryNotFound = errors.New("delivery not found")
	ErrPaymentNotFound  = errors.New("payment not found")
)

type ShardConnection struct {
	Primary  *sql.DB
	Replicas []*sql.DB
}

type PostgresRepository struct {
	shards   map[string]*ShardConnection
	lookupDB *sql.DB
	logger   *logrus.Entry
}

func connectToDB(ctx context.Context, dsn config.DSN) (*sql.DB, error) {
	db, err := sql.Open("postgres", string(dsn))
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func NewPostgresRepository(ctx context.Context, cfg *config.PostgresConfig, logger *logrus.Entry) (*PostgresRepository, error) {
	lookupBD, err := connectToDB(ctx, cfg.LookupDbDsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to lookup BD: %w", err)
	}

	shardMap := make(map[string]*ShardConnection)

	for shardKey, shardCfg := range cfg.Shards {
		primaryDB, err := connectToDB(ctx, shardCfg.Primary)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to primary DB for shard key %s: %w", shardKey, err)
		}

		replicasDB := make([]*sql.DB, len(shardCfg.Replicas))
		for _, replica := range shardCfg.Replicas {
			replicaDB, err := connectToDB(ctx, replica)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to replica DB for shard key %s: %w", shardKey, err)
			}
			replicasDB = append(replicasDB, replicaDB)
		}
		shardMap[shardKey] = &ShardConnection{
			Primary:  primaryDB,
			Replicas: replicasDB,
		}
	}
	return &PostgresRepository{
		shards:   shardMap,
		lookupDB: lookupBD,
		logger:   logger,
	}, nil
}

func (r *PostgresRepository) Save(ctx context.Context, order *domain.Order) error {
	shardKey := order.ShardKey
	shardConn, ok := r.shards[shardKey]
	if !ok {
		return fmt.Errorf("shard with key %s not found", shardKey)
	}

	db := shardConn.Primary
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			r.logger.WithFields(logrus.Fields{
				"order_uid": order.OrderUID,
			}).Errorf("failed to save orders: %w", err)
		}
	}(tx)

	sqlQuery := `INSERT INTO orders (
                    order_uid, track_number, entry, locale, internal_signature, customer_id, 
                    delivery_service, shardkey, sm_id, date_created, oof_shard) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	if _, err = tx.ExecContext(ctx, sqlQuery, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID,
		order.DateCreated, order.OOFShard); err != nil {
		r.logger.WithFields(logrus.Fields{
			"order_uid": order.OrderUID,
		}).Errorf("failed to save orders: %w", err)
		return err
	}

	if order.Delivery == nil {
		return fmt.Errorf("order uid %s delivery is empty", order.OrderUID)
	}
	sqlQuery = `INSERT INTO deliveries (
                        order_uid, name, phone, zip, city, address, region, email) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := tx.ExecContext(ctx, sqlQuery, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email); err != nil {
		r.logger.WithFields(logrus.Fields{
			"order_uid": order.OrderUID,
		}).Errorf("failed to insert into deliveries: %v", err)
		return err
	}

	if order.Payment == nil {
		return fmt.Errorf("order uid %s payment is empty", order.OrderUID)
	}
	sqlQuery = `INSERT INTO payments (
                      order_uid, transaction, request_id, currency, provider, amount, 
                      payment_dt, bank, delivery_cost, goods_total, custom_fee) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	if _, err := tx.ExecContext(ctx, sqlQuery, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee); err != nil {
		r.logger.WithFields(logrus.Fields{
			"order_uid": order.OrderUID,
		}).Errorf("failed to insert into payments: %v", err)
		return err
	}

	sqlQuery = `INSERT INTO items (
                   chrt_id, order_uid, track_number,price, rid, name, sale, size, 
                   total_price, nm_id, brand, status) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for _, item := range order.Items {
		if _, err = tx.ExecContext(ctx, sqlQuery, item.ChrtID, order.OrderUID, item.TrackNumber, item.Price, item.RID,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status); err != nil {
			r.logger.WithFields(logrus.Fields{
				"order_uid": order.OrderUID,
				"chrt_id":   item.ChrtID,
			}).Error("failed to insert into items: %v", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"order_uid": order.OrderUID,
		}).Errorf("failed to commit transaction: %w", err)
		return err
	}

	sqlQuery = "INSERT INTO order_shard_mapping (order_uid, shard_key) VALUES ($1, $2)"
	_, err = r.lookupDB.ExecContext(ctx, sqlQuery, order.OrderUID, order.ShardKey)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"order_uid": order.OrderUID,
			"shard_key": order.ShardKey,
		}).Errorf("CRITICAL: failed to update look up table: %w", err)
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"order_uid": order.OrderUID,
		"shard_key": order.ShardKey,
	}).Info("Order saved successfully")
	return nil
}

func (r *PostgresRepository) GetByUID(ctx context.Context, uid uuid.UUID) (*domain.Order, error) {
	sqlQuery := `SELECT shard_key FROM order_shard_mapping WHERE order_uid = $1`
	var shardKey string
	if err := r.lookupDB.QueryRowContext(ctx, sqlQuery, uid).Scan(&shardKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Error("shard not found in look up table")
			return nil, ErrShardNotFound
		} else {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Errorf("failed to look up order: %v", err)
			return nil, err
		}
	}

	shardConn, ok := r.shards[string(shardKey)]
	if !ok {
		r.logger.WithFields(logrus.Fields{
			"shard_key": shardKey,
		}).Error("CRITICAL: failed configuration loading shards")
		return nil, ErrShardNotFound
	}

	var db *sql.DB

	if len(shardConn.Replicas) > 0 {
		db = shardConn.Replicas[rand.Intn(len(shardConn.Replicas))]
	} else {
		db = shardConn.Primary
	}

	sqlQuery = `SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, 
                    delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid = $1`
	row := db.QueryRowContext(ctx, sqlQuery, uid)
	order := &domain.Order{}
	if err := row.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID,
		&order.DateCreated, &order.OOFShard); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Error("order not found in orders")
			return nil, ErrOrderNotFound
		} else {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Errorf("failed to find order: %v", err)
			return nil, err
		}
	}

	sqlQuery = `SELECT name, phone, zip, city, address, region, email FROM deliveries WHERE order_uid = $1`
	row = db.QueryRowContext(ctx, sqlQuery, uid)
	delivery := &domain.Delivery{}
	if err := row.Scan(&delivery.Name, &delivery.Phone, &delivery.Zip,
		&delivery.City, &delivery.Address, &delivery.Region, &delivery.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Error("delivery not found in orders")
			return nil, ErrDeliveryNotFound
		} else {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Errorf("failed to find delivery: %v", err)
			return nil, err
		}
	}

	sqlQuery = `SELECT transaction, request_id, currency, provider, amount, 
                      payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payments WHERE order_uid = $1`
	row = db.QueryRowContext(ctx, sqlQuery, uid)
	payment := &domain.Payment{}
	if err := row.Scan(&payment.Transaction, &payment.RequestID,
		&payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDT, &payment.Bank,
		&payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Error("payment not found in orders")
			return nil, ErrPaymentNotFound
		} else {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Errorf("failed to find payment: %v", err)
			return nil, err
		}
	}

	sqlQuery = `SELECT chrt_id, track_number,price, rid, name, sale, size, 
                   total_price, nm_id, brand, status FROM items WHERE order_uid = $1`
	rows, err := db.QueryContext(ctx, sqlQuery, uid)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"uid": uid,
		}).Errorf("failed to find order: %v", err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Errorf("failed to close rows: %v", err)
		}
	}(rows)

	items := make([]domain.Item, 0)
	for rows.Next() {
		var item domain.Item
		if err = rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
			r.logger.WithFields(logrus.Fields{
				"uid": uid,
			}).Errorf("failed to scan item: %v", err)
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		r.logger.WithFields(logrus.Fields{
			"uid": uid,
		}).Errorf("failed to scan rows: %v", err)
		return nil, err
	}
	order.Items = items
	order.Delivery = delivery
	order.Payment = payment
	return order, nil
}

func (r *PostgresRepository) GetLatest(ctx context.Context, limit uint) ([]*domain.Order, error) {
	r.logger.WithField("limit", limit).Info("Fetching latest orders to warm up cache...")

	sqlQuery := `SELECT order_uid FROM orders ORDER BY date_created DESC LIMIT $1`
	rows, err := r.lookupDB.QueryContext(ctx, sqlQuery, limit)
	if err != nil {
		r.logger.Errorf("failed to query latest order uids: %v", err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			r.logger.Errorf("failed to close rows: %v", err)
		}
	}(rows)

	var orderUIDs []uuid.UUID
	for rows.Next() {
		var uid uuid.UUID
		if err := rows.Scan(&uid); err != nil {
			r.logger.Errorf("failed to scan order uid during cache warming: %v", err)
			return nil, err
		}
		orderUIDs = append(orderUIDs, uid)
	}
	if err := rows.Err(); err != nil {
		r.logger.Errorf("error during latest order uids iteration: %v", err)
		return nil, err
	}

	orders := make([]*domain.Order, 0, len(orderUIDs))
	for _, uid := range orderUIDs {
		order, err := r.GetByUID(ctx, uid)
		if err != nil {
			r.logger.WithField("order_uid", uid).Warnf("failed to get full order during cache warming: %v", err)
			continue
		}
		orders = append(orders, order)
	}

	r.logger.Infof("cache warming complete. Fetched %d orders.", len(orders))
	return orders, nil
}
