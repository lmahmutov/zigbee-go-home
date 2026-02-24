package coordinator

import (
	"context"
	"fmt"

	"zigbee-go-home/internal/ncp"
	"zigbee-go-home/internal/zcl"
)

// AttributeResult holds a decoded attribute read result.
type AttributeResult struct {
	AttrID   uint16      `json:"attr_id"`
	AttrName string      `json:"attr_name"`
	TypeID   uint8       `json:"type_id"`
	TypeName string      `json:"type_name"`
	Value    interface{} `json:"value"`
	Status   uint8       `json:"status"`
	Error    string      `json:"error,omitempty"`
}

// ReadAttributes reads attributes from a device endpoint/cluster.
func (c *Coordinator) ReadAttributes(ctx context.Context, shortAddr uint16, endpoint uint8, clusterID uint16, attrIDs []uint16) ([]AttributeResult, error) {
	responses, err := c.ncp.ReadAttributes(ctx, ncp.ReadAttributesRequest{
		DstAddr:   shortAddr,
		DstEP:     endpoint,
		ClusterID: clusterID,
		AttrIDs:   attrIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("read attributes: %w", err)
	}

	cluster := c.registry.Get(clusterID)
	var results []AttributeResult
	for _, r := range responses {
		result := AttributeResult{
			AttrID: r.AttrID,
			Status: r.Status,
			TypeID: r.DataType,
			TypeName: zcl.TypeName(r.DataType),
		}
		if cluster != nil {
			if attr := cluster.FindAttribute(r.AttrID); attr != nil {
				result.AttrName = attr.Name
			}
		}
		if result.AttrName == "" {
			result.AttrName = fmt.Sprintf("0x%04X", r.AttrID)
		}
		if r.Status != 0 {
			result.Error = fmt.Sprintf("status 0x%02X", r.Status)
		} else if len(r.Value) > 0 {
			val, _, err := zcl.DecodeValue(r.DataType, r.Value)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Value = val
			}
		}
		results = append(results, result)
	}
	return results, nil
}

// WriteAttribute writes a single attribute value.
func (c *Coordinator) WriteAttribute(ctx context.Context, shortAddr uint16, endpoint uint8, clusterID uint16, attrID uint16, dataType uint8, value interface{}) error {
	encoded, err := zcl.EncodeValue(dataType, value)
	if err != nil {
		return fmt.Errorf("encode value: %w", err)
	}
	return c.ncp.WriteAttributes(ctx, ncp.WriteAttributesRequest{
		DstAddr:   shortAddr,
		DstEP:     endpoint,
		ClusterID: clusterID,
		Records: []ncp.WriteRecord{
			{AttrID: attrID, DataType: dataType, Value: encoded},
		},
	})
}

// SendClusterCommand sends a cluster-specific command.
func (c *Coordinator) SendClusterCommand(ctx context.Context, shortAddr uint16, endpoint uint8, clusterID uint16, commandID uint8, payload []byte) error {
	return c.ncp.SendCommand(ctx, ncp.ClusterCommandRequest{
		DstAddr:   shortAddr,
		DstEP:     endpoint,
		ClusterID: clusterID,
		CommandID: commandID,
		Payload:   payload,
	})
}

// ConfigureReporting sets up attribute reporting on a device.
func (c *Coordinator) ConfigureReporting(ctx context.Context, shortAddr uint16, endpoint uint8, clusterID uint16, attrID uint16, dataType uint8, minInterval, maxInterval uint16, reportableChange []byte) error {
	return c.ncp.ConfigureReporting(ctx, ncp.ConfigureReportingRequest{
		DstAddr:      shortAddr,
		DstEP:        endpoint,
		ClusterID:    clusterID,
		AttrID:       attrID,
		DataType:     dataType,
		MinInterval:  minInterval,
		MaxInterval:  maxInterval,
		ReportChange: reportableChange,
	})
}
