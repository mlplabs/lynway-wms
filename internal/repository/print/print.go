package print

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainErr "github.com/mlplabs/common-go-pkg/pkg/http/errors/custom"
	"github.com/mlplabs/lynway-wms/internal/model"
	"github.com/mlplabs/lynway-wms/pkg/dto"
)

type pgPrint struct {
	conn *sql.DB
}

func NewPrintRepository(conn *sql.DB) *pgPrint {
	return &pgPrint{
		conn: conn,
	}
}

func (r *pgPrint) GetAgent(ctx context.Context, token string) (*model.Agent, error) {
	res := model.Agent{}
	query := `SELECT id, name, token FROM agents WHERE token = $1;`
	err := r.conn.QueryRowContext(ctx, query, token).Scan(&res.Id, &res.Name, &res.Token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErr.NewErrorNoRows(err)
		}
		return nil, err
	}
	return &res, nil
}

func (r *pgPrint) GetPrintTaskWithLock(ctx context.Context, agentId int) (int, error) {
	var res int
	query := `UPDATE print_queue SET status = $3 WHERE id = (
			  	SELECT id  
					FROM print_queue 
           			WHERE printer_id IN (
						SELECT id FROM printers WHERE agent_id = $1
					) 
					AND status = $2
					ORDER BY id ASC LIMIT 1 
					FOR UPDATE SKIP LOCKED
			  ) RETURNING id`
	err := r.conn.QueryRowContext(ctx, query, agentId, int(dto.PrintTaskStatusPending), int(dto.PrintTaskStatusProcessing)).
		Scan(&res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domainErr.NewErrorNoRows(err)
		}
		return 0, err
	}
	return res, nil
}

func (r *pgPrint) GetPrintTaskById(ctx context.Context, taskId int) (*model.PrintQueueTask, error) {
	res := model.PrintQueueTask{
		Printer: &model.Printer{},
	}

	query := `SELECT q.id, q.print_type, q.entity_id, q.entity_type_id, q.printer_id, p.name, p.agent_id 
				FROM print_queue q
				LEFT JOIN printers p ON p.id = q.printer_id
				WHERE q.id = $1;`
	err := r.conn.QueryRowContext(ctx, query, taskId).
		Scan(&res.Id, &res.PrintType, &res.EntityId, &res.EntityTypeId,
			&res.Printer.Id, &res.Printer.Name, &res.Printer.AgentId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErr.NewErrorNoRows(err)
		}

		return nil, err
	}
	return &res, nil
}

// UpdateAgentInfo обновление информации об агенте и принтерах
func (r *pgPrint) UpdateAgentInfo(ctx context.Context, agentInfo *model.AgentInfo) error {
	var agentId, printerId int
	query := `INSERT INTO agents (name, token) VALUES ('unnamed', $1) 
                  	ON CONFLICT (token) DO UPDATE SET name = excluded.name
				RETURNING id`

	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	err = tx.QueryRowContext(ctx, query, agentInfo.Token).Scan(&agentId)
	if err != nil {
		tx.Rollback()
		return err
	}
	if agentInfo.Printers != nil {
		for _, prn := range agentInfo.Printers {
			queryPrn := `INSERT INTO printers (agent_id, name) VALUES ($1, $2)
					ON CONFLICT (agent_id, name) DO UPDATE SET name = excluded.name RETURNING id`
			err = tx.QueryRowContext(ctx, queryPrn, agentId, prn.Name).Scan(&printerId)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to printer update, agent: %d, printer: %s, err: %v", agentId, prn.Name, err)
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit agent info, agent: %d, printer: %d, err: %v", agentId, printerId, err)
	}
	return nil
}

// UpdatePrintTaskStatus обновление статуса задания на печать
func (r *pgPrint) UpdatePrintTaskStatus(ctx context.Context, status *model.PrintTaskStatus) error {
	query := `UPDATE print_queue SET status = $1 WHERE id = $2;`
	_, err := r.conn.ExecContext(ctx, query, status.Status, status.Id)
	if err != nil {
		return fmt.Errorf("failed to update print task status, task: %d, err: %v", status.Id, err)
	}

	return nil
}

// UpdateAgentToken обновляет токен агента по ID
func (r *pgPrint) UpdateAgentToken(ctx context.Context, agentId int, token string) error {
	query := `UPDATE agents SET token = $2 WHERE id = $1;`
	_, err := r.conn.ExecContext(ctx, query, agentId, token)
	if err != nil {
		return fmt.Errorf("failed to update agent token, task: %d, err: %v", agentId, err)
	}
	return nil
}
