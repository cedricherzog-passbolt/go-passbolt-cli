package resource

import "time"

type ResourceJSONOutput struct {
	ID                *string        `json:"id,omitempty"`
	FolderParentID    *string        `json:"folder_parent_id,omitempty"`
	Name              *string        `json:"name,omitempty"`
	Username          *string        `json:"username,omitempty"`
	URI               *string        `json:"uri,omitempty"`
	Password          *string        `json:"password,omitempty"`
	Description       *string        `json:"description,omitempty"`
	CreatedTimestamp  *time.Time     `json:"created_timestamp,omitempty"`
	ModifiedTimestamp *time.Time     `json:"modified_timestamp,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	Secret            map[string]any `json:"secret,omitempty"`
	// Non-pointer bool without omitempty so `false` reaches output.
	Deleted bool `json:"deleted"`
	Expired bool `json:"expired"`
	// Nullable timestamp: omitted when Resource.Expired is nil.
	ExpiredAt      *time.Time `json:"expired_at,omitempty"`
	ResourceTypeID *string    `json:"resource_type_id,omitempty"`
}
