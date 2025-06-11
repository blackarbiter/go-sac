package migration

import (
	"fmt"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"gorm.io/gorm"
)

// AutoMigrate 执行数据库迁移
func AutoMigrate(db *gorm.DB) error {
	// 定义所有需要迁移的模型
	models := []interface{}{
		&model.BaseAsset{},
		&model.RequirementAsset{},
		&model.DesignDocumentAsset{},
		&model.RepositoryAsset{},
		&model.UploadedFileAsset{},
		&model.ImageAsset{},
		&model.DomainAsset{},
		&model.IPAsset{},
	}

	logger.Logger.Info("auto migrate start...")
	// 执行迁移
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", m, err)
		}
	}

	// 添加外键约束
	foreignKeys := []struct {
		table    string
		column   string
		refTable string
		refCol   string
	}{
		{"assets_requirement", "id", "assets_base", "id"},
		{"assets_design_document", "id", "assets_base", "id"},
		{"assets_repository", "id", "assets_base", "id"},
		{"assets_uploaded_file", "id", "assets_base", "id"},
		{"assets_image", "id", "assets_base", "id"},
		{"assets_domain", "id", "assets_base", "id"},
		{"assets_ip", "id", "assets_base", "id"},
	}

	for _, fk := range foreignKeys {
		// 检查外键是否已存在
		var count int64
		db.Raw(`
			SELECT COUNT(*)
			FROM information_schema.table_constraints
			WHERE constraint_name = ? AND table_name = ?
		`, fmt.Sprintf("fk_%s_%s", fk.table, fk.refTable), fk.table).Count(&count)

		if count == 0 {
			// 添加外键约束
			sql := fmt.Sprintf(`
				ALTER TABLE %s
				ADD CONSTRAINT fk_%s_%s
				FOREIGN KEY (%s) REFERENCES %s(%s)
			`, fk.table, fk.table, fk.refTable, fk.column, fk.refTable, fk.refCol)

			if err := db.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to add foreign key constraint for %s: %w", fk.table, err)
			}
		}
	}

	return nil
}
