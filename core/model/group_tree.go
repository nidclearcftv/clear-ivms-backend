package model

// GroupTreeNode is one node of a GroupTree: a Group plus its child nodes
// and the vehicles assigned directly to it (not to any descendant).
// Assembled in-memory by GroupService.GetTree — there is no "tree" table;
// persistence only ever stores flat Group/Vehicle rows.
type GroupTreeNode struct {
	Group    Group
	Children []GroupTreeNode
	Vehicles []Vehicle
}

// GroupTree is an organization's whole fleet hierarchy: every top-level
// (ParentID == nil) group as a root, plus every vehicle whose GroupID is
// nil, gathered separately since they don't belong to any node.
type GroupTree struct {
	Roots              []GroupTreeNode
	UnassignedVehicles []Vehicle
}
