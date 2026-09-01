package repomanager

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/adapter/cmsv6"
	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type VehicleRepositoryOptions struct {
	CMSV6 *cmsv6.Server `validate:"required"`
}

// VehicleRepository implements port.VehicleRepository by routing to the
// underlying IVMS-specific repository for a vehicle's source. Right now
// that's only cmsv6, but this is the layer that dispatches by
// model.IVMSType once more sources exist.
type VehicleRepository struct {
	cmsv6 *cmsv6.Server
}

func NewVehicleRepository(opts VehicleRepositoryOptions) (*VehicleRepository, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &VehicleRepository{cmsv6: opts.CMSV6}, nil
}

func (r *VehicleRepository) ListVehicles(ctx context.Context, filters model.VehicleFilters) (model.List[model.Vehicle], error) {
	return r.cmsv6.ListVehicles(ctx, filters)
}

var _ port.VehicleRepository = (*VehicleRepository)(nil)
