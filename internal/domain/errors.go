package domain

type ErrorCode string

const (
	TeamExists    ErrorCode = "TEAM_EXISTS"
	PRExists      ErrorCode = "PR_EXISTS"
	PRMerged      ErrorCode = "PR_MERGED"
	NotAssigned   ErrorCode = "NOT_ASSIGNED"
	NoCandidate   ErrorCode = "NO_CANDIDATE"
	NotFound      ErrorCode = "NOT_FOUND"
	NoMembers     ErrorCode = "NO_MEMBERS"
	NonUniqueUser ErrorCode = "NONUNIQUE_USER"
)

type AppError struct {
	Code    ErrorCode
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewError(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func NewTeamExistsError() *AppError {
	return NewError(TeamExists, "team_name already exists")
}

func NewPRExistsError() *AppError {
	return NewError(PRExists, "PR id already exists")
}

func NewPRMergedError() *AppError {
	return NewError(PRMerged, "cannot reassign on merged PR")
}

func NewNotAssignedError() *AppError {
	return NewError(NotAssigned, "reviewer is not assigned to this PR")
}

func NewNoCandidateError() *AppError {
	return NewError(NoCandidate, "no active replacement candidate in team")
}

func NewNotFoundError(resource string) *AppError {
	return NewError(NotFound, resource+" not found")
}

func NewNoMembersError() *AppError {
	return NewError(NoMembers, "members list cannot be empty")
}

func NewNonUniqueUserError() *AppError {
	return NewError(NonUniqueUser, "members list contains at least 1 non-unique user_id")
}
