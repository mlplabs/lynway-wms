package repository

import (
	"context"

	"github.com/mlplabs/lynway-wms/internal/model"
)

func (r *pgRepository) GetContainerById(ctx context.Context, id int) (*model.Container, error) {
	q := `SELECT id, name, type, parent_id, root_id, depth FROM containers WHERE id = $1`
	ct := model.Container{}

	err := r.conn.QueryRowContext(ctx, q, id).Scan(&ct.Id, &ct.Name, &ct.Type, &ct.Pid, &ct.Rid, &ct.Depth)
	if err != nil {
		return nil, err
	}

	return &ct, nil
}

// CreateContainer cоздает контейнер (в базе, получить новый ID)
// У нового контейнера нет родительского элемента, глубина 0 и ID корневого контейнера равно самому себе
func (r *pgRepository) CreateContainer(ctx context.Context, name *string, containerType *int) (*model.Container, error) {
	var id int

	q := `INSERT INTO containers (id, name, type, parent_id, root_id, depth)
			SELECT sub.row_id, $1, $2, 0, sub.row_id, 0 FROM
    			(SELECT nextval('containers_id_seq'::regclass) as row_id) AS sub
			RETURNING id`

	err := r.conn.QueryRowContext(ctx, q, name, containerType).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.Container{Id: id, Pid: 0, Name: name, Rid: id}, nil
}

// PutContainer положить контейнер в другой контейнер
// Содержимое контейнера не меняется
func (r *pgRepository) PutContainer(ctx context.Context, this *model.Container, target *model.Container) error {
	q := `UPDATE containers 
				SET parent_id = $2, root_id = $3, depth = $4
				WHERE id = $1`
	_, err := r.conn.ExecContext(ctx, q, this.Id, target.Id, target.Rid, target.Depth+1)
	return err
}

// GetContainer достать контейнер
// Теряет родителей, root_id, depth...
func (r *pgRepository) GetContainer(ctx context.Context, this *model.Container) error {
	q := `UPDATE containers 
				SET parent_id = $2, root_id = $3, depth = $4
				WHERE id = $1`
	_, err := r.conn.ExecContext(ctx, q, this.Id, 0, this.Id, 0)
	return err
}

// GetChildrenContainers получить вложенные контейнеры
// Отбираем контейнеры по root_id с глубиной ниже
func (r *pgRepository) GetChildrenContainers(ctx context.Context, this *model.Container) ([]model.Container, error) {
	q := `SELECT id, name, type, parent_id, root_id, depth FROM containers WHERE root_id = $1 AND depth > $2`
	rows, err := r.conn.QueryContext(ctx, q, this.Rid, this.Depth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []model.Container

	for rows.Next() {
		container := model.Container{
			Pid: this.Id,
		}
		if err := rows.Scan(&container.Id, &container.Name, &container.Type, &container.Pid, &container.Rid, &container.Depth); err != nil {
			return nil, err
		}

		containers = append(containers, container)
	}
	return containers, nil
}

// GetParentContainers получить все родительские контейнеры
// Отбираем контейнеры по root_id с глубиной выше
func (r *pgRepository) GetParentContainers(ctx context.Context, this *model.Container) ([]model.Container, error) {
	q := `SELECT id, name, type, parent_id, root_id, depth FROM containers WHERE root_id = $1 AND depth < $2`
	rows, err := r.conn.QueryContext(ctx, q, this.Rid, this.Depth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []model.Container

	for rows.Next() {
		container := model.Container{
			Pid: this.Id,
		}
		if err := rows.Scan(&container.Id, &container.Name, &container.Type, &container.Pid, &container.Rid, &container.Depth); err != nil {
			return nil, err
		}

		containers = append(containers, container)
	}
	return containers, nil
}

// GetParentContainer получить родительский контейнер
func (r *pgRepository) GetParentContainer(ctx context.Context, this *model.Container) (*model.Container, error) {
	q := `SELECT id, name, type, parent_id, root_id, depth FROM containers WHERE parent_id = $1`
	parent := model.Container{}
	err := r.conn.QueryRowContext(ctx, q, this.Pid).Scan(&parent.Id, &parent.Name, &parent.Type, &parent.Pid, &parent.Rid, &parent.Depth)
	if err != nil {
		return nil, err
	}
	return &parent, nil
}

func (r *pgRepository) GetContainerProducts(ctx context.Context, this *model.Container) ([]model.RemainingProduct, error) {
	q := `SELECT
				rst.container_id,
				rst.product_id,
				coalesce(p.name, '') AS product_name,
				coalesce(p.item_number, '') AS item_number,
				coalesce(p.manufacturer_id, 0) AS manufacturer_id,
				coalesce(m.name, '') AS manufacturer_name,
				rst.quantity
			FROM (
				SELECT s.container_id, s.product_id, SUM(s.quantity) AS quantity
					FROM storage_feature s
					LEFT JOIN products p ON s.product_id = p.id
					WHERE container_id IN (SELECT id
										 FROM containers ct
										 WHERE ct.root_id = $1 AND ct.depth >= $2)
				GROUP BY s.container_id, s.product_id) rst
			LEFT JOIN products p ON rst.product_id = p.id
			LEFT JOIN manufacturers m ON p.manufacturer_id = m.id`
	rows, err := r.conn.QueryContext(ctx, q, this.Rid, this.Depth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]model.RemainingProduct, 0)

	for rows.Next() {
		p := model.RemainingProduct{
			Product: model.Product{Manufacturer: &model.Manufacturer{}},
		}

		err = rows.Scan(&p.ContainerId, &p.Id, &p.Name, &p.ItemNumber, &p.Manufacturer.ID, &p.Manufacturer.Name, &p.Quantity)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}
	return res, nil
}
