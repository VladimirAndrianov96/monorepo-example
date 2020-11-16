package user

type (
	// AlreadyExists signifies a user with a specified email address already exists in the system.
	AlreadyExists struct{}

	// IsActive signifies a user fails being inactive invariant.
	IsActive struct{}

	// IsInactive signifies a user fails being active invariant.
	IsInactive struct{}

	// UserNotFound signifies a user is not found.
	UserNotFound struct{}
)

func (err AlreadyExists) Error() string {
	return "User already exists"
}

func (err IsActive) Error() string {
	return "User is active"
}

func (err IsInactive) Error() string {
	return "User is inactive"
}

func (err UserNotFound) Error() string {
	return "User is unverified"
}
