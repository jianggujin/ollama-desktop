package vulcan

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"sort"
	"strings"
	"time"
)

var SqliteDataTypeMappers = map[string]string{
	Varchar:   "VARCHAR",
	Char:      "CHARACTER",
	Text:      "TEXT",
	Clob:      "CLOB",
	Boolean:   "TINYINT",
	Tinyint:   "TINYINT",
	Smallint:  "SMALLINT",
	Int:       "INTEGER",
	Bigint:    "INTEGER",
	Decimal:   "DECIMAL",
	Date:      "DATE",
	Time:      "TIME",
	Timestamp: "DATETIME",
	Blob:      "BLOB",
}

type SqliteMigrator struct {
	driver          Driver
	changeTableName string
	inited          bool
}

func NewSqliteMigrator(driver Driver) *SqliteMigrator {
	return &SqliteMigrator{
		driver:          driver,
		changeTableName: "database_change_log",
	}
}

func (m *SqliteMigrator) Name() string {
	return "sqlite"
}

// 初始化变更记录表
func (m *SqliteMigrator) initChangeLogTable(ctx context.Context) error {
	if m.inited {
		return nil
	}
	rows, err := m.driver.Query(ctx, "SELECT name FROM sqlite_master WHERE type = 'table'")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		if m.changeTableName == strings.ToLower(tableName) {
			m.inited = true
			return nil
		}
	}
	err = m.driver.Execute(ctx, "CREATE TABLE "+m.changeTableName+"(id INTEGER PRIMARY KEY AUTOINCREMENT, change_version VARCHAR(255) NOT NULL, is_success TINYINT DEFAULT 0 NOT NULL, created_at DATETIME, updated_at DATETIME)")
	if err != nil {
		return err
	}
	m.inited = true
	return nil
}

