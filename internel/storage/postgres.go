package storage

import (
	"context"
	"github.com/fcraft/open-chat/internel/models"
	"github.com/jackc/pgx/v5"
)

type PostgresStore struct {
	conn *pgx.Conn
}

func (s *PostgresStore) CreateSession(session *models.Session) error {
	_, err := s.conn.Exec(context.Background(),
		`INSERT INTO sessions 
         (id, user_id, enable_context, model_params)
         VALUES ($1, $2, $3, $4)`,
		session.ID, session.UserID,
		session.EnableContext, session.ModelParams)
	return err
}

func (s *PostgresStore) SaveMessage(msg *models.Message) error {
	_, err := s.conn.Exec(context.Background(),
		`INSERT INTO messages 
         (id, session_id, role, content)
         VALUES ($1, $2, $3, $4)`,
		msg.ID, msg.SessionID, msg.Role, msg.Content)
	return err
}
