package repository

import (
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"

	"github.com/avarian/primbon-ajaib-backend/model"
	"gorm.io/gorm"
)

type ChatboxMessageRepository struct {
	db *gorm.DB
}

func NewChatboxMessageRepository(db *gorm.DB) *ChatboxMessageRepository {
	return &ChatboxMessageRepository{
		db: db,
	}
}

func (s *ChatboxMessageRepository) FilterScope(r *http.Request) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db
	}
}

func (s *ChatboxMessageRepository) PaginateScope(r *http.Request) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		if page == 0 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(q.Get("page_size"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		sortBy := q.Get("sort_by")
		if sortBy == "" {
			sortBy = "id"
		}

		direction := q.Get("direction")
		if direction == "" {
			direction = "desc"
		}

		sort := sortBy + " " + direction

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize).Order(sort)
	}
}

func (s *ChatboxMessageRepository) MetaPaginate(r *http.Request) map[string]interface{} {
	q := r.URL.Query()
	var totalRows int64
	s.db.Model(model.ChatboxMessage{}).Scopes(s.FilterScope(r)).Count(&totalRows)

	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}
	totalPages := int(math.Ceil(float64(totalRows) / float64(pageSize)))
	page, _ := strconv.Atoi(q.Get("page"))
	if page == 0 {
		page = 1
	}
	meta := map[string]interface{}{
		"page":        page,
		"page_size":   pageSize,
		"total_rows":  totalRows,
		"total_pages": totalPages,
	}
	return meta
}

func (s *ChatboxMessageRepository) Index(r *http.Request, preload ...string) ([]model.ChatboxMessage, *gorm.DB) {
	var table []model.ChatboxMessage
	tx := s.db.Scopes(s.FilterScope(r), s.PaginateScope(r))
	for _, v := range preload {
		tx = tx.Preload(v)
	}
	query := tx.Find(&table)

	return table, query
}

func (s *ChatboxMessageRepository) All(r *http.Request, preload ...string) ([]model.ChatboxMessage, *gorm.DB) {
	var table []model.ChatboxMessage
	tx := s.db.Scopes(s.FilterScope(r))
	for _, v := range preload {
		tx = tx.Preload(v)
	}
	query := tx.Find(&table)

	return table, query
}

func (s *ChatboxMessageRepository) One(r *http.Request, preload ...string) (model.ChatboxMessage, *gorm.DB) {
	var table model.ChatboxMessage
	tx := s.db.Scopes(s.FilterScope(r))
	for _, v := range preload {
		tx = tx.Preload(v)
	}
	query := tx.Find(&table)

	return table, query
}

func (s *ChatboxMessageRepository) OneById(id int, preload ...string) (model.ChatboxMessage, *gorm.DB) {
	var table model.ChatboxMessage
	tx := s.db.Where("id = ?", id)
	for _, v := range preload {
		tx = tx.Preload(v)
	}
	query := tx.Find(&table)

	return table, query
}

func (s *ChatboxMessageRepository) Create(data model.ChatboxMessage) (model.ChatboxMessage, *gorm.DB) {
	var table model.ChatboxMessage
	s.AssignData(&table, data)
	query := s.db.Create(&table)
	return table, query
}

func (s *ChatboxMessageRepository) Update(id int, data model.ChatboxMessage) (model.ChatboxMessage, *gorm.DB) {
	var table model.ChatboxMessage
	table, result := s.OneById(id)
	if result.RowsAffected == 0 {
		result.Error = fmt.Errorf("data not found with id = %d", id)
		return table, result
	}
	s.AssignData(&table, data)
	query := s.db.Save(&table)
	return table, query
}

func (s *ChatboxMessageRepository) Delete(id int, isHard bool) *gorm.DB {
	tx := s.db
	if isHard {
		tx = tx.Unscoped()
	}
	query := tx.Delete(&model.ChatboxMessage{}, id)
	return query
}

func (s *ChatboxMessageRepository) AssignData(table *model.ChatboxMessage, data model.ChatboxMessage) {
	dataRV := reflect.ValueOf(data)
	tableRV := reflect.ValueOf(table)
	tableRVE := tableRV.Elem()

	for i := 0; i < dataRV.NumField(); i++ {
		if !dataRV.Field(i).IsZero() && (tableRVE.Field(i) != dataRV.Field(i)) {
			fv := tableRVE.FieldByName(dataRV.Type().Field(i).Name)
			fv.Set(dataRV.Field(i))
		}
	}
}

func (s *ChatboxMessageRepository) AllByChatboxCode(chatboxCode string, preload ...string) ([]model.ChatboxMessage, *gorm.DB) {
	var table []model.ChatboxMessage
	tx := s.db.Where("chatbox_code = ?", chatboxCode).Order("created_at ASC")
	for _, v := range preload {
		tx = tx.Preload(v)
	}
	query := tx.Find(&table)

	return table, query
}
