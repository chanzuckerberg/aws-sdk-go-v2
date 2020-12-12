// Code generated by smithy-go-codegen DO NOT EDIT.

package types

type BackupStatus string

// Enum values for BackupStatus
const (
	BackupStatusInProgress BackupStatus = "IN_PROGRESS"
	BackupStatusOk         BackupStatus = "OK"
	BackupStatusFailed     BackupStatus = "FAILED"
	BackupStatusDeleting   BackupStatus = "DELETING"
)

// Values returns all known values for BackupStatus. Note that this can be expanded
// in the future, and so it is only as up to date as the client. The ordering of
// this slice is not guaranteed to be stable across updates.
func (BackupStatus) Values() []BackupStatus {
	return []BackupStatus{
		"IN_PROGRESS",
		"OK",
		"FAILED",
		"DELETING",
	}
}

type BackupType string

// Enum values for BackupType
const (
	BackupTypeAutomated BackupType = "AUTOMATED"
	BackupTypeManual    BackupType = "MANUAL"
)

// Values returns all known values for BackupType. Note that this can be expanded
// in the future, and so it is only as up to date as the client. The ordering of
// this slice is not guaranteed to be stable across updates.
func (BackupType) Values() []BackupType {
	return []BackupType{
		"AUTOMATED",
		"MANUAL",
	}
}

type MaintenanceStatus string

// Enum values for MaintenanceStatus
const (
	MaintenanceStatusSuccess MaintenanceStatus = "SUCCESS"
	MaintenanceStatusFailed  MaintenanceStatus = "FAILED"
)

// Values returns all known values for MaintenanceStatus. Note that this can be
// expanded in the future, and so it is only as up to date as the client. The
// ordering of this slice is not guaranteed to be stable across updates.
func (MaintenanceStatus) Values() []MaintenanceStatus {
	return []MaintenanceStatus{
		"SUCCESS",
		"FAILED",
	}
}

type NodeAssociationStatus string

// Enum values for NodeAssociationStatus
const (
	NodeAssociationStatusSuccess    NodeAssociationStatus = "SUCCESS"
	NodeAssociationStatusFailed     NodeAssociationStatus = "FAILED"
	NodeAssociationStatusInProgress NodeAssociationStatus = "IN_PROGRESS"
)

// Values returns all known values for NodeAssociationStatus. Note that this can be
// expanded in the future, and so it is only as up to date as the client. The
// ordering of this slice is not guaranteed to be stable across updates.
func (NodeAssociationStatus) Values() []NodeAssociationStatus {
	return []NodeAssociationStatus{
		"SUCCESS",
		"FAILED",
		"IN_PROGRESS",
	}
}

type ServerStatus string

// Enum values for ServerStatus
const (
	ServerStatusBackingUp        ServerStatus = "BACKING_UP"
	ServerStatusConnectionLost   ServerStatus = "CONNECTION_LOST"
	ServerStatusCreating         ServerStatus = "CREATING"
	ServerStatusDeleting         ServerStatus = "DELETING"
	ServerStatusModifying        ServerStatus = "MODIFYING"
	ServerStatusFailed           ServerStatus = "FAILED"
	ServerStatusHealthy          ServerStatus = "HEALTHY"
	ServerStatusRunning          ServerStatus = "RUNNING"
	ServerStatusRestoring        ServerStatus = "RESTORING"
	ServerStatusSetup            ServerStatus = "SETUP"
	ServerStatusUnderMaintenance ServerStatus = "UNDER_MAINTENANCE"
	ServerStatusUnhealthy        ServerStatus = "UNHEALTHY"
	ServerStatusTerminated       ServerStatus = "TERMINATED"
)

// Values returns all known values for ServerStatus. Note that this can be expanded
// in the future, and so it is only as up to date as the client. The ordering of
// this slice is not guaranteed to be stable across updates.
func (ServerStatus) Values() []ServerStatus {
	return []ServerStatus{
		"BACKING_UP",
		"CONNECTION_LOST",
		"CREATING",
		"DELETING",
		"MODIFYING",
		"FAILED",
		"HEALTHY",
		"RUNNING",
		"RESTORING",
		"SETUP",
		"UNDER_MAINTENANCE",
		"UNHEALTHY",
		"TERMINATED",
	}
}