package cmsv6

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

var _ port.VehicleRepository = (*Server)(nil)

// ListVehicles implements port.VehicleRepository, fetching the account's
// device list from /vehicle/list.do and mapping it into the domain model.
//
// /vehicle/list.do has no server-side pagination, so filters.Pagination is
// applied in-memory after fetching and mapping the full list.
func (s *Server) ListVehicles(ctx context.Context, filters model.VehicleFilters) (model.List[model.Vehicle], error) {
	resp, err := s.VehicleList(ctx)
	if err != nil {
		return model.List[model.Vehicle]{}, err
	}

	vehicles := make([]model.Vehicle, 0, len(resp.UserDeviceList))
	for _, v := range resp.UserDeviceList {
		vehicle, err := toModelVehicle(v)
		if err != nil {
			return model.List[model.Vehicle]{}, fmt.Errorf("cmsv6: failed to map vehicle %d: %w", v.ID, err)
		}
		vehicles = append(vehicles, vehicle)
	}

	return model.List[model.Vehicle]{Items: vehicles, Total: len(vehicles)}, nil
}

// toModelVehicle maps a cmsv6 Vehicle into the domain model.Vehicle. Fields
// without a direct equivalent in the domain model are preserved under
// Attributes instead of being discarded.
func toModelVehicle(v Vehicle) (model.Vehicle, error) {
	return model.Vehicle{
		ID:          EncodeID(v.ID),
		FleetID:     EncodeID(v.GroupId),
		PlateNumber: v.PlateIDNO,
		IVMSType:    model.IVMSTypeCMSV6,
	}, nil
}

// toAttributes round-trips v through JSON into a model.JSON map, preserving
// every cmsv6-specific field (under its original JSON key) that doesn't
// have a typed home in model.Vehicle.
func toAttributes(v Vehicle) (model.JSON, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal vehicle: %w", err)
	}

	var attrs model.JSON
	if err := json.Unmarshal(raw, &attrs); err != nil {
		return nil, fmt.Errorf("unmarshal vehicle attributes: %w", err)
	}

	return attrs, nil
}
