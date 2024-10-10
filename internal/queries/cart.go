package queries

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/vuisme/litecart/internal/models"
	"github.com/vuisme/litecart/pkg/errors"
	"github.com/vuisme/litecart/pkg/litepay"
)

// CartQueries is a struct that embeds a pointer to an sql.DB.
// This allows for direct access to all the methods of sql.DB through CartQueries.
type CartQueries struct {
	*sql.DB
}

// PaymentList retrieves the status of different payment methods from the database.
func (q *CartQueries) PaymentList(ctx context.Context) (map[string]bool, error) {
	payments := map[string]bool{}
	keys := []any{
		"stripe_active", "paypal_active", "spectrocoin_active",
	}

	query := fmt.Sprintf("SELECT key, value FROM setting WHERE key IN (%s)", strings.Repeat("?, ", len(keys)-1)+"?")
	rows, err := q.DB.QueryContext(ctx, query, keys...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}

		vBool, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		name := strings.ReplaceAll(key, "_active", "")
		payments[name] = vBool
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

// Carts retrieves a list of carts from the database.
func (q *CartQueries) Carts(ctx context.Context) ([]*models.Cart, error) {
	carts := []*models.Cart{}

	query := `
	SELECT 
		id, 
		email, 
		amount_total,
		currency,
		payment_id,
		payment_status,
		payment_system,
		strftime('%s', created),
		strftime('%s', updated)
	FROM cart
`

	rows, err := q.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var email, paymentID sql.NullString
		var updated sql.NullInt64
		cart := &models.Cart{}

		err := rows.Scan(
			&cart.ID,
			&email,
			&cart.AmountTotal,
			&cart.Currency,
			&paymentID,
			&cart.PaymentStatus,
			&cart.PaymentSystem,
			&cart.Created,
			&updated,
		)
		if err != nil {
			return nil, err
		}

		cart.Email = email.String
		cart.PaymentID = paymentID.String
		if updated.Valid {
			cart.Updated = updated.Int64
		}

		carts = append(carts, cart)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return carts, nil
}

// Cart retrieves a cart from the database using the provided cartId.
func (q *CartQueries) Cart(ctx context.Context, cartId string) (*models.Cart, error) {
	query := `
	SELECT 
    id, 
    email, 
    amount_total,
    currency,
    payment_id,
    payment_status,
    strftime('%s', created),
    strftime('%s', updated)
	FROM cart
	WHERE id = ?
	`

	var email, paymentID sql.NullString
	var created, updated sql.NullInt64
	cart := &models.Cart{}

	err := q.DB.QueryRowContext(ctx, query, cartId).
		Scan(
			&cart.ID,
			&email,
			&cart.AmountTotal,
			&cart.Currency,
			&paymentID,
			&cart.PaymentStatus,
			&created,
			&updated,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrProductNotFound
		}
		return nil, err
	}

	cart.Email = email.String
	cart.PaymentID = paymentID.String
	if created.Valid {
		cart.Created = created.Int64
	}

	if updated.Valid {
		cart.Updated = updated.Int64
	}

	return cart, nil
}

