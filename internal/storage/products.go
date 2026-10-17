package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go-backend-project/internal/models"
	"go-backend-project/internal/service"
)

type ProductDB struct{ DB *sql.DB }

func NewProductDB(db *sql.DB) *ProductDB { return &ProductDB{DB: db} }
func (s *ProductDB) List(ctx context.Context, f models.ProductFilter) (models.ProductPage, error) {
	page := models.ProductPage{Items: make([]models.Product, 0)}
	clauses := []string{"deleted_at IS NULL"}
	args := []any{}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		clauses = append(clauses, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", len(args), len(args)))
	}
	for _, filter := range []struct {
		col, op string
		value   *float64
	}{{"price", ">=", f.MinPrice}, {"price", "<=", f.MaxPrice}, {"rating", ">=", f.MinRating}} {
		if filter.value != nil {
			args = append(args, *filter.value)
			clauses = append(clauses, fmt.Sprintf("%s %s $%d", filter.col, filter.op, len(args)))
		}
	}
	where := strings.Join(clauses, " AND ")
	if err := s.DB.QueryRowContext(ctx, "SELECT count(*) FROM products WHERE "+where, args...).Scan(&page.Meta.TotalItems); err != nil {
		return page, err
	}
	page.Meta.TotalPages = (page.Meta.TotalItems + f.Limit - 1) / f.Limit
	page.Meta.CurrentPage = f.Page
	page.Meta.Limit = f.Limit
	args = append(args, f.Limit, (f.Page-1)*f.Limit)
	query := fmt.Sprintf("SELECT id,title,description,price,rating,to_json(images),stock_quantity,created_at FROM products WHERE %s ORDER BY %s %s, id ASC LIMIT $%d OFFSET $%d", where, f.SortBy, f.Order, len(args)-1, len(args))
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.Product
		var images []byte
		if err = rows.Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.Rating, &images, &p.StockQuantity, &p.CreatedAt); err != nil {
			return page, err
		}
		if err = json.Unmarshal(images, &p.Images); err != nil {
			return page, err
		}
		page.Items = append(page.Items, p)
	}
	return page, rows.Err()
}
func (s *ProductDB) Create(ctx context.Context, v models.ProductInput) (int64, error) {
	images, err := json.Marshal(v.Images)
	if err != nil {
		return 0, err
	}
	var id int64
	err = s.DB.QueryRowContext(ctx, `INSERT INTO products(title,description,price,rating,images,stock_quantity) SELECT $1,$2,$3,$4,ARRAY(SELECT jsonb_array_elements_text($5::jsonb)),$6 RETURNING id`, v.Title, v.Description, *v.Price, *v.Rating, string(images), *v.Stock).Scan(&id)
	return id, err
}
func (s *ProductDB) Update(ctx context.Context, id int64, v models.ProductInput) error {
	images, err := json.Marshal(v.Images)
	if err != nil {
		return err
	}
	result, err := s.DB.ExecContext(ctx, `UPDATE products SET title=$1,description=$2,price=$3,rating=$4,images=ARRAY(SELECT jsonb_array_elements_text($5::jsonb)),stock_quantity=$6 WHERE id=$7 AND deleted_at IS NULL`, v.Title, v.Description, *v.Price, *v.Rating, string(images), *v.Stock, id)
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
func (s *ProductDB) Delete(ctx context.Context, id int64) error {
	result, err := s.DB.ExecContext(ctx, "UPDATE products SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
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
