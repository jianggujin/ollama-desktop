package vulcan

import (
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
func (m *SqliteMigrator) initChangeLogTable() error {
	if m.inited {
		return nil
	}
	rows, err := m.driver.Query("SELECT name FROM sqlite_master WHERE type = 'table'")
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
	err = m.driver.Execute("CREATE TABLE " + m.changeTableName + "(id INTEGER PRIMARY KEY AUTOINCREMENT, change_version VARCHAR(255) NOT NULL, is_success TINYINT DEFAULT 0 NOT NULL, created_at DATETIME, updated_at DATETIME)")
	if err != nil {
		return err
	}
	m.inited = true
	return nil
}

func (m *SqliteMigrator) LastVersion() (*version.Version, error) {
	if err := m.initChangeLogTable(); err != nil {
		return nil, err
	}
	rows, err := m.driver.Query("SELECT change_version FROM " + m.changeTableName + " WHERE is_success = 1")
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

func (m *SqliteMigrator) Migrate(nodes []Node, version *version.Version) error {
	if err := m.initChangeLogTable(); err != nil {
		return err
	}
	err := m.driver.Execute("INSERT INTO "+m.changeTableName+"(change_version, is_success, created_at, updated_at) VALUES(?, 0, ?, ?)", version.Original(), time.Now(), time.Now())
	if err != nil {
		return err
	}
	for _, node := range nodes {
		switch n := node.(type) {
		case *CreateTableNode:
			err = m.createTable(n)
		case *CreateIndexNode:
			err = m.createIndex(n)
		case *CreatePrimaryKeyNode:
			err = m.createPrimaryKey(n)
		case *DropTableNode:
			err = m.dropTable(n)
		case *DropIndexNode:
			err = m.dropIndex(n)
		case *AddColumnNode:
			err = m.addColumn(n)
		case *AlterColumnNode:
			err = m.alterColumn(n)
		case *DropColumnNode:
			err = m.dropColumn(n)
		case *DropPrimaryKeyNode:
			err = m.dropPrimaryKey(n)
		case *RenameTableNode:
			err = m.renameTable(n)
		case *AlterTableRemarksNode:
			err = m.alterTableRemarks(n)
		case *ScriptNode:
			err = m.script(n)
		}
		if err != nil {
			return err
		}
	}
	return m.driver.Execute("UPDATE "+m.changeTableName+" SET is_success = 1, updated_at = ? WHERE change_version = ? AND is_success = 0", time.Now(), version.Original())
}

func (m *SqliteMigrator) createTable(node *CreateTableNode) error {
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
	return m.driver.Execute(builder.String())
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

func (m *SqliteMigrator) createIndex(node *CreateIndexNode) error {
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
	return m.driver.Execute(builder.String())
}

func (m *SqliteMigrator) createPrimaryKey(node *CreatePrimaryKeyNode) error {
	// 查询表结构
	info, err := m.tableStruct(node.TableName)
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
	return m.copyTable(builder.String(), columnNames, tmpTableName, node.TableName, info.indexs)
}

func (m *SqliteMigrator) copyTable(createSql string, columnNames []string, tmpTableName, tableName string, indexSqls []string) error {
	if err := m.driver.Execute(createSql); err != nil {
		return nil
	}
	columnNameStr := strings.Join(columnNames, ", ")
	if err := m.driver.Execute(fmt.Sprintf("INSERT INTO %s(%s) SELECT %s FROM %s", tmpTableName, columnNameStr, columnNameStr, tableName)); err != nil {
		return nil
	}
	if err := m.driver.Execute(fmt.Sprintf("DROP TABLE %s", tableName)); err != nil {
		return nil
	}
	if err := m.driver.Execute(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tmpTableName, tableName)); err != nil {
		return nil
	}
	for _, insexSql := range indexSqls {
		if err := m.driver.Execute(insexSql); err != nil {
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

func (m *SqliteMigrator) tableStruct(tableName string) (*sqliteTableStruct, error) {
	columns, err := m.parseColumns(tableName)
	if err != nil {
		return nil, err
	}
	indexSqls, err := m.parseIndexSqls(tableName)
	if err != nil {
		return nil, err
	}
	return &sqliteTableStruct{
		columns: columns,
		indexs:  indexSqls,
	}, nil
}

func (m *SqliteMigrator) parseColumns(tableName string) ([]*sqliteColumnStruct, error) {
	// 查询表结构
	rows, err := m.driver.Query("PRAGMA table_info (?)", tableName)
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

func (m *SqliteMigrator) parseIndexSqls(tableName string) ([]string, error) {
	rows, err := m.driver.Query("select sql from sqlite_master where sql is not null and type = 'index' and lower(tbl_name) = ?", strings.ToLower(tableName))
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

func (m *SqliteMigrator) dropTable(node *DropTableNode) error {
	return m.driver.Execute(fmt.Sprintf("DROP TABLE %s", node.TableName))
}

func (m *SqliteMigrator) dropIndex(node *DropIndexNode) error {
	return m.driver.Execute(fmt.Sprintf("DROP INDEX %s", node.IndexName))
}

func (m *SqliteMigrator) addColumn(node *AddColumnNode) error {
	for _, column := range node.Columns {
		var builder strings.Builder
		builder.WriteString("ALTER TABLE ")
		builder.WriteString(node.TableName)
		builder.WriteString(" ADD ")
		m.createTableColumn(column, &builder)
		if err := m.driver.Execute(builder.String()); err != nil {
			return err
		}
	}
	return nil
}

func (m *SqliteMigrator) alterColumn(node *AlterColumnNode) error {
	info, err := m.tableStruct(node.TableName)
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
	return m.copyTable(builder.String(), columnNames, tmpTableName, node.TableName, info.indexs)
}

func (m *SqliteMigrator) dropColumn(node *DropColumnNode) error {
	// 查询表结构
	info, err := m.tableStruct(node.TableName)
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
	return m.copyTable(builder.String(), columnNames, tmpTableName, node.TableName, info.indexs)
}

func (m *SqliteMigrator) dropPrimaryKey(node *DropPrimaryKeyNode) error {
	// 查询表结构
	info, err := m.tableStruct(node.TableName)
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
	return m.copyTable(builder.String(), columnNames, tmpTableName, node.TableName, info.indexs)
}

func (m *SqliteMigrator) renameTable(node *RenameTableNode) error {
	return m.driver.Execute(fmt.Sprintf("alter table %s rename to %s", node.TableName, node.NewTableName))
}

func (m *SqliteMigrator) alterTableRemarks(node *AlterTableRemarksNode) error {
	// 不支持
	return nil
}

func (m *SqliteMigrator) script(node *ScriptNode) error {
	if (node.Dialect != "" && node.Dialect != m.Name()) || node.Value == "" {
		return nil
	}
	for _, statement := range splitSQLStatements(node.Value) {
		if err := m.driver.Execute(statement); err != nil {
			return err
		}
	}
	return nil
}
