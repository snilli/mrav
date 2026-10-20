package itemrepo

import (
	"context"

	"orm"
	"orm/ent"
	entitem "orm/ent/item"

	"app/internal/core/item"
	"app/internal/core/item/domain"
	"app/internal/infrastructure/memo"
)

// Repository reads the game's list of items through Ent's query builder. The list never changes while the viewer runs, so it
// is read once and kept; what the pages are given is shared, so they must not write to it.
type Repository struct {
	client *orm.Client
	items  memo.Cache[struct{}, []domain.Item]
}

func New(client *orm.Client) item.ItemRepository {
	return &Repository{client: client}
}

func (r *Repository) ListItems(ctx context.Context) ([]domain.Item, error) {
	return r.items.Get(struct{}{}, func() ([]domain.Item, error) {
		rows, err := r.client.Item.Query().Order(ent.Asc(entitem.FieldCode)).All(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]domain.Item, len(rows))
		for i, it := range rows {
			out[i] = domain.Item{
				Code: it.Code, Name: it.Name, DesireName: it.DesireName, Kind: it.Kind, Description: it.Description,
				Effect: it.Effect, SalePrice: it.SalePrice, PurchasePrice: it.PurchasePrice,
			}
		}
		return out, nil
	})
}
