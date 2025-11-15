package repository

import (
	"backend/internal/model"
	"context"
)

type ProductRepository struct {
	db DBTX
}

func NewProductRepository(db DBTX) *ProductRepository {
	return &ProductRepository{db: db}
}

// 商品一覧をSQLレベルでページング処理を行う
func (r *ProductRepository) ListProducts(ctx context.Context, userID int, req model.ListRequest) ([]model.Product, int, error) {
	var products []model.Product

	// 件数取得用クエリ
	countQuery := "SELECT COUNT(*) FROM products"
	args := []interface{}{}

	// 商品取得用クエリ
	baseQuery := `
		SELECT product_id, name, value, weight, image, description
		FROM products
	`

	// 検索条件の追加
	whereClause := ""
	if req.Search != "" {
		whereClause = " WHERE (name LIKE ? OR description LIKE ?)"
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// 総件数の取得
	var total int
	err := r.db.GetContext(ctx, &total, countQuery+whereClause, args...)
	if err != nil {
		return nil, 0, err
	}

	// ソートとページング（SQLで実行）
	queryArgs := make([]interface{}, len(args))
	copy(queryArgs, args)
	baseQuery += whereClause
	baseQuery += " ORDER BY " + req.SortField + " " + req.SortOrder + ", product_id ASC"
	baseQuery += " LIMIT ? OFFSET ?"
	queryArgs = append(queryArgs, req.PageSize, req.Offset)

	err = r.db.SelectContext(ctx, &products, baseQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
