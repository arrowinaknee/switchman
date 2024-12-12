package auth

type store interface {
	loadData() (*usersData, error)
	saveData(*usersData) error
}

type fileStore struct {
	path string
}

func getFileStore(path string) store {
	return &fileStore{path}
}

func (f *fileStore) loadData() (*usersData, error) {
	panic("unimplemented")
}

func (f *fileStore) saveData(*usersData) error {
	panic("unimplemented")
}
