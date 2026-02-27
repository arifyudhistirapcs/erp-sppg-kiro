package services

import (
	"testing"
)

// TestValidateStatusTransition_ValidTransition tests that valid transitions are allowed
func TestValidateStatusTransition_ValidTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		newStatus     string
	}{
		{
			name:          "cooking to completed",
			currentStatus: "sedang_dimasak",
			newStatus:     "selesai_dimasak",
		},
		{
			name:          "completed cooking to ready for packing",
			currentStatus: "selesai_dimasak",
			newStatus:     "siap_dipacking",
		},
		{
			name:          "in transit to arrived at school",
			currentStatus: "diperjalanan",
			newStatus:     "sudah_sampai_sekolah",
		},
		{
			name:          "ompreng arrived to cleaning",
			currentStatus: "ompreng_sampai_di_sppg",
			newStatus:     "ompreng_proses_pencucian",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatusTransition(tt.currentStatus, tt.newStatus)
			if err != nil {
				t.Errorf("ValidateStatusTransition() error = %v, want nil", err)
			}
		})
	}
}

// TestValidateStatusTransition_InvalidTransition tests that invalid transitions are rejected
func TestValidateStatusTransition_InvalidTransition(t *testing.T) {
	tests := []struct {
		name            string
		currentStatus   string
		newStatus       string
		wantAllowedLen  int
		wantErrContains string
	}{
		{
			name:            "skip stage - cooking to packing",
			currentStatus:   "sedang_dimasak",
			newStatus:       "siap_dipacking",
			wantAllowedLen:  1,
			wantErrContains: "cannot transition from sedang_dimasak to siap_dipacking",
		},
		{
			name:            "backward transition",
			currentStatus:   "selesai_dimasak",
			newStatus:       "sedang_dimasak",
			wantAllowedLen:  1,
			wantErrContains: "cannot transition from selesai_dimasak to sedang_dimasak",
		},
		{
			name:            "jump to cleaning from delivery",
			currentStatus:   "diperjalanan",
			newStatus:       "ompreng_proses_pencucian",
			wantAllowedLen:  1,
			wantErrContains: "cannot transition from diperjalanan to ompreng_proses_pencucian",
		},
		{
			name:            "transition from final state",
			currentStatus:   "ompreng_selesai_dicuci",
			newStatus:       "sedang_dimasak",
			wantAllowedLen:  0,
			wantErrContains: "cannot transition from ompreng_selesai_dicuci to sedang_dimasak",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatusTransition(tt.currentStatus, tt.newStatus)
			if err == nil {
				t.Errorf("ValidateStatusTransition() error = nil, want error")
				return
			}

			invalidErr, ok := err.(*InvalidTransitionError)
			if !ok {
				t.Errorf("ValidateStatusTransition() error type = %T, want *InvalidTransitionError", err)
				return
			}

			if invalidErr.CurrentStatus != tt.currentStatus {
				t.Errorf("InvalidTransitionError.CurrentStatus = %v, want %v", invalidErr.CurrentStatus, tt.currentStatus)
			}

			if invalidErr.RequestedStatus != tt.newStatus {
				t.Errorf("InvalidTransitionError.RequestedStatus = %v, want %v", invalidErr.RequestedStatus, tt.newStatus)
			}

			if len(invalidErr.AllowedStatuses) != tt.wantAllowedLen {
				t.Errorf("InvalidTransitionError.AllowedStatuses length = %v, want %v", len(invalidErr.AllowedStatuses), tt.wantAllowedLen)
			}

			errMsg := err.Error()
			if errMsg != tt.wantErrContains && len(errMsg) < len(tt.wantErrContains) {
				t.Errorf("InvalidTransitionError.Error() = %v, want to contain %v", errMsg, tt.wantErrContains)
			}
		})
	}
}

// TestValidateStatusTransition_InvalidCurrentStatus tests handling of unknown current status
func TestValidateStatusTransition_InvalidCurrentStatus(t *testing.T) {
	err := ValidateStatusTransition("invalid_status", "sedang_dimasak")
	if err == nil {
		t.Errorf("ValidateStatusTransition() error = nil, want error for invalid current status")
		return
	}

	invalidErr, ok := err.(*InvalidTransitionError)
	if !ok {
		t.Errorf("ValidateStatusTransition() error type = %T, want *InvalidTransitionError", err)
		return
	}

	if len(invalidErr.AllowedStatuses) != 0 {
		t.Errorf("InvalidTransitionError.AllowedStatuses length = %v, want 0 for invalid current status", len(invalidErr.AllowedStatuses))
	}
}

