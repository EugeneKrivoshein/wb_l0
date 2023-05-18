package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
	csh  *Cache
	name string
}

func NewDB() *DB {
	db := DB{}
	db.Init()
	return &db
}

func (db *DB) SetCacheInstance(csh *Cache) {
	db.csh = csh
}

func (db *DB) GetCacheState(bufSize int) (map[int64]Order, []int64, int, error) {
	buffer := make(map[int64]Order, bufSize)
	que := make([]int64, bufSize)
	var queIndex int

	query := fmt.Sprintf("SELECT order_id FROM cache WHERE app_key = '%s' ORDER BY id DESC LIMIT %d", os.Getenv("APP_KEY"), bufSize)
	rows, err := db.pool.Query(context.Background(), query)
	if err != nil {
		log.Printf("%v: error getting order_id from db: %v\n", db.name, err)
	}
	defer rows.Close()

	var orderid int64
	for rows.Next() {
		if err := rows.Scan(&orderid); err != nil {
			log.Printf("%v: error getting order_id from db row: %v\n", db.name, err)
			return buffer, que, queIndex, errors.New("error getting order_id from db row")
		}
		que[queIndex] = orderid
		queIndex++

		o, err := db.GetOrderByID(orderid)
		if err != nil {
			log.Printf("%v: error getting order from db: %v\n", db.name, err)
			continue
		}
		buffer[orderid] = o
	}

	if queIndex == 0 {
		return buffer, que, queIndex, errors.New("cache is empty")
	}

	for i := 0; i < int(queIndex/2); i++ {
		que[i], que[queIndex-i-1] = que[queIndex-i-1], que[i]
	}

	return buffer, que, queIndex, nil
}

func (db *DB) GetOrderByID(oid int64) (Order, error) {
	var o Order
	var delivery_id_fk int64
	var payment_id_fk int64

	err := db.pool.QueryRow(context.Background(), `SELECT OrderUID, Entry, InternalSignature, delivery_id_fk, payment_id_fk, Locale, CustomerID, 
	TrackNumber, DeliveryService, Shardkey, SmID, Total FROM orders WHERE id = $1`, oid).Scan(&o.OrderUID, &o.Entry,
		&o.InternalSignature, &delivery_id_fk, &payment_id_fk, &o.Locale, &o.CustomerID, &o.TrackNumber, &o.DeliveryService, &o.Shardkey,
		&o.SmID, &o.Total)
	if err != nil {
		return o, errors.New("error getting order from db")
	}

	err = db.pool.QueryRow(context.Background(), `SELECT Name, Phone, Zip, City, Address, Region, Email
	FROM delivery WHERE id = $1`, delivery_id_fk).Scan(&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City, &o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email)
	if err != nil {
		log.Printf("%v: unable to get delivery from database: %v\n", db.name, err)
		return o, errors.New("error getting delivery from db")
	}

	err = db.pool.QueryRow(context.Background(), `SELECT Transaction,Request_id, Currency, Provider, Amount, Payment_dt, Bank, Delivery_cost,
	Goods_total, Custom_fee FROM payment WHERE id = $1`, payment_id_fk).Scan(&o.Payment.Transaction, &o.Payment.Request_id, &o.Payment.Currency, &o.Payment.Provider,
		&o.Payment.Amount, &o.Payment.Payment_dt, &o.Payment.Bank, &o.Payment.Delivery_cost, &o.Payment.Goods_total, &o.Payment.Custom_fee)
	if err != nil {
		log.Printf("%v: unable to get payment from database: %v\n", db.name, err)
		return o, errors.New("error getting payment from db")
	}

	rowsItems, err := db.pool.Query(context.Background(), "SELECT item_id_fk FROM order_items WHERE order_id_fk = $1", oid)
	if err != nil {
		return o, errors.New("error getting items id list from db")
	}
	defer rowsItems.Close()

	var itemID int64
	for rowsItems.Next() {
		var item Items
		if err := rowsItems.Scan(&itemID); err != nil {
			return o, errors.New("error getting itemID from db row")
		}

		err = db.pool.QueryRow(context.Background(), `SELECT Chrt_id, Track_number, Price, Rid, Name, Sale, Size, TotalPrice, Nm_id, Brand, Status 
		FROM items WHERE id = $1`, itemID).Scan(&item.Chrt_id, &item.Track_number, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.Nm_id, &item.Brand, &item.Status)
		if err != nil {
			return o, errors.New("error getting item from db")
		}
		o.Items = append(o.Items, item)
	}
	return o, nil
}

