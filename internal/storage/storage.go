package storage

import(
	"errors"
)

var (
	ErrTeamNameExists = errors.New("team_name already exists")
	ErrTeamNameNotFound = errors.New("team_name not found")
	ErrUserNotFound = errors.New("user not found")
	ErrPullRequestExists = errors.New("pull_request already exists")
	ErrAuthorNotFound = errors.New("author not found")
	ErrPullRequestNotFound = errors.New("pull request not found")
	ErrCannotReassingnOnMergedPr = errors.New("cannot reassign on merged PR")
	ErrPrReviewerNotFound = errors.New("pr reviewer not found")
	ErrReviewerIsNotAssignedToThisPr = errors.New("reviewer is not assigned to this pr")
	ErrActiveReplacementCandidateInTeam = errors.New("no active replacement candidate in team")
	ErrUserExists = errors.New("user already exists")
	ErrUserAlreadyDeleted = errors.New("user already deleted")
	ErrUserNotDeleted = errors.New("user is not deleted")
)
