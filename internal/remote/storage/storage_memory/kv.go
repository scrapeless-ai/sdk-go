package storage_memory

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/scrapeless-ai/sdk-go/internal/remote/storage/models"
	"github.com/scrapeless-ai/sdk-go/scrapeless/log"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const MaxExpireTime = 24 * 60 * 60 * 7

func (c *LocalClient) GetNamespace(ctx context.Context, namespaceId string) (*models.KvNamespaceItem, error) {
	nsPath := filepath.Join(storageDir, keyValueDir, namespaceId)
	ok := isDirExists(nsPath)
	if !ok {
		return nil, ErrResourceNotFound
	}
	metaDataPath := filepath.Join(nsPath, metadataFile)
	file, err := os.ReadFile(metaDataPath)
	if err != nil {
		return nil, fmt.Errorf("read file %s failed: %v", metaDataPath, err)
	}

	var namespace models.KvNamespaceItem
	if err = json.Unmarshal(file, &namespace); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %s", err)
	}
	now := time.Now()
	stats := models.Stats{}
	_ = filepath.WalkDir(nsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == metadataFile {
			return nil
		}
		kvFile, err := os.ReadFile(filepath.Join(nsPath, d.Name()))
		if err != nil {
			log.Warnf("read file %s failed: %v", path, err)
			return nil
		}
		if d.Name() == "INPUT.json" {
			stats.Size += uint64(len(kvFile))
			stats.Count++
			return nil
		}
		var kv models.SetValueLocal
		err = json.Unmarshal(kvFile, &kv)
		if err != nil {
			log.Warnf("json unmarshal failed: %s", err)
			return nil
		}
		if kv.ExpireAt.Before(now) {
			return nil
		}
		stats.Count++
		stats.Size += uint64(kv.Size)
		return nil
	})
	namespace.Stats = stats

	return &namespace, nil
}

func (c *LocalClient) ListNamespaces(ctx context.Context, page int64, pageSize int64, desc bool) (*models.KvNamespace, error) {
	dirPath := filepath.Join(storageDir, keyValueDir)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir: %v", err)
	}

	var allNamespaces []models.KvNamespaceItem

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		metaPath := filepath.Join(dirPath, name, metadataFile)

		file, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		var meta models.KvNamespaceItem
		if err = json.Unmarshal(file, &meta); err != nil {
			continue
		}

		allNamespaces = append(allNamespaces, meta)
	}

	// sort
	sort.Slice(allNamespaces, func(i, j int) bool {
		if desc {
			return allNamespaces[i].CreatedAt > allNamespaces[j].CreatedAt
		}
		return allNamespaces[i].CreatedAt < allNamespaces[j].CreatedAt
	})

	total := int64(len(allNamespaces))

	// page
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	pagedItems := allNamespaces[start:end]
	now := time.Now()
	for i := range pagedItems {
		stats := models.Stats{}
		nsPath := filepath.Join(dirPath, pagedItems[i].Id)
		_ = filepath.WalkDir(nsPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || d.Name() == metadataFile {
				return nil
			}

			kvFile, err := os.ReadFile(path)
			if err != nil {
				log.Warnf("read file %s failed: %v", path, err)
				return nil
			}
			if d.Name() == "INPUT.json" {
				stats.Size += uint64(len(kvFile))
				stats.Count++
				return nil
			}
			var kv models.SetValueLocal
			err = json.Unmarshal(kvFile, &kv)
			if err != nil {
				log.Warnf("json unmarshal failed: %s", err)
				return nil
			}
			if kv.ExpireAt.Before(now) {
				return nil
			}
			stats.Count++
			stats.Size += uint64(kv.Size)
			return nil
		})
		pagedItems[i].Stats = stats
	}

	return &models.KvNamespace{
		Items:     pagedItems,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage(total, pageSize),
	}, nil
}

func (c *LocalClient) CreateNamespace(ctx context.Context, req *models.CreateKvNamespaceRequest) (namespaceId string, err error) {
	id := uuid.NewString()
	path := filepath.Join(storageDir, keyValueDir, id)

	exists, err := isNameExists(filepath.Join(storageDir, keyValueDir), req.Name)
	if err != nil {
		return "", err
	}
	if exists {
		return "", fmt.Errorf("namespace %s already exists", req.Name)
	}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("create dataset failed, cause: %v", err)
	}

	namespace := models.KvNamespaceItem{
		Id:        id,
		Name:      req.Name,
		RunId:     req.RunId,
		ActorId:   req.ActorId,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	marshal, err := json.Marshal(&namespace)
	if err != nil {
		return "", fmt.Errorf("marshal namespace failed, cause: %v", err)
	}
	metaFile := filepath.Join(path, metadataFile)

	err = os.WriteFile(metaFile, marshal, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("write file %s failed, cause: %v", metaFile, err)
	}
	return id, nil
}

func (c *LocalClient) DelNamespace(ctx context.Context, namespaceId string) (bool, error) {
	absPath := filepath.Join(storageDir, keyValueDir, namespaceId)
	err := os.RemoveAll(absPath)
	if err != nil {
		return false, fmt.Errorf("delete namespace failed, cause: %v", err)
	}
	return true, nil
}

