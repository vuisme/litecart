package queries

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vuisme/litecart/internal/models"
	"github.com/vuisme/litecart/pkg/errors"
	"github.com/vuisme/litecart/pkg/security"
)

// PageQueries is a struct that embeds a pointer to an sql.DB.
// This allows for direct access to database methods on the PageQueries struct,
// effectively extending it with all the methods of *sql.DB.
type PageQueries struct {
	*sql.DB
}

// IsPage checks if a page with the given slug exists in the database.
// It uses the context provided for any query-related timeouts or cancellations.
func (q *PageQueries) IsPage(ctx context.Context, slug string) bool {
	var exists bool
	err := q.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM page WHERE slug = ?)`, slug).Scan(&exists)
	return err == nil && exists
}

// ListPages retrieves a list of pages from the database.
// It filters out private pages unless `private` is set to true,
// and can also filter by a list of page IDs if provided.
func (q *PageQueries) ListPages(ctx context.Context, private bool, idList ...string) ([]models.Page, error) {
	pages := []models.Page{}

	query := `SELECT id, name, slug, position, active, seo, strftime('%s', created), strftime('%s', updated) FROM page`
	if !private {
		query = query + ` WHERE active = 1`
	}

	stmt, err := q.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var page models.Page
		var seo sql.NullString
		var updated sql.NullInt64

		err := rows.Scan(&page.ID, &page.Name, &page.Slug, &page.Position, &page.Active, &seo, &page.Created, &updated)
		if err != nil {
			return nil, err
		}

		if updated.Valid {
			page.Updated = updated.Int64
		}

		if seo.Valid {
			if err = json.Unmarshal([]byte(seo.String), &page.Seo); err != nil {
				return nil, err
			}
		}

		pages = append(pages, page)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

// Page retrieves a single page from the database based on its slug.
func (q *PageQueries) Page(ctx context.Context, slug string) (*models.Page, error) {
	page := models.Page{
		Slug: slug,
	}

	var content, seo sql.NullString
	query := `SELECT id, name, content, active, seo FROM page WHERE slug = ?`
	err := q.DB.QueryRowContext(ctx, query, slug).Scan(&page.ID, &page.Name, &content, &page.Active, &seo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrPageNotFound
		}
		return nil, err
	}

	page.Content = &content.String
	if seo.Valid {
		if err = json.Unmarshal([]byte(seo.String), &page.Seo); err != nil {
			return nil, err
		}
	}

	return &page, nil
}

// AddPage inserts a new page into the database and returns the created page or an error.
func (q *PageQueries) AddPage(ctx context.Context, page *models.Page) (*models.Page, error) {
	page.ID = security.RandomString()
	page.Active = false

	query := `INSERT INTO page (id, name, slug, position) VALUES (?, ?, ?, ?) RETURNING strftime('%s', created)`
	stmt, err := q.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, query, page.ID, page.Name, page.Slug, page.Position).Scan(&page.Created)
	if err != nil {
		return nil, err
	}

	return page, nil
}

// UpdatePage updates the details of a page in the database.
func (q *PageQueries) UpdatePage(ctx context.Context, page *models.Page) error {
	seo, err := json.Marshal(page.Seo)
	if err != nil {
		return err
	}

	query := `UPDATE page SET name = ?, slug = ?, position = ?, seo = ?, updated = datetime('now') WHERE id = ?`
	_, err = q.DB.ExecContext(ctx, query, page.Name, page.Slug, page.Position, seo, page.ID)
	return err
}

// DeletePage method belongs to the PageQueries struct. This method is responsible for deleting a page from the database.
func (q *PageQueries) DeletePage(ctx context.Context, id string) error {
	query := `DELETE FROM page WHERE id = ?`
	_, err := q.DB.ExecContext(ctx, query, id)
	return err
}

// UpdatePageContent updates the content of an existing page in the database.
func (q *PageQueries) UpdatePageContent(ctx context.Context, page *models.Page) error {
	query := `UPDATE page SET content = ?, updated = datetime('now') WHERE id = ? `
	_, err := q.DB.ExecContext(ctx, query, page.Content, page.ID)
	return err
}

// UpdatePageActive toggles the active status of a page with the given ID.
// It updates the 'active' field to its logical negation (i.e., if it was true, it becomes false and vice versa)
func (q *ProductQueries) UpdatePageActive(ctx context.Context, id string) error {
	query := `UPDATE page SET active = NOT active, updated = datetime('now') WHERE id = ?`
	_, err := q.DB.ExecContext(ctx, query, id)
	return err
}
