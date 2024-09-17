# nodechat

how to install

下載下來之後執行

`docker-compose up --build`

先打/v1/register進行註冊
之後登入才有使用者/v1/login

查詢get /v1/item之前先post /v1/item寫入物品列表
post的參數如這個物件
type Item struct {
	ItemID     int    `gorm:"primaryKey" json:"itemid"`
	ItemName   string `json:"item_name"`
	Category   string `json:"category"`
	Username   string `json:"username"`
	ItemType   string `json:"item_type"`
}

另外玩家登入之後會有 jwt

不過後續 api 如果要求檢查 jwt 的話使用者操作會太麻煩

這次作業我就不檢查了
