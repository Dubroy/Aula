// item_handlers.go

package main

import (
	"encoding/json"
	"net/http"
)

// Item 结构体用于表示单个物品
type Item struct {
	ItemID     int    `gorm:"primaryKey" json:"itemid"`
	ItemName   string `json:"item_name"`
	Category   string `json:"category"`
	Username   string `json:"username"`
	ItemType   string `json:"item_type"`
}
type ItemList struct {
	Itemid     int    `gorm:"primaryKey" json:"itemid"`
	ItemName   string `json:"item_name"`
	Category   string `json:"category"`
}

// ListItemHandler 处理 GET /v1/item 请求
func ListItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只允許 GET 請求", http.StatusMethodNotAllowed)
		return
	}

	// 从查询参数中获取值
	username := r.URL.Query().Get("username")
	itemType := r.URL.Query().Get("item")

	// 创建查询构建器
	query := db.Model(&Item{})
	query.Select("itemid,item_name,category")
	
	// 根据传递的参数构建条件
	if username != "" {
		query = query.Where("username = ?", username)
	}
	if itemType != "" {
		query = query.Where("item_type = ?", itemType)
	}

	// 执行查询
	var items []ItemList
	result := query.Find(&items)
	if result.Error != nil {
		http.Error(w, "查詢數據失敗", http.StatusInternalServerError)
		return
	}

	// 构造响应
	response := map[string]interface{}{
		"status": "success",
		"items":  items,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "編碼 JSON 失敗", http.StatusInternalServerError)
	}
}

// CreateItemHandler 处理 POST /v1/item 请求
func CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允許 POST 請求", http.StatusMethodNotAllowed)
		return
	}

	var item Item
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		http.Error(w, "無法解析請求主體", http.StatusBadRequest)
		return
	}

	// 验证输入数据
	if item.ItemName == "" || item.Category == "" || item.Username == "" || item.ItemType == "" {
		http.Error(w, "缺少必需的字段", http.StatusBadRequest)
		return
	}

	// 使用 GORM 插入物品
	result := db.Create(&item)
	if result.Error != nil {
		http.Error(w, "無法儲存物品", http.StatusInternalServerError)
		return
	}

	// 构造响应
	response := map[string]interface{}{
		"status":      "success",
		"description": "物品已成功創建",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "編碼 JSON 失敗", http.StatusInternalServerError)
	}
}