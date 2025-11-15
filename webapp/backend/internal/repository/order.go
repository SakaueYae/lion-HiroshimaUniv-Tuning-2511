package repository

import (
	"backend/internal/model"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db DBTX
}

func NewOrderRepository(db DBTX) *OrderRepository {
	return &OrderRepository{db: db}
}

// 注文を作成し、生成された注文IDを返す
func (r *OrderRepository) Create(ctx context.Context, order *model.Order) (string, error) {
	query := `INSERT INTO orders (user_id, product_id, shipped_status, created_at) VALUES (?, ?, 'shipping', NOW())`
	result, err := r.db.ExecContext(ctx, query, order.UserID, order.ProductID)
	if err != nil {
		return "", err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

// 複数の注文IDのステータスを一括で更新
// 主に配送ロボットが注文を引き受けた際に一括更新をするために使用
func (r *OrderRepository) UpdateStatuses(ctx context.Context, orderIDs []int64, newStatus string) error {
	if len(orderIDs) == 0 {
		return nil
	}
	query, args, err := sqlx.In("UPDATE orders SET shipped_status = ? WHERE order_id IN (?)", newStatus, orderIDs)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// 配送中(shipped_status:shipping)の注文一覧を取得
func (r *OrderRepository) GetShippingOrders(ctx context.Context) ([]model.Order, error) {
	var orders []model.Order
	query := `
        SELECT
            o.order_id,
            p.weight,
            p.value
        FROM orders o
        JOIN products p ON o.product_id = p.product_id
        WHERE o.shipped_status = 'shipping'
    `
	err := r.db.SelectContext(ctx, &orders, query)
	return orders, err
}

// 注文履歴一覧を取得（SQLで検索・ソート・ページング）
func (r *OrderRepository) ListOrders(ctx context.Context, userID int, req model.ListRequest) ([]model.Order, int, error) {
	type orderRow struct {
		OrderID       int          `db:"order_id"`
		ProductID     int          `db:"product_id"`
		ProductName   string       `db:"product_name"`
		ShippedStatus string       `db:"shipped_status"`
		CreatedAt     sql.NullTime `db:"created_at"`
		ArrivedAt     sql.NullTime `db:"arrived_at"`
	}

	// WHERE句の構築
	whereClause := "WHERE o.user_id = ?"
	args := []interface{}{userID}

	// 検索条件の追加
	if req.Search != "" {
		if req.Type == "prefix" {
			whereClause += " AND p.name LIKE ?"
			args = append(args, req.Search+"%")
		} else {
			// partial (部分一致)
			whereClause += " AND p.name LIKE ?"
			args = append(args, "%"+req.Search+"%")
		}
	}

	// 総件数を取得
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM orders o
		JOIN products p ON o.product_id = p.product_id
		%s
	`, whereClause)

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// ソートフィールドのマッピング（カラム名への変換）
	sortColumn := "o.order_id"
	switch req.SortField {
	case "product_name":
		sortColumn = "p.name"
	case "created_at":
		sortColumn = "o.created_at"
	case "shipped_status":
		sortColumn = "o.shipped_status"
	case "arrived_at":
		sortColumn = "o.arrived_at"
	case "order_id":
		sortColumn = "o.order_id"
	default:
		sortColumn = "o.order_id"
	}

	// ソート順序の検証
	sortOrder := "ASC"
	if strings.ToUpper(req.SortOrder) == "DESC" {
		sortOrder = "DESC"
	}

	// データ取得クエリ（検索・ソート・ページング全てSQLで実行）
	dataQuery := fmt.Sprintf(`
		SELECT o.order_id, o.product_id, p.name as product_name, 
		       o.shipped_status, o.created_at, o.arrived_at
		FROM orders o
		JOIN products p ON o.product_id = p.product_id
		%s
		ORDER BY %s %s, o.order_id ASC
		LIMIT ? OFFSET ?
	`, whereClause, sortColumn, sortOrder)

	queryArgs := append(args, req.PageSize, req.Offset)

	var ordersRaw []orderRow
	if err := r.db.SelectContext(ctx, &ordersRaw, dataQuery, queryArgs...); err != nil {
		return nil, 0, err
	}

	// モデルへの変換
	orders := make([]model.Order, 0, len(ordersRaw))
	for _, o := range ordersRaw {
		orders = append(orders, model.Order{
			OrderID:       int64(o.OrderID),
			ProductID:     o.ProductID,
			ProductName:   o.ProductName,
			ShippedStatus: o.ShippedStatus,
			CreatedAt:     o.CreatedAt.Time,
			ArrivedAt:     o.ArrivedAt,
		})
	}

	return orders, total, nil
}
