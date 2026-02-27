package services

import "fmt"

// statusTransitionRules defines the allowed status transitions for the logistics monitoring process.
// Each key represents a current status, and the value is a slice of allowed next statuses.
// The lifecycle consists of 15 stages:
//   - Delivery stages (1-8): cooking through school receipt
//   - Collection stages (9-13): driver assignment through SPPG arrival
//   - Cleaning stages (14-15): cleaning process through completion
//
// Requirements: 14.1, 14.2
var statusTransitionRules = map[string][]string{
	// Delivery stages (1-8)
	"sedang_dimasak":               {"selesai_dimasak"},
	"selesai_dimasak":              {"siap_dipacking"},
	"siap_dipacking":               {"selesai_dipacking"},
	"selesai_dipacking":            {"siap_dikirim"},
	"siap_dikirim":                 {"diperjalanan"},
	"diperjalanan":                 {"sudah_sampai_sekolah"},
	"sudah_sampai_sekolah":         {"sudah_diterima_pihak_sekolah"},
	"sudah_diterima_pihak_sekolah": {"driver_ditugaskan_mengambil_ompreng"},

	// Collection stages (9-13)
	"driver_ditugaskan_mengambil_ompreng": {"driver_menuju_sekolah"},
	"driver_menuju_sekolah":               {"driver_sampai_di_sekolah"},
	"driver_sampai_di_sekolah":            {"ompreng_telah_diambil"},
	"ompreng_telah_diambil":               {"ompreng_sampai_di_sppg"},
	"ompreng_sampai_di_sppg":              {"ompreng_proses_pencucian"},

	// Cleaning stages (14-15)
	"ompreng_proses_pencucian": {"ompreng_selesai_dicuci"},
	"ompreng_selesai_dicuci":   {}, // Final state - no further transitions allowed
}

// InvalidTransitionError represents an error when a status transition is not allowed.
// It provides information about the current status, requested status, and allowed transitions.
//
// Requirements: 14.1, 14.2
type InvalidTransitionError struct {
	CurrentStatus   string
	RequestedStatus string
	AllowedStatuses []string
}

// Error implements the error interface for InvalidTransitionError.
// It returns a formatted error message with the current status, requested status,
// and the list of allowed transitions.
func (e *InvalidTransitionError) Error() string {
	return fmt.Sprintf("cannot transition from %s to %s. Allowed transitions: %v",
		e.CurrentStatus, e.RequestedStatus, e.AllowedStatuses)
}

// ValidateStatusTransition checks if a status transition is allowed based on the
// statusTransitionRules map. It returns an InvalidTransitionError if the transition
// is not allowed, providing the current status, requested status, and allowed statuses
// for user guidance.
//
// Parameters:
//   - currentStatus: The current status of the delivery record
//   - newStatus: The requested new status
//
// Returns:
//   - error: nil if the transition is allowed, InvalidTransitionError otherwise
//
// Requirements: 14.1, 14.2
func ValidateStatusTransition(currentStatus, newStatus string) error {
	// Look up the allowed transitions for the current status
	allowedStatuses, exists := statusTransitionRules[currentStatus]

	// If the current status doesn't exist in the rules, it's invalid
	if !exists {
		return &InvalidTransitionError{
			CurrentStatus:   currentStatus,
			RequestedStatus: newStatus,
			AllowedStatuses: []string{},
		}
	}

	// Check if the new status is in the allowed transitions list
	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return nil // Transition is allowed
		}
	}

	// Transition is not allowed, return error with guidance
	return &InvalidTransitionError{
		CurrentStatus:   currentStatus,
		RequestedStatus: newStatus,
		AllowedStatuses: allowedStatuses,
	}
}

// StageGroup represents the lifecycle stage group
type StageGroup int

const (
	StageGroupDelivery StageGroup = iota
	StageGroupCollection
	StageGroupCleaning
	StageGroupUnknown
)

