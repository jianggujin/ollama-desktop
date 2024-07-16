package vulcan

import (
	"context"
	"database/sql"
	"github.com/hashicorp/go-version"
)

const (
	Varchar   = "VARCHAR"
	Char      = "CHAR"
	Text      = "TEXT"
	Clob      = "CLOB"
	Boolean   = "BOOLEAN"
	Tinyint   = "TINYINT"
	Smallint  = "SMALLINT"
	Int       = "INT"
	Bigint    = "BIGINT"
	Decimal   = "DECIMAL"
	Date      = "DATE"
	Time      = "TIME"
	Timestamp = "TIMESTAMP"
	Blob      = "BLOB"
)

type Node interface {
}
type CreateTableNode struct {
	TableName string        `xml:"tableName,attr"`
	Remarks   string        `xml:"remarks,attr"`
	Columns   []*ColumnNode `xml:"column"`
}
type ColumnNode struct {
	ColumnName         string               `xml:"columnName,attr"`
	DataType           string               `xml:"dataType,attr"`
	MaxLength          int                  `xml:"maxLength,attr"`
	NumericScale       int                  `xml:"numericScale,attr"`
	Nullable           bool                 `xml:"nullable,attr"`
	Unique             bool                 `xml:"unique,attr"`
	PrimaryKey         bool                 `xml:"primaryKey,attr"`
	DefaultValue       string               `xml:"defaultValue,attr"`
	DefaultOriginValue string               `xml:"defaultOriginValue,attr"`
	Remarks            string               `xml:"remarks,attr"`
	Dialects           []*ColumnDialectNode `xml:"columnDialect"`
}
type ColumnDialectNode struct {
	Dialect            string `xml:"dialect,attr"`
	DataType           string `xml:"dataType,attr"`
	DefaultValue       string `xml:"defaultValue,attr"`
	DefaultOriginValue string `xml:"defaultOriginValue,attr"`
}
type CreateIndexNode struct {
	TableName string             `xml:"tableName,attr"`
	IndexName string             `xml:"indexName,attr"`
	Unique    bool               `xml:"unique,attr"`
	Columns   []*IndexColumnNode `xml:"indexColumn"`
}
type IndexColumnNode struct {
	ColumnName string `xml:"columnName,attr"`
}
type CreatePrimaryKeyNode struct {
	TableName string           `xml:"tableName,attr"`
	KeyName   string           `xml:"keyName,attr"`
	Column    *IndexColumnNode `xml:"indexColumn"`
}
type DropTableNode struct {
	TableName string `xml:"tableName,attr"`
}
type DropIndexNode struct {
	TableName string `xml:"tableName,attr"`
	IndexName string `xml:"indexName,attr"`
}
type AddColumnNode struct {
	TableName string        `xml:"tableName,attr"`
	Columns   []*ColumnNode `xml:"column"`
}
type AlterColumnNode struct {
	TableName  string      `xml:"tableName,attr"`
	ColumnName string      `xml:"columnName,attr"`
	Column     *ColumnNode `xml:"column"`
}
type DropColumnNode struct {
	TableName  string `xml:"tableName,attr"`
	ColumnName string `xml:"columnName,attr"`
}
type DropPrimaryKeyNode struct {
	TableName string `xml:"tableName,attr"`
}
type RenameTableNode struct {
	TableName    string `xml:"tableName,attr"`
	NewTableName string `xml:"newTableName,attr"`
}
type AlterTableRemarksNode struct {
	TableName string `xml:"tableName,attr"`
	Remarks   string `xml:"remarks,attr"`
}
type ScriptNode struct {
	Dialect string `xml:"dialect,attr"`
	Value   string `xml:",chardata"`
}

type Migrator interface {
	// 最后一次版本信息
	LastVersion(context.Context) (*version.Version, error)
	// 合并指定版本
	Migrate(context.Context, []Node, *version.Version) error
}

type Driver interface {
	Execute(context.Context, string, ...interface{}) error
	Query(context.Context, string, ...interface{}) (*sql.Rows, error)
}
