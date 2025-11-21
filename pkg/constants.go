package pkg

// service

type StatusTransfer int

const (
	StatusTransferNotFound StatusTransfer = iota

	StatusTransferDraft
	StatusTransferInitiated
	StatusTransferSuccess
	StatusTransferFailed
	StatusTransferTimeout
	StatusTransferCoreFailed
	StatusW4ChallengeRequest
)

// names maps each Status to its string representation.
var StatusTransfers = map[StatusTransfer]string{

	StatusTransferNotFound:   "NOT_FOUND",
	StatusTransferDraft:      "DRAFT",
	StatusTransferInitiated:  "INITIATED",
	StatusTransferSuccess:    "SUCCESS",
	StatusTransferFailed:     "FAILED",
	StatusTransferTimeout:    "TIMEOUT",
	StatusTransferCoreFailed: "CORE_FAILED",
	StatusW4ChallengeRequest: "W4_CHALLENGE_REQUEST",
}

func (s StatusTransfer) String() string {
	if name, ok := StatusTransfers[s]; ok {
		return name
	}
	return "unknown"
}

// app environment

var (
	AppEnvDevelopment = "development"
	AppEnvProduction  = "production"
)

type ContextKey string

const MetadataKey ContextKey = "metadata"
