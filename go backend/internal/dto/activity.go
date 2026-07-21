package dto

// RecordActivityRequest — kunlik faollikni yozish/yangilash so'rovi.
// date bo'sh bo'lsa — bugungi kun. Qiymatlar o'sha kunning JAMI (kümülativ) hisobi.
type RecordActivityRequest struct {
	Date      string  `json:"date" binding:"omitempty,datetime=2006-01-02"`
	Steps     int     `json:"steps" binding:"gte=0,lte=300000"`
	Calories  float64 `json:"calories" binding:"gte=0,lte=50000"`
	DistanceM float64 `json:"distance_m" binding:"gte=0,lte=500000"`
	ActiveMin int     `json:"active_min" binding:"gte=0,lte=1440"`
	Source    string  `json:"source" binding:"omitempty,max=50"`
}

// RecordActivityBatchRequest — bir necha kunlik faollikni bitta so'rovda
// yozish ("backfill"): ilova ochilganda telefondagi oxirgi kunlar qayta
// yuboriladi, shuning uchun foydalanuvchi bir necha kun ilovani ochmasa
// ham hech qanday kun yo'qolmaydi.
type RecordActivityBatchRequest struct {
	Items []RecordActivityRequest `json:"items" binding:"required,min=1,max=31,dive"`
}