func (c *LocalClient) RenameNamespace(ctx context.Context, namespaceId string, name string) (ok bool, err error) {
	nsPath := filepath.Join(storageDir, keyValueDir, namespaceId)
	exists := isDirExists(nsPath)
	if !exists {
		return false, ErrResourceNotFound
	}
	filePath := filepath.Join(nsPath, metadataFile)
	file, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("read file %s failed: %v", filePath, err)
	}

	var old models.KvNamespaceItem
	if err = json.Unmarshal(file, &old); err != nil {
		return false, fmt.Errorf("json unmarshal failed: %s", err)
	}
	old.Name = name

	marshal, err := json.Marshal(&old)
	if err != nil {
		return false, fmt.Errorf("json marshal failed: %s", err)
	}

	err = os.WriteFile(filePath, marshal, 0644)
	if err != nil {
		return false, fmt.Errorf("write file %s failed: %v", filePath, err)
	}
	return true, nil
}

func (c *LocalClient) SetValue(ctx context.Context, req *models.SetValue) (bool, error) {
	if req.Key == "INPUT" && req.NamespaceId == "default" {
		return false, nil
	}
	keyFile := fmt.Sprintf("%s.json", req.Key)
	if keyFile == metadataFile {
		return false, fmt.Errorf("key name can't use 'metadata'")
	}
	path := filepath.Join(storageDir, keyValueDir, req.NamespaceId)
	file := filepath.Join(path, keyFile)
	if req.Expiration == 0 {
		req.Expiration = MaxExpireTime
	}
	local := models.SetValueLocal{
		SetValue: models.SetValue{
			Expiration:  req.Expiration,
			Key:         req.Key,
			Value:       req.Value,
			NamespaceId: req.NamespaceId,
		},
		ExpireAt: time.Now().Add(time.Duration(req.Expiration) * time.Second),
		Size:     len([]byte(req.Value)),
	}

	marshal, err := json.Marshal(local)
	if err != nil {
		return false, fmt.Errorf("json marshal failed: %s", err)
	}
	err = os.WriteFile(file, marshal, os.ModePerm)

	return true, nil
}

func (c *LocalClient) ListKeys(ctx context.Context, req *models.ListKeyInfo) (*models.KvKeys, error) {
	dirPath := filepath.Join(storageDir, keyValueDir, req.NamespaceId)
	var keys []map[string]any

	now := time.Now()
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == metadataFile {
			return nil
		}

		kvFile, err := os.ReadFile(filepath.Join(dirPath, d.Name()))
		if err != nil {
			return fmt.Errorf("read file %s failed: %v", path, err)
		}
		var kv models.SetValueLocal
		err = json.Unmarshal(kvFile, &kv)
		if err != nil {
			return fmt.Errorf("json unmarshal failed: %s", err)
		}

		if kv.ExpireAt.Before(now) {
			return nil
		}

		keys = append(keys, map[string]any{
			"key":  kv.Key,
			"size": kv.Size,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}
	total := int64(len(keys))
	kvKeys := &models.KvKeys{
		Total:     total,
		Page:      req.Page,
		PageSize:  req.Size,
		TotalPage: totalPage(total, req.Size),
	}

	start := (req.Page - 1) * req.Size
	if start >= total {
		return kvKeys, nil
	}

	end := start + req.Size
	if end > total {
		end = total
	}
	kvKeys.Items = keys[start:end]
	return kvKeys, nil
}

func (c *LocalClient) BulkSetValue(ctx context.Context, req *models.BulkSet) (int64, error) {
	var success int64
	for i := range req.Items {
		ok, _ := c.SetValue(ctx, &models.SetValue{
			NamespaceId: req.NamespaceId,
			Key:         req.Items[i].Key,
			Value:       req.Items[i].Value,
			Expiration:  req.Items[i].Expiration,
		})
		if ok {
			success++
		}
	}

	return success, nil
}

func (c *LocalClient) DelValue(ctx context.Context, namespaceId string, key string) (bool, error) {
	file := fmt.Sprintf("%s.json", key)
	if file == metadataFile {
		return true, nil
	}
	path := filepath.Join(storageDir, keyValueDir, namespaceId, file)
	err := os.Remove(path)
	if err != nil {
		return false, fmt.Errorf("delete file %s failed: %v", path, err)
	}

	return true, nil
}

func (c *LocalClient) BulkDelValue(ctx context.Context, namespaceId string, keys []string) (bool, error) {
	for i := range keys {
		_, err := c.DelValue(ctx, namespaceId, keys[i])
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (c *LocalClient) GetValue(ctx context.Context, namespaceId string, key string) (string, error) {
	namespacePath := filepath.Join(storageDir, keyValueDir, namespaceId)
	if !isDirExists(namespacePath) {
		return "", ErrResourceNotFound
	}
	file := fmt.Sprintf("%s.json", key)
	path := filepath.Join(namespacePath, file)
	buff, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file %s failed: %v", path, err)
	}

	if key == "INPUT" && namespaceId == "default" {
		return string(buff), nil
	}

	var kv models.SetValueLocal
	if err := json.Unmarshal(buff, &kv); err != nil {
		return "", fmt.Errorf("json unmarshal failed: %s", err)
	}
	if kv.ExpireAt.Before(time.Now()) {
		return "", nil
	}
	return kv.Value, nil
}
