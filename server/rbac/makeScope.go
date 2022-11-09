package rbac

// ResourceScope is the resource scope parsed from request
type ResourceScope struct {
	Type string
	// Labels is a map used to filter resource permissions during pre verification.
	// If a key of permission set is missing in the Labels, pre verification will pass this key
	Labels []map[string]string
	// Verb is the apply resource action, e.g. "get", "create"
	Verb string
}

func CreateConfigResourceScope() *ResourceScope {
	return &ResourceScope{
		Type: "config",
		Verb: "create",
	}
}

func DeleteConfigResourceScope() *ResourceScope {
	return &ResourceScope{
		Type: "config",
		Verb: "delete",
	}
}

func UpdateConfigResourceScope() *ResourceScope {
	return &ResourceScope{
		Type: "config",
		Verb: "update",
	}
}

func GetConfigResourceScope() *ResourceScope {
	return &ResourceScope{
		Type: "config",
		Verb: "get",
	}
}
