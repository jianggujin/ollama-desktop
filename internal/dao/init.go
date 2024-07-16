package dao

import (
	"context"
	"database/sql"
	"embed"
	_ "github.com/glebarez/go-sqlite"
	"github.com/hashicorp/go-version"
	"ollama-desktop/internal/vulcan"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const dbHistoryTable = "t_db_history"

//go:embed sql
var sqlFiles embed.FS

type DbDao struct {
	ctx context.Context
	db  *sql.DB
}

var Dao *DbDao

func (d *DbDao) startup(ctx context.Context) {
	d.ctx = ctx
	if err := d.init(); err != nil {
		panic(err)
	}
}

func (d *DbDao) shutdown(ctx context.Context) {
	if d.db != nil {
		d.db.Close()
		d.db = nil
	}
}

func (d *DbDao) init() error {
	if d.db != nil {
		return nil
	}
	// 连接到SQLite数据库
	path := "config/ollama-desktop.db"
	dir := filepath.Dir(path)
	os.MkdirAll(dir, os.ModePerm)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	db.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	db.SetMaxOpenConns(50)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	db.SetConnMaxLifetime(time.Duration(3600) * time.Second)
	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
	db.SetConnMaxIdleTime(time.Duration(1800) * time.Second)
	return d.migrate()
}

func (d *DbDao) migrate() error {
	migrator := vulcan.NewSqliteMigrator(vulcan.NewSqlDriver(d.db))
	migrate := vulcan.NewVulcan(migrator, &vulcan.EmbedFSSource{
		Fs:    sqlFiles,
		Paths: []string{"sql"},
	})
	return migrate.Migrate()
}

func (d *DbDao) initDbHistory() error {
	rows, err := d.db.QueryContext(d.ctx, "SELECT name FROM sqlite_master WHERE type = 'table'")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		if dbHistoryTable == strings.ToLower(tableName) {
			return nil
		}
	}

	_, err = d.db.ExecContext(d.ctx, "CREATE TABLE "+dbHistoryTable+"(id INTEGER PRIMARY KEY AUTOINCREMENT, db_version VARCHAR(255) NOT NULL, is_success TINYINT DEFAULT 0 NOT NULL, created_at DATETIME, updated_at DATETIME)")
	if err != nil {
		return err
	}
	return nil
}

func (d *DbDao) lastVersion() (*version.Version, error) {
	if err := d.initDbHistory(); err != nil {
		return nil, err
	}
	rows, err := d.db.QueryContext(d.ctx, "SELECT db_version FROM "+dbHistoryTable+" WHERE is_success = 1")
	if err != nil {
		return nil, err
	}
	var versions []*version.Version
	defer rows.Close()
	for rows.Next() {
		var dbVersion string
		if err := rows.Scan(&dbVersion); err != nil {
			return nil, err
		}
		ver, err := version.NewVersion(dbVersion)
		if err != nil {
			return nil, err
		}
		versions = append(versions, ver)
	}
	if len(versions) == 0 {
		return nil, nil
	}
	sort.Sort(version.Collection(versions))
	return versions[len(versions)-1], nil
}
