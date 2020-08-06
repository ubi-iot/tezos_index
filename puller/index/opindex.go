// Copyright (c) 2020 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package index

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"tezos_index/puller/models"
)

const (
	OpPackSizeLog2         = 15 // 32k packs
	OpJournalSizeLog2      = 16 // 64k
	OpCacheSize            = 4
	OpFillLevel            = 100
	OpIndexPackSizeLog2    = 15 // 16k packs (32k split size)
	OpIndexJournalSizeLog2 = 16 // 64k
	OpIndexCacheSize       = 128
	OpIndexFillLevel       = 90
	OpIndexKey             = "op"
	OpTableKey             = "op"
)

var (
	ErrNoOpEntry = errors.New("op not indexed")
)

type OpIndex struct {
	db *gorm.DB
}

func NewOpIndex(db *gorm.DB) *OpIndex {
	return &OpIndex{db}
}

func (idx *OpIndex) DB() *gorm.DB {
	return idx.db
}

func (idx *OpIndex) ConnectBlock(ctx context.Context, block *models.Block, _ models.BlockBuilder) error {
	ops := make([]*models.Op, 0, len(block.Ops))
	for _, op := range block.Ops {
		ops = append(ops, op)
	}
	// todo batch insert
	tx := idx.DB().Begin()
	for _, op := range ops {
		if err := tx.Create(op).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (idx *OpIndex) DisconnectBlock(ctx context.Context, block *models.Block, _ models.BlockBuilder) error {
	return idx.DeleteBlock(ctx, block.Height)
}

func (idx *OpIndex) DeleteBlock(ctx context.Context, height int64) error {
	log.Debugf("Rollback deleting ops at height %d", height)

	return idx.DB().Where("height = ?", height).Delete(&models.Op{}).Error
}
