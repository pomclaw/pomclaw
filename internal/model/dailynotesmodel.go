package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DailyNotesModel = (*customDailyNotesModel)(nil)

type (
	// DailyNotesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDailyNotesModel.
	DailyNotesModel interface {
		dailyNotesModel
		withSession(session sqlx.Session) DailyNotesModel
	}

	customDailyNotesModel struct {
		*defaultDailyNotesModel
	}
)

// NewDailyNotesModel returns a model for the database table.
func NewDailyNotesModel(conn sqlx.SqlConn) DailyNotesModel {
	return &customDailyNotesModel{
		defaultDailyNotesModel: newDailyNotesModel(conn),
	}
}

func (m *customDailyNotesModel) withSession(session sqlx.Session) DailyNotesModel {
	return NewDailyNotesModel(sqlx.NewSqlConnFromSession(session))
}
