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
	// common permissions
	Permission_Claim_Entity.Name: Permission_Claim_Entity,
	Permission_Signup_User.Name:  Permission_Signup_User,

	// topic permissions
	Permission_Topic_Publish.Name: Permission_Topic_Publish,
	Permission_Topic_Tail.Name:    Permission_Topic_Tail,
	Permission_Topic_Delete.Name:  Permission_Topic_Delete,
	Permission_Topic_Drain.Name:   Permission_Topic_Drain,
	Permission_Topic_Pause.Name:   Permission_Topic_Pause,

	// channel permissions
	Permission_Channel_Pause.Name:  Permission_Channel_Pause,
	Permission_Channel_Empty.Name:  Permission_Channel_Empty,
	Permission_Channel_Delete.Name: Permission_Channel_Delete,
}

var TopicActionPermissions = []Permission{
	Permission_Topic_Publish,
	Permission_Topic_Tail,
	Permission_Topic_Drain,
	Permission_Topic_Pause,
	Permission_Topic_Delete,
}

var (
	Permission_Channel_Pause = Permission{
		Name:        "pause",
		Description: "Pause a channel",
	}
	Permission_Channel_Empty = Permission{
		Name:        "empty",
		Description: "Empty a channel",
	}
	Permission_Channel_Delete = Permission{
		Name:        "delete",
		Description: "Delete a channel",
	}
)

var ChannelActionPermissions = []Permission{
	Permission_Channel_Pause,
	Permission_Channel_Empty,
	Permission_Channel_Delete,
}
