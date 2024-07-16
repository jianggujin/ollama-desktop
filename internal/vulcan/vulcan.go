package vulcan

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"sort"
)

type Vulcan struct {
	migrator Migrator
	source   Source
}

func NewVulcan(migrator Migrator, source Source) *Vulcan {
	return &Vulcan{
		migrator: migrator,
		source:   source,
	}
}

func (v *Vulcan) Migrate() error {
	return v.MigrateContext(context.Background())
}

func (v *Vulcan) MigrateContext(ctx context.Context) error {
	// 1 查找版本
	sources, err := v.source.Scan()
	if err != nil {
		return err
	}
	// 1.1 不存在版本则直接返回
	if len(sources) == 0 {
		return nil
	}
	// 1.2 对版本信息进行排序
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Version.LessThan(sources[j].Version)
	})

	// 2 查询历史执行版本
	lastVersion, err := v.migrator.LastVersion(ctx)
	if err != nil {
		return nil
	}
	// 3 合并版本
	for _, source := range sources {
		// 3.1 如果版本小于最后一次版本则忽略
		if lastVersion != nil && source.Version.LessThanOrEqual(lastVersion) {
			continue
		}
		// 3.2 执行升级
		nodes, err := v.parseContent(source.Uid)
		if err != nil {
			return err
		}
		if err := v.migrator.Migrate(ctx, nodes, source.Version); err != nil {
			return err
		}
	}
	return nil
}

func (v *Vulcan) parseContent(uid string) ([]Node, error) {
	content, err := v.source.Read(uid)
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(bytes.NewReader(content))
	var nodes []Node
	// 遍历 XML 树
	for {
		// 读取下一个 token
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nodes, nil
			}
			return nil, err
		}

		// 判断 token 的类型
		switch token := token.(type) {
		case xml.StartElement:
			var node Node
			switch token.Name.Local {
			case "createTable":
				node = &CreateTableNode{}
			case "createIndex":
				node = &CreateIndexNode{}
			case "createPrimaryKey":
				node = &CreatePrimaryKeyNode{}
			case "dropTable":
				node = &DropTableNode{}
			case "dropIndex":
				node = &DropIndexNode{}
			case "addColumn":
				node = &AddColumnNode{}
			case "alterColumn":
				node = &AlterColumnNode{}
			case "dropColumn":
				node = &DropColumnNode{}
			case "dropPrimaryKey":
				node = &DropPrimaryKeyNode{}
			case "renameTable":
				node = &RenameTableNode{}
			case "alterTableRemarks":
				node = &AlterTableRemarksNode{}
			case "script":
				node = &ScriptNode{}
			}
			if node != nil {
				if err := decoder.DecodeElement(node, &token); err != nil {
					return nil, err
				}
				nodes = append(nodes, node)
			}
		}
	}
}