func (m *SqliteMigrator) LastVersion(ctx context.Context) (*version.Version, error) {
	if err := m.initChangeLogTable(ctx); err != nil {
		return nil, err
	}
	rows, err := m.driver.Query(ctx, "SELECT change_version FROM "+m.changeTableName+" WHERE is_success = 1")
	if err != nil {
		return nil, err
	}
	var versions []*version.Version
	defer rows.Close()
	for rows.Next() {
		var changeVersion string
		if err := rows.Scan(&changeVersion); err != nil {
			return nil, err
		}
		ver, err := version.NewVersion(changeVersion)
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

func (m *SqliteMigrator) Migrate(ctx context.Context, nodes []Node, version *version.Version) error {
	if err := m.initChangeLogTable(ctx); err != nil {
		return err
	}
	err := m.driver.Execute(ctx, "INSERT INTO "+m.changeTableName+"(change_version, is_success, created_at, updated_at) VALUES(?, 0, ?, ?)", version.Original(), time.Now(), time.Now())
	if err != nil {
		return err
	}
	for _, node := range nodes {
		switch n := node.(type) {
		case *CreateTableNode:
			err = m.createTable(ctx, n)
		case *CreateIndexNode:
			err = m.createIndex(ctx, n)
		case *CreatePrimaryKeyNode:
			err = m.createPrimaryKey(ctx, n)
		case *DropTableNode:
			err = m.dropTable(ctx, n)
		case *DropIndexNode:
			err = m.dropIndex(ctx, n)
		case *AddColumnNode:
			err = m.addColumn(ctx, n)
		case *AlterColumnNode:
			err = m.alterColumn(ctx, n)
		case *DropColumnNode:
			err = m.dropColumn(ctx, n)
		case *DropPrimaryKeyNode:
			err = m.dropPrimaryKey(ctx, n)
		case *RenameTableNode:
			err = m.renameTable(ctx, n)
		case *AlterTableRemarksNode:
			err = m.alterTableRemarks(ctx, n)
		case *ScriptNode:
			err = m.script(ctx, n)
		}
		if err != nil {
			return err
		}
	}
	return m.driver.Execute(ctx, "UPDATE "+m.changeTableName+" SET is_success = 1, updated_at = ? WHERE change_version = ? AND is_success = 0", time.Now(), version.Original())
}

func (m *SqliteMigrator) createTable(ctx context.Context, node *CreateTableNode) error {
	var builder strings.Builder
	builder.WriteString("CREATE TABLE ")
	builder.WriteString(node.TableName)
	builder.WriteString("\n(\n")
	size := len(node.Columns)
	for index, column := range node.Columns {
		builder.WriteString("  ")
		m.createTableColumn(column, &builder)
		if index < size-1 {
			builder.WriteString(",\n")
		}
	}
	builder.WriteString("\n)")
	return m.driver.Execute(ctx, builder.String())
}

func (m *SqliteMigrator) createTableColumn(node *ColumnNode, builder *strings.Builder) {
	var dialectNode *ColumnDialectNode
	// 查找方言
	for _, dialect := range node.Dialects {
		if dialect.Dialect == m.Name() {
			dialectNode = dialect
			break
		}
	}
	builder.WriteString(node.ColumnName)
	builder.WriteString(" ")
	var defaultValue string
	if dialectNode != nil {
		builder.WriteString(dialectNode.DataType)
		if dialectNode.DefaultOriginValue != "" {
			defaultValue = dialectNode.DefaultOriginValue
		} else if dialectNode.DefaultValue != "" {
			defaultValue = fmt.Sprintf("'%s'", strings.ReplaceAll(dialectNode.DefaultValue, "'", "''"))
		}
	} else {
		builder.WriteString(columnType(node.DataType, SqliteDataTypeMappers[node.DataType], node.MaxLength, node.NumericScale))
		if node.DefaultOriginValue != "" {
			defaultValue = node.DefaultOriginValue
		} else if node.DefaultValue != "" {
			defaultValue = fmt.Sprintf("'%s'", strings.ReplaceAll(node.DefaultValue, "'", "''"))
		}
	}
	if node.PrimaryKey {
		builder.WriteString(" PRIMARY KEY")
	} else {
		if defaultValue != "" {
			builder.WriteString(" DEFAULT ")
			builder.WriteString(defaultValue)
		}
		if node.Unique {
			builder.WriteString(" UNIQUE")
		}
		if !node.Nullable {
			builder.WriteString(" NOT NULL")
		}
	}
}

func (m *SqliteMigrator) createIndex(ctx context.Context, node *CreateIndexNode) error {
	var builder strings.Builder
	builder.WriteString("CREATE")
	if node.Unique {
		builder.WriteString(" UNIQUE")
	}
	builder.WriteString(" INDEX ")
	builder.WriteString(node.IndexName)
	builder.WriteString(" ON ")
	builder.WriteString(node.TableName)
	builder.WriteString(" (")
	var columns []string
	for _, columnNode := range node.Columns {
		columns = append(columns, columnNode.ColumnName)
	}
	builder.WriteString(strings.Join(columns, ", "))
	builder.WriteString(")")
	return m.driver.Execute(ctx, builder.String())
}

func (m *SqliteMigrator) createPrimaryKey(ctx context.Context, node *CreatePrimaryKeyNode) error {
	// 查询表结构
	info, err := m.tableStruct(ctx, node.TableName)
	if err != nil {
		return err
	}
	pkColumnName := strings.ToLower(node.Column.ColumnName)
	tmpTableName := node.TableName + "_vulcan"
	var builder strings.Builder
	builder.WriteString("CREATE TABLE ")
	builder.WriteString(tmpTableName)
	builder.WriteString("\n(\n")
	size := len(info.columns)
	var columnNames []string
	for index, column := range info.columns {
		columnNames = append(columnNames, column.Name)
		builder.WriteString("  ")
		builder.WriteString(column.Name)
		builder.WriteString(" ")
		builder.WriteString(column.Type)
		if column.DfltValue != "" {
			builder.WriteString(" DEFAULT ")
			builder.WriteString(column.DfltValue)
		}
		if column.Notnull {
			builder.WriteString(" NOT NULL")
		}
		if pkColumnName == strings.ToLower(column.Name) {
			builder.WriteString("CONSTRAINT ")
			builder.WriteString(node.KeyName)
			builder.WriteString(" PRIMARY KEY")
		}
		if index < size-1 {
			builder.WriteString(",\n")
		}
	}
	builder.WriteString("\n)")
	return m.copyTable(ctx, builder.String(), columnNames, tmpTableName, node.TableName, info.indexs)
}

func (m *SqliteMigrator) copyTable(ctx context.Context, createSql string, columnNames []string, tmpTableName, tableName string, indexSqls []string) error {
	if err := m.driver.Execute(ctx, createSql); err != nil {
		return nil
	}
	columnNameStr := strings.Join(columnNames, ", ")
	if err := m.driver.Execute(ctx, fmt.Sprintf("INSERT INTO %s(%s) SELECT %s FROM %s", tmpTableName, columnNameStr, columnNameStr, tableName)); err != nil {
		return nil
	}
	if err := m.driver.Execute(ctx, fmt.Sprintf("DROP TABLE %s", tableName)); err != nil {
		return nil
	}
	if err := m.driver.Execute(ctx, fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tmpTableName, tableName)); err != nil {
		return nil
	}
	for _, insexSql := range indexSqls {
		if err := m.driver.Execute(ctx, insexSql); err != nil {
			return nil
		}
	}
	return nil
}

type sqliteTableStruct struct {
	columns []*sqliteColumnStruct
	indexs  []string
}
type sqliteColumnStruct struct {
	Cid       int
	Name      string
	Type      string
	Notnull   bool
	DfltValue string
	Pk        bool
}

func (m *SqliteMigrator) tableStruct(ctx context.Context, tableName string) (*sqliteTableStruct, error) {
	columns, err := m.parseColumns(ctx, tableName)
	if err != nil {
		return nil, err
	}
	indexSqls, err := m.parseIndexSqls(ctx, tableName)
	if err != nil {
		return nil, err
	}
	return &sqliteTableStruct{
		columns: columns,
		indexs:  indexSqls,
	}, nil
}

func (m *SqliteMigrator) parseColumns(ctx context.Context, tableName string) ([]*sqliteColumnStruct, error) {
	// 查询表结构
	rows, err := m.driver.Query(ctx, "PRAGMA table_info (?)", tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []*sqliteColumnStruct
	for rows.Next() {
		var column sqliteColumnStruct
		if err := rows.Scan(&column.Cid, &column.Name, &column.Type, &column.Notnull, &column.DfltValue, &column.Pk); err != nil {
			return nil, err
		}
		columns = append(columns, &column)
	}
	return columns, nil
}

func (m *SqliteMigrator) parseIndexSqls(ctx context.Context, tableName string) ([]string, error) {
	rows, err := m.driver.Query(ctx, "select sql from sqlite_master where sql is not null and type = 'index' and lower(tbl_name) = ?", strings.ToLower(tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sqls []string
	for rows.Next() {
		var sqlStr string
		if err := rows.Scan(&sqlStr); err != nil {
			return nil, err
		}
		sqls = append(sqls, sqlStr)
	}
	return sqls, nil
}

func (m *SqliteMigrator) dropTable(ctx context.Context, node *DropTableNode) error {
	return m.driver.Execute(ctx, fmt.Sprintf("DROP TABLE %s", node.TableName))
}

func (m *SqliteMigrator) dropIndex(ctx context.Context, node *DropIndexNode) error {
	return m.driver.Execute(ctx, fmt.Sprintf("DROP INDEX %s", node.IndexName))
}

func (m *SqliteMigrator) addColumn(ctx context.Context, node *AddColumnNode) error {
	for _, column := range node.Columns {
		var builder strings.Builder
		builder.WriteString("ALTER TABLE ")
		builder.WriteString(node.TableName)
		builder.WriteString(" ADD ")
		m.createTableColumn(column, &builder)
		if err := m.driver.Execute(ctx, builder.String()); err != nil {
			return err
		}
	}
	return nil
}

func (m *SqliteMigrator) alterColumn(ctx context.Context, node *AlterColumnNode) error {
	info, err := m.tableStruct(ctx, node.TableName)
	if err != nil {
		return err
	}
	columnName := strings.ToLower(node.ColumnName)
	tmpTableName := node.TableName + "_vulcan"
	var builder strings.Builder
	builder.WriteString("CREATE TABLE ")
	builder.WriteString(tmpTableName)
	builder.WriteString("\n(\n")
	size := len(info.columns)
	var columnNames []string
	for index, column := range info.columns {
		columnNames = append(columnNames, column.Name)
		builder.WriteString("  ")
		if columnName == strings.ToLower(column.Name) {
			m.createTableColumn(node.Column, &builder)
		} else {
			builder.WriteString(column.Name)
			builder.WriteString(" ")
			builder.WriteString(column.Type)
			if column.DfltValue != "" {
				builder.WriteString(" DEFAULT ")
				builder.WriteString(column.DfltValue)
			}
			if column.Notnull {
				builder.WriteString(" NOT NULL")
			}
		}
		if index < size-1 {
			builder.WriteString(",\n")
		}
	}
	builder.WriteString("\n)")
	return m.copyTable(ctx, builder.String(), columnNames, tmpTableName, node.TableName, info.indexs)
}

func (m *SqliteMigrator) dropColumn(ctx context.Context, node *DropColumnNode) error {
	// 查询表结构
	info, err := m.tableStruct(ctx, node.TableName)
	if err != nil {
		return err
	}
	columnName := strings.ToLower(node.ColumnName)
	tmpTableName := node.TableName + "_vulcan"
	var builder strings.Builder
	builder.WriteString("CREATE TABLE ")
	builder.WriteString(tmpTableName)
	builder.WriteString("\n(\n")
	size := len(info.columns)
	var columnNames []string
	for index, column := range info.columns {
		if columnName == strings.ToLower(column.Name) {
			continue
		}
		columnNames = append(columnNames, column.Name)
		builder.WriteString("  ")
		builder.WriteString(column.Name)
		builder.WriteString(" ")
		builder.WriteString(column.Type)
		if column.DfltValue != "" {
			builder.WriteString(" DEFAULT ")
			builder.WriteString(column.DfltValue)
		}
		if column.Notnull {
			builder.WriteString(" NOT NULL")
		}
		if index < size-1 {
			builder.WriteString(",\n")
		}
	}
	builder.WriteString("\n)")
	return m.copyTable(ctx, builder.String(), columnNames, tmpTableName, node.TableName, info.indexs)
}

func (m *SqliteMigrator) dropPrimaryKey(ctx context.Context, node *DropPrimaryKeyNode) error {
	// 查询表结构
	info, err := m.tableStruct(ctx, node.TableName)
	if err != nil {
		return err
	}
	tmpTableName := node.TableName + "_vulcan"
	var builder strings.Builder
	builder.WriteString("CREATE TABLE ")
	builder.WriteString(tmpTableName)
	builder.WriteString("\n(\n")
	size := len(info.columns)
	var columnNames []string
	for index, column := range info.columns {
		columnNames = append(columnNames, column.Name)
		builder.WriteString("  ")
		builder.WriteString(column.Name)
		builder.WriteString(" ")
		builder.WriteString(column.Type)
		if column.DfltValue != "" {
			builder.WriteString(" DEFAULT ")
			builder.WriteString(column.DfltValue)
		}
		if column.Notnull {
			builder.WriteString(" NOT NULL")
		}
		if index < size-1 {
			builder.WriteString(",\n")
		}
	}
	builder.WriteString("\n)")
	return m.copyTable(ctx, builder.String(), columnNames, tmpTableName, node.TableName, info.indexs)
}

func (m *SqliteMigrator) renameTable(ctx context.Context, node *RenameTableNode) error {
	return m.driver.Execute(ctx, fmt.Sprintf("alter table %s rename to %s", node.TableName, node.NewTableName))
}

func (m *SqliteMigrator) alterTableRemarks(ctx context.Context, node *AlterTableRemarksNode) error {
	// 不支持
	return nil
}

func (m *SqliteMigrator) script(ctx context.Context, node *ScriptNode) error {
	if (node.Dialect != "" && node.Dialect != m.Name()) || node.Value == "" {
		return nil
	}
	for _, statement := range splitSQLStatements(node.Value) {
		if err := m.driver.Execute(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
