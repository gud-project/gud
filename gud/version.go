package gud

type Version struct {
	Major, Minor, Patch uint
}

func GetVersion() Version {
	return Version{
		Major: 0,
		Minor: 0,
		Patch: 0,
	}
}
