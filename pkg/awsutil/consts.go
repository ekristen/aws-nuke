package awsutil

// Default is a generic constant for the word default
const Default = "default"

// DefaultPrefix matches AWS-managed "default.*" resource IDs (e.g. ElastiCache's
// auto-created "default.iam-user" user and "default.iam-user-group" group).
// Real user-supplied IDs can never contain a "." (AWS rejects such characters),
// so this prefix safely identifies permanent, non-deletable AWS defaults without
// risking a false-positive match against a legitimately named resource.
const DefaultPrefix = Default + "."

// Custom is a generic constant for the word custom
const Custom = "custom"

// StateDeleted is a generic constant for the word deleted for state
const StateDeleted = "deleted"

// StateDeleting is a generic constant for the word deleting for state
const StateDeleting = "deleting"

// StateRejected is a generic constant for the word rejected for state
const StateRejected = "rejected"

// StateExpired is a generic constant for the word expired for state
const StateExpired = "expired"

// StateFailed is a generic constant for the word failed for state
const StateFailed = "failed"
