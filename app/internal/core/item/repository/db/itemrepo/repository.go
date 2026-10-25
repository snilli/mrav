package itemrepo

import (
	"context"

	"orm"
	"orm/ent"
	entitem "orm/ent/item"

	"app/internal/core/item"
	"app/internal/core/item/domain"
	monster "app/internal/core/monster/domain"
	"app/internal/infrastructure/memo"
)

type Repository struct {
	client *orm.Client
	items  memo.Cache[monster.Version, []domain.Item]
}

func New(client *orm.Client) item.ItemRepository {
	return &Repository{client: client}
}

func (r *Repository) ListItems(ctx context.Context, version monster.Version) ([]domain.Item, error) {
	return r.items.Get(version, func() ([]domain.Item, error) {
		rows, err := r.client.Item.Query().Where(entitem.VersionEQ(entitem.Version(version))).Order(ent.Asc(entitem.FieldCode)).All(ctx)
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
