package gud

type PackageVersion struct {
	Major, Minor, Patch uint
}

// GetVersion returns the current version of Gud you are using.
func GetVersion() PackageVersion {
	return PackageVersion{
		Major: 0,
		Minor: 0,
		Patch: 0,
	}
}
