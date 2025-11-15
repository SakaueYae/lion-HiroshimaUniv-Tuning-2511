package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service/utils"
	"context"
	"log"
)

type RobotService struct {
	store *repository.Store
}

func NewRobotService(store *repository.Store) *RobotService {
	return &RobotService{store: store}
}

// 注意：このメソッドは、現在、ordersテーブルのshipped_statusが"shipping"になっている注文"全件"を対象に配送計画を立てます。
// 注文の取得件数を制限した場合、ペナルティの対象になります。
func (s *RobotService) GenerateDeliveryPlan(ctx context.Context, robotID string, capacity int) (*model.DeliveryPlan, error) {
	var plan model.DeliveryPlan

	err := utils.WithTimeout(ctx, func(ctx context.Context) error {
		return s.store.ExecTx(ctx, func(txStore *repository.Store) error {
			orders, err := txStore.OrderRepo.GetShippingOrders(ctx)
			if err != nil {
				return err
			}
			plan, err = selectOrdersForDelivery(ctx, orders, robotID, capacity)
			if err != nil {
				return err
			}
			if len(plan.Orders) > 0 {
				orderIDs := make([]int64, len(plan.Orders))
				for i, order := range plan.Orders {
					orderIDs[i] = order.OrderID
				}

				if err := txStore.OrderRepo.UpdateStatuses(ctx, orderIDs, "delivering"); err != nil {
					return err
				}
				log.Printf("Updated status to 'delivering' for %d orders", len(orderIDs))
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *RobotService) UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) error {
	return utils.WithTimeout(ctx, func(ctx context.Context) error {
		return s.store.OrderRepo.UpdateStatuses(ctx, []int64{orderID}, newStatus)
	})
}

func selectOrdersForDelivery(ctx context.Context, orders []model.Order, robotID string, robotCapacity int) (model.DeliveryPlan, error) {
	n := len(orders)

	// エッジケースの処理
	if n == 0 || robotCapacity <= 0 {
		return model.DeliveryPlan{
			RobotID:     robotID,
			TotalWeight: 0,
			TotalValue:  0,
			Orders:      []model.Order{},
		}, nil
	}

	// 動的計画法テーブルの初期化
	// dp[i][w] = 最初のi個の注文を考慮して、重量w以下での最大価値
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, robotCapacity+1)
	}

	checkEvery := 100 // コンテキストチェックの間隔

	// 各注文について動的計画法を実行
	for i := 1; i <= n; i++ {
		// 定期的にコンテキストキャンセルをチェック（最初の反復も含む）
		if i == 1 || i%checkEvery == 0 {
			select {
			case <-ctx.Done():
				return model.DeliveryPlan{}, ctx.Err()
			default:
			}
		}

		order := orders[i-1]
		weight := order.Weight
		value := order.Value

		for w := 0; w <= robotCapacity; w++ {
			// この注文を含めない場合
			dp[i][w] = dp[i-1][w]

			// この注文を含める場合（重量が許す限り）
			if w >= weight {
				includeValue := dp[i-1][w-weight] + value
				if includeValue > dp[i][w] {
					dp[i][w] = includeValue
				}
			}
		}
	}

	// 最大価値を取得
	maxValue := dp[n][robotCapacity]

	// 選択された注文を復元
	var selectedOrders []model.Order
	currentWeight := robotCapacity

	for i := n; i > 0 && currentWeight > 0; i-- {
		// dp[i][currentWeight] != dp[i-1][currentWeight] なら、注文i-1を選択している
		if dp[i][currentWeight] != dp[i-1][currentWeight] {
			order := orders[i-1]
			selectedOrders = append(selectedOrders, order)
			currentWeight -= order.Weight
		}
	}

	// 復元は逆順なので、元の順序に戻す
	for i, j := 0, len(selectedOrders)-1; i < j; i, j = i+1, j-1 {
		selectedOrders[i], selectedOrders[j] = selectedOrders[j], selectedOrders[i]
	}

	// 実際の総重量を計算
	var totalWeight int
	for _, o := range selectedOrders {
		totalWeight += o.Weight
	}

	return model.DeliveryPlan{
		RobotID:     robotID,
		TotalWeight: totalWeight,
		TotalValue:  maxValue,
		Orders:      selectedOrders,
	}, nil
}
