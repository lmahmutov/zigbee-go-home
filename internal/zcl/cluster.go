package zcl

// Access flags
const (
	AccessRead   uint8 = 0x01
	AccessWrite  uint8 = 0x02
	AccessReport uint8 = 0x04
)

// AttributeDef defines a ZCL attribute.
type AttributeDef struct {
	ID     uint16 `json:"id"`
	Name   string `json:"name"`
	Type   uint8  `json:"type"`
	Access uint8  `json:"access"` // bitmask: 1=read, 2=write, 4=reportable
}

// IsReadable returns true if the attribute can be read.
func (a *AttributeDef) IsReadable() bool {
	return a.Access&AccessRead != 0
}

// IsWritable returns true if the attribute can be written.
func (a *AttributeDef) IsWritable() bool {
	return a.Access&AccessWrite != 0
}

// IsReportable returns true if the attribute supports reporting.
func (a *AttributeDef) IsReportable() bool {
	return a.Access&AccessReport != 0
}

// CommandDirection indicates the direction of a cluster command.
type CommandDirection string

const (
	DirectionToServer CommandDirection = "toServer"
	DirectionToClient CommandDirection = "toClient"
)

// CommandDef defines a cluster-specific command.
type CommandDef struct {
	ID        uint8            `json:"id"`
	Name      string           `json:"name"`
	Direction CommandDirection `json:"direction"`
}

// ClusterDef defines a ZCL cluster with its attributes and commands.
type ClusterDef struct {
	ID         uint16         `json:"id"`
	Name       string         `json:"name"`
	Attributes []AttributeDef `json:"attributes,omitempty"`
	Commands   []CommandDef   `json:"commands,omitempty"`
}

// FindAttribute looks up an attribute by ID.
func (c *ClusterDef) FindAttribute(id uint16) *AttributeDef {
	for i := range c.Attributes {
		if c.Attributes[i].ID == id {
			return &c.Attributes[i]
		}
	}
	return nil
}

// FindCommand looks up a command by ID and direction.
func (c *ClusterDef) FindCommand(id uint8, dir CommandDirection) *CommandDef {
	for i := range c.Commands {
		if c.Commands[i].ID == id && c.Commands[i].Direction == dir {
			return &c.Commands[i]
		}
	}
	return nil
}

// DeepCopy returns a deep copy of the cluster definition.
func (c *ClusterDef) DeepCopy() *ClusterDef {
	cp := *c
	if c.Attributes != nil {
		cp.Attributes = make([]AttributeDef, len(c.Attributes))
		copy(cp.Attributes, c.Attributes)
	}
	if c.Commands != nil {
		cp.Commands = make([]CommandDef, len(c.Commands))
		copy(cp.Commands, c.Commands)
	}
	return &cp
}

// Merge adds attributes and commands from another definition (for custom.json overlay).
func (c *ClusterDef) Merge(other *ClusterDef) {
	for _, attr := range other.Attributes {
		if c.FindAttribute(attr.ID) == nil {
			c.Attributes = append(c.Attributes, attr)
		}
	}
	for _, cmd := range other.Commands {
		if c.FindCommand(cmd.ID, cmd.Direction) == nil {
			c.Commands = append(c.Commands, cmd)
		}
	}
}