// getStageGroup determines which stage group a status belongs to.
// Returns:
//   - StageGroupDelivery for stages 1-8 (cooking through school receipt)
//   - StageGroupCollection for stages 9-13 (driver assignment through SPPG arrival)
//   - StageGroupCleaning for stages 14-15 (cleaning process through completion)
//   - StageGroupUnknown for invalid statuses
//
// Requirements: 14.3, 14.4
func getStageGroup(status string) StageGroup {
	// Delivery stages (1-8)
	deliveryStages := []string{
		"sedang_dimasak",
		"selesai_dimasak",
		"siap_dipacking",
		"selesai_dipacking",
		"siap_dikirim",
		"diperjalanan",
		"sudah_sampai_sekolah",
		"sudah_diterima_pihak_sekolah",
	}

	// Collection stages (9-13)
	collectionStages := []string{
		"driver_ditugaskan_mengambil_ompreng",
		"driver_menuju_sekolah",
		"driver_sampai_di_sekolah",
		"ompreng_telah_diambil",
		"ompreng_sampai_di_sppg",
	}

	// Cleaning stages (14-15)
	cleaningStages := []string{
		"ompreng_proses_pencucian",
		"ompreng_selesai_dicuci",
	}

	// Check which group the status belongs to
	for _, s := range deliveryStages {
		if s == status {
			return StageGroupDelivery
		}
	}

	for _, s := range collectionStages {
		if s == status {
			return StageGroupCollection
		}
	}

	for _, s := range cleaningStages {
		if s == status {
			return StageGroupCleaning
		}
	}

	return StageGroupUnknown
}

// StageSequenceError represents an error when stage sequence is violated.
// It indicates that a transition is attempting to skip a required stage group.
//
// Requirements: 14.3, 14.4
type StageSequenceError struct {
	CurrentStatus string
	NewStatus     string
	Message       string
}

// Error implements the error interface for StageSequenceError.
func (e *StageSequenceError) Error() string {
	return fmt.Sprintf("stage sequence violation: %s (current: %s, requested: %s)",
		e.Message, e.CurrentStatus, e.NewStatus)
}

// ValidateStageSequence validates that the lifecycle stages follow the correct sequence:
//   - Delivery stages (1-8) must occur before collection stages (9-13)
//   - Collection stages (9-13) must occur before cleaning stages (14-15)
//
// This prevents transitions that skip stage groups, such as going directly from
// delivery to cleaning without completing collection.
//
// Parameters:
//   - currentStatus: The current status of the delivery record
//   - newStatus: The requested new status
//
// Returns:
//   - error: nil if the sequence is valid, StageSequenceError if violated
//
// Requirements: 14.3, 14.4
func ValidateStageSequence(currentStatus, newStatus string) error {
	currentGroup := getStageGroup(currentStatus)
	newGroup := getStageGroup(newStatus)

	// If either status is unknown, return error
	if currentGroup == StageGroupUnknown {
		return &StageSequenceError{
			CurrentStatus: currentStatus,
			NewStatus:     newStatus,
			Message:       "current status is not a valid lifecycle stage",
		}
	}

	if newGroup == StageGroupUnknown {
		return &StageSequenceError{
			CurrentStatus: currentStatus,
			NewStatus:     newStatus,
			Message:       "requested status is not a valid lifecycle stage",
		}
	}

	// Validate sequence: delivery → collection → cleaning
	// Moving backwards or skipping groups is not allowed

	// From delivery stage
	if currentGroup == StageGroupDelivery {
		// Can only move to delivery or collection, not cleaning
		if newGroup == StageGroupCleaning {
			return &StageSequenceError{
				CurrentStatus: currentStatus,
				NewStatus:     newStatus,
				Message:       "cannot skip collection stages: delivery stages must be followed by collection stages before cleaning",
			}
		}
	}

	// From collection stage
	if currentGroup == StageGroupCollection {
		// Cannot move back to delivery
		if newGroup == StageGroupDelivery {
			return &StageSequenceError{
				CurrentStatus: currentStatus,
				NewStatus:     newStatus,
				Message:       "cannot move back to delivery stages from collection stages",
			}
		}
	}

	// From cleaning stage
	if currentGroup == StageGroupCleaning {
		// Cannot move back to delivery or collection
		if newGroup == StageGroupDelivery || newGroup == StageGroupCollection {
			return &StageSequenceError{
				CurrentStatus: currentStatus,
				NewStatus:     newStatus,
				Message:       "cannot move back to earlier stages from cleaning stages",
			}
		}
	}

	return nil
}