func (db *DB) AddOrder(o Order) (int64, error) {
	var lastInsertId int64
	var itemsIds []int64 = []int64{}

	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(context.Background())

	for _, item := range o.Items {
		err := tx.QueryRow(context.Background(), `INSERT INTO items (Chrt_id, Track_number, Price, Rid, Name, Sale, Size, TotalPrice, Nm_id, Brand, Status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`, item.Chrt_id, item.Track_number, item.Price, item.Rid, item.Name, item.Sale, item.Size,
			item.TotalPrice, item.Nm_id, item.Brand, item.Status).Scan(&lastInsertId)
		if err != nil {
			log.Printf("%v: err insert data: %v\n", db.name, err)
			return -1, err
		}
		itemsIds = append(itemsIds, lastInsertId)
	}

	err = tx.QueryRow(context.Background(), `INSERT INTO delivery (Name, Phone, Zip, City, Address, Region, Email) 
		values ($1, $2, $3, $4, $5, $6, $7) RETURNING id`, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City, o.Delivery.Address,
		o.Delivery.Region, o.Delivery.Email).Scan(&lastInsertId)
	if err != nil {
		log.Printf("%v: err insert data: %v\n", db.name, err)
		return -1, err
	}
	deliveryIdFk := lastInsertId

	err = tx.QueryRow(context.Background(), `INSERT INTO payment (Transaction, Request_id, Currency, Provider, Amount, Payment_dt, Bank, Delivery_cost,
		 Goods_total, Custom_fee) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`, o.Payment.Transaction, o.Payment.Request_id, o.Payment.Currency, o.Payment.Provider,
		o.Payment.Amount, o.Payment.Payment_dt, o.Payment.Bank, o.Payment.Delivery_cost, o.Payment.Goods_total, o.Payment.Custom_fee).Scan(&lastInsertId)
	if err != nil {
		log.Printf("%v: err insert data: %v\n", db.name, err)
		return -1, err
	}
	paymentIdFk := lastInsertId

	err = tx.QueryRow(context.Background(), `INSERT INTO orders (OrderUID, Entry, InternalSignature, delivery_id_fk, payment_id_fk, Locale, 
		CustomerID, TrackNumber, DeliveryService, Shardkey, SmID, Total) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`,
		o.OrderUID, o.Entry, o.InternalSignature, deliveryIdFk, paymentIdFk, o.Locale, o.CustomerID, o.TrackNumber, o.DeliveryService,
		o.Shardkey, o.SmID, o.Total).Scan(&lastInsertId)
	if err != nil {
		log.Printf("%v: err insert data: %v\n", db.name, err)
		return -1, err
	}
	orderIdFk := lastInsertId

	for _, itemId := range itemsIds {
		_, err := tx.Exec(context.Background(), `INSERT INTO order_items (order_id_fk, item_id_fk) values ($1, $2)`,
			orderIdFk, itemId)
		if err != nil {
			log.Printf("%v: err insert data: %v\n", db.name, err)
			return -1, err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return 0, err
	}

	log.Printf("%v: order add to db\n", db.name)
	db.csh.SetOrder(orderIdFk, o)
	return orderIdFk, nil
}

func (db *DB) SendOrderIDToCache(oid int64) {
	db.pool.QueryRow(context.Background(), `INSERT INTO cache (order_id, app_key) VALUES ($1, $2)`, oid, os.Getenv("APP_KEY"))
	log.Printf("%v: order_id add to cache\n", db.name)
}

func (db *DB) ClearCache() {
	_, err := db.pool.Exec(context.Background(), `DELETE FROM cache WHERE app_key = $1`, os.Getenv("APP_KEY"))
	if err != nil {
		log.Printf("%v: clear cache error: %s\n", db.name, err)
	}
	log.Printf("%v: cache cleared from db\n", db.name)
}
