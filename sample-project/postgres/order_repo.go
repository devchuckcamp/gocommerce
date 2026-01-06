package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/devchuckcamp/gocommerce/orders"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*orders.Order, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, order_number, user_id, status,
			subtotal_amount, subtotal_currency,
			discount_amount, COALESCE(discount_currency, subtotal_currency),
			tax_amount, COALESCE(tax_currency, subtotal_currency),
			shipping_amount, COALESCE(shipping_currency, subtotal_currency),
			total_amount, COALESCE(total_currency, subtotal_currency),
			COALESCE(payment_method_id,''),
			COALESCE(notes,''),
			COALESCE(ip_address,''),
			COALESCE(user_agent,''),
			COALESCE(shipping_address, '{}'::jsonb),
			COALESCE(billing_address, '{}'::jsonb),
			created_at, updated_at, completed_at, canceled_at
		FROM orders
		WHERE id = $1
	`, id)

	var o orders.Order
	var status string
	var subtotalAmt, discountAmt, taxAmt, shippingAmt, totalAmt int64
	var subtotalCur, discountCur, taxCur, shippingCur, totalCur string
	var shippingAddr, billingAddr []byte
	var completedAt, canceledAt sql.NullTime

	if err := row.Scan(
		&o.ID,
		&o.OrderNumber,
		&o.UserID,
		&status,
		&subtotalAmt,
		&subtotalCur,
		&discountAmt,
		&discountCur,
		&taxAmt,
		&taxCur,
		&shippingAmt,
		&shippingCur,
		&totalAmt,
		&totalCur,
		&o.PaymentMethodID,
		&o.Notes,
		&o.IPAddress,
		&o.UserAgent,
		&shippingAddr,
		&billingAddr,
		&o.CreatedAt,
		&o.UpdatedAt,
		&completedAt,
		&canceledAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, orders.ErrOrderNotFound
		}
		return nil, err
	}

	o.Status = orders.OrderStatus(status)
	o.Subtotal, _ = moneyFrom(subtotalAmt, subtotalCur)
	o.DiscountTotal, _ = moneyFrom(discountAmt, discountCur)
	o.TaxTotal, _ = moneyFrom(taxAmt, taxCur)
	o.ShippingTotal, _ = moneyFrom(shippingAmt, shippingCur)
	o.Total, _ = moneyFrom(totalAmt, totalCur)
	o.CompletedAt = scanNullTime(completedAt)
	o.CanceledAt = scanNullTime(canceledAt)
	_ = fromJSONB(shippingAddr, &o.ShippingAddress)
	_ = fromJSONB(billingAddr, &o.BillingAddress)

	items, err := r.findItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *OrderRepository) FindByOrderNumber(ctx context.Context, orderNumber string) (*orders.Order, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id FROM orders WHERE order_number = $1`, orderNumber)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, orders.ErrOrderNotFound
		}
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *OrderRepository) FindByUserID(ctx context.Context, userID string, filter orders.OrderFilter) ([]*orders.Order, error) {
	q := `SELECT id FROM orders WHERE user_id = $1`
	args := []any{userID}

	if filter.Status != nil {
		args = append(args, string(*filter.Status))
		q += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if filter.DateFrom != nil {
		args = append(args, *filter.DateFrom)
		q += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if filter.DateTo != nil {
		args = append(args, *filter.DateTo)
		q += fmt.Sprintf(" AND created_at <= $%d", len(args))
	}

	q += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		q += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if filter.Offset > 0 {
		args = append(args, filter.Offset)
		q += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]*orders.Order, 0, len(ids))
	for _, id := range ids {
		o, err := r.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

func (r *OrderRepository) Save(ctx context.Context, o *orders.Order) error {
	if o == nil {
		return errors.New("order is nil")
	}
	if o.ID == "" {
		return errors.New("order ID is required")
	}
	if o.OrderNumber == "" {
		return errors.New("order number is required")
	}

	shipAddr, err := toJSONB(o.ShippingAddress)
	if err != nil {
		return err
	}
	billAddr, err := toJSONB(o.BillingAddress)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (
			id, order_number, user_id, status,
			subtotal_amount, subtotal_currency,
			discount_amount, tax_amount, shipping_amount, total_amount,
			discount_currency, tax_currency, shipping_currency, total_currency,
			payment_method_id, notes, ip_address, user_agent,
			shipping_address, billing_address,
			created_at, updated_at, completed_at, canceled_at
		) VALUES (
			$1,$2,$3,$4,
			$5,$6,
			$7,$8,$9,$10,
			$11,$12,$13,$14,
			NULLIF($15,''),$16,NULLIF($17,''),$18,
			$19,$20,
			COALESCE($21, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP, $22, $23
		)
		ON CONFLICT (id) DO UPDATE SET
			order_number = EXCLUDED.order_number,
			user_id = EXCLUDED.user_id,
			status = EXCLUDED.status,
			subtotal_amount = EXCLUDED.subtotal_amount,
			subtotal_currency = EXCLUDED.subtotal_currency,
			discount_amount = EXCLUDED.discount_amount,
			tax_amount = EXCLUDED.tax_amount,
			shipping_amount = EXCLUDED.shipping_amount,
			total_amount = EXCLUDED.total_amount,
			discount_currency = EXCLUDED.discount_currency,
			tax_currency = EXCLUDED.tax_currency,
			shipping_currency = EXCLUDED.shipping_currency,
			total_currency = EXCLUDED.total_currency,
			payment_method_id = EXCLUDED.payment_method_id,
			notes = EXCLUDED.notes,
			ip_address = EXCLUDED.ip_address,
			user_agent = EXCLUDED.user_agent,
			shipping_address = EXCLUDED.shipping_address,
			billing_address = EXCLUDED.billing_address,
			completed_at = EXCLUDED.completed_at,
			canceled_at = EXCLUDED.canceled_at,
			updated_at = CURRENT_TIMESTAMP
	`,
		o.ID,
		o.OrderNumber,
		o.UserID,
		string(o.Status),
		o.Subtotal.Amount,
		o.Subtotal.Currency,
		o.DiscountTotal.Amount,
		o.TaxTotal.Amount,
		o.ShippingTotal.Amount,
		o.Total.Amount,
		o.DiscountTotal.Currency,
		o.TaxTotal.Currency,
		o.ShippingTotal.Currency,
		o.Total.Currency,
		o.PaymentMethodID,
		o.Notes,
		o.IPAddress,
		o.UserAgent,
		shipAddr,
		billAddr,
		nullTime(o.CreatedAt),
		o.CompletedAt,
		o.CanceledAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, o.ID)
	if err != nil {
		return err
	}

	for _, item := range o.Items {
		attrs, err := toJSONB(item.Attributes)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO order_items (
				id, order_id, product_id, variant_id, sku, name,
				unit_price_amount, unit_price_currency,
				quantity,
				discount_amount, discount_currency,
				tax_amount, tax_currency,
				total_amount, total_currency,
				attributes
			) VALUES (
				$1,$2,$3,$4,$5,$6,
				$7,$8,
				$9,
				$10,$11,
				$12,$13,
				$14,$15,
				$16
			)
		`,
			item.ID,
			o.ID,
			item.ProductID,
			item.VariantID,
			item.SKU,
			item.Name,
			item.UnitPrice.Amount,
			item.UnitPrice.Currency,
			item.Quantity,
			item.DiscountAmount.Amount,
			item.DiscountAmount.Currency,
			item.TaxAmount.Amount,
			item.TaxAmount.Currency,
			item.Total.Amount,
			item.Total.Currency,
			attrs,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *OrderRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, id)
	return err
}

func (r *OrderRepository) findItems(ctx context.Context, orderID string) ([]orders.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, variant_id, sku, name,
			unit_price_amount, unit_price_currency,
			quantity,
			discount_amount, discount_currency,
			tax_amount, tax_currency,
			total_amount, total_currency,
			COALESCE(attributes, '{}'::jsonb)
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`, orderID)
	if err != nil {
		// If the table doesn't exist yet (older schema), treat as no items.
		msg := err.Error()
		if strings.Contains(msg, "order_items") && strings.Contains(msg, "does not exist") {
			return []orders.OrderItem{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	items := make([]orders.OrderItem, 0)
	for rows.Next() {
		var it orders.OrderItem
		var variantID sql.NullString
		var unitAmt, discAmt, taxAmt, totalAmt int64
		var unitCur, discCur, taxCur, totalCur string
		var attrsRaw []byte

		if err := rows.Scan(
			&it.ID,
			&it.ProductID,
			&variantID,
			&it.SKU,
			&it.Name,
			&unitAmt,
			&unitCur,
			&it.Quantity,
			&discAmt,
			&discCur,
			&taxAmt,
			&taxCur,
			&totalAmt,
			&totalCur,
			&attrsRaw,
		); err != nil {
			return nil, err
		}

		if variantID.Valid {
			v := variantID.String
			it.VariantID = &v
		}
		it.UnitPrice, _ = moneyFrom(unitAmt, unitCur)
		it.DiscountAmount, _ = moneyFrom(discAmt, discCur)
		it.TaxAmount, _ = moneyFrom(taxAmt, taxCur)
		it.Total, _ = moneyFrom(totalAmt, totalCur)
		_ = fromJSONB(attrsRaw, &it.Attributes)

		items = append(items, it)
	}
	return items, rows.Err()
}
