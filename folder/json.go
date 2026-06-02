package folder

import "time"

type FolderJSONOutput struct {
	ID                *string    `json:"id,omitempty"`
	FolderParentID    *string    `json:"folder_parent_id,omitempty"`
	Name              *string    `json:"name,omitempty"`
	CreatedTimestamp  *time.Time `json:"created_timestamp,omitempty"`
	ModifiedTimestamp *time.Time `json:"modified_timestamp,omitempty"`
	// Non-pointer bool without omitempty so `false` reaches output and is
	// filterable via `--filter 'personal == false'`.
	Personal bool `json:"personal"`
}
