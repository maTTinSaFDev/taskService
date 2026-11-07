package transport

import (
	"core/model"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

const boardIDFilterName = "board_id"

func taskFiltersFromRequest(r *http.Request, authenticatedUserID int64) (model.TaskFilters, error) {
	filters := model.TaskFilters{UserID: &authenticatedUserID}
	rawBoardID := strings.TrimSpace(r.URL.Query().Get(boardIDFilterName))
	if rawBoardID != "" {
		boardID, err := strconv.ParseInt(rawBoardID, 10, 64)
		if err != nil {
			return filters, errors.New(`query parameter "board_id" must be a valid integer`)
		}
		filters.BoardID = &boardID
	}

	if err := filters.Validate(); err != nil {
		return filters, err
	}

	return filters, nil
}
