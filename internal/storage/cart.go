package storage

import (
	"context"
	"database/sql"
	"errors"

	"go-backend-project/internal/models"
	"go-backend-project/internal/service"
)

type CartDB struct{ DB *sql.DB }

func NewCartDB(db *sql.DB) *CartDB { return &CartDB{DB: db} }

func (s *CartDB) Get(ctx context.Context, userID int64) (models.Cart, error) {
	cart := models.Cart{Items: make([]models.CartItem, 0)}
	rows, err := s.DB.QueryContext(ctx, `SELECT ci.id,p.id,p.title,p.price,ci.quantity FROM cart_items ci JOIN products p ON p.id=ci.product_id WHERE ci.user_id=$1 AND p.deleted_at IS NULL ORDER BY ci.id`, userID)
	if err != nil {
		return cart, err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.CartItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Title, &item.Price, &item.Quantity); err != nil {
			return cart, err
		}
		cart.Items = append(cart.Items, item)
	}
	if err := rows.Err(); err != nil {
		return cart, err
	}
	err = s.DB.QueryRowContext(ctx, `SELECT COALESCE(p.discount_percentage,0) FROM carts c LEFT JOIN promo_codes p ON p.id=c.promo_id AND p.is_active WHERE c.user_id=$1`, userID).Scan(&cart.DiscountPercentage)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return cart, err
	}
	return cart, nil
}
func (s *CartDB) Add(ctx context.Context, userID, productID int64, quantity int) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT INTO carts(user_id) VALUES($1) ON CONFLICT DO NOTHING", userID); err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRowContext(ctx, `INSERT INTO cart_items(user_id,product_id,quantity) SELECT $1,id,$3 FROM products WHERE id=$2 AND deleted_at IS NULL AND stock_quantity >= $3 ON CONFLICT(user_id,product_id) DO UPDATE SET quantity=cart_items.quantity+EXCLUDED.quantity WHERE cart_items.quantity+EXCLUDED.quantity <= (SELECT stock_quantity FROM products WHERE id=EXCLUDED.product_id) RETURNING id`, userID, productID, quantity).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrUnavailable
	}
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}
func (s *CartDB) Update(ctx context.Context, userID, itemID int64, quantity int) error {
	result, err := s.DB.ExecContext(ctx, `UPDATE cart_items ci SET quantity=$1 FROM products p WHERE ci.id=$2 AND ci.user_id=$3 AND p.id=ci.product_id AND p.deleted_at IS NULL AND p.stock_quantity >= $1`, quantity, itemID, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return service.ErrNotFound
	}
	return nil
}
func (s *CartDB) Delete(ctx context.Context, userID, itemID int64) error {
	result, err := s.DB.ExecContext(ctx, "DELETE FROM cart_items WHERE id=$1 AND user_id=$2", itemID, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return service.ErrNotFound
	}
	return nil
}
func (s *CartDB) ApplyPromo(ctx context.Context, userID int64, code string) error {
	var id int64
	err := s.DB.QueryRowContext(ctx, "SELECT id FROM promo_codes WHERE code=$1 AND is_active", code).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrInvalidPromo
	}
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, "INSERT INTO carts(user_id,promo_id) VALUES($1,$2) ON CONFLICT(user_id) DO UPDATE SET promo_id=EXCLUDED.promo_id", userID, id)
	return err
}
