package acl

type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

var (
	Permission_Topic_Publish = Permission{
		Name:        "publish",
		Description: "Publish a topic",
	}
	Permission_Topic_Tail = Permission{
		Name:        "tail",
		Description: "Tail a topic",
	}
	Permission_Topic_Delete = Permission{
		Name:        "delete",
		Description: "Delete a topic",
	}
	Permission_Topic_Drain = Permission{
		Name:        "drain",
		Description: "Drain a topic",
	}
	Permission_Topic_Pause = Permission{
		Name:        "pause",
		Description: "Pause a topic",
	}

	Permission_Claim_Entity = Permission{
		Name:        "claim_entity",
		Description: "Claim an entity",
	}
	Permission_Signup_User = Permission{
		Name:        "signup",
		Description: "Signup a user",
	}
)

var PermissionList = map[string]Permission{
	Permission_Topic_Publish.Name: Permission_Topic_Publish,
	Permission_Topic_Tail.Name:    Permission_Topic_Tail,
	Permission_Topic_Delete.Name:  Permission_Topic_Delete,
	Permission_Topic_Drain.Name:   Permission_Topic_Drain,
	Permission_Topic_Pause.Name:   Permission_Topic_Pause,
	Permission_Claim_Entity.Name:  Permission_Claim_Entity,
	Permission_Signup_User.Name:   Permission_Signup_User,
}
