package tenant

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/zoekt/grpc/propagator"
	"github.com/sourcegraph/zoekt/internal/tenant/internal/tenanttype"
)

const (
	// headerKeyTenantID is the header key for the tenant ID.
	headerKeyTenantID = "X-Sourcegraph-Tenant-ID"

	// headerValueNoTenant indicates the request has no tenant.
	headerValueNoTenant = "none"
)

// Propagator implements the propagator.Propagator interface
// for propagating tenants across RPC calls. This is modeled directly on
// the HTTP middleware in this package, and should work exactly the same.
type Propagator struct{}

var _ propagator.Propagator = &Propagator{}

func (Propagator) FromContext(ctx context.Context) metadata.MD {
	md := make(metadata.MD)
	tenant, err := tenanttype.FromContext(ctx)
	if err != nil {
		md.Append(headerKeyTenantID, headerValueNoTenant)
	} else {
		md.Append(headerKeyTenantID, strconv.Itoa(tenant.ID()))
	}
	return md
}

func (Propagator) InjectContext(ctx context.Context, md metadata.MD) (context.Context, error) {
	var raw string
	if vals := md.Get(headerKeyTenantID); len(vals) > 0 {
		raw = vals[0]
	}
	switch raw {
	case "", headerValueNoTenant:
		// Nothing to do, empty tenant.
		return ctx, nil
	default:
		tenant, err := tenanttype.Unmarshal(raw)
		if err != nil {
			// The tenant value is invalid.
			return ctx, status.New(codes.InvalidArgument, fmt.Errorf("bad tenant value in metadata: %w", err).Error()).Err()
		}
		return tenanttype.WithTenant(ctx, tenant), nil
	}
}
