package user

import "time"

type UserJSONOutput struct {
	ID                *string    `json:"id,omitempty"`
	Username          *string    `json:"username,omitempty"`
	FirstName         *string    `json:"first_name,omitempty"`
	LastName          *string    `json:"last_name,omitempty"`
	Role              *string    `json:"role,omitempty"`
	CreatedTimestamp  *time.Time `json:"created_timestamp,omitempty"`
	ModifiedTimestamp *time.Time `json:"modified_timestamp,omitempty"`
	// Lifecycle flags use non-pointer bool without omitempty so that `false`
	// values reach JSON output and remain distinguishable from "absent" when
	// users do --filter 'active == false'.
	Active   bool `json:"active"`
	Deleted  bool `json:"deleted"`
	Disabled bool `json:"disabled"`
}
