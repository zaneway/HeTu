package util

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// HistoryRecord 历史记录结构
type HistoryRecord struct {
	ID        int64
	TabName   string
	Content   string
	CreatedAt time.Time
}

// HistoryDB 历史记录数据库管理器
type HistoryDB struct {
	db *sql.DB
}

var (
	historyDBInstance *HistoryDB
)

// GetHistoryDB 获取历史记录数据库实例（单例模式）
func GetHistoryDB() *HistoryDB {
	if historyDBInstance == nil {
		historyDBInstance = &HistoryDB{}
		historyDBInstance.init()
	}
	return historyDBInstance
}

// init 初始化数据库连接
func (h *HistoryDB) init() {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("无法获取用户主目录: " + err.Error())
	}

	// 创建 .hetu 目录
	hetuDir := filepath.Join(homeDir, ".hetu")
	err = os.MkdirAll(hetuDir, 0755)
	if err != nil {
		panic("无法创建 .hetu 目录: " + err.Error())
	}

	// 数据库文件路径
	dbPath := filepath.Join(hetuDir, "history.db")

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic("无法打开数据库: " + err.Error())
	}

	h.db = db

	// 创建表（如果不存在）
	err = h.createTable()
	if err != nil {
		panic("无法创建表: " + err.Error())
	}
}

// createTable 创建历史记录表
func (h *HistoryDB) createTable() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tab_name TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_tab_name ON history(tab_name);
	CREATE INDEX IF NOT EXISTS idx_created_at ON history(created_at DESC);
	`

	_, err := h.db.Exec(createTableSQL)
	return err
}

// AddHistory 添加历史记录
func (h *HistoryDB) AddHistory(tabName string, content string) error {
	insertSQL := `INSERT INTO history (tab_name, content, created_at) VALUES (?, ?, ?)`
	_, err := h.db.Exec(insertSQL, tabName, content, time.Now())
	return err
}

// GetHistory 获取指定标签页的历史记录
func (h *HistoryDB) GetHistory(tabName string, limit int) ([]HistoryRecord, error) {
	querySQL := `
		SELECT id, tab_name, content, created_at 
		FROM history 
		WHERE tab_name = ? 
		ORDER BY created_at DESC 
		LIMIT ?
	`

	rows, err := h.db.Query(querySQL, tabName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []HistoryRecord
	for rows.Next() {
		var record HistoryRecord
		err := rows.Scan(&record.ID, &record.TabName, &record.Content, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

// ClearHistory 清除指定标签页的历史记录
func (h *HistoryDB) ClearHistory(tabName string) error {
	deleteSQL := `DELETE FROM history WHERE tab_name = ?`
	_, err := h.db.Exec(deleteSQL, tabName)
	return err
}

// Close 关闭数据库连接
func (h *HistoryDB) Close() error {
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}
