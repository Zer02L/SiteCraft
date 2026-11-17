package data

import (
	"context"
	"encoding/json"
	"fmt"

	"sitecraft/internal/db"
	"sitecraft/internal/models"

	"github.com/google/uuid"
)

// GetPageBlocks returns active page blocks ordered by index.
func (r *Repository) GetPageBlocks(ctx context.Context) ([]models.PageBlock, error) {
	c := r.Cache()
	if cached, ok := c.GetBlocks(); ok {
		return cached, nil
	}

	rows, err := r.queries.ListPageBlocks(ctx, "")
	if err != nil {
		return nil, err
	}

	blocks := make([]models.PageBlock, len(rows))
	for i, row := range rows {
		var payload map[string]interface{}
		if row.Payload != "" {
			_ = json.Unmarshal([]byte(row.Payload), &payload)
		}
		blocks[i] = models.PageBlock{
			ID:      row.ID,
			PageID:  row.PageID,
			Type:    row.Type,
			Title:   row.Title,
			Content: row.Content,
			Order:   int(row.OrderIndex),
			Active:  row.Active == 1,
			Payload: payload,
		}
	}
	c.SetBlocks(blocks)
	return blocks, nil
}

// GetAllPageBlocks returns all page blocks (both active and inactive) ordered by index for admin panel.
func (r *Repository) GetAllPageBlocks(ctx context.Context) ([]models.PageBlock, error) {
	rows, err := r.queries.ListAllPageBlocks(ctx)
	if err != nil {
		return nil, err
	}

	blocks := make([]models.PageBlock, len(rows))
	for i, row := range rows {
		var payload map[string]interface{}
		if row.Payload != "" {
			_ = json.Unmarshal([]byte(row.Payload), &payload)
		}
		blocks[i] = models.PageBlock{
			ID:      row.ID,
			PageID:  row.PageID,
			Type:    row.Type,
			Title:   row.Title,
			Content: row.Content,
			Order:   int(row.OrderIndex),
			Active:  row.Active == 1,
			Payload: payload,
		}
	}
	return blocks, nil
}

// GetPageBlockByID retrieves a page block by ID.
func (r *Repository) GetPageBlockByID(ctx context.Context, id string) (models.PageBlock, error) {
	row, err := r.queries.GetPageBlockByID(ctx, id)
	if err != nil {
		return models.PageBlock{}, err
	}
	var payload map[string]interface{}
	if row.Payload != "" {
		_ = json.Unmarshal([]byte(row.Payload), &payload)
	}
	return models.PageBlock{
		ID:      row.ID,
		PageID:  row.PageID,
		Type:    row.Type,
		Title:   row.Title,
		Content: row.Content,
		Order:   int(row.OrderIndex),
		Active:  row.Active == 1,
		Payload: payload,
	}, nil
}

// CreatePageBlock inserts a page block and invalidates blocks cache.
func (r *Repository) CreatePageBlock(ctx context.Context, block models.PageBlock) (models.PageBlock, error) {
	if block.ID == "" {
		block.ID = uuid.New().String()
	}
	payloadBytes, _ := json.Marshal(block.Payload)
	act := int64(1)
	if !block.Active {
		act = 0
	}
	_, err := r.queries.CreatePageBlock(ctx, db.CreatePageBlockParams{
		ID:         block.ID,
		PageID:     block.PageID,
		Type:       block.Type,
		Title:      block.Title,
		Content:    block.Content,
		OrderIndex: int64(block.Order),
		Active:     act,
		Payload:    string(payloadBytes),
	})
	if err != nil {
		return block, err
	}
	r.Cache().InvalidateBlocks()
	return block, nil
}

// CreatePage inserts a new page into the pages table.
func (r *Repository) CreatePage(ctx context.Context, slug, title, seoTitle, seoDescription string, isPublished int64) (db.Page, error) {
	return r.queries.CreatePage(ctx, db.CreatePageParams{
		ID:             uuid.New().String(),
		Slug:           slug,
		Title:          title,
		SeoTitle:       seoTitle,
		SeoDescription: seoDescription,
		IsPublished:    isPublished,
	})
}

// TogglePageBlock toggles block visibility and invalidates blocks cache.
func (r *Repository) TogglePageBlock(ctx context.Context, id string) error {
	block, err := r.queries.GetPageBlockByID(ctx, id)
	if err != nil {
		return err
	}
	newActive := int64(1)
	if block.Active == 1 {
		newActive = 0
	}
	_, err = r.queries.UpdatePageBlock(ctx, db.UpdatePageBlockParams{
		ID:         block.ID,
		Type:       block.Type,
		Title:      block.Title,
		Content:    block.Content,
		OrderIndex: block.OrderIndex,
		Active:     newActive,
		Payload:    block.Payload,
	})
	if err != nil {
		return err
	}
	r.Cache().InvalidateBlocks()
	return nil
}

// UpdatePageBlock updates title, content, type, or payload of a page block and invalidates cache.
func (r *Repository) UpdatePageBlock(ctx context.Context, block models.PageBlock) error {
	existing, err := r.queries.GetPageBlockByID(ctx, block.ID)
	if err != nil {
		return err
	}
	payloadStr := existing.Payload
	if block.Payload != nil {
		if b, err := json.Marshal(block.Payload); err == nil {
			payloadStr = string(b)
		}
	}
	active := int64(0)
	if block.Active {
		active = 1
	}
	blockType := block.Type
	if blockType == "" {
		blockType = existing.Type
	}
	title := block.Title
	if title == "" {
		title = existing.Title
	}

	_, err = r.queries.UpdatePageBlock(ctx, db.UpdatePageBlockParams{
		ID:         block.ID,
		Type:       blockType,
		Title:      title,
		Content:    block.Content,
		OrderIndex: existing.OrderIndex,
		Active:     active,
		Payload:    payloadStr,
	})
	if err != nil {
		return err
	}
	r.Cache().InvalidateBlocks()
	return nil
}

// SwapPageBlockOrder swaps order of two blocks atomically inside a transaction and invalidates cache.
func (r *Repository) SwapPageBlockOrder(ctx context.Context, id1 string, order1 int64, id2 string, order2 int64) error {
	tx, err := r.rawDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	if err := qtx.UpdatePageBlockOrder(ctx, db.UpdatePageBlockOrderParams{
		ID:         id1,
		OrderIndex: order2,
	}); err != nil {
		return fmt.Errorf("failed to update block order %s: %w", id1, err)
	}
	if err := qtx.UpdatePageBlockOrder(ctx, db.UpdatePageBlockOrderParams{
		ID:         id2,
		OrderIndex: order1,
	}); err != nil {
		return fmt.Errorf("failed to update block order %s: %w", id2, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.Cache().InvalidateBlocks()
	return nil
}

// DeletePageBlock deletes a block by ID and invalidates cache.
func (r *Repository) DeletePageBlock(ctx context.Context, id string) error {
	if err := r.queries.DeletePageBlock(ctx, id); err != nil {
		return err
	}
	r.Cache().InvalidateBlocks()
	return nil
}
