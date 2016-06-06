package common

func (f *Friend) GetName() string {
	return f.Name
}

func (f *Friend) GetUserId() int64 {
	return f.UserId
}

func (f *Friend) GetPictureDigest() []byte {
	return f.PictureDigest
}
