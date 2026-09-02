package cmsv6

import (
	"encoding/hex"
	"fmt"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

func EncodeID(id int) model.ID {
	return model.ID(hex.EncodeToString(fmt.Appendf(nil, "%d:%d", model.IVMSTypeCMSV6, id)))
}

func DecodeID(encoded model.ID) (int, error) {
	data, err := hex.DecodeString(string(encoded))
	if err != nil {
		return 0, err
	}

	var t model.IVMSType
	var id int
	_, err = fmt.Sscanf(string(data), "%d:%d", &t, &id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
