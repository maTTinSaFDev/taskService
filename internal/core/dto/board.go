package dto

type BoardCreateDto struct {
	OwnerUserId int64  `json:"owner_user_id"`
	Title       string `json:"title"`
}

type BoardColumnCreateDto struct {
	Title    string `json:"title"`
	Position int32  `json:"position"`
}

type BoardColumnUpdateDto struct {
	Title    string `json:"title"`
	Position int32  `json:"position"`
}

type BoardColumnOrderDto struct {
	ColumnIds []int64 `json:"column_ids"`
}