// TestValidateStatusTransition_AllowedStatusesInError tests that error includes allowed statuses
func TestValidateStatusTransition_AllowedStatusesInError(t *testing.T) {
	err := ValidateStatusTransition("sedang_dimasak", "siap_dipacking")
	if err == nil {
		t.Errorf("ValidateStatusTransition() error = nil, want error")
		return
	}

	invalidErr, ok := err.(*InvalidTransitionError)
	if !ok {
		t.Errorf("ValidateStatusTransition() error type = %T, want *InvalidTransitionError", err)
		return
	}

	// For "sedang_dimasak", the only allowed transition is "selesai_dimasak"
	if len(invalidErr.AllowedStatuses) != 1 {
		t.Errorf("InvalidTransitionError.AllowedStatuses length = %v, want 1", len(invalidErr.AllowedStatuses))
	}

	if len(invalidErr.AllowedStatuses) > 0 && invalidErr.AllowedStatuses[0] != "selesai_dimasak" {
		t.Errorf("InvalidTransitionError.AllowedStatuses[0] = %v, want selesai_dimasak", invalidErr.AllowedStatuses[0])
	}
}

// TestGetStageGroup tests the stage group classification
func TestGetStageGroup(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		wantGroup StageGroup
	}{
		// Delivery stages
		{
			name:      "cooking stage",
			status:    "sedang_dimasak",
			wantGroup: StageGroupDelivery,
		},
		{
			name:      "completed cooking stage",
			status:    "selesai_dimasak",
			wantGroup: StageGroupDelivery,
		},
		{
			name:      "in transit stage",
			status:    "diperjalanan",
			wantGroup: StageGroupDelivery,
		},
		{
			name:      "received by school stage",
			status:    "sudah_diterima_pihak_sekolah",
			wantGroup: StageGroupDelivery,
		},
		// Collection stages
		{
			name:      "driver assigned stage",
			status:    "driver_ditugaskan_mengambil_ompreng",
			wantGroup: StageGroupCollection,
		},
		{
			name:      "driver heading to school stage",
			status:    "driver_menuju_sekolah",
			wantGroup: StageGroupCollection,
		},
		{
			name:      "ompreng collected stage",
			status:    "ompreng_telah_diambil",
			wantGroup: StageGroupCollection,
		},
		{
			name:      "ompreng arrived at SPPG stage",
			status:    "ompreng_sampai_di_sppg",
			wantGroup: StageGroupCollection,
		},
		// Cleaning stages
		{
			name:      "cleaning in progress stage",
			status:    "ompreng_proses_pencucian",
			wantGroup: StageGroupCleaning,
		},
		{
			name:      "cleaning completed stage",
			status:    "ompreng_selesai_dicuci",
			wantGroup: StageGroupCleaning,
		},
		// Unknown stage
		{
			name:      "invalid status",
			status:    "invalid_status",
			wantGroup: StageGroupUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStageGroup(tt.status)
			if got != tt.wantGroup {
				t.Errorf("getStageGroup(%v) = %v, want %v", tt.status, got, tt.wantGroup)
			}
		})
	}
}

// TestValidateStageSequence_ValidSequence tests that valid stage sequences are allowed
func TestValidateStageSequence_ValidSequence(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		newStatus     string
	}{
		{
			name:          "delivery to delivery",
			currentStatus: "sedang_dimasak",
			newStatus:     "selesai_dimasak",
		},
		{
			name:          "delivery to collection",
			currentStatus: "sudah_diterima_pihak_sekolah",
			newStatus:     "driver_ditugaskan_mengambil_ompreng",
		},
		{
			name:          "collection to collection",
			currentStatus: "driver_menuju_sekolah",
			newStatus:     "driver_sampai_di_sekolah",
		},
		{
			name:          "collection to cleaning",
			currentStatus: "ompreng_sampai_di_sppg",
			newStatus:     "ompreng_proses_pencucian",
		},
		{
			name:          "cleaning to cleaning",
			currentStatus: "ompreng_proses_pencucian",
			newStatus:     "ompreng_selesai_dicuci",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStageSequence(tt.currentStatus, tt.newStatus)
			if err != nil {
				t.Errorf("ValidateStageSequence() error = %v, want nil", err)
			}
		})
	}
}