// AddCart inserts a new cart into the database.
func (q *CartQueries) AddCart(ctx context.Context, cart *models.Cart) error {
	byteCart, err := json.Marshal(cart.Cart)
	if err != nil {
		return err
	}

	query := `INSERT INTO cart (id, email, cart, amount_total, currency, payment_status, payment_system) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = q.DB.ExecContext(ctx, query, cart.ID, cart.Email, string(byteCart), cart.AmountTotal, cart.Currency, cart.PaymentStatus, cart.PaymentSystem)
	return err
}

// UpdateCart updates the cart details in the database.
func (q *CartQueries) UpdateCart(ctx context.Context, cart *models.Cart) error {
	var (
		args []interface{}
		sql  strings.Builder
	)

	sql.WriteString("UPDATE cart SET ")

	if cart.PaymentID != "" {
		sql.WriteString("payment_id = ?, ")
		args = append(args, cart.PaymentID)
	}

	if cart.PaymentStatus != "" {
		sql.WriteString("payment_status = ?, ")
		args = append(args, cart.PaymentStatus)
	}

	sql.WriteString("updated = datetime('now') WHERE id = ?")
	args = append(args, cart.ID)

	_, err := q.DB.ExecContext(ctx, sql.String(), args...)
	return err
}

// CartLetterPayment is ...
func (q *CartQueries) CartLetterPayment(ctx context.Context, email, amountPayment, paymentURL string) (*models.MessageMail, error) {
	mailLetter, err := db.GetSettingByKey(ctx, "site_name", "mail_letter_payment")
	if err != nil {
		return nil, err
	}
	letterTemplate := models.Letter{}
	if err := json.Unmarshal([]byte(mailLetter["mail_letter_payment"].Value.(string)), &letterTemplate); err != nil {
		return nil, err
	}

	mail := &models.MessageMail{
		To:     email,
		Letter: letterTemplate,
		Data: map[string]string{
			"Payment_URL":    paymentURL,
			"Site_Name":      mailLetter["site_name"].Value.(string),
			"Amount_Payment": amountPayment,
		},
	}

	return mail, nil
}

// CartLetterPurchase is ...
func (q *CartQueries) CartLetterPurchase(ctx context.Context, cartID string) (*models.MessageMail, error) {
	mail := &models.MessageMail{}

	// Fetch the email, cart information, and 'email' setting in one query.
	var cartJSON string
	err := q.QueryRowContext(ctx, `
        SELECT email, cart
        FROM cart
        WHERE payment_status = ? AND id = ?
    `, litepay.PAID, cartID).Scan(&mail.To, &cartJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrPageNotFound
		}
		return nil, err
	}

	// Unmarshal the products from the cart JSON.
	products := []models.CartProduct{}
	if err := json.Unmarshal([]byte(cartJSON), &products); err != nil {
		return nil, err
	}

	// Begin a transaction.
	tx, err := q.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	keys := []models.Data{}
	files := []models.File{}
	for _, cart := range products {
		var digitalType string
		err := tx.QueryRowContext(ctx, `SELECT digital FROM product WHERE id = ?`, cart.ProductID).Scan(&digitalType)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.ErrPageNotFound
			}
			return nil, err
		}

		switch digitalType {
		case "file":
			rows, err := tx.QueryContext(ctx, `SELECT id, name, ext, orig_name FROM digital_file WHERE product_id = ?`, cart.ProductID)
			if err != nil {
				return nil, err
			}
			for rows.Next() {
				file := models.File{}
				if err := rows.Scan(&file.ID, &file.Name, &file.Ext, &file.OrigName); err != nil {
					rows.Close()
					return nil, err
				}
				files = append(files, file)
			}
			rows.Close()
		case "data":
			key := models.Data{}
			err := tx.QueryRowContext(ctx, `SELECT id, content FROM digital_data WHERE cart_id = ?`, cartID).Scan(&key.ID, &key.Content)
			if err != nil {
				if err == sql.ErrNoRows {
					err = tx.QueryRowContext(ctx, `SELECT id, content FROM digital_data WHERE cart_id IS NULL AND product_id = ? LIMIT 1`, cart.ProductID).Scan(&key.ID, &key.Content)
					if err != nil {
						if err == sql.ErrNoRows {
							return nil, errors.ErrPageNotFound
						}
						return nil, err
					}
					if _, err := tx.ExecContext(ctx, `UPDATE digital_data SET cart_id = ? WHERE id = ?`, cartID, key.ID); err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			}
			keys = append(keys, key)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Construct the purchases information.
	var purchases strings.Builder
	count := 1
	if len(keys) > 0 {
		purchases.WriteString("Keys:\n")
		for _, key := range keys {
			purchases.WriteString(fmt.Sprintf("%v: %s\n", count, key.Content))
			count++
		}
	}
	if len(files) > 0 {
		purchases.WriteString("Files:\n")
		for _, file := range files {
			purchases.WriteString(fmt.Sprintf("%v: %s\n", count, file.OrigName))
			count++
		}
	}

	// Fetch the 'mail_letter_purchase' setting value.
	mailLetter, err := db.GetSettingByKey(ctx, "email", "mail_letter_purchase")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(mailLetter["mail_letter_purchase"].Value.(string)), &mail.Letter); err != nil {
		return nil, err
	}

	mail.Data = map[string]string{
		"Purchases":   purchases.String(),
		"Admin_Email": mailLetter["email"].Value.(string),
	}
	mail.Files = files

	return mail, nil
}
