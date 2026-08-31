package cmsv6

import (
	"context"
	"fmt"
)

// VehicleList fetches the authenticated account's device/vehicle list from
// /vehicle/list.do. The session established by Login (or lazily on first
// use) is attached automatically as the JSESSIONID cookie.
func (s *Server) VehicleList(ctx context.Context) (*VehicleListResponse, error) {
	var result VehicleListResponse

	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/vehicle/list.do")
	if err != nil {
		return nil, fmt.Errorf("cmsv6: vehicle list request failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("cmsv6: vehicle list failed with status %d", resp.StatusCode())
	}

	return &result, nil
}
