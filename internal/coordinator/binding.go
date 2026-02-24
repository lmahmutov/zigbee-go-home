package coordinator

import (
	"context"
	"fmt"

	"zigbee-go-home/internal/ncp"
)

// Bind creates a binding on the target device.
func (c *Coordinator) Bind(ctx context.Context, targetShortAddr uint16, srcIEEE string, srcEP uint8, clusterID uint16, dstIEEE string, dstEP uint8) error {
	srcAddr, err := ParseIEEE(srcIEEE)
	if err != nil {
		return fmt.Errorf("parse src ieee: %w", err)
	}
	dstAddr, err := ParseIEEE(dstIEEE)
	if err != nil {
		return fmt.Errorf("parse dst ieee: %w", err)
	}
	return c.ncp.Bind(ctx, ncp.BindRequest{
		TargetShortAddr: targetShortAddr,
		SrcIEEE:         srcAddr,
		SrcEP:           srcEP,
		ClusterID:       clusterID,
		DstIEEE:         dstAddr,
		DstEP:           dstEP,
	})
}

// Unbind removes a binding from the target device.
func (c *Coordinator) Unbind(ctx context.Context, targetShortAddr uint16, srcIEEE string, srcEP uint8, clusterID uint16, dstIEEE string, dstEP uint8) error {
	srcAddr, err := ParseIEEE(srcIEEE)
	if err != nil {
		return fmt.Errorf("parse src ieee: %w", err)
	}
	dstAddr, err := ParseIEEE(dstIEEE)
	if err != nil {
		return fmt.Errorf("parse dst ieee: %w", err)
	}
	return c.ncp.Unbind(ctx, ncp.BindRequest{
		TargetShortAddr: targetShortAddr,
		SrcIEEE:         srcAddr,
		SrcEP:           srcEP,
		ClusterID:       clusterID,
		DstIEEE:         dstAddr,
		DstEP:           dstEP,
	})
}
