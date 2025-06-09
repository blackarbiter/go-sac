package repository

import (
	"context"
	"testing"

	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 执行迁移
	err = db.AutoMigrate(
		&model.BaseAsset{},
		&model.RequirementAsset{},
		&model.DesignDocumentAsset{},
		&model.RepositoryAsset{},
		&model.UploadedFileAsset{},
		&model.ImageAsset{},
		&model.DomainAsset{},
		&model.IPAsset{},
	)
	require.NoError(t, err)

	return db
}

func TestGormRepository_CreateBase(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	base := &model.BaseAsset{
		AssetType:      "Requirement",
		Name:           "Test Asset",
		Status:         "active",
		CreatedBy:      "test",
		UpdatedBy:      "test",
		OrganizationID: 1,
	}

	err := repo.CreateBase(ctx, base)
	require.NoError(t, err)
	assert.NotZero(t, base.ID)

	// 验证数据是否正确保存
	var saved model.BaseAsset
	err = db.First(&saved, base.ID).Error
	require.NoError(t, err)
	assert.Equal(t, base.Name, saved.Name)
	assert.Equal(t, base.AssetType, saved.AssetType)
}

func TestGormRepository_CreateRequirement(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	base := &model.BaseAsset{
		AssetType:      "Requirement",
		Name:           "Test Requirement",
		Status:         "active",
		CreatedBy:      "test",
		UpdatedBy:      "test",
		OrganizationID: 1,
	}

	ext := &model.RequirementAsset{
		BusinessValue: "Test Business Value",
		Priority:      1,
		Version:       "1.0",
	}

	err := repo.CreateRequirement(ctx, base, ext)
	require.NoError(t, err)
	assert.NotZero(t, base.ID)
	assert.Equal(t, base.ID, ext.ID)

	// 验证数据是否正确保存
	var savedBase model.BaseAsset
	var savedExt model.RequirementAsset
	err = db.First(&savedBase, base.ID).Error
	require.NoError(t, err)
	err = db.First(&savedExt, ext.ID).Error
	require.NoError(t, err)

	assert.Equal(t, base.Name, savedBase.Name)
	assert.Equal(t, ext.BusinessValue, savedExt.BusinessValue)
}

func TestGormRepository_ListBase(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 创建测试数据
	bases := []*model.BaseAsset{
		{
			AssetType:      "Requirement",
			Name:           "Test 1",
			Status:         "active",
			CreatedBy:      "test",
			UpdatedBy:      "test",
			OrganizationID: 1,
		},
		{
			AssetType:      "Requirement",
			Name:           "Test 2",
			Status:         "active",
			CreatedBy:      "test",
			UpdatedBy:      "test",
			OrganizationID: 1,
		},
	}

	for _, base := range bases {
		err := repo.CreateBase(ctx, base)
		require.NoError(t, err)
	}

	// 测试分页查询
	results, total, err := repo.ListBase(ctx, map[string]interface{}{
		"asset_type": "Requirement",
	}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

func TestGormRepository_UpdateBase(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 创建测试数据
	base := &model.BaseAsset{
		AssetType:      "Requirement",
		Name:           "Test Asset",
		Status:         "active",
		CreatedBy:      "test",
		UpdatedBy:      "test",
		OrganizationID: 1,
	}

	err := repo.CreateBase(ctx, base)
	require.NoError(t, err)

	// 更新数据
	base.Name = "Updated Name"
	err = repo.UpdateBase(ctx, base)
	require.NoError(t, err)

	// 验证更新
	var updated model.BaseAsset
	err = db.First(&updated, base.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
}

func TestGormRepository_DeleteBase(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 创建测试数据
	base := &model.BaseAsset{
		AssetType:      "Requirement",
		Name:           "Test Asset",
		Status:         "active",
		CreatedBy:      "test",
		UpdatedBy:      "test",
		OrganizationID: 1,
	}

	err := repo.CreateBase(ctx, base)
	require.NoError(t, err)

	// 删除数据
	err = repo.DeleteBase(ctx, base.ID)
	require.NoError(t, err)

	// 验证删除
	var deleted model.BaseAsset
	err = db.First(&deleted, base.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
