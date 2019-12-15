package gud

type PackageVersion struct {
	Major, Minor, Patch uint
}

func GetVersion() PackageVersion {
	return PackageVersion{
		Major: 0,
		Minor: 0,
		Patch: 0,
	}
}