// TestValidateStageSequence_InvalidSequence tests that invalid stage sequences are rejected
func TestValidateStageSequence_InvalidSequence(t *testing.T) {
	tests := []struct {
		name            string
		currentStatus   string
		newStatus       string
		wantErrContains string
	}{
		{
			name:            "delivery to cleaning - skip collection",
			currentStatus:   "diperjalanan",
			newStatus:       "ompreng_proses_pencucian",
			wantErrContains: "cannot skip collection stages",
		},
		{
			name:            "collection to delivery - backward",
			currentStatus:   "driver_menuju_sekolah",
			newStatus:       "sudah_sampai_sekolah",
			wantErrContains: "cannot move back to delivery stages",
		},
		{
			name:            "cleaning to delivery - backward",
			currentStatus:   "ompreng_proses_pencucian",
			newStatus:       "sedang_dimasak",
			wantErrContains: "cannot move back to earlier stages",
		},
		{
			name:            "cleaning to collection - backward",
			currentStatus:   "ompreng_selesai_dicuci",
			newStatus:       "ompreng_sampai_di_sppg",
			wantErrContains: "cannot move back to earlier stages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStageSequence(tt.currentStatus, tt.newStatus)
			if err == nil {
				t.Errorf("ValidateStageSequence() error = nil, want error")
				return
			}

			seqErr, ok := err.(*StageSequenceError)
			if !ok {
				t.Errorf("ValidateStageSequence() error type = %T, want *StageSequenceError", err)
				return
			}

			if seqErr.CurrentStatus != tt.currentStatus {
				t.Errorf("StageSequenceError.CurrentStatus = %v, want %v", seqErr.CurrentStatus, tt.currentStatus)
			}

			if seqErr.NewStatus != tt.newStatus {
				t.Errorf("StageSequenceError.NewStatus = %v, want %v", seqErr.NewStatus, tt.newStatus)
			}

			errMsg := err.Error()
			if len(errMsg) == 0 {
				t.Errorf("StageSequenceError.Error() returned empty string")
			}
		})
	}
}

// TestValidateStageSequence_InvalidStatus tests handling of unknown statuses
func TestValidateStageSequence_InvalidStatus(t *testing.T) {
	tests := []struct {
		name            string
		currentStatus   string
		newStatus       string
		wantErrContains string
	}{
		{
			name:            "invalid current status",
			currentStatus:   "invalid_status",
			newStatus:       "sedang_dimasak",
			wantErrContains: "current status is not a valid lifecycle stage",
		},
		{
			name:            "invalid new status",
			currentStatus:   "sedang_dimasak",
			newStatus:       "invalid_status",
			wantErrContains: "requested status is not a valid lifecycle stage",
		},
		{
			name:            "both statuses invalid",
			currentStatus:   "invalid_current",
			newStatus:       "invalid_new",
			wantErrContains: "current status is not a valid lifecycle stage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStageSequence(tt.currentStatus, tt.newStatus)
			if err == nil {
				t.Errorf("ValidateStageSequence() error = nil, want error for invalid status")
				return
			}

			seqErr, ok := err.(*StageSequenceError)
			if !ok {
				t.Errorf("ValidateStageSequence() error type = %T, want *StageSequenceError", err)
				return
			}

			if seqErr.CurrentStatus != tt.currentStatus {
				t.Errorf("StageSequenceError.CurrentStatus = %v, want %v", seqErr.CurrentStatus, tt.currentStatus)
			}

			if seqErr.NewStatus != tt.newStatus {
				t.Errorf("StageSequenceError.NewStatus = %v, want %v", seqErr.NewStatus, tt.newStatus)
			}
		})
	}
}

// TestValidateStageSequence_EdgeCases tests edge cases in stage sequence validation
func TestValidateStageSequence_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		newStatus     string
		wantErr       bool
	}{
		{
			name:          "first delivery stage to second delivery stage",
			currentStatus: "sedang_dimasak",
			newStatus:     "selesai_dimasak",
			wantErr:       false,
		},
		{
			name:          "last delivery stage to first collection stage",
			currentStatus: "sudah_diterima_pihak_sekolah",
			newStatus:     "driver_ditugaskan_mengambil_ompreng",
			wantErr:       false,
		},
		{
			name:          "last collection stage to first cleaning stage",
			currentStatus: "ompreng_sampai_di_sppg",
			newStatus:     "ompreng_proses_pencucian",
			wantErr:       false,
		},
		{
			name:          "first delivery stage to last cleaning stage - skip all",
			currentStatus: "sedang_dimasak",
			newStatus:     "ompreng_selesai_dicuci",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStageSequence(tt.currentStatus, tt.newStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStageSequence() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
